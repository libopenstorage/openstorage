package alerts

import (
	"time"

	"github.com/libopenstorage/openstorage/api"
)

// Filter defines behavior of a filter.
type Filter interface {
	GetFilterType() FilterType
	GetFilterValue() interface{}
	Match(alert *api.Alert) bool
}

// FilterType contains a filter type.
type FilterType string

// FilterType constants.
const (
	MatchTimeFilter         FilterType = "time"
	MatchAlertTypeFilter    FilterType = "alertType"
	MatchResourceTypeFilter FilterType = "resourceType"
	MatchResourceIDFilter   FilterType = "resourceID"
)

// filter implements Filter interface.
type filter struct {
	filter FilterType
	value  interface{}
}

// timeZone contains information about time window.
type timeZone struct {
	start time.Time
	stop  time.Time
}

func (f *filter) GetFilterType() FilterType {
	return f.filter
}

func (f *filter) GetFilterValue() interface{} {
	return f.value
}

func (f *filter) Match(alert *api.Alert) bool {
	switch f.filter {
	case MatchAlertTypeFilter:
		if alert.AlertType == f.value.(int64) {
			return true
		} else {
			return false
		}
	case MatchResourceTypeFilter:
		if alert.Resource == f.value.(api.ResourceType) {
			return true
		} else {
			return false
		}
	case MatchResourceIDFilter:
		if alert.ResourceId == f.value.(string) {
			return true
		} else {
			return false
		}
	case MatchTimeFilter:
		timeSlot := f.value.(timeZone)
		if alert.FirstSeen.Seconds >= int64(timeSlot.start.Second()) &&
			alert.Timestamp.Seconds <= int64(timeSlot.stop.Second()) {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

// MatchAlertType creates a filter that matches on alert type.
func NewMatchAlertTypeFilter(alertType int64) Filter {
	return &filter{filter: MatchAlertTypeFilter, value: alertType}
}

// MatchResourceType creates a filter that matches on resource type.
func NewMatchResourceTypeFilter(resourceType api.ResourceType) Filter {
	return &filter{filter: MatchResourceTypeFilter, value: resourceType}
}

// MatchResourceID creates a filter that matches on resource id.
func NewMatchResourceIDFilter(resourceID string) Filter {
	return &filter{filter: MatchResourceIDFilter, value: resourceID}
}

// MatchTime creates a filter that matches on alert raised in a given time window.
func NewMatchTimeFilter(start, stop time.Time) Filter {
	return &filter{filter: MatchTimeFilter, value: timeZone{start: start, stop: stop}}
}
