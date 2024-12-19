package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	imagev1 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	
	c, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		panic(err.Error())
	}

	// List ImagePolicies from all namespaces
	var imagePolicies imagev1.ImagePolicyList
	if err := c.List(context.Background(), &imagePolicies); err != nil {
		panic(err.Error())
	}

	// Print ImagePolicies
	fmt.Printf("\nFound %d ImagePolicies:\n", len(imagePolicies.Items))
	for _, policy := range imagePolicies.Items {
		fmt.Printf("\nName: %s\n", policy.Name)
		fmt.Printf("Namespace: %s\n", policy.Namespace)
		if policy.Status.LatestImage != "" {
			fmt.Printf("Latest Image: %s\n", policy.Status.LatestImage)
		}
		if policy.Spec.Policy.SemVer != nil {
			fmt.Printf("Semver Range: %s\n", policy.Spec.Policy.SemVer.Range)
		}
	}
}
