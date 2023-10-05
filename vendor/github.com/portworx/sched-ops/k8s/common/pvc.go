package common

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/storage/v1"
)

// IsPVCShared returns true if the PersistentVolumeClaim has been configured for use by multiple clients.
func IsPVCShared(pvc *corev1.PersistentVolumeClaim) bool {
	for _, mode := range pvc.Spec.AccessModes {
		if mode == corev1.ReadOnlyMany || mode == corev1.ReadWriteMany {
			return true
		}
	}

	return false
}

// GetStorageClassForPVC tries to find a storage class by pvc spec definitions or by pvc annotations.
func GetStorageClassForPVC(client v1.StorageV1Interface, pvc *corev1.PersistentVolumeClaim) (*storagev1.StorageClass, error) {
	var scName string
	if pvc.Spec.StorageClassName != nil && len(*pvc.Spec.StorageClassName) > 0 {
		scName = *pvc.Spec.StorageClassName
	} else {
		scName = pvc.Annotations[corev1.BetaStorageClassAnnotation]
	}

	if len(scName) == 0 {
		return nil, fmt.Errorf("PVC: %s does not have a storage class", pvc.Name)
	}

	return client.StorageClasses().Get(context.TODO(), scName, metav1.GetOptions{})
}
