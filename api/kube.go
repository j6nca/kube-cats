package main

import (
	"fmt"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Clients bundles the core API client with the optional metrics client.
// Metrics is nil when metrics-server is not reachable; callers must handle
// that, since the visualizer is expected to work without it.
type Clients struct {
	Core    *kubernetes.Clientset
	Metrics *metricsv.Clientset
}

func getConfig() (*rest.Config, error) {
	// Detect dev mode
	dev := os.Getenv("ENVIRONMENT") != "production"

	if dev {
		log.Println("Running in DEV mode: using kubeconfig")
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
		}

		proxy := os.Getenv("KUBE_PROXY_URL")
		if proxy == "" {
			proxy = "http://127.0.0.1:8087" // point to kubectl proxy
		}

		return clientcmd.BuildConfigFromFlags(proxy, kubeconfig)
	}

	log.Println("Running in-cluster")
	return rest.InClusterConfig()
}

func getClients() (*Clients, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	core, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	clients := &Clients{Core: core}

	// metrics-server is optional. Failing to build the client is not fatal;
	// pressure calculations fall back to requests vs allocatable.
	metrics, err := metricsv.NewForConfig(config)
	if err != nil {
		log.Printf("Metrics client unavailable, falling back to resource requests: %v", err)
		return clients, nil
	}
	clients.Metrics = metrics

	return clients, nil
}
