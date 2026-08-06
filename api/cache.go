package main

import (
	"context"
	"log"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// PressureSource records how a Pressure value was derived, so the UI (and
// anyone reading the API) can tell live usage apart from a static estimate.
type PressureSource string

const (
	// SourceMetrics means the ratio came from live metrics-server usage.
	SourceMetrics PressureSource = "metrics"
	// SourceRequests means metrics-server was unavailable and the ratio was
	// estimated from resource requests instead.
	SourceRequests PressureSource = "requests"
	// SourceUnknown means there was nothing to divide by (no limits, no
	// requests, no allocatable) and the ratio is meaningless.
	SourceUnknown PressureSource = "unknown"
)

// Pressure is a 0..1-ish utilisation ratio per resource. Values can exceed 1
// when usage overshoots the denominator (e.g. a pod above its requests), so
// consumers should clamp before using it for display.
type Pressure struct {
	CPU    float64        `json:"cpu"`
	Memory float64        `json:"memory"`
	Source PressureSource `json:"source"`
}

type SimplifiedPod struct {
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Node      string             `json:"node"`
	Phase     string             `json:"phase"`
	Ready     bool               `json:"ready"`
	Restarts  int32              `json:"restarts"`
	Resources PodResourceSummary `json:"resources"`
	Pressure  Pressure           `json:"pressure"`
}

type PodResourceSummary struct {
	CPURequests    string `json:"cpuRequests"`
	CPULimits      string `json:"cpuLimits"`
	MemoryRequests string `json:"memoryRequests"`
	MemoryLimits   string `json:"memoryLimits"`
	// Numeric forms so the frontend does not have to parse Kubernetes
	// quantity strings like "100m" or "128Mi".
	CPURequestsMilli int64 `json:"cpuRequestsMilli"`
	CPULimitsMilli   int64 `json:"cpuLimitsMilli"`
	MemoryRequestsB  int64 `json:"memoryRequestsBytes"`
	MemoryLimitsB    int64 `json:"memoryLimitsBytes"`
	CPUUsageMilli    int64 `json:"cpuUsageMilli"`
	MemoryUsageB     int64 `json:"memoryUsageBytes"`
}

type SimplifiedNode struct {
	Name      string              `json:"name"`
	Ready     bool                `json:"ready"`
	Resources NodeResourceSummary `json:"resources"`
	Pressure  Pressure            `json:"pressure"`
}

type NodeResourceSummary struct {
	CPUCapacity       string `json:"cpuCapacity"`
	CPUAllocatable    string `json:"cpuAllocatable"`
	MemoryCapacity    string `json:"memoryCapacity"`
	MemoryAllocatable string `json:"memoryAllocatable"`
	// Numeric forms, as above.
	CPUAllocatableMilli int64 `json:"cpuAllocatableMilli"`
	MemoryAllocatableB  int64 `json:"memoryAllocatableBytes"`
	CPUUsageMilli       int64 `json:"cpuUsageMilli"`
	MemoryUsageB        int64 `json:"memoryUsageBytes"`
}

type Cache struct {
	Pods  []SimplifiedPod
	Nodes []SimplifiedNode
	sync.RWMutex
}

// podUsage keys live pod metrics by namespace/name.
type podUsage map[string]struct{ cpuMilli, memBytes int64 }

// nodeUsage keys live node metrics by name.
type nodeUsage map[string]struct{ cpuMilli, memBytes int64 }

func getKubeInfo(clients *Clients, cache *Cache) {
	log.Println(" Scraping Kubernetes API...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodes, err := clients.Core.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to get nodes: %v", err)
		return
	}

	pods, err := clients.Core.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to get pods: %v", err)
		return
	}

	// Optional: live usage. A failure here degrades pressure to the
	// requests-based estimate rather than dropping the whole scrape.
	pUsage, nUsage, haveMetrics := getUsage(ctx, clients)

	log.Println("Updating Cache")
	trimmedNodes := simplifyNodes(nodes.Items, nUsage, haveMetrics)
	trimmedPods := simplifyPods(pods.Items, nodes.Items, pUsage, haveMetrics)

	// Node pressure without metrics is the sum of the requests of the pods
	// scheduled onto it, so it can only be computed once pods are known.
	if !haveMetrics {
		applyRequestedNodePressure(trimmedNodes, trimmedPods)
	}

	cache.Lock()
	cache.Pods = trimmedPods
	cache.Nodes = trimmedNodes
	cache.Unlock()

	log.Printf("Cache updated: %d pods, %d nodes (metrics: %t)", len(trimmedPods), len(trimmedNodes), haveMetrics)
}

// getUsage pulls live pod and node usage from metrics-server. The bool result
// reports whether the data is usable; on any error it is false and the maps
// are empty.
func getUsage(ctx context.Context, clients *Clients) (podUsage, nodeUsage, bool) {
	if clients.Metrics == nil {
		return nil, nil, false
	}

	nodeMetrics, err := clients.Metrics.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("metrics-server unavailable for nodes, using requests instead: %v", err)
		return nil, nil, false
	}

	podMetrics, err := clients.Metrics.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("metrics-server unavailable for pods, using requests instead: %v", err)
		return nil, nil, false
	}

	nu := make(nodeUsage, len(nodeMetrics.Items))
	for _, n := range nodeMetrics.Items {
		cpu := n.Usage[corev1.ResourceCPU]
		mem := n.Usage[corev1.ResourceMemory]
		nu[n.Name] = struct{ cpuMilli, memBytes int64 }{cpu.MilliValue(), mem.Value()}
	}

	pu := make(podUsage, len(podMetrics.Items))
	for _, p := range podMetrics.Items {
		var cpu, mem resource.Quantity
		for _, c := range p.Containers {
			addContainerUsage(&cpu, &mem, c)
		}
		pu[p.Namespace+"/"+p.Name] = struct{ cpuMilli, memBytes int64 }{cpu.MilliValue(), mem.Value()}
	}

	return pu, nu, true
}

func addContainerUsage(cpu, mem *resource.Quantity, c metricsv1beta1.ContainerMetrics) {
	if v, ok := c.Usage[corev1.ResourceCPU]; ok {
		cpu.Add(v)
	}
	if v, ok := c.Usage[corev1.ResourceMemory]; ok {
		mem.Add(v)
	}
}

// ratio divides guarding against a zero or negative denominator.
func ratio(used, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(used) / float64(total)
}

// While not publishing sensitive data, trim and simplify responses
func simplifyPods(pods []corev1.Pod, nodes []corev1.Node, usage podUsage, haveMetrics bool) []SimplifiedPod {
	// Pods without limits are measured against the capacity of the node they
	// landed on, so a pod hogging a whole node still reads as hot.
	nodeAlloc := make(map[string]struct{ cpuMilli, memBytes int64 }, len(nodes))
	for _, node := range nodes {
		cpu := node.Status.Allocatable[corev1.ResourceCPU]
		mem := node.Status.Allocatable[corev1.ResourceMemory]
		nodeAlloc[node.Name] = struct{ cpuMilli, memBytes int64 }{cpu.MilliValue(), mem.Value()}
	}

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

		used := usage[pod.Namespace+"/"+pod.Name]
		alloc := nodeAlloc[pod.Spec.NodeName]

		summary := PodResourceSummary{
			CPURequests:      cpuRequests.String(),
			CPULimits:        cpuLimits.String(),
			MemoryRequests:   memRequests.String(),
			MemoryLimits:     memLimits.String(),
			CPURequestsMilli: cpuRequests.MilliValue(),
			CPULimitsMilli:   cpuLimits.MilliValue(),
			MemoryRequestsB:  memRequests.Value(),
			MemoryLimitsB:    memLimits.Value(),
			CPUUsageMilli:    used.cpuMilli,
			MemoryUsageB:     used.memBytes,
		}

		simplifiedPods = append(simplifiedPods, SimplifiedPod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Node:      pod.Spec.NodeName,
			Phase:     string(pod.Status.Phase),
			Ready:     podReady(pod),
			Restarts:  podRestarts(pod),
			Resources: summary,
			Pressure:  podPressure(summary, alloc, haveMetrics),
		})
	}

	return simplifiedPods
}

// podPressure measures a pod against its own limits when it has them, and
// against its node's allocatable capacity when it does not.
func podPressure(s PodResourceSummary, alloc struct{ cpuMilli, memBytes int64 }, haveMetrics bool) Pressure {
	pick := func(limit, nodeTotal int64) int64 {
		if limit > 0 {
			return limit
		}
		return nodeTotal
	}

	cpuTotal := pick(s.CPULimitsMilli, alloc.cpuMilli)
	memTotal := pick(s.MemoryLimitsB, alloc.memBytes)

	if cpuTotal <= 0 && memTotal <= 0 {
		return Pressure{Source: SourceUnknown}
	}

	if haveMetrics {
		return Pressure{
			CPU:    ratio(s.CPUUsageMilli, cpuTotal),
			Memory: ratio(s.MemoryUsageB, memTotal),
			Source: SourceMetrics,
		}
	}

	return Pressure{
		CPU:    ratio(s.CPURequestsMilli, cpuTotal),
		Memory: ratio(s.MemoryRequestsB, memTotal),
		Source: SourceRequests,
	}
}

func podReady(pod corev1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

func podRestarts(pod corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

func simplifyNodes(nodes []corev1.Node, usage nodeUsage, haveMetrics bool) []SimplifiedNode {
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

		used := usage[node.Name]

		summary := NodeResourceSummary{
			CPUCapacity:         cpuCap.String(),
			CPUAllocatable:      cpuAlloc.String(),
			MemoryCapacity:      memCap.String(),
			MemoryAllocatable:   memAlloc.String(),
			CPUAllocatableMilli: cpuAlloc.MilliValue(),
			MemoryAllocatableB:  memAlloc.Value(),
			CPUUsageMilli:       used.cpuMilli,
			MemoryUsageB:        used.memBytes,
		}

		pressure := Pressure{Source: SourceUnknown}
		if haveMetrics {
			pressure = Pressure{
				CPU:    ratio(summary.CPUUsageMilli, summary.CPUAllocatableMilli),
				Memory: ratio(summary.MemoryUsageB, summary.MemoryAllocatableB),
				Source: SourceMetrics,
			}
		}

		simplifiedNodes = append(simplifiedNodes, SimplifiedNode{
			Name:      node.Name,
			Ready:     nodeReady(node),
			Resources: summary,
			Pressure:  pressure,
		})
	}

	return simplifiedNodes
}

func nodeReady(node corev1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

// applyRequestedNodePressure fills in node pressure from the summed requests of
// the pods scheduled onto each node. Used only when metrics-server is absent.
func applyRequestedNodePressure(nodes []SimplifiedNode, pods []SimplifiedPod) {
	type totals struct{ cpuMilli, memBytes int64 }
	requested := make(map[string]totals, len(nodes))

	for _, pod := range pods {
		t := requested[pod.Node]
		t.cpuMilli += pod.Resources.CPURequestsMilli
		t.memBytes += pod.Resources.MemoryRequestsB
		requested[pod.Node] = t
	}

	for i := range nodes {
		t := requested[nodes[i].Name]
		nodes[i].Pressure = Pressure{
			CPU:    ratio(t.cpuMilli, nodes[i].Resources.CPUAllocatableMilli),
			Memory: ratio(t.memBytes, nodes[i].Resources.MemoryAllocatableB),
			Source: SourceRequests,
		}
	}
}
