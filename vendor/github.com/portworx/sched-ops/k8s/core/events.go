package core

import (
	"context"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/tools/record"
)

// EventOps is an interface to put and get k8s events
type EventOps interface {
	// CreateEvent puts an event into k8s etcd
	CreateEvent(event *corev1.Event) (*corev1.Event, error)
	// UpdateEvent updates an event in k8s etcd
	UpdateEvent(event *corev1.Event) (*corev1.Event, error)
	// ListEvents retrieves all events registered with kubernetes
	ListEvents(namespace string, opts metav1.ListOptions) (*corev1.EventList, error)
	// WatchEvents sets up a watcher that listens for events in given namespace or all namespaces if the namespace is empty
	WatchEvents(namespace string, fn WatchFunc, listOptions metav1.ListOptions) error
}

// CreateEvent puts an event into k8s etcd
func (c *Client) CreateEvent(event *corev1.Event) (*corev1.Event, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Events(event.Namespace).Create(context.TODO(), event, metav1.CreateOptions{})
}

// UpdateEvent updates an event in k8s etcd
func (c *Client) UpdateEvent(event *corev1.Event) (*corev1.Event, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Events(event.Namespace).Update(context.TODO(), event, metav1.UpdateOptions{})
}

// ListEvents retrieves all events registered with kubernetes
func (c *Client) ListEvents(namespace string, opts metav1.ListOptions) (*corev1.EventList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Events(namespace).List(context.TODO(), opts)
}

// WatchEvents sets up a watcher that listens for events in given namespace or all namespaces if the namespace is empty
func (c *Client) WatchEvents(namespace string, fn WatchFunc, listOptions metav1.ListOptions) error {
	if err := c.initClient(); err != nil {
		return err
	}

	listOptions.Watch = true
	watchInterface, err := c.kubernetes.CoreV1().Events(namespace).Watch(context.TODO(), listOptions)
	if err != nil {
		logrus.WithError(err).Error("error invoking the watch api for events")
		return err
	}

	// fire off watch function
	go c.handleWatch(
		watchInterface,
		&corev1.Event{},
		namespace,
		fn,
		listOptions)

	return nil
}

// RecorderOps is an interface to record k8s events
type RecorderOps interface {
	// RecordEventf records a new-style event into k8s using client-go's new events.EventRecorder interface.
	RecordEventf(
		reportingController string, regarding runtime.Object, related runtime.Object, eventType, reason, action string,
		note string, args ...interface{},
	)

	// RecordEvent is deprecated. Use RecordEventf instead.
	// This function translates old-style event params to record a new-style event. Kept for backwards compatibility.
	// New callers should use RecordEventf directly.
	RecordEvent(source corev1.EventSource, object runtime.Object, eventtype, reason, message string)

	// RecordEventLegacy records an old-style event into k8s using client-go's record.EventRecorder (deprecated) interface.
	// It takes the event source and the object on which the event is being raised.
	// Deprecated. Use RecordEventf instead.
	RecordEventLegacy(source corev1.EventSource, object runtime.Object, eventtype, reason, message string)
}

// RecordEventLegacy see RecorderOps interface
func (c *Client) RecordEventLegacy(source corev1.EventSource, object runtime.Object, eventtype, reason, message string) {
	if err := c.initClient(); err != nil {
		return
	}
	c.eventRecordersLock.Lock()
	if len(c.eventRecordersLegacy) == 0 {
		c.eventRecordersLegacy = make(map[string]record.EventRecorder)
		c.eventBroadcasterLegacy = record.NewBroadcaster()
		c.eventBroadcasterLegacy.StartRecordingToSink(
			&typedcorev1.EventSinkImpl{
				Interface: c.kubernetes.CoreV1().Events(""), // use the namespace from the object
			},
		)
	}
	key := source.Component + "-" + source.Host
	eventRecorder, exists := c.eventRecordersLegacy[key]
	if !exists {
		eventRecorder = c.eventBroadcasterLegacy.NewRecorder(
			scheme.Scheme,
			source,
		)
		c.eventRecordersLegacy[key] = eventRecorder
	}
	c.eventRecordersLock.Unlock()
	eventRecorder.Event(object, eventtype, reason, message)
}

// RecordEvent see RecorderOps interface
func (c *Client) RecordEvent(source corev1.EventSource, object runtime.Object, eventtype, reason, message string) {
	c.RecordEventf(source.Component, object, nil, eventtype, reason, "Unspecified", message)
}

// RecordEventf see RecorderOps interface
func (c *Client) RecordEventf(
	reportingController string, regarding runtime.Object, related runtime.Object, eventType, reason, action string,
	note string, args ...interface{},
) {
	if err := c.initClient(); err != nil {
		return
	}
	c.eventRecordersLock.Lock()
	if c.eventRecordersNew == nil {
		c.eventRecordersNew = make(map[string]events.EventRecorder)
		eventSink := &events.EventSinkImpl{Interface: c.kubernetes.EventsV1()}
		c.eventBroadcasterNew = events.NewBroadcaster(eventSink)
		stopCh := make(chan struct{}) // TODO: should we use stopCh someplace before exiting?
		c.eventBroadcasterNew.StartRecordingToSink(stopCh)
	}
	eventRecorder, exists := c.eventRecordersNew[reportingController]
	if !exists {
		eventRecorder = c.eventBroadcasterNew.NewRecorder(scheme.Scheme, reportingController)
		c.eventRecordersNew[reportingController] = eventRecorder
	}
	c.eventRecordersLock.Unlock()
	eventRecorder.Eventf(regarding, related, eventType, reason, action, note, args...)
}
