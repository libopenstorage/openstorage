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

// NewFilterDeleter obtains instance of FilterDeleter for alerts enumeration and deletion.
func NewFilterDeleter(kv kvdb.Kvdb, options ...Option) (FilterDeleter, error) {
	return newManager(kv, options...)
}

// Option API

// NewTTLOption provides an option to be used in manager creation.
func NewTTLOption(ttl uint64) Option {
	return &option{optionType: ttlOption, value: ttl}
}

// NewTimeSpanOption provides an option to be used in filter definition.
// Filters that take options, apply options only during matching alerts.
func NewTimeSpanOption(start, stop time.Time) Option {
	return &option{optionType: timeSpanOption, value: NewTimeSpanFilter(start, stop)}
}

// NewCountSpanOption provides an option to be used in filter definition that
// accept options. Only filters that are efficient in querying kvdb accept options
// and apply these options during matching alerts.
func NewCountSpanOption(minCount, maxCount int64) Option {
	return &option{optionType: countSpanOption, value: NewCountSpanFilter(minCount, maxCount)}
}

// NewMinSeverityOption provides an option to be used during filter creation that
// accept such options. Only filters that are efficient in querying kvdb accept options
// and apply these options during matching alerts.
func NewMinSeverityOption(minSev api.SeverityType) Option {
	return &option{optionType: minSeverityOption, value: NewMinSeverityFilter(minSev)}
}

// NewFlagCheckOptions provides an option to be used during filter creation that
// accept such options. Only filters that are efficient in querying kvdb accept options
// and apply these options during matching alerts.
func NewFlagCheckOption(flag bool) Option {
	return &option{optionType: flagCheckOption, value: NewFlagCheckFilter(flag)}
}

// NewResourceIdOption provides an option to be used during filter creation that
// accept such options. Only filters that are efficient in querying kvdb accept options
// and apply these options during matching alerts.
func NewResourceIdOption(resourceId string) Option {
	return &option{optionType: resourceIdOption, value: NewMatchResourceIDFilter(resourceId)}
}

// Filter API

// NewResourceTypeFilter creates a filter that matches on <resourceType>
func NewResourceTypeFilter(resourceType api.ResourceType, options ...Option) Filter {
	return &filter{filterType: resourceTypeFilter, value: resourceType, options: options}
}

// NewAlertTypeFilter creates a filter that matches on <resourceType>/<alertType>
func NewAlertTypeFilter(alertType int64, resourceType api.ResourceType, options ...Option) Filter {
	return &filter{filterType: alertTypeFilter, value: alertInfo{
		alertType: alertType, resourceType: resourceType},
		options: options}
}

// NewResourceIDFilter creates a filter that matches on <resourceType>/<alertType>/<resourceID>
func NewResourceIDFilter(resourceID string, alertType int64, resourceType api.ResourceType, options ...Option) Filter {
	return &filter{filterType: resourceIDFilter, value: alertInfo{
		resourceID: resourceID, alertType: alertType, resourceType: resourceType},
		options: options}
}

// NewTimeSpanFilter creates a filter that matches on alert raised in a given time window.
func NewTimeSpanFilter(start, stop time.Time) Filter {
	return &filter{filterType: timeSpanFilter, value: timeZone{start: start, stop: stop}}
}

// NewMatchAlertTypeFilter provides a filter that matches on alert type.
// Please use NewAlertTypeFilter if other inputs are known.
func NewMatchAlertTypeFilter(alertType int64) Filter {
	return &filter{filterType: matchAlertTypeFilter, value: alertType}
}

// NewMatchResourceIDFilter provides a filter that matches on resource id.
// Please use NewResourceIDFilter if other inputs are known.
func NewMatchResourceIDFilter(resourceID string) Filter {
	return &filter{filterType: matchResourceIDFilter, value: resourceID}
}

// NewCountSpanFilter provides a filter that matches on alert count.
func NewCountSpanFilter(minCount, maxCount int64) Filter {
	return &filter{filterType: countSpanFilter, value: []int64{minCount, maxCount}}
}

// NewMinSeverityFilter provides a filter that matches on alert when severity is greater than
// or equal to the minSev value.
func NewMinSeverityFilter(minSev api.SeverityType) Filter {
	return &filter{filterType: minSeverityFilter, value: minSev}
}

// NewFlagCheckFilter provides a filter that matches on alert clear flag.
func NewFlagCheckFilter(flag bool) Filter {
	return &filter{filterType: flagCheckFilter, value: flag}
}

// NewCustomFilter creates a filter that matches on UDF (user defined function)
func NewCustomFilter(f func(alert *api.Alert) (bool, error)) Filter {
	return &filter{filterType: CustomFilter, value: f}
}

// Action API

// NewDeleteAction deletes alert entries based on filters.
func NewDeleteAction(filters ...Filter) Action {
	return &action{action: deleteAction, filters: filters, f: deleteActionFunc}
}

// NewClearAction marks alert entries as cleared that get deleted after half a day of life in kvdb.
func NewClearAction(filters ...Filter) Action {
	return &action{action: clearAction, filters: filters, f: clearActionFunc}
}

// NewCustomAction takes custom action using user defined function.
func NewCustomAction(f func(manager Manager, filters ...Filter) error, filters ...Filter) Action {
	return &action{action: CustomAction, filters: filters, f: f}
}

// Rule API

// NewRaiseRule creates a rule that runs action when a raised alerts matche filter.
// Action happens before incoming alert is raised.
func NewRaiseRule(name string, filter Filter, action Action) Rule {
	return &rule{name: name, event: raiseEvent, filter: filter, action: action}
}

// NewDeleteRule creates a rule that runs action when deleted alerts matche filter.
// Action happens after matching alerts are deleted.
func NewDeleteRule(name string, filter Filter, action Action) Rule {
	return &rule{name: name, event: deleteEvent, filter: filter, action: action}
}
