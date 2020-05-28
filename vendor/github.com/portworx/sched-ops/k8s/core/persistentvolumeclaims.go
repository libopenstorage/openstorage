package core

import (
	"fmt"
	"time"

	"github.com/portworx/sched-ops/k8s/common"
	schederrors "github.com/portworx/sched-ops/k8s/errors"
	"github.com/portworx/sched-ops/task"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeClaimOps is an interface to perform k8s PVC operations
type PersistentVolumeClaimOps interface {
	// CreatePersistentVolumeClaim creates the given persistent volume claim
	CreatePersistentVolumeClaim(*corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error)
	// UpdatePersistentVolumeClaim updates an existing persistent volume claim
	UpdatePersistentVolumeClaim(*corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error)
	// DeletePersistentVolumeClaim deletes the given persistent volume claim
	DeletePersistentVolumeClaim(name, namespace string) error
	// ValidatePersistentVolumeClaim validates the given pvc
	ValidatePersistentVolumeClaim(vv *corev1.PersistentVolumeClaim, timeout, retryInterval time.Duration) error
	// ValidatePersistentVolumeClaimSize validates the given pvc size
	ValidatePersistentVolumeClaimSize(vv *corev1.PersistentVolumeClaim, expectedPVCSize int64, timeout, retryInterval time.Duration) error
	// GetPersistentVolumeClaim returns the PVC for given name and namespace
	GetPersistentVolumeClaim(pvcName string, namespace string) (*corev1.PersistentVolumeClaim, error)
	// GetPersistentVolumeClaims returns all PVCs in given namespace and that match the optional labelSelector
	GetPersistentVolumeClaims(namespace string, labelSelector map[string]string) (*corev1.PersistentVolumeClaimList, error)
	// CreatePersistentVolume creates the given PV
	CreatePersistentVolume(pv *corev1.PersistentVolume) (*corev1.PersistentVolume, error)
	// GetPersistentVolume returns the PV for given name
	GetPersistentVolume(pvName string) (*corev1.PersistentVolume, error)
	// DeletePersistentVolume deletes the PV for given name
	DeletePersistentVolume(pvName string) error
	// GetPersistentVolumes returns all PVs in cluster
	GetPersistentVolumes() (*corev1.PersistentVolumeList, error)
	// GetVolumeForPersistentVolumeClaim returns the volumeID for the given PVC
	GetVolumeForPersistentVolumeClaim(*corev1.PersistentVolumeClaim) (string, error)
	// GetPersistentVolumeClaimParams fetches custom parameters for the given PVC
	GetPersistentVolumeClaimParams(*corev1.PersistentVolumeClaim) (map[string]string, error)
	// GetPersistentVolumeClaimStatus returns the status of the given pvc
	GetPersistentVolumeClaimStatus(*corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaimStatus, error)
	// GetPVCsUsingStorageClass returns all PVCs that use the given storage class
	GetPVCsUsingStorageClass(scName string) ([]corev1.PersistentVolumeClaim, error)
	// GetStorageProvisionerForPVC returns storage provisioner for given PVC if it exists
	GetStorageProvisionerForPVC(pvc *corev1.PersistentVolumeClaim) (string, error)
	// GetStorageClassForPVC returns the appropriate storage class object for a certain pvc
	GetStorageClassForPVC(pvc *corev1.PersistentVolumeClaim) (*storagev1.StorageClass, error)
}

// CreatePersistentVolumeClaim creates the given persistent volume claim
func (c *Client) CreatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	ns := pvc.Namespace
	if len(ns) == 0 {
		ns = corev1.NamespaceDefault
	}

	return c.kubernetes.CoreV1().PersistentVolumeClaims(ns).Create(pvc)
}

// UpdatePersistentVolumeClaim updates an existing persistent volume claim
func (c *Client) UpdatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	ns := pvc.Namespace
	if len(ns) == 0 {
		ns = corev1.NamespaceDefault
	}

	return c.kubernetes.CoreV1().PersistentVolumeClaims(ns).Update(pvc)
}

// DeletePersistentVolumeClaim deletes the given persistent volume claim
func (c *Client) DeletePersistentVolumeClaim(name, namespace string) error {
	if err := c.initClient(); err != nil {
		return err
	}

	return c.kubernetes.CoreV1().PersistentVolumeClaims(namespace).Delete(name, &metav1.DeleteOptions{})
}

// ValidatePersistentVolumeClaim validates the given pvc
func (c *Client) ValidatePersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim, timeout, retryInterval time.Duration) error {
	t := func() (interface{}, bool, error) {
		if err := c.initClient(); err != nil {
			return "", true, err
		}

		result, err := c.kubernetes.CoreV1().
			PersistentVolumeClaims(pvc.Namespace).
			Get(pvc.Name, metav1.GetOptions{})
		if err != nil {
			return "", true, err
		}

		if result.Status.Phase == corev1.ClaimBound {
			return "", false, nil
		}

		return "", true, &schederrors.ErrPVCNotReady{
			ID:    result.Name,
			Cause: fmt.Sprintf("PVC expected status: %v PVC actual status: %v", corev1.ClaimBound, result.Status.Phase),
		}
	}

	if _, err := task.DoRetryWithTimeout(t, timeout, retryInterval); err != nil {
		return err
	}
	return nil
}

// ValidatePersistentVolumeClaimSize validates the given pvc size
func (c *Client) ValidatePersistentVolumeClaimSize(pvc *corev1.PersistentVolumeClaim, expectedPVCSize int64, timeout, retryInterval time.Duration) error {
	t := func() (interface{}, bool, error) {
		if err := c.initClient(); err != nil {
			return "", true, err
		}

		result, err := c.kubernetes.CoreV1().
			PersistentVolumeClaims(pvc.Namespace).
			Get(pvc.Name, metav1.GetOptions{})
		if err != nil {
			return "", true, err
		}

		capacity, ok := result.Status.Capacity[corev1.ResourceName(corev1.ResourceStorage)]
		if !ok {
			return "", true, fmt.Errorf("failed to get storage size for pvc: %v", pvc.Name)
		}

		if capacity.Value() == expectedPVCSize {
			return "", false, nil
		}

		return "", true, &schederrors.ErrValidatePVCSize{
			ID:    result.Name,
			Cause: fmt.Sprintf("PVC expected size: %v actual size: %v", expectedPVCSize, capacity.Value()),
		}
	}

	if _, err := task.DoRetryWithTimeout(t, timeout, retryInterval); err != nil {
		return err
	}
	return nil
}

// CreatePersistentVolume creates the given PV
func (c *Client) CreatePersistentVolume(pv *corev1.PersistentVolume) (*corev1.PersistentVolume, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().PersistentVolumes().Create(pv)
}

// GetPersistentVolumeClaim returns the PVC for given name and namespace
func (c *Client) GetPersistentVolumeClaim(pvcName string, namespace string) (*corev1.PersistentVolumeClaim, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().PersistentVolumeClaims(namespace).
		Get(pvcName, metav1.GetOptions{})
}

// GetPersistentVolumeClaims returns all PVCs in given namespace and that match the optional labelSelector
func (c *Client) GetPersistentVolumeClaims(namespace string, labelSelector map[string]string) (*corev1.PersistentVolumeClaimList, error) {
	return c.getPVCsWithListOptions(namespace, metav1.ListOptions{
		LabelSelector: mapToCSV(labelSelector),
	})
}

func (c *Client) getPVCsWithListOptions(namespace string, listOpts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().PersistentVolumeClaims(namespace).List(listOpts)
}

// GetPersistentVolume returns the PV for given name
func (c *Client) GetPersistentVolume(pvName string) (*corev1.PersistentVolume, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
}

// DeletePersistentVolume deletes the PV for given name
func (c *Client) DeletePersistentVolume(pvName string) error {
	if err := c.initClient(); err != nil {
		return err
	}

	return c.kubernetes.CoreV1().PersistentVolumes().Delete(pvName, &metav1.DeleteOptions{
		PropagationPolicy: &deleteForegroundPolicy,
	})
}

// GetPersistentVolumes returns all PVs in cluster
func (c *Client) GetPersistentVolumes() (*corev1.PersistentVolumeList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
}

// GetVolumeForPersistentVolumeClaim returns the volumeID for the given PVC
func (c *Client) GetVolumeForPersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) (string, error) {
	result, err := c.GetPersistentVolumeClaim(pvc.Name, pvc.Namespace)
	if err != nil {
		return "", err
	}

	return result.Spec.VolumeName, nil
}

// GetPersistentVolumeClaimStatus returns the status of the given pvc
func (c *Client) GetPersistentVolumeClaimStatus(pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaimStatus, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	result, err := c.kubernetes.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(pvc.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &result.Status, nil
}

// GetPersistentVolumeClaimParams fetches custom parameters for the given PVC
func (c *Client) GetPersistentVolumeClaimParams(pvc *corev1.PersistentVolumeClaim) (map[string]string, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	params := make(map[string]string)

	result, err := c.kubernetes.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(pvc.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	capacity, ok := result.Spec.Resources.Requests[corev1.ResourceName(corev1.ResourceStorage)]
	if !ok {
		return nil, fmt.Errorf("failed to get storage resource for pvc: %v", result.Name)
	}

	// We explicitly send the unit with so the client can compare it with correct units
	requestGB := uint64(roundUpSize(capacity.Value(), 1024*1024*1024))
	params["size"] = fmt.Sprintf("%dG", requestGB)

	sc, err := c.GetStorageClassForPVC(result)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage class for pvc: %v", result.Name)
	}

	for key, value := range sc.Parameters {
		params[key] = value
	}

	return params, nil
}

// GetPVCsUsingStorageClass returns all PVCs that use the given storage class
func (c *Client) GetPVCsUsingStorageClass(scName string) ([]corev1.PersistentVolumeClaim, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	var retList []corev1.PersistentVolumeClaim
	pvcs, err := c.kubernetes.CoreV1().PersistentVolumeClaims("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pvc := range pvcs.Items {
		sc, err := c.GetStorageClassForPVC(&pvc)
		if err == nil && sc.Name == scName {
			retList = append(retList, pvc)
		}
	}

	return retList, nil
}

// GetStorageProvisionerForPVC returns storage provisioner for given PVC if it exists
func (c *Client) GetStorageProvisionerForPVC(pvc *corev1.PersistentVolumeClaim) (string, error) {
	// first try to get the provisioner directly from the annotations
	provisionerName, present := pvc.Annotations[pvcStorageProvisionerKey]
	if present {
		return provisionerName, nil
	}

	sc, err := c.GetStorageClassForPVC(pvc)
	if err != nil {
		return "", err
	}

	return sc.Provisioner, nil
}

// isPVCShared returns true if the PersistentVolumeClaim has been configured for use by multiple clients
func (c *Client) isPVCShared(pvc *corev1.PersistentVolumeClaim) bool {
	for _, mode := range pvc.Spec.AccessModes {
		if mode == corev1.PersistentVolumeAccessMode(corev1.ReadOnlyMany) ||
			mode == corev1.PersistentVolumeAccessMode(corev1.ReadWriteMany) {
			return true
		}
	}

	return false
}

// GetStorageClassForPVC returns the appropriate storage class object for a certain pvc
func (c *Client) GetStorageClassForPVC(pvc *corev1.PersistentVolumeClaim) (*storagev1.StorageClass, error) {
	return common.GetStorageClassForPVC(c.kubernetes.StorageV1(), pvc)
}
