package core

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/portworx/sched-ops/task"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
)

const (
	masterLabelKey           = "node-role.kubernetes.io/master"
	pvcStorageProvisionerKey = "volume.beta.kubernetes.io/storage-provisioner"
	labelUpdateMaxRetries    = 5
)

var (
	instance Ops
	once     sync.Once

	deleteForegroundPolicy = metav1.DeletePropagationForeground
)

// Ops is an interface to perform kubernetes related operations on the core resources.
type Ops interface {
	ConfigMapOps
	EndpointsOps
	EventOps
	RecorderOps
	NamespaceOps
	NodeOps
	PersistentVolumeClaimOps
	PodOps
	SecretOps
	ServiceOps
	ServiceAccountOps
	LimitRangeOps

	// SetConfig sets the config and resets the client
	SetConfig(config *rest.Config)
	// GetVersion gets the version from the kubernetes cluster
	GetVersion() (*version.Info, error)
	// ResourceExists returns true if given resource type exists in kubernetes API server
	ResourceExists(schema.GroupVersionKind) (bool, error)
}

// Instance returns a singleton instance of the client.
func Instance() Ops {
	once.Do(func() {
		if instance == nil {
			instance = &Client{}
		}
	})
	return instance
}

// SetInstance replaces the instance with the provided one. Should be used only for testing purposes.
func SetInstance(i Ops) {
	instance = i
}

// New builds a new client.
func New(kubernetes kubernetes.Interface) *Client {
	return &Client{
		kubernetes: kubernetes,
	}
}

// NewForConfig builds a new client for the given config.
func NewForConfig(c *rest.Config) (*Client, error) {
	kubernetes, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return &Client{
		kubernetes: kubernetes,
	}, nil
}

// NewInstanceFromConfigFile returns new instance of client by using given
// config file
func NewInstanceFromConfigFile(config string) (Ops, error) {
	newInstance := &Client{}
	err := newInstance.loadClientFromKubeconfig(config)
	if err != nil {
		return nil, err
	}
	return newInstance, nil
}

// Client is a wrapper for kubernetes core client.
type Client struct {
	config     *rest.Config
	kubernetes kubernetes.Interface
	// eventRecorders is a map of component to event recorders
	eventRecorders     map[string]record.EventRecorder
	eventRecordersLock sync.Mutex
	eventBroadcaster   record.EventBroadcaster
}

// SetConfig sets the config and resets the client.
func (c *Client) SetConfig(cfg *rest.Config) {
	c.config = cfg
	c.kubernetes = nil
}

// GetVersion returns server version
func (c *Client) GetVersion() (*version.Info, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.Discovery().ServerVersion()
}

// ResourceExists checks if resource already exists
func (c *Client) ResourceExists(gvk schema.GroupVersionKind) (bool, error) {
	if err := c.initClient(); err != nil {
		return false, err
	}
	_, apiLists, err := c.kubernetes.Discovery().ServerGroupsAndResources()
	if err != nil {
		return false, err
	}
	for _, apiList := range apiLists {
		if apiList.GroupVersion == gvk.GroupVersion().String() {
			for _, r := range apiList.APIResources {
				if r.Kind == gvk.Kind {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// initClient the k8s client if uninitialized
func (c *Client) initClient() error {
	if c.kubernetes != nil {
		return nil
	}

	return c.setClient()
}

// setClient instantiates a client.
func (c *Client) setClient() error {
	var err error

	if c.config != nil {
		err = c.loadClient()
	} else {
		kubeconfig := os.Getenv("KUBECONFIG")
		if len(kubeconfig) > 0 {
			err = c.loadClientFromKubeconfig(kubeconfig)
		} else {
			err = c.loadClientFromServiceAccount()
		}

	}
	return err
}

// loadClientFromServiceAccount loads a k8s client from a ServiceAccount specified in the pod running px
func (c *Client) loadClientFromServiceAccount() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	c.config = config
	return c.loadClient()
}

func (c *Client) loadClientFromKubeconfig(kubeconfig string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	c.config = config
	return c.loadClient()
}

func (c *Client) loadClient() error {
	if c.config == nil {
		return fmt.Errorf("rest config is not provided")
	}

	var err error

	c.kubernetes, err = kubernetes.NewForConfig(c.config)
	if err != nil {
		return err
	}
	return nil
}

// WatchFunc is a callback provided to the Watch functions
// which is invoked when the given object is changed.
type WatchFunc func(object runtime.Object) error

// handleWatch is internal function that handles the watch.  On channel shutdown (ie. stop watch),
// it'll attempt to reestablish its watch function.
func (c *Client) handleWatch(
	watchInterface watch.Interface,
	object runtime.Object,
	namespace string,
	fn WatchFunc,
	listOptions metav1.ListOptions) {
	defer watchInterface.Stop()
	for {
		select {
		case event, more := <-watchInterface.ResultChan():
			if !more {
				logrus.Debug("Kubernetes watch closed (attempting to re-establish)")

				t := func() (interface{}, bool, error) {
					var err error
					if node, ok := object.(*corev1.Node); ok {
						err = c.WatchNode(node, fn)
					} else if cm, ok := object.(*corev1.ConfigMap); ok {
						err = c.WatchConfigMap(cm, fn)
					} else if _, ok := object.(*corev1.Pod); ok {
						err = c.WatchPods(namespace, fn, listOptions)
					} else if sc, ok := object.(*corev1.Secret); ok {
						err = c.WatchSecret(sc, fn)
					} else {
						return "", false, fmt.Errorf("unsupported object: %v given to handle watch", object)
					}

					return "", true, err
				}

				if _, err := task.DoRetryWithTimeout(t, 10*time.Minute, 10*time.Second); err != nil {
					logrus.WithError(err).Error("Could not re-establish the watch")
				} else {
					logrus.Debug("watch re-established")
				}
				return
			}

			fn(event.Object)
		}
	}
}
