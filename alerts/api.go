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

// NewQueryResourceTypeFilter creates a filter that matches on <resourceType>
func NewQueryResourceTypeFilter(resourceType api.ResourceType) Filter {
	return &filter{filterType: QueryResourceTypeFilter, value: resourceType}
}

// NewQueryAlertTypeFilter creates a filter that matches on <resourceType>/<alertType>
func NewQueryAlertTypeFilter(alertType int64, resourceType api.ResourceType) Filter {
	return &filter{filterType: QueryAlertTypeFilter, value: alertInfo{
		alertType: alertType, resourceType: resourceType}}
}

// NewQueryResourceIDFilter creates a filter that matches on <resourceType>/<alertType>/<resourceID>
func NewQueryResourceIDFilter(resourceID string, alertType int64, resourceType api.ResourceType) Filter {
	return &filter{filterType: QueryResourceIDFilter, value: alertInfo{
		resourceID: resourceID, alertType: alertType, resourceType: resourceType}}
}

// NewTimeFilter creates a filter that matches on alert raised in a given time window.
func NewTimeFilter(start, stop time.Time) Filter {
	return &filter{filterType: TimeFilter, value: timeZone{start: start, stop: stop}}
}

// NewAlertTypeFilter provides a filter that matches on alerty type.
// Please note that if you are filtering for alert types for a given resource type, it is better
// to use NewQueryAlertTypeFilter instead of this filter since it is more efficient in terms of
// kvdb access.
func NewAlertTypeFilter(alertType int64) Filter {
	return &filter{filterType: AlertTypeFilter, value: alertType}
}

// NewResourceIDFilter provides a filter that matches on resource id.
func NewResourceIDFilter(resourceID string) Filter {
	return &filter{filterType: ResourceIDFilter, value: resourceID}
}

// NewCountFilter provides a filter that matches on alert count.
func NewCountFilter(count int64) Filter {
	return &filter{filterType: CountFilter, value: count}
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

// NewRaiseRule creates a rule that runs action when a raised alerts matche filter.
// Action happens before incoming alert is raised.
func NewRaiseRule(name string, filter Filter, action Action) Rule {
	return &rule{name: name, event: RaiseEvent, filter: filter, action: action}
}

// NewDeleteRule creates a rule that runs action when deleted alerts matche filter.
// Action happens after matching alerts are deleted.
func NewDeleteRule(name string, filter Filter, action Action) Rule {
	return &rule{name: name, event: DeleteEvent, filter: filter, action: action}
}
