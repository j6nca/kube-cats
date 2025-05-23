package main

import (
	"context"
   	"log"
	"time"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Cache struct {
	Pods  []v1.Pod
	Nodes []v1.Node
	sync.RWMutex
}

func getKubeInfo(clientset *kubernetes.Clientset, cache *Cache) {
	log.Println(" Scraping Kubernetes API...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to get nodes: %v", err)
		return
	}

	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to get pods: %v", err)
		return
	}

	log.Println("Updating Cache")
    cache.Lock()
	if err == nil {
		cache.Pods = pods.Items
	}
	if err == nil {
		cache.Nodes = nodes.Items
	}
	cache.Unlock()

	log.Printf("Cache updated: %d pods, %d nodes", len(cache.Pods), len(cache.Nodes))
}