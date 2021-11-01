package core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
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
	return c.kubernetes.CoreV1().Events(event.Namespace).Create(context.TODO(), event, metav1.CreateOptions{})
}

// ListEvents retrieves all events registered with kubernetes
func (c *Client) ListEvents(namespace string, opts metav1.ListOptions) (*corev1.EventList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Events(namespace).List(context.TODO(), opts)
}

// RecorderOps is an interface to record k8s events
type RecorderOps interface {
	// RecordEvent records an event into k8s using client-go's EventRecorder inteface
	// It takes the event source and the object on which the event is being raised.
	RecordEvent(source v1.EventSource, object runtime.Object, eventtype, reason, message string)
}

func (c *Client) RecordEvent(source v1.EventSource, object runtime.Object, eventtype, reason, message string) {
	if err := c.initClient(); err != nil {
		return
	}
	c.eventRecordersLock.Lock()
	if len(c.eventRecorders) == 0 {
		c.eventRecorders = make(map[string]record.EventRecorder)
		c.eventBroadcaster = record.NewBroadcaster()
		c.eventBroadcaster.StartRecordingToSink(
			&typedcorev1.EventSinkImpl{
				Interface: c.kubernetes.CoreV1().Events(""), // use the namespace from the object
			},
		)
	}
	key := source.Component + "-" + source.Host
	eventRecorder, exists := c.eventRecorders[key]
	if !exists {
		eventRecorder = c.eventBroadcaster.NewRecorder(
			scheme.Scheme,
			source,
		)
		c.eventRecorders[key] = eventRecorder
	}
	c.eventRecordersLock.Unlock()
	eventRecorder.Event(object, eventtype, reason, message)
}
