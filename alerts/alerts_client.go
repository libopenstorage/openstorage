package alerts

import (
	"github.com/libopenstorage/openstorage/api"

	"go.pedge.io/dlog"
)

// AlertsInstance is a singleton which should be used to raise/clear alerts
type alertsInstance struct {
	NodeId    string
	ClusterId string
	Version   string
	kva       AlertsClient
}

// Singleton AlertsInstance
var inst *alertsInstance

// Clear clears an alert
func (ai *alertsInstance) Clear(resource Resource, resourceId string, alertID int64) {
	resourceType, _ := ai.getResource(resource, resourceId)
	if err := ai.kva.Clear(resourceType, alertID); err != nil {
		dlog.Errorf("Failed to clear alert, type: %v, id: %v", resourceType,
			alertID)
	}
}

// Alarm raises an alert with severity : ALARM
func (ai *alertsInstance) Alarm(name string, msg string, resource Resource, resourceId string) (int64, error) {
	alert, err := ai.raiseAlert(name, msg, resource, resourceId, api.SeverityType_ALARM)
	return alert.Id, err
}

// Notify raises an alert with severity : NOTIFY
func (ai *alertsInstance) Notify(name string, msg string, resource Resource, resourceId string) (int64, error) {
	alert, err := ai.raiseAlert(name, msg, resource, resourceId, api.SeverityType_NOTIFY)
	return alert.Id, err
}

// Warn raises an alert with severity : WARNING
func (ai *alertsInstance) Warn(name string, msg string, resource Resource, resourceId string) (int64, error) {
	alert, err := ai.raiseAlert(name, msg, resource, resourceId, api.SeverityType_WARNING)
	return alert.Id, err
}

// Alert :  Keeping this function for backward compatibility until we remove all calls to this function
func (ai *alertsInstance) Alert(name string, msg string) error {
	// Do nothing
	return nil
}

// Sets the new singleton alerts instance
func newAlertsInstance(nodeId, clusterId, version string, kva AlertsClient) {
	inst = &alertsInstance{
		NodeId:    nodeId,
		ClusterId: clusterId,
		Version:   version,
		kva:       kva,
	}
}

// Instance returns the singleton AlertsInstance
func instance() *alertsInstance {
	return inst
}


func (ai *alertsInstance) getResource(resource Resource, id string) (resourceType api.ResourceType, resourceId string) {
	resourceId = id
	if resource == Volume {
		resourceType = api.ResourceType_VOLUMES
	} else if resource == Node {
		resourceType = api.ResourceType_NODE
		if resourceId == "" {
			resourceId = ai.NodeId
		}
	} else if resource == Cluster {
		resourceType = api.ResourceType_CLUSTER
		if resourceId == "" {
			resourceId = ai.ClusterId
		}
	} else {
		resourceType = api.ResourceType_UNKNOWN_RESOURCE
	}
	return resourceType, resourceId
}

func (ai *alertsInstance) raiseAlert(name string, msg string, resource Resource, resourceId string, severity api.SeverityType) (api.Alerts, error) {
	resourceType, resourceId := ai.getResource(resource, resourceId)
	alert, err := ai.kva.Raise(api.Alerts{
		Resource:   resourceType,
		ResourceId: resourceId,
		Severity:   severity,
		Name:       name,
		Message:    msg,
	})
	return alert, err
}
