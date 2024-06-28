package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/portworx/sched-ops/k8s/common"
	schederrors "github.com/portworx/sched-ops/k8s/errors"
	"github.com/portworx/sched-ops/task"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// PodOps is an interface to perform k8s pod operations
type PodOps interface {
	// CreatePod creates the given pod.
	CreatePod(pod *corev1.Pod) (*corev1.Pod, error)
	// UpdatePod updates the given pod
	UpdatePod(pod *corev1.Pod) (*corev1.Pod, error)
	// ListPods returns pods from all namespaces matching the given label
	ListPods(map[string]string) (*corev1.PodList, error)
	// GetPods returns pods for the given namespace
	GetPods(string, map[string]string) (*corev1.PodList, error)
	// GetPodsByNode returns all pods in given namespace and given k8s node name.
	//  If namespace is empty, it will return pods from all namespaces.
	GetPodsByNode(nodeName, namespace string) (*corev1.PodList, error)
	// GetPodsByNodeAndLabels returns all pods in given namespace and given k8s node name
	// with a given label selector.
	//  If namespace is empty, it will return pods from all namespaces.
	GetPodsByNodeAndLabels(nodeName, namespace string, labelSelector map[string]string) (*corev1.PodList, error)
	// GetPodsByOwner returns pods for the given owner and namespace
	GetPodsByOwner(types.UID, string) ([]corev1.Pod, error)
	// GetPodsUsingPV returns all pods in cluster using given pv
	GetPodsUsingPV(pvName string) ([]corev1.Pod, error)
	// GetPodsUsingPVByNodeName returns all pods running on the node using the given pv
	GetPodsUsingPVByNodeName(pvName, nodeName string) ([]corev1.Pod, error)
	// GetPodsUsingPVC returns all pods in cluster using given pvc
	GetPodsUsingPVC(pvcName, pvcNamespace string) ([]corev1.Pod, error)
	// GetPodsUsingPVCByNodeName returns all pods running on the node using given pvc
	GetPodsUsingPVCByNodeName(pvcName, pvcNamespace, nodeName string) ([]corev1.Pod, error)
	// GetPodsUsingVolumePlugin returns all pods who use PVCs provided by the given volume plugin
	GetPodsUsingVolumePlugin(plugin string) ([]corev1.Pod, error)
	// GetPodsUsingVolumePluginByNodeName returns all pods who use PVCs provided by the given volume plugin on the given node
	GetPodsUsingVolumePluginByNodeName(nodeName, plugin string) ([]corev1.Pod, error)
	// GetPodByName returns pod for the given pod name and namespace
	GetPodByName(string, string) (*corev1.Pod, error)
	// GetPodByUID returns pod with the given UID, or error if nothing found
	GetPodByUID(types.UID, string) (*corev1.Pod, error)
	// DeletePod deletes the given pod
	DeletePod(string, string, bool) error
	// DeletePods deletes the given pods
	DeletePods([]corev1.Pod, bool) error
	// DeletePodsByLabels deletes pods for the given labels and namespace
	DeletePodsByLabels(namespace string, labelSelector map[string]string, timeout time.Duration) error
	// IsPodRunning checks if all containers in a pod are in running state
	IsPodRunning(corev1.Pod) bool
	// IsPodCompleted checks if the pod is in completed state
	IsPodCompleted(corev1.Pod) bool
	// IsPodReady checks if all containers in a pod are ready (passed readiness probe)
	IsPodReady(corev1.Pod) bool
	// IsPodBeingManaged returns true if the pod is being managed by a controller
	IsPodBeingManaged(corev1.Pod) bool
	// WaitForPodDeletion waits for given timeout for given pod to be deleted
	WaitForPodDeletion(uid types.UID, namespace string, timeout time.Duration) error
	// RunCommandInPod runs given command in the given pod
	RunCommandInPod(cmds []string, podName, containerName, namespace string) (string, error)
	// RunCommandInPodEx is extended version of RunCommandInPod
	RunCommandInPodEx(*RunCommandInPodExRequest) error
	// ValidatePod validates the given pod if it's ready
	ValidatePod(pod *corev1.Pod, timeout, retryInterval time.Duration) error
	// WatchPods sets up a watcher that listens for the changes to pods in given namespace
	WatchPods(namespace string, fn WatchFunc, listOptions metav1.ListOptions) error
	// GetPodLogs returns the logs of a POD as a string
	GetPodLog(podName string, namespace string, podLogOptions *corev1.PodLogOptions) (string, error)
}

// RunCommandInPodExRequest is a request structure for the RunCommandInPodEx func
type RunCommandInPodExRequest struct {
	Command       []string
	PODName       string
	ContainerName string
	Namespace     string
	UseTTY        bool
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
}

// CreatePod creates the given pod.
func (c *Client) CreatePod(pod *corev1.Pod) (*corev1.Pod, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
}

// UpdatePod updates the given pod
func (c *Client) UpdatePod(pod *corev1.Pod) (*corev1.Pod, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
}

// ListPods returns pods from all namespaces matching the given label
func (c *Client) ListPods(labelSelector map[string]string) (*corev1.PodList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	opts := metav1.ListOptions{
		LabelSelector: mapToCSV(labelSelector),
	}
	return c.kubernetes.CoreV1().Pods("").List(context.TODO(), opts)
}

// GetPods returns pods for the given namespace
func (c *Client) GetPods(namespace string, labelSelector map[string]string) (*corev1.PodList, error) {
	return c.getPodsWithListOptions(namespace, metav1.ListOptions{
		LabelSelector: mapToCSV(labelSelector),
	})
}

// GetPodsByNode returns all pods in given namespace and given k8s node name.
// If namespace is empty, it will return pods from all namespaces
func (c *Client) GetPodsByNode(nodeName, namespace string) (*corev1.PodList, error) {
	if len(nodeName) == 0 {
		return nil, fmt.Errorf("node name is required for this API")
	}

	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	}

	return c.getPodsWithListOptions(namespace, listOptions)
}

// GetPodsByNodeAndLabels returns all pods in given namespace and given k8s node name for the given labels
// If namespace is empty, it will return pods from all namespaces
func (c *Client) GetPodsByNodeAndLabels(nodeName, namespace string, labels map[string]string) (*corev1.PodList, error) {
	if len(nodeName) == 0 {
		return nil, fmt.Errorf("node name is required for this API")
	}

	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
		LabelSelector: mapToCSV(labels),
	}

	return c.getPodsWithListOptions(namespace, listOptions)
}

// GetPodsByOwner returns pods for the given owner and namespace
func (c *Client) GetPodsByOwner(ownerUID types.UID, namespace string) ([]corev1.Pod, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return common.GetPodsByOwner(c.kubernetes.CoreV1(), ownerUID, namespace)
}

// GetPodsUsingPV returns all pods in cluster using given pv
func (c *Client) GetPodsUsingPV(pvName string) ([]corev1.Pod, error) {
	return c.getPodsUsingPVWithListOptions(pvName, metav1.ListOptions{})
}

// GetPodsUsingPVByNodeName returns all pods running on the node using the given pv
func (c *Client) GetPodsUsingPVByNodeName(pvName, nodeName string) ([]corev1.Pod, error) {
	if len(nodeName) == 0 {
		return nil, fmt.Errorf("node name is required for this API")
	}

	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	}
	return c.getPodsUsingPVWithListOptions(pvName, listOptions)
}

// GetPodsUsingPVC returns all pods in cluster using given pvc
func (c *Client) GetPodsUsingPVC(pvcName, pvcNamespace string) ([]corev1.Pod, error) {
	return c.getPodsUsingPVCWithListOptions(pvcName, pvcNamespace, metav1.ListOptions{})
}

// GetPodsUsingPVCByNodeName returns all pods running on the node using given pvc
func (c *Client) GetPodsUsingPVCByNodeName(pvcName, pvcNamespace, nodeName string) ([]corev1.Pod, error) {
	if len(nodeName) == 0 {
		return nil, fmt.Errorf("node name is required for this API")
	}

	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	}
	return c.getPodsUsingPVCWithListOptions(pvcName, pvcNamespace, listOptions)
}

func (c *Client) getPodsWithListOptions(namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Pods(namespace).List(context.TODO(), opts)
}

func (c *Client) getPodsUsingPVWithListOptions(pvName string, opts metav1.ListOptions) ([]corev1.Pod, error) {
	pv, err := c.GetPersistentVolume(pvName)
	if err != nil {
		return nil, err
	}

	if pv.Status.Phase == corev1.VolumeBound {
		// In some k8s installations, we have seen that the Kind is not populated in the claim ref object. This
		// should be ok since the kind is implicit for "ClaimRef".
		if pv.Spec.ClaimRef != nil && (pv.Spec.ClaimRef.Kind == "PersistentVolumeClaim" || pv.Spec.ClaimRef.Kind == "") {
			return c.getPodsUsingPVCWithListOptions(pv.Spec.ClaimRef.Name, pv.Spec.ClaimRef.Namespace, opts)
		}
	} // else the volume is not bound so cannot rely on stale claim ref objects

	return nil, nil
}

func (c *Client) getPodsUsingPVCWithListOptions(pvcName, pvcNamespace string, opts metav1.ListOptions) ([]corev1.Pod, error) {
	pods, err := c.getPodsWithListOptions(pvcNamespace, opts)
	if err != nil {
		return nil, err
	}

	retList := make([]corev1.Pod, 0)
	for _, p := range pods.Items {
		for _, v := range p.Spec.Volumes {
			if v.PersistentVolumeClaim != nil && v.PersistentVolumeClaim.ClaimName == pvcName {
				// Along PVC present in the volume list, we also checking whether any of the container in the
				// pod is really using it by mount them.
			containerLoop:
				for _, container := range p.Spec.Containers {
					for _, mount := range container.VolumeMounts {
						if mount.Name == v.Name {
							retList = append(retList, p)
							break containerLoop
						}
					}
					// adding check for rawblock volume devices
					for _, device := range container.VolumeDevices {
						if device.Name == v.Name {
							retList = append(retList, p)
							break containerLoop
						}
					}
				}
			}
		}
	}
	return retList, nil
}

// GetPodsUsingVolumePlugin returns all pods who use PVCs provided by the given volume plugin
func (c *Client) GetPodsUsingVolumePlugin(plugin string) ([]corev1.Pod, error) {
	return c.listPluginPodsWithOptions(metav1.ListOptions{}, plugin)
}

// GetPodsUsingVolumePluginByNodeName returns all pods who use PVCs provided by the given volume plugin on the given node
func (c *Client) GetPodsUsingVolumePluginByNodeName(nodeName, plugin string) ([]corev1.Pod, error) {
	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	}

	return c.listPluginPodsWithOptions(listOptions, plugin)
}

func (c *Client) listPluginPodsWithOptions(opts metav1.ListOptions, plugin string) ([]corev1.Pod, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	nodePods, err := c.kubernetes.CoreV1().Pods("").List(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	var retList []corev1.Pod
	for _, p := range nodePods.Items {
		if ok := c.isAnyVolumeUsingVolumePlugin(p.Spec.Volumes, p.Namespace, plugin); ok {
			retList = append(retList, p)
		}
	}

	return retList, nil
}

// GetPodByName returns pod for the given pod name and namespace
func (c *Client) GetPodByName(podName string, namespace string) (*corev1.Pod, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	pod, err := c.kubernetes.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, schederrors.ErrPodsNotFound
	}

	return pod, nil
}

// GetPodByUID returns pod with the given UID, or error if nothing found
func (c *Client) GetPodByUID(uid types.UID, namespace string) (*corev1.Pod, error) {
	pods, err := c.GetPods(namespace, nil)
	if err != nil {
		return nil, err
	}

	pUID := types.UID(uid)
	for _, pod := range pods.Items {
		if pod.UID == pUID {
			return &pod, nil
		}
	}

	return nil, schederrors.ErrPodsNotFound
}

// DeletePods deletes the given pods
func (c *Client) DeletePods(pods []corev1.Pod, force bool) error {
	return common.DeletePods(c.kubernetes.CoreV1(), pods, force)
}

// DeletePod deletes the given pod
func (c *Client) DeletePod(name string, ns string, force bool) error {
	if err := c.initClient(); err != nil {
		return err
	}

	deleteOptions := metav1.DeleteOptions{}
	if force {
		gracePeriodSec := int64(0)
		deleteOptions.GracePeriodSeconds = &gracePeriodSec
	}

	return c.kubernetes.CoreV1().Pods(ns).Delete(context.TODO(), name, deleteOptions)
}

// DeletePodsByLabels deletes pods for the given labels and namespace
func (c *Client) DeletePodsByLabels(namespace string, listOptions map[string]string, timeout time.Duration) error {
	pods, err := c.GetPods(namespace, listOptions)
	if err != nil {
		return err
	}

	var podsNamesToDelete []string
	var podsToDelete []corev1.Pod
	for _, pod := range pods.Items {
		podsNamesToDelete = append(podsNamesToDelete, pod.Name)
		podsToDelete = append(podsToDelete, pod)
	}

	if err := c.DeletePods(pods.Items, false); err != nil {
		return err
	}

	if err := common.WaitForPodsToBeDeleted(c.kubernetes.CoreV1(), podsToDelete, timeout); err != nil {
		return fmt.Errorf("Failed to wait for pods to be deleted: %s, Err: %v", podsNamesToDelete, err)
	}

	return nil
}

// IsPodRunning checks if all containers in a pod are in running state
func (c *Client) IsPodRunning(pod corev1.Pod) bool {
	return common.IsPodRunning(pod)
}

// IsPodCompleted checks if the pod is in completed state
func (c *Client) IsPodCompleted(pod corev1.Pod) bool {
	return common.IsPodCompleted(pod)
}

// IsPodReady checks if all containers in a pod are ready (passed readiness probe)
func (c *Client) IsPodReady(pod corev1.Pod) bool {
	return common.IsPodReady(pod)
}

// IsPodBeingManaged returns true if the pod is being managed by a controller
func (c *Client) IsPodBeingManaged(pod corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Controller != nil && *owner.Controller {
			// We are assuming that if a pod has a owner who has set itself as
			// a controller, the pod is managed. We are not checking for specific
			// contollers like ReplicaSet, StatefulSet as that is
			// 1) requires changes when new controllers get added
			// 2) not handle customer controllers like operators who create pods
			//    directly
			return true
		}
	}
	return false
}

// ValidatePod validates the given pod if it's ready
func (c *Client) ValidatePod(pod *corev1.Pod, timeout, retryInterval time.Duration) error {
	t := func() (interface{}, bool, error) {
		currPod, err := c.GetPodByUID(pod.UID, pod.Namespace)
		if err != nil {
			return "", true, fmt.Errorf("Could not get Pod [%s] %s", pod.Namespace, pod.Name)
		}

		ready := c.IsPodReady(*currPod)
		if !ready {
			return "", true, fmt.Errorf("Pod %s, ID: %s  is not ready. Status %v", currPod.Name, currPod.UID, currPod.Status.Phase)
		}

		return "", false, nil
	}
	if _, err := task.DoRetryWithTimeout(t, timeout, retryInterval); err != nil {
		return err
	}
	return nil
}

// WatchPods sets up a watcher that listens for the changes to pods in given namespace
func (c *Client) WatchPods(namespace string, fn WatchFunc, listOptions metav1.ListOptions) error {
	if err := c.initClient(); err != nil {
		return err
	}

	listOptions.Watch = true
	watchInterface, err := c.kubernetes.CoreV1().Pods(namespace).Watch(context.TODO(), listOptions)
	if err != nil {
		logrus.WithError(err).Error("error invoking the watch api for pods")
		return err
	}

	// fire off watch function
	go c.handleWatch(
		watchInterface,
		&corev1.Pod{},
		namespace,
		fn,
		listOptions)

	return nil
}

// WaitForPodDeletion waits for given timeout for given pod to be deleted
func (c *Client) WaitForPodDeletion(uid types.UID, namespace string, timeout time.Duration) error {
	t := func() (interface{}, bool, error) {
		if err := c.initClient(); err != nil {
			return nil, true, err
		}

		p, err := c.GetPodByUID(uid, namespace)
		if err != nil {
			if err == schederrors.ErrPodsNotFound {
				return nil, false, nil
			}

			return nil, true, err
		}

		if p != nil {
			return nil, true, fmt.Errorf("pod %s:%s (%s) still present in the system", namespace, p.Name, uid)
		}

		return nil, false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, timeout, 5*time.Second); err != nil {
		return err
	}

	return nil
}

// RunCommandInPodEx runs given command in the given pod  (extended syntax)
func (c *Client) RunCommandInPodEx(req *RunCommandInPodExRequest) error {
	if c == nil || req == nil {
		return os.ErrInvalid
	}

	err := c.initClient()
	if err != nil {
		return err
	}

	if len(req.ContainerName) == 0 {
		pod, err := c.kubernetes.CoreV1().Pods(req.Namespace).Get(context.TODO(), req.PODName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if len(pod.Spec.Containers) != 1 {
			return fmt.Errorf("could not determine which container to use")
		}

		req.ContainerName = pod.Spec.Containers[0].Name
	}

	post := c.kubernetes.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(req.PODName).
		Namespace(req.Namespace).
		SubResource("exec")

	post.VersionedParams(&corev1.PodExecOptions{
		Container: req.ContainerName,
		Command:   req.Command,
		Stdin:     (req.Stdin != nil),
		Stdout:    (req.Stdout != nil),
		Stderr:    (req.Stderr != nil),
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", post.URL())
	if err != nil {
		return fmt.Errorf("failed to init executor: %v", err)
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  req.Stdin,
		Stdout: req.Stdout,
		Stderr: req.Stderr,
		Tty:    req.UseTTY,
	})

	return err
}

// RunCommandInPod runs given command in the given pod  (simplified syntax)
func (c *Client) RunCommandInPod(cmds []string, podName, containerName, namespace string) (string, error) {
	var execOut, execErr bytes.Buffer

	err := c.RunCommandInPodEx(&RunCommandInPodExRequest{
		cmds, podName, containerName, namespace, false, nil, &execOut, &execErr,
	})
	if err != nil {
		return execErr.String(), fmt.Errorf("could not execute: %v: %v %v", err, execErr.String(), execOut.String())
	}

	if execErr.Len() > 0 {
		return execErr.String(), nil
	}

	return execOut.String(), nil
}

// GetPodLog returns the logs of a POD as a string
func (c *Client) GetPodLog(podName string, ns string, podLogOptions *corev1.PodLogOptions) (string, error) {
	if err := c.initClient(); err != nil {
		return "", err
	}

	l := c.kubernetes.CoreV1().Pods(ns).GetLogs(podName, podLogOptions)
	buf := new(bytes.Buffer)
	stream, err := l.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer stream.Close()

	_, err = io.Copy(buf, stream)
	if err != nil {
		return "", err
	}
	str := buf.String()

	return str, err
}

// isAnyVolumeUsingVolumePlugin returns true if any of the given volumes is using a storage class for the given plugin
// In case errors are found while looking up a particular volume, the function ignores the errors as the goal is to
// find if there is any match or not
func (c *Client) isAnyVolumeUsingVolumePlugin(volumes []corev1.Volume, volumeNamespace, plugin string) bool {
	for _, v := range volumes {
		if v.PersistentVolumeClaim != nil {
			pvc, err := c.GetPersistentVolumeClaim(v.PersistentVolumeClaim.ClaimName, volumeNamespace)
			if err == nil && pvc != nil {
				provisioner, err := c.GetStorageProvisionerForPVC(pvc)
				if err == nil {
					if provisioner == plugin {
						return true
					}
				}
			}
		}
	}

	return false
}
