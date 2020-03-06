package core

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretOps is an interface to perform k8s Secret operations
type SecretOps interface {
	// GetSecret gets the secrets object given its name and namespace
	GetSecret(name string, namespace string) (*corev1.Secret, error)
	// CreateSecret creates the given secret
	CreateSecret(*corev1.Secret) (*corev1.Secret, error)
	// UpdateSecret updates the given secret
	UpdateSecret(*corev1.Secret) (*corev1.Secret, error)
	// UpdateSecretData updates or creates a new secret with the given data
	UpdateSecretData(string, string, map[string][]byte) (*corev1.Secret, error)
	// DeleteSecret deletes the given secret
	DeleteSecret(name, namespace string) error
}

// GetSecret gets the secrets object given its name and namespace
func (c *Client) GetSecret(name string, namespace string) (*corev1.Secret, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}

// CreateSecret creates the given secret
func (c *Client) CreateSecret(secret *corev1.Secret) (*corev1.Secret, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Secrets(secret.Namespace).Create(secret)
}

// UpdateSecret updates the given secret
func (c *Client) UpdateSecret(secret *corev1.Secret) (*corev1.Secret, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().Secrets(secret.Namespace).Update(secret)
}

// UpdateSecretData updates or creates a new secret with the given data
func (c *Client) UpdateSecretData(name string, ns string, data map[string][]byte) (*corev1.Secret, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	secret, err := c.GetSecret(name, ns)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.CreateSecret(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: ns,
					},
					Data: data,
				})
		}
		return nil, err
	}

	// This only adds/updates the key value pairs; does not remove the existing.
	for k, v := range data {
		secret.Data[k] = v
	}
	return c.UpdateSecret(secret)
}

// DeleteSecret deletes the given secret
func (c *Client) DeleteSecret(name, namespace string) error {
	if err := c.initClient(); err != nil {
		return err
	}

	return c.kubernetes.CoreV1().Secrets(namespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deleteForegroundPolicy,
	})
}
