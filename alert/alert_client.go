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
func (ai *alertInstance) Clear(resourceType api.ResourceType, resourceId string, alertID int64) {
	if err := ai.kva.Clear(resourceType, alertID); err != nil {
		dlog.Errorf("Failed to clear alert, type: %v, id: %v", resourceType, alertID)
	}
}

// Alarm raises an alert with severity : ALARM
func (ai *alertInstance) Alarm(name string, msg string, resourceType api.ResourceType, resourceId string) (int64, error) {
	alert, err := ai.raiseAlert(name, msg, resourceType, resourceId, api.SeverityType_SEVERITY_TYPE_ALARM)
	return alert.Id, err
}

// Notify raises an alert with severity : NOTIFY
func (ai *alertInstance) Notify(name string, msg string, resourceType api.ResourceType, resourceId string) (int64, error) {
	alert, err := ai.raiseAlert(name, msg, resourceType, resourceId, api.SeverityType_SEVERITY_TYPE_NOTIFY)
	return alert.Id, err
}

// Warn raises an alert with severity : WARNING
func (ai *alertInstance) Warn(name string, msg string, resourceType api.ResourceType, resourceId string) (int64, error) {
	alert, err := ai.raiseAlert(name, msg, resourceType, resourceId, api.SeverityType_SEVERITY_TYPE_WARNING)
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

func (ai *alertInstance) raiseAlert(name string, msg string, resourceType api.ResourceType, resourceId string, severity api.SeverityType) (api.Alert, error) {
	alert, err := ai.kva.Raise(api.Alert{
		Resource:   resourceType,
		ResourceId: resourceId,
		Severity:   severity,
		Name:       name,
		Message:    msg,
	})
	return alert, err
}
