package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	imagev1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
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

func main() {
	// Setup kubernetes config
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
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

	// Print ImagePolicies and check pods
	fmt.Printf("\nFound %d ImagePolicies:\n", len(imagePolicies.Items))
	for _, policy := range imagePolicies.Items {
		fmt.Printf("\nName: %s\n", policy.Name)
		fmt.Printf("Namespace: %s\n", policy.Namespace)
		
		latestImage := policy.Status.LatestImage
		if latestImage != "" {
			fmt.Printf("Latest Image: %s\n", latestImage)
			baseLatestImage := getBaseImage(latestImage)

			// Check pods using this image
			for _, pod := range pods.Items {
				for _, container := range pod.Spec.Containers {
					if getBaseImage(container.Image) == baseLatestImage {
						fmt.Printf("Pod %s/%s using image: %s\n", 
							pod.Namespace, 
							pod.Name, 
							container.Image)
						if container.Image != latestImage {
							fmt.Printf("  -> Not using latest version!\n")
						}
					}
				}
			}
		}
		
		if policy.Spec.Policy.SemVer != nil {
			fmt.Printf("Semver Range: %s\n", policy.Spec.Policy.SemVer.Range)
		}
	}
}
