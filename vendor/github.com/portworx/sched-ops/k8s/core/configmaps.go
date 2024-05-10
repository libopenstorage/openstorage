package core

import (
	"context"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

// ConfigMapOps is an interface to perform k8s ConfigMap operations
type ConfigMapOps interface {
	// GetConfigMap gets the config map object for the given name and namespace
	GetConfigMap(name string, namespace string) (*corev1.ConfigMap, error)
	// CreateConfigMap creates a new config map object if it does not already exist.
	CreateConfigMap(configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	// DeleteConfigMap deletes the given config map
	DeleteConfigMap(name, namespace string) error
	// UpdateConfigMap updates the given config map object
	UpdateConfigMap(configMap *corev1.ConfigMap) (*corev1.ConfigMap, error)
	// WatchConfigMap sets up a watcher that listens for changes on the config map
	WatchConfigMap(configMap *corev1.ConfigMap, fn WatchFunc) error
	//ListConfigMap returns the list of ConfigMaps
	ListConfigMap(namespace string, filterOptions metav1.ListOptions) (*corev1.ConfigMapList, error)
}

// GetConfigMap gets the config map object for the given name and namespace
func (c *Client) GetConfigMap(name string, namespace string) (*corev1.ConfigMap, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// CreateConfigMap creates a new config map object if it does not already exist.
func (c *Client) CreateConfigMap(configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	ns := configMap.Namespace
	if len(ns) == 0 {
		ns = corev1.NamespaceDefault
	}

	return c.kubernetes.CoreV1().ConfigMaps(ns).Create(context.TODO(), configMap, metav1.CreateOptions{})
}

// DeleteConfigMap deletes the given config map
func (c *Client) DeleteConfigMap(name, namespace string) error {
	if err := c.initClient(); err != nil {
		return err
	}

	if len(namespace) == 0 {
		namespace = corev1.NamespaceDefault
	}

	return c.kubernetes.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deleteForegroundPolicy,
	})
}

// UpdateConfigMap updates the given config map object
func (c *Client) UpdateConfigMap(configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	ns := configMap.Namespace
	if len(ns) == 0 {
		ns = corev1.NamespaceDefault
	}

	return c.kubernetes.CoreV1().ConfigMaps(ns).Update(context.TODO(), configMap, metav1.UpdateOptions{})
}

// WatchConfigMap sets up a watcher that listens for changes on the config map
func (c *Client) WatchConfigMap(configMap *corev1.ConfigMap, fn WatchFunc) error {
	if err := c.initClient(); err != nil {
		return err
	}

	listOptions := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", configMap.Name).String(),
		Watch:         true,
	}

	watchInterface, err := c.kubernetes.CoreV1().ConfigMaps(configMap.Namespace).Watch(context.TODO(), listOptions)
	if err != nil {
		logrus.WithError(err).Error("error invoking the watch api for config maps")
		return err
	}

	// fire off watch function
	go c.handleWatch(watchInterface, configMap, "", fn, listOptions)
	return nil
}

// ListConfigMap returns the list of ConfigMaps
func (c *Client) ListConfigMap(namespace string, filterOptions metav1.ListOptions) (*corev1.ConfigMapList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().ConfigMaps(namespace).List(context.TODO(), filterOptions)

}
