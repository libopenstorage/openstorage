package v1beta1

import (
	portworx "github.com/portworx/talisman/pkg/apis/portworx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: portworx.GroupName, Version: portworx.Version}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder is the scheme builder for the types
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme applies all the stored functions to the scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Cluster{},
		&ClusterList{},
		&VolumePlacementStrategy{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
