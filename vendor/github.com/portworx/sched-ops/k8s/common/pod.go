package common

import (
	"bytes"
	"fmt"

	schederrors "github.com/portworx/sched-ops/k8s/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// GetPodsByOwner returns pods for the given owner and namespace
func GetPodsByOwner(client v1.CoreV1Interface, ownerUID types.UID, namespace string) ([]corev1.Pod, error) {
	podList, err := client.Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []corev1.Pod
	for _, pod := range podList.Items {
		for _, owner := range pod.OwnerReferences {
			if owner.UID == ownerUID {
				result = append(result, pod)
			}
		}
	}

	if len(result) == 0 {
		return nil, schederrors.ErrPodsNotFound
	}

	return result, nil
}

// GeneratePodsOverviewString returns a string description of pods.
func GeneratePodsOverviewString(pods []corev1.Pod) string {
	var buffer bytes.Buffer
	for _, p := range pods {
		running := IsPodRunning(p)
		ready := IsPodReady(p)
		podString := fmt.Sprintf("  pod name:%s namespace:%s running:%v ready:%v node:%s\n", p.Name, p.Namespace, running, ready, p.Status.HostIP)
		buffer.WriteString(podString)
	}

	return buffer.String()
}

// IsPodReady checks if all containers in a pod are ready (passed readiness probe).
func IsPodReady(pod corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodSucceeded {
		return false
	}

	// If init containers are running, return false since the actual container would not have started yet
	for _, c := range pod.Status.InitContainerStatuses {
		if c.State.Running != nil {
			return false
		}
	}

	for _, c := range pod.Status.ContainerStatuses {
		if c.State.Terminated != nil &&
			c.State.Terminated.ExitCode == 0 &&
			c.State.Terminated.Reason == "Completed" {
			continue // container has exited successfully
		}

		if c.State.Running == nil {
			return false
		}

		if !c.Ready {
			return false
		}
	}

	return true
}

// IsPodRunning checks if all containers in a pod are in running state.
func IsPodRunning(pod corev1.Pod) bool {
	// If init containers are running, return false since the actual container would not have started yet
	for _, c := range pod.Status.InitContainerStatuses {
		if c.State.Running != nil {
			return false
		}
	}

	for _, c := range pod.Status.ContainerStatuses {
		if c.State.Running == nil {
			return false
		}
	}

	return true
}
