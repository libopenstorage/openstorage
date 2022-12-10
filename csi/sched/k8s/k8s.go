package k8s

import (
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/libopenstorage/openstorage/csi/sched"
	"github.com/portworx/sched-ops/k8s/core"
	v1 "k8s.io/api/core/v1"
)

const (
	// Openstorage specific parameters
	osdParameterPrefix   = "csi.openstorage.org/"
	osdPvcNameKey        = osdParameterPrefix + "pvc-name"
	osdPvcNamespaceKey   = osdParameterPrefix + "pvc-namespace"
	osdPvcAnnotationsKey = osdParameterPrefix + "pvc-annotations"
	osdPvcLabelsKey      = osdParameterPrefix + "pvc-labels"

	// CSI keys for PVC metadata
	csiPVCNameKey      = "csi.storage.k8s.io/pvc/name"
	csiPVCNamespaceKey = "csi.storage.k8s.io/pvc/namespace"
)

type k8s struct{}

// PreVolumeCreate takes a CSI request and modifies it for k8s scheduler
func (k *k8s) PreVolumeCreate(req *csi.CreateVolumeRequest) (*csi.CreateVolumeRequest, error) {
	var err error

	// Get PVC
	pvcName, ok := req.Parameters[csiPVCNameKey]
	if !ok {
		return nil, fmt.Errorf("CSI PVC Name/Namespace not provided to this request")
	}
	pvcNamespace, ok := req.Parameters[csiPVCNamespaceKey]
	if !ok {
		return nil, fmt.Errorf("CSI PVC Name/Namespace not provided to this request")
	}
	pvc, err := core.Instance().GetPersistentVolumeClaim(pvcName, pvcNamespace)
	if err != nil {
		return nil, err
	}

	// add pvc name, namespace, annotations and labels in the parameters
	req.Parameters, err = addPVCMetadataParams(req.Parameters, pvc)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewInterceptor returns a interceptor wrapper around the k8s implementation
func NewInterceptor() sched.FilterInterceptor {
	return sched.FilterInterceptor{
		Filter: &k8s{},
	}
}

func addPVCMetadataParams(params map[string]string, pvc *v1.PersistentVolumeClaim) (map[string]string, error) {
	if pvc.Labels == nil {
		pvc.Labels = make(map[string]string)
	}
	if pvc.Annotations == nil {
		pvc.Annotations = make(map[string]string)
	}
	if params == nil {
		params = make(map[string]string)
	}

	// add all annotations to labels. Annotations take precedence, so we will overwrite
	// labels with annotations if they have overlapping keys.
	for k, v := range pvc.Annotations {
		pvc.Labels[k] = v
	}
	for k, v := range pvc.Labels {
		params[k] = v
	}
	params[osdPvcNameKey] = pvc.Name
	params[osdPvcNamespaceKey] = pvc.Namespace

	return params, nil
}
