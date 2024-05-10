package core

import (
	"context"

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
	// ListEndpoints retrieves endpoints for a given namespace
	ListEndpoints(string, metav1.ListOptions) (*corev1.EndpointsList, error)
	// PatchEndpoints applies a patch for a given endpoints.
	PatchEndpoints(name, namespace string, pt types.PatchType, jsonPatch []byte, subresources ...string) (*corev1.Endpoints, error)
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
	return c.kubernetes.CoreV1().Endpoints(endpoints.Namespace).Create(context.TODO(), endpoints, metav1.CreateOptions{})
}

// GetEndpoints retrieves endpoints for a given namespace/name.
func (c *Client) GetEndpoints(name, ns string) (*corev1.Endpoints, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Endpoints(ns).Get(context.TODO(), name, metav1.GetOptions{})
}

// ListEndpoints retrieves endpoints for a given namespace
func (c *Client) ListEndpoints(ns string, opts metav1.ListOptions) (*corev1.EndpointsList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Endpoints(ns).List(context.TODO(), opts)
}

// PatchEndpoints applies a patch for a given endpoints.
func (c *Client) PatchEndpoints(name, ns string, pt types.PatchType, jsonPatch []byte, subresources ...string) (*corev1.Endpoints, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Endpoints(ns).Patch(context.TODO(), name, pt, jsonPatch, metav1.PatchOptions{}, subresources...)
}

// DeleteEndpoints retrieves endpoints for a given namespace/name.
func (c *Client) DeleteEndpoints(name, ns string) error {
	if err := c.initClient(); err != nil {
		return err
	}
	return c.kubernetes.CoreV1().Endpoints(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// UpdateEndpoints updates the given endpoint.
func (c *Client) UpdateEndpoints(endpoints *corev1.Endpoints) (*corev1.Endpoints, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Endpoints(endpoints.Namespace).Update(context.TODO(), endpoints, metav1.UpdateOptions{})
}
