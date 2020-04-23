package core

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAccountOps is an interface to perform operations on role resources.
type ServiceAccountOps interface {
	// CreateServiceAccount creates the given service account
	CreateServiceAccount(account *corev1.ServiceAccount) (*corev1.ServiceAccount, error)
	// GetServiceAccount gets the given service account
	GetServiceAccount(name, namespace string) (*corev1.ServiceAccount, error)
	// UpdateServiceAccount updates the given service account
	UpdateServiceAccount(account *corev1.ServiceAccount) (*corev1.ServiceAccount, error)
	// DeleteServiceAccount deletes the given service account
	DeleteServiceAccount(accountName, namespace string) error
}

// CreateServiceAccount creates the given service account
func (c *Client) CreateServiceAccount(account *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().ServiceAccounts(account.Namespace).Create(account)
}

// GetServiceAccount gets the given service account
func (c *Client) GetServiceAccount(name, namespace string) (*corev1.ServiceAccount, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
}

// UpdaeServiceAccount updates the given service account
func (c *Client) UpdateServiceAccount(account *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().ServiceAccounts(account.Namespace).Update(account)
}

// DeleteServiceAccount deletes the given service account
func (c *Client) DeleteServiceAccount(accountName, namespace string) error {
	if err := c.initClient(); err != nil {
		return err
	}

	return c.kubernetes.CoreV1().ServiceAccounts(namespace).Delete(accountName, &metav1.DeleteOptions{
		PropagationPolicy: &deleteForegroundPolicy,
	})
}
