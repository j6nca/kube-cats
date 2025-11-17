package main

import (
	"fmt"
  "log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getClientset() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Detect dev mode
	dev := os.Getenv("ENVIRONMENT") != "production"

	if dev {
		log.Println("Running in DEV mode: using kubeconfig")
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
		}

		config, err = clientcmd.BuildConfigFromFlags("http://127.0.0.1:8000", kubeconfig) // point to kubectl proxy
	} else {
		log.Println("Running in-cluster")
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
