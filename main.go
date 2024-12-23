package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"path/filepath"
	"strings"

	imagev1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getBaseImage(image string) string {
	// Remove tag or digest
	if strings.Contains(image, "@sha256:") {
		return strings.Split(image, "@sha256:")[0]
	}
	if strings.Contains(image, ":") {
		return strings.Split(image, ":")[0]
	}
	return image
}

var (
	updateAvailable = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "update_available",
			Help: "Indicates if a newer version is available (1) or not (0)",
		},
		[]string{"namespace", "pod", "container"},
	)
)

func init() {
	prometheus.MustRegister(updateAvailable)
}

func checkUpdates(c client.Client) {
	// List ImagePolicies from all namespaces
	var imagePolicies imagev1.ImagePolicyList
	if err := c.List(context.Background(), &imagePolicies); err != nil {
		log.Printf("Error listing ImagePolicies: %v", err)
		return
	}

	// List all pods
	var pods corev1.PodList
	if err := c.List(context.Background(), &pods); err != nil {
		log.Printf("Error listing Pods: %v", err)
		return
	}

	// Check ImagePolicies and pods
	log.Printf("Found %d ImagePolicies", len(imagePolicies.Items))
	for _, policy := range imagePolicies.Items {
		log.Printf("ImagePolicy: name=%s namespace=%s", policy.Name, policy.Namespace)
		
		latestImage := policy.Status.LatestImage
		if latestImage != "" {
			log.Printf("Latest Image: %s", latestImage)
			baseLatestImage := getBaseImage(latestImage)

			// Check pods using this image
			for _, pod := range pods.Items {
				for _, container := range pod.Spec.Containers {
					if getBaseImage(container.Image) == baseLatestImage {
						log.Printf("Pod %s/%s using image: %s", 
							pod.Namespace, 
							pod.Name, 
							container.Image)
						
						// Update metric
						value := 0.0
						if container.Image != latestImage {
							log.Printf("  -> Pod %s/%s container %s not using latest version", 
								pod.Namespace, 
								pod.Name,
								container.Name)
							value = 1.0
						}
						updateAvailable.With(prometheus.Labels{
							"namespace": pod.Namespace,
							"pod":      pod.Name,
							"container": container.Name,
						}).Set(value)
					}
				}
			}
		}
		
		if policy.Spec.Policy.SemVer != nil {
			log.Printf("Semver Range: %s", policy.Spec.Policy.SemVer.Range)
		}
	}
}

func main() {
	// Configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	var (
		kubeconfig *string
		interval   = flag.Duration("interval", 5*time.Minute, "Polling interval")
		addr      = flag.String("listen-address", ":2112", "The address to listen on for HTTP requests.")
	)
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Try in-cluster config first, fall back to kubeconfig file
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	// Create controller-runtime client
	scheme := runtime.NewScheme()
	imagev1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	
	c, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		panic(err.Error())
	}

	// List ImagePolicies from all namespaces
	var imagePolicies imagev1.ImagePolicyList
	if err := c.List(context.Background(), &imagePolicies); err != nil {
		panic(err.Error())
	}

	// List all pods
	var pods corev1.PodList
	if err := c.List(context.Background(), &pods); err != nil {
		panic(err.Error())
	}

	// Start initial check
	checkUpdates(c)

	// Setup metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Start polling routine
	go func() {
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()

		for range ticker.C {
			checkUpdates(c)
		}
	}()

	// Start HTTP server
	log.Printf("Starting metrics server on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		panic(err)
	}
}
