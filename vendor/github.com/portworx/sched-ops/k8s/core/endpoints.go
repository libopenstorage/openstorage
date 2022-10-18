package core

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EndpointsOps is an interface to deal with kubernetes endpoints.
type EndpointsOps interface {
	// CreateEndpoints creates a given endpoints.
	CreateEndpoints(endpoints *corev1.Endpoints) (*corev1.Endpoints, error)
	// GetEndpoints retrieves endpoints for a given namespace/name.
	GetEndpoints(name, namespace string) (*corev1.Endpoints, error)
	// PatchEndpoints applies a patch for a given endpoints.
	PatchEndpoints(name, namespace string, pt types.PatchType, jsonPatch []byte) (*corev1.Endpoints, error)
	// DeleteEndpoints removes endpoints for a given namespace/name.
	DeleteEndpoints(name, namespace string) error
	// UpdateEndpoints updates the given endpoint
	UpdateEndpoints(endpoints *corev1.Endpoints) (*corev1.Endpoints, error)
}

// CreateEndpoints creates a given endpoints.
func (c *Client) CreateEndpoints(endpoints *corev1.Endpoints) (*corev1.Endpoints, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Endpoints(endpoints.Namespace).Create(endpoints)
}

// GetEndpoints retrieves endpoints for a given namespace/name.
func (c *Client) GetEndpoints(name, ns string) (*corev1.Endpoints, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Endpoints(ns).Get(name, metav1.GetOptions{})
}

// PatchEndpoints applies a patch for a given endpoints.
func (c *Client) PatchEndpoints(name, ns string, pt types.PatchType, jsonPatch []byte) (*corev1.Endpoints, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Endpoints(ns).Patch(name, pt, jsonPatch)
}

// DeleteEndpoints retrieves endpoints for a given namespace/name.
func (c *Client) DeleteEndpoints(name, ns string) error {
	if err := c.initClient(); err != nil {
		return err
	}
	return c.kubernetes.CoreV1().Endpoints(ns).Delete(name, nil)
}

// UpdateEndpoints updates the given endpoint.
func (c *Client) UpdateEndpoints(endpoints *corev1.Endpoints) (*corev1.Endpoints, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Endpoints(endpoints.Namespace).Update(endpoints)
}
