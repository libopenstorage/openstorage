package alerts

import (
	"time"

	"github.com/libopenstorage/openstorage/api"
)

// Filters is a list of Filters that can be sorted
type Filters []Filter

func (fs Filters) Len() int {
	return len(fs)
}

func (fs Filters) Less(i, j int) bool {
	return (fs)[i].GetFilterType() < (fs)[j].GetFilterType()
}

func (fs Filters) Swap(i, j int) {
	(fs)[i], (fs)[j] = (fs)[j], (fs)[i]
}

// Filter is something that matches an alert.
type Filter interface {
	GetFilterType() FilterType
	GetValue() interface{}
	Match(alert *api.Alert) (bool, error)
}

// FilterType contains a filter type.
type FilterType int

// FilterType constants.
// Please note that the ordering of these filter type constants is very important
// since filters are sorted based on this ordering and such sorting is done to
// properly query kvdb tree structure.
//
// kvdb tree struct is defined as alerts/<resourceType>/<resourceID>/<alertType>/data
//
// Input filters are sorted based
const (
	// CustomFilter
	CustomFilter FilterType = iota
	TimeFilter
	ResourceTypeFilter
	ResourceIDFilter
	AlertTypeFilter
)

// filter implements Filter interface.
type filter struct {
	filterType FilterType
	value      interface{}
}

// timeZone contains information about time window.
type timeZone struct {
	start time.Time
	stop  time.Time
}

type resourceInfo struct {
	resourceID   string
	resourceType api.ResourceType
}

type alertInfo struct {
	alertType    int64
	resourceID   string
	resourceType api.ResourceType
}

func (f *filter) GetFilterType() FilterType {
	return f.filterType
}

func (f *filter) GetValue() interface{} {
	return f.value
}

func (f *filter) Match(alert *api.Alert) (bool, error) {
	switch f.filterType {
	case CustomFilter:
		v, ok := f.value.(func(alert *api.Alert) (bool, error))
		if !ok {
			return false, typeAssertionError
		}
		return v(alert)
	case TimeFilter:
		v, ok := f.value.(timeZone)
		if !ok {
			return false, typeAssertionError
		}
		timeSlot := v
		if alert.FirstSeen.Seconds >= int64(timeSlot.start.Second()) &&
			alert.Timestamp.Seconds <= int64(timeSlot.stop.Second()) {
			return true, nil
		} else {
			return false, nil
		}
	case ResourceTypeFilter:
		v, ok := f.value.(api.ResourceType)
		if !ok {
			return false, typeAssertionError
		}
		if alert.Resource == v {
			return true, nil
		} else {
			return false, nil
		}
	case ResourceIDFilter:
		v, ok := f.value.(resourceInfo)
		if !ok {
			return false, typeAssertionError
		}
		if alert.ResourceId == v.resourceID &&
			alert.Resource == v.resourceType {
			return true, nil
		} else {
			return false, nil
		}
	case AlertTypeFilter:
		v, ok := f.value.(alertInfo)
		if !ok {
			return false, typeAssertionError
		}
		if alert.AlertType == v.alertType &&
			alert.Resource == v.resourceType &&
			alert.ResourceId == v.resourceID {
			return true, nil
		} else {
			return false, nil
		}
	default:
		return false, invalidFilterType
	}
}

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
