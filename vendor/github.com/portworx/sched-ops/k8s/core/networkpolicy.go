package core

import (
	"context"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NetworkPolicyOps is an interface to deal with kubernetes NetworkPolicy.
type NetworkPolicyOps interface {
	// CreateNetworkPolicy creates a given network policy.
	CreateNetworkPolicy(NetworkPolicy *v1.NetworkPolicy) (*v1.NetworkPolicy, error)
	// GetNetworkPolicy retrieves NetworkPolicy for a given namespace/name.
	GetNetworkPolicy(name, namespace string) (*v1.NetworkPolicy, error)
	// ListNetworkPolicy retrieves NetworkPolicy for a given namespace/name.
	ListNetworkPolicy(namespace string, listOptions metav1.ListOptions) (*v1.NetworkPolicyList, error)
	// PatchNetworkPolicy applies a patch for a given NetworkPolicy.
	PatchNetworkPolicy(name, namespace string, pt types.PatchType, jsonPatch []byte, subresources ...string) (*v1.NetworkPolicy, error)
	// DeleteNetworkPolicy removes NetworkPolicy for a given namespace/name.
	DeleteNetworkPolicy(name, namespace string) error
	// UpdateNetworkPolicy updates the given networkpolicy
	UpdateNetworkPolicy(NetworkPolicy *v1.NetworkPolicy) (*v1.NetworkPolicy, error)
}

// CreateNetworkPolicy creates a given NetworkPolicy.
func (c *Client) CreateNetworkPolicy(NetworkPolicy *v1.NetworkPolicy) (*v1.NetworkPolicy, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.NetworkingV1().NetworkPolicies(NetworkPolicy.Namespace).Create(context.TODO(), NetworkPolicy, metav1.CreateOptions{})
}

// GetNetworkPolicy retrieves NetworkPolicy for a given namespace/name.
func (c *Client) GetNetworkPolicy(name, ns string) (*v1.NetworkPolicy, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.NetworkingV1().NetworkPolicies(ns).Get(context.TODO(), name, metav1.GetOptions{})
}

// ListNetworkPolicy retrieves NetworkPolicies for a given namespace.
func (c *Client) ListNetworkPolicy(ns string, opts metav1.ListOptions) (*v1.NetworkPolicyList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.NetworkingV1().NetworkPolicies(ns).List(context.TODO(), opts)
}

// PatchNetworkPolicy applies a patch for a given NetworkPolicy.
func (c *Client) PatchNetworkPolicy(name, ns string, pt types.PatchType, jsonPatch []byte, subresources ...string) (*v1.NetworkPolicy, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.NetworkingV1().NetworkPolicies(ns).Patch(context.TODO(), name, pt, jsonPatch, metav1.PatchOptions{}, subresources...)
}

// DeleteNetworkPolicy retrieves NetworkPolicy for a given namespace/name.
func (c *Client) DeleteNetworkPolicy(name, ns string) error {
	if err := c.initClient(); err != nil {
		return err
	}
	return c.kubernetes.NetworkingV1().NetworkPolicies(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// UpdateNetworkPolicy updates the given network policy.
func (c *Client) UpdateNetworkPolicy(NetworkPolicy *v1.NetworkPolicy) (*v1.NetworkPolicy, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.NetworkingV1().NetworkPolicies(NetworkPolicy.Namespace).Update(context.TODO(), NetworkPolicy, metav1.UpdateOptions{})
}
