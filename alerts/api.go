package alerts

import (
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
)

// NewManager obtains instance of Manager for alerts management.
func NewManager(kv kvdb.Kvdb, options ...Option) (Manager, error) {
	return newManager(kv, options...)
}

func NewTTLOption(ttl uint64) Option {
	return &option{optionType: TTLOption, value: ttl}
}

// Filter API

// NewResourceTypeFilter creates a filter that matches on <resourceType>
func NewResourceTypeFilter(resourceType api.ResourceType) Filter {
	return &filter{filterType: ResourceTypeFilter, value: resourceType}
}

// NewAlertTypeFilter creates a filter that matches on <resourceType>/<alertType>
func NewAlertTypeFilter(alertType int64, resourceType api.ResourceType) Filter {
	return &filter{filterType: AlertTypeFilter, value: alertInfo{
		alertType: alertType, resourceType: resourceType}}
}

// NewResourceIDFilter creates a filter that matches on <resourceType>/<alertType>/<resourceID>
func NewResourceIDFilter(resourceID string, alertType int64, resourceType api.ResourceType) Filter {
	return &filter{filterType: ResourceIDFilter, value: alertInfo{
		resourceID: resourceID, alertType: alertType, resourceType: resourceType}}
}

// NewTimeSpanFilter creates a filter that matches on alert raised in a given time window.
func NewTimeSpanFilter(start, stop time.Time) Filter {
	return &filter{filterType: TimeSpanFilter, value: timeZone{start: start, stop: stop}}
}

// NewInefficientResourceIDFilter provides a filter that matches on resource id.
func NewInefficientResourceIDFilter(resourceID string) Filter {
	return &filter{filterType: InefficientResourceIDFilter, value: resourceID}
}

// NewCountSpanFilter provides a filter that matches on alert count.
func NewCountSpanFilter(minCount, maxCount int64) Filter {
	return &filter{filterType: CountSpanFilter, value: []int64{minCount, maxCount}}
}

// NewCustomFilter creates a filter that matches on UDF (user defined function)
func NewCustomFilter(f func(alert *api.Alert) (bool, error)) Filter {
	return &filter{filterType: CustomFilter, value: f}
}

// Action API

// NewDeleteAction deletes alert entries based on filters.
func NewDeleteAction(filters ...Filter) Action {
	return &action{action: DeleteAction, filters: filters, f: deleteActionFunc}
}

// NewClearAction marks alert entries as cleared that get deleted after half a day of life in kvdb.
func NewClearAction(filters ...Filter) Action {
	return &action{action: ClearAction, filters: filters, f: clearActionFunc}
}

// NewCustomAction takes custom action using user defined function.
func NewCustomAction(f func(manager Manager, filters ...Filter) error, filters ...Filter) Action {
	return &action{action: CustomAction, filters: filters, f: f}
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
