package core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceOps is an interface to perform namespace operations
type NamespaceOps interface {
	// ListNamespaces returns all the namespaces
	ListNamespaces(labelSelector map[string]string) (*corev1.NamespaceList, error)
	// GetNamespace returns a namespace object for given name
	GetNamespace(name string) (*corev1.Namespace, error)
	// CreateNamespace creates a namespace with given name and metadata
	CreateNamespace(*corev1.Namespace) (*corev1.Namespace, error)
	// UpdateNamespace update a namespace with given metadata
	UpdateNamespace(*corev1.Namespace) (*corev1.Namespace, error)
	// DeleteNamespace deletes a namespace with given name
	DeleteNamespace(name string) error
}

// ListNamespaces returns all the namespaces
func (c *Client) ListNamespaces(labelSelector map[string]string) (*corev1.NamespaceList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		LabelSelector: mapToCSV(labelSelector),
	})
}

// GetNamespace returns a namespace object for given name
func (c *Client) GetNamespace(name string) (*corev1.Namespace, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
}

// CreateNamespace creates a namespace with given name and metadata
func (c *Client) CreateNamespace(namespace *corev1.Namespace) (*corev1.Namespace, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
}

// DeleteNamespace deletes a namespace with given name
func (c *Client) DeleteNamespace(name string) error {
	if err := c.initClient(); err != nil {
		return err
	}

	return c.kubernetes.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// UpdateNamespace updates a namespace with given metadata
func (c *Client) UpdateNamespace(namespace *corev1.Namespace) (*corev1.Namespace, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
}
