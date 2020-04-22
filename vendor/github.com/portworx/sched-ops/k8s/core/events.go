package core

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventOps is an interface to put and get k8s events
type EventOps interface {
	// CreateEvent puts an event into k8s etcd
	CreateEvent(event *corev1.Event) (*corev1.Event, error)
	// ListEvents retrieves all events registered with kubernetes
	ListEvents(namespace string, opts metav1.ListOptions) (*corev1.EventList, error)
}

// CreateEvent puts an event into k8s etcd
func (c *Client) CreateEvent(event *corev1.Event) (*corev1.Event, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Events(event.Namespace).Create(event)
}

// ListEvents retrieves all events registered with kubernetes
func (c *Client) ListEvents(namespace string, opts metav1.ListOptions) (*corev1.EventList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Events(namespace).List(opts)
}
