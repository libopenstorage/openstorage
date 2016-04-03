package alert

import (
	"github.com/libopenstorage/openstorage/api"

	"go.pedge.io/dlog"
)

// AlertInstance is a singleton which should be used to raise/clear alert
type alertInstance struct {
	NodeId    string
	ClusterId string
	Version   string
	kva       AlertClient
}

// Singleton AlertInstance
var inst *alertInstance

// Clear clears an alert
func (ai *alertInstance) Clear(resource Resource, resourceId string, alertID int64) {
	resourceType, _ := ai.getResource(resource, resourceId)
	if err := ai.kva.Clear(resourceType, alertID); err != nil {
		dlog.Errorf("Failed to clear alert, type: %v, id: %v", resourceType, alertID)
	}
}

// Alarm raises an alert with severity : ALARM
func (ai *alertInstance) Alarm(name string, msg string, resource Resource, resourceId string) (int64, error) {
	alert, err := ai.raiseAlert(name, msg, resource, resourceId, api.SeverityType_SEVERITY_TYPE_ALARM)
	return alert.Id, err
}

// Notify raises an alert with severity : NOTIFY
func (ai *alertInstance) Notify(name string, msg string, resource Resource, resourceId string) (int64, error) {
	alert, err := ai.raiseAlert(name, msg, resource, resourceId, api.SeverityType_SEVERITY_TYPE_NOTIFY)
	return alert.Id, err
}

// Warn raises an alert with severity : WARNING
func (ai *alertInstance) Warn(name string, msg string, resource Resource, resourceId string) (int64, error) {
	alert, err := ai.raiseAlert(name, msg, resource, resourceId, api.SeverityType_SEVERITY_TYPE_WARNING)
	return alert.Id, err
}

// Alert :  Keeping this function for backward compatibility until we remove all calls to this function
func (ai *alertInstance) Alert(name string, msg string) error {
	// Do nothing
	return nil
}

// Sets the new singleton alert instance
func newAlertInstance(nodeId, clusterId, version string, kva AlertClient) {
	inst = &alertInstance{
		NodeId:    nodeId,
		ClusterId: clusterId,
		Version:   version,
		kva:       kva,
	}
}

// Instance returns the singleton AlertInstance
func instance() *alertInstance {
	return inst
}

func (ai *alertInstance) getResource(resource Resource, id string) (resourceType api.ResourceType, resourceId string) {
	resourceId = id
	if resource == Volume {
		resourceType = api.ResourceType_RESOURCE_TYPE_VOLUME
	} else if resource == Node {
		resourceType = api.ResourceType_RESOURCE_TYPE_NODE
		if resourceId == "" {
			resourceId = ai.NodeId
		}
	} else if resource == Cluster {
		resourceType = api.ResourceType_RESOURCE_TYPE_CLUSTER
		if resourceId == "" {
			resourceId = ai.ClusterId
		}
	} else {
		resourceType = api.ResourceType_RESOURCE_TYPE_NONE
	}
	return resourceType, resourceId
}

func (ai *alertInstance) raiseAlert(name string, msg string, resource Resource, resourceId string, severity api.SeverityType) (api.Alert, error) {
	resourceType, resourceId := ai.getResource(resource, resourceId)
	alert, err := ai.kva.Raise(api.Alert{
		Resource:   resourceType,
		ResourceId: resourceId,
		Severity:   severity,
		Name:       name,
		Message:    msg,
	})
	return alert, err
}
