package main

import (
	"context"
   	"log"
	"time"
	"sync"

	corev1 "k8s.io/api/core/v1"
  "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SimplifiedPod struct {
  Name      string               `json:"name"`
  Namespace string               `json:"namespace"`
  Resources PodResourceSummary `json:"resources"`
}

type PodResourceSummary struct {
  CPURequests    string `json:"cpuRequests"`
  CPULimits      string `json:"cpuLimits"`
  MemoryRequests string `json:"memoryRequests"`
  MemoryLimits   string `json:"memoryLimits"`
}

type SimplifiedNode struct {
  Name      string              `json:"name"`
  Resources NodeResourceSummary `json:"resources"`
}

type NodeResourceSummary struct {
  CPUCapacity      string `json:"cpuCapacity"`
  CPUAllocatable   string `json:"cpuAllocatable"`
  MemoryCapacity   string `json:"memoryCapacity"`
  MemoryAllocatable string `json:"memoryAllocatable"`
}


type Cache struct {
	Pods  []SimplifiedPod
	Nodes []SimplifiedNode
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
	if err == nil {
      trimmedPods := simplifyPods(pods.Items)
      trimmedNodes := simplifyNodes(nodes.Items)

      cache.Lock()
      cache.Pods = trimmedPods
      cache.Nodes = trimmedNodes
      cache.Unlock()
	}

	log.Printf("Cache updated: %d pods, %d nodes", len(cache.Pods), len(cache.Nodes))
}

// While not publishing sensitive data, trim and simplify responses
func simplifyPods(pods []corev1.Pod) []SimplifiedPod {
  var simplifiedPods []SimplifiedPod

  for _, pod := range pods {
    var cpuRequests, cpuLimits, memRequests, memLimits resource.Quantity
    for _, container := range pod.Spec.Containers {
      if container.Resources.Requests != nil {
        if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
          cpuRequests.Add(cpu)
        }
        if mem, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
          memRequests.Add(mem)
        }
      }
      if container.Resources.Limits != nil {
        if cpu, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
          cpuLimits.Add(cpu)
        }
        if mem, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
          memLimits.Add(mem)
        }
      }
    }
    simplifiedPods = append(simplifiedPods, SimplifiedPod{
      Name:      pod.Name,
      Namespace: pod.Namespace,
      Resources: PodResourceSummary{
        CPURequests:    cpuRequests.String(),
        CPULimits:      cpuLimits.String(),
        MemoryRequests: memRequests.String(),
        MemoryLimits:   memLimits.String(),
      },
    }) 
  }

  return simplifiedPods
}

func simplifyNodes(nodes []corev1.Node) []SimplifiedNode {
  var simplifiedNodes []SimplifiedNode
  
  for _, node := range nodes {
    var cpuCap, cpuAlloc, memCap, memAlloc resource.Quantity

    if cpu, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
      cpuCap = cpu
    }
    if cpu, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
      cpuAlloc = cpu
    }
    if mem, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
      memCap = mem
    }
    if mem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
      memAlloc = mem
    }

    simplifiedNodes = append(simplifiedNodes, SimplifiedNode{
      Name: node.Name,
      Resources: NodeResourceSummary{
        CPUCapacity:      cpuCap.String(),
        CPUAllocatable:   cpuAlloc.String(),
        MemoryCapacity:   memCap.String(),
        MemoryAllocatable: memAlloc.String(),
      },
    })
  }

  return simplifiedNodes
}
