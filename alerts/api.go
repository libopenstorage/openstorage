package alerts

import (
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
)

// NewManager obtains instance of Manager for alerts management.
func NewManager(kv kvdb.Kvdb) Manager {
	return newManager(kv)
}

// Filter API

// NewResourceTypeFilter creates a filter that matches on resource type.
func NewResourceTypeFilter(resourceType api.ResourceType) Filter {
	return &filter{filterType: ResourceTypeFilter, value: resourceType}
}

// NewResourceIDFilter creates a filter that matches on resource id.
func NewResourceIDFilter(resourceID string, resourceType api.ResourceType) Filter {
	return &filter{filterType: ResourceIDFilter, value: resourceInfo{
		resourceID: resourceID, resourceType: resourceType}}
}

// NewAlertTypeFilter creates a filter that matches on alert type.
func NewAlertTypeFilter(alertType int64, resourceID string, resourceType api.ResourceType) Filter {
	return &filter{filterType: AlertTypeFilter, value: alertInfo{
		alertType: alertType, resourceID: resourceID, resourceType: resourceType}}
}

// NewTimeFilter creates a filter that matches on alert raised in a given time window.
func NewTimeFilter(start, stop time.Time) Filter {
	return &filter{filterType: TimeFilter, value: timeZone{start: start, stop: stop}}
}

// NewCustomFilter creates a filter that matches on UDF (user defined function)
func NewCustomFilter(f func(alert *api.Alert) (bool, error)) Filter {
	return &filter{filterType: CustomFilter, value: f}
}

// Action API

// NewDeleteAction deletes alert entries based on filters.
func NewDeleteAction(filter ...Filter) Action {
	return &action{action: DeleteAction, filters: filter, f: deleteAction}
}

// NewCustomAction takes custom action using user defined function.
func NewCustomAction(f func(manager Manager, filter ...Filter) error, filter ...Filter) Action {
	return &action{action: CustomAction, filters: filter, f: f}
}

// Rule API

// NewRaiseRule creates a rule that activates when a raised alerts matches filter.
func NewRaiseRule(filter Filter, action Action) Rule {
	return &rule{event: RaiseEvent, filter: filter, action: action}
}

// NewDeleteRule creates a rule that activates when deleted alert matches filter.
func NewDeleteRule(filter Filter, action Action) Rule {
	return &rule{event: DeleteEvent, filter: filter, action: action}
}
