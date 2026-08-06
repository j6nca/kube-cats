package main

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func node(name, cpu, mem string) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(cpu),
				corev1.ResourceMemory: resource.MustParse(mem),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(cpu),
				corev1.ResourceMemory: resource.MustParse(mem),
			},
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			},
		},
	}
}

func pod(name, nodeName string, reqCPU, limCPU string) corev1.Pod {
	res := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse(reqCPU)},
	}
	if limCPU != "" {
		res.Limits = corev1.ResourceList{corev1.ResourceCPU: resource.MustParse(limCPU)}
	}

	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: corev1.PodSpec{
			NodeName:   nodeName,
			Containers: []corev1.Container{{Name: "c", Resources: res}},
		},
		Status: corev1.PodStatus{
			Phase:      corev1.PodRunning,
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
			ContainerStatuses: []corev1.ContainerStatus{
				{RestartCount: 2},
				{RestartCount: 3},
			},
		},
	}
}

const epsilon = 0.0001

func closeTo(a, b float64) bool { return a-b < epsilon && b-a < epsilon }

// With metrics, a pod that has limits is measured against its own limits.
func TestPodPressureUsesLimitsWhenMetricsPresent(t *testing.T) {
	nodes := []corev1.Node{node("n1", "4", "8Gi")}
	pods := []corev1.Pod{pod("p1", "n1", "100m", "1")}
	usage := podUsage{"default/p1": {cpuMilli: 750, memBytes: 0}}

	got := simplifyPods(pods, nodes, usage, true)[0]

	if got.Pressure.Source != SourceMetrics {
		t.Errorf("source = %q, want %q", got.Pressure.Source, SourceMetrics)
	}
	// 750m used of a 1000m limit
	if !closeTo(got.Pressure.CPU, 0.75) {
		t.Errorf("cpu pressure = %v, want 0.75", got.Pressure.CPU)
	}
}

// A pod with no limits is measured against the capacity of its node instead.
func TestPodPressureFallsBackToNodeAllocatable(t *testing.T) {
	nodes := []corev1.Node{node("n1", "4", "8Gi")}
	pods := []corev1.Pod{pod("p1", "n1", "100m", "")}
	usage := podUsage{"default/p1": {cpuMilli: 2000}}

	got := simplifyPods(pods, nodes, usage, true)[0]

	// 2000m used of a 4000m node
	if !closeTo(got.Pressure.CPU, 0.5) {
		t.Errorf("cpu pressure = %v, want 0.5", got.Pressure.CPU)
	}
}

// Without metrics the ratio is estimated from requests and labelled as such.
func TestPodPressureWithoutMetricsUsesRequests(t *testing.T) {
	nodes := []corev1.Node{node("n1", "4", "8Gi")}
	pods := []corev1.Pod{pod("p1", "n1", "500m", "1")}

	got := simplifyPods(pods, nodes, nil, false)[0]

	if got.Pressure.Source != SourceRequests {
		t.Errorf("source = %q, want %q", got.Pressure.Source, SourceRequests)
	}
	// 500m requested against a 1000m limit
	if !closeTo(got.Pressure.CPU, 0.5) {
		t.Errorf("cpu pressure = %v, want 0.5", got.Pressure.CPU)
	}
}

// An unscheduled pod with no limits has nothing to divide by.
func TestPodPressureUnknownWhenNoDenominator(t *testing.T) {
	pods := []corev1.Pod{pod("p1", "", "0", "")}

	got := simplifyPods(pods, nil, nil, true)[0]

	if got.Pressure.Source != SourceUnknown {
		t.Errorf("source = %q, want %q", got.Pressure.Source, SourceUnknown)
	}
	if got.Pressure.CPU != 0 || got.Pressure.Memory != 0 {
		t.Errorf("pressure = %+v, want zeroes", got.Pressure)
	}
}

func TestPodHealthFields(t *testing.T) {
	nodes := []corev1.Node{node("n1", "4", "8Gi")}
	got := simplifyPods([]corev1.Pod{pod("p1", "n1", "100m", "1")}, nodes, nil, false)[0]

	if got.Phase != "Running" {
		t.Errorf("phase = %q, want Running", got.Phase)
	}
	if !got.Ready {
		t.Error("ready = false, want true")
	}
	if got.Restarts != 5 {
		t.Errorf("restarts = %d, want 5 (summed across containers)", got.Restarts)
	}
}

// Node pressure without metrics is the sum of scheduled pod requests.
func TestApplyRequestedNodePressure(t *testing.T) {
	nodes := simplifyNodes([]corev1.Node{node("n1", "4", "8Gi")}, nil, false)
	pods := simplifyPods([]corev1.Pod{
		pod("p1", "n1", "1", ""),
		pod("p2", "n1", "2", ""),
	}, []corev1.Node{node("n1", "4", "8Gi")}, nil, false)

	applyRequestedNodePressure(nodes, pods)

	// 3000m requested of a 4000m node
	if !closeTo(nodes[0].Pressure.CPU, 0.75) {
		t.Errorf("cpu pressure = %v, want 0.75", nodes[0].Pressure.CPU)
	}
	if nodes[0].Pressure.Source != SourceRequests {
		t.Errorf("source = %q, want %q", nodes[0].Pressure.Source, SourceRequests)
	}
}

func TestNodePressureFromMetrics(t *testing.T) {
	usage := nodeUsage{"n1": {cpuMilli: 1000, memBytes: 4 << 30}}

	got := simplifyNodes([]corev1.Node{node("n1", "4", "8Gi")}, usage, true)[0]

	if !closeTo(got.Pressure.CPU, 0.25) {
		t.Errorf("cpu pressure = %v, want 0.25", got.Pressure.CPU)
	}
	if !closeTo(got.Pressure.Memory, 0.5) {
		t.Errorf("memory pressure = %v, want 0.5", got.Pressure.Memory)
	}
	if !got.Ready {
		t.Error("ready = false, want true")
	}
}

func TestRatioGuardsZeroDenominator(t *testing.T) {
	if got := ratio(100, 0); got != 0 {
		t.Errorf("ratio(100, 0) = %v, want 0", got)
	}
	if got := ratio(100, -5); got != 0 {
		t.Errorf("ratio(100, -5) = %v, want 0", got)
	}
}
