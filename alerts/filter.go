package alerts

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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
// kvdb tree struct is defined as alerts/<resourceType>/<alertType>/<resourceID>/data
//
// Input filters are sorted based
const (
	// This group of filters, when used to fetch directly from kvdb, target all entries in kvdb.
	// Therefore, these filters are considered inefficient in fetching. They fetch _all_ entries
	// and then filter out the fetched entries. Due to this reason these filters need to be handled
	// individually in the Delete method implementation.

	// CustomFilter is based on a user defined function (UDF). All alert entries are fetched from kvdb
	// and then passed through UDF. Entries that match are returned. Therefore, this filter does not query
	// kvdb efficiently.
	CustomFilter FilterType = iota
	// timeSpanFilter is based on a start and end timestamp. All alert entries are fetched from kvdb
	// and then parsed to see if the timestamp for each entry falls within the start and end timestamp
	// of this filter. Matching entries are returned.
	// This filter is not an efficient filter.
	timeSpanFilter
	// countSpanFilter parses on the count value of alert entries. This filter requires pulling all entries
	// and is, therefore, not an efficient filter.
	countSpanFilter
	// minSeverityFilter matches on the alert severity if it is more than or equal to severity set in this filter.
	// This filter can be used for fetching alerts directly but it not an efficient filter. Therefore, the
	// recommended approach is to fetch alerts from kvdb using one of the efficient filters, then
	// filter the fetched alerts using this filter.
	minSeverityFilter
	// flagCheckFilter matches on the clear alert flag. This filter should be used for filtering the fetched
	// alerts and not directly for fetching alerts from kvdb since this is not an efficient filter.
	flagCheckFilter
	// matchResourceIDFilter takes only one argument, i.e., resource id. It fetches all entries from kvdb
	// then parses them to see resource id's are matching. Matching entries are returned.
	// This filter is not an efficient filter since it requires pulling all entries.
	// Recommend to use resourceIDFilter if resource type and alert type info is also known for
	// efficient querying.
	matchResourceIDFilter
	// matchAlertTypeFilter takes only one argume, i.e., an alert type. It fetches all entries from kvdb
	// then parses them to see alert type is matching. Matching entries are returned.
	// This filter is not an efficient filter since it requires pulling all entries.
	// Recommend to use alertTypeFilter if resource type info is also known for
	// efficient querying.
	matchAlertTypeFilter

	// Filter types listed below provide more efficient querying into kvdb by directly querying kvdb sub tree.
	// These filters reach a sub tree in kvdb and only fetch some alerts, therefore, these are called efficient
	// filters.

	// resourceTypeFilter takes resource type and fetches all alert entries under that resource type prefix.
	// Since resource type is a top level indexing of data, it always performs querying efficiently without
	// requiring any further filtering.
	resourceTypeFilter
	// alertTypeFilter draws alert entries under a prefix based on resourceType/alertType.
	// In other words, to use this filter you would need to provide both inputs.
	// This filter also fetches only the required contents which do not require further filtering.
	alertTypeFilter
	// resourceIDFilter, similarly, requires three inputs, i.e., resourceID, alert type and
	// resource type. This filter also fetches efficiently based on prefix defined by these three
	// inputs and requires no further filtering on the fetched contents.
	resourceIDFilter
)

// filter implements Filter interface.
type filter struct {
	filterType FilterType
	value      interface{}
	options    []Option
}

// timeZone contains information about time window.
type timeZone struct {
	start time.Time
	stop  time.Time
}

type alertInfo struct {
	alertType    int64
	resourceType api.ResourceType
	resourceID   string
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
			return false, typeAssertionError.
				Tag("custom filter").
				Tag("func Match")
		}
		return v(alert)
	case timeSpanFilter:
		v, ok := f.value.(timeZone)
		if !ok {
			return false, typeAssertionError.
				Tag("timeSpanFilter").
				Tag("func Match")
		}
		if alert.Timestamp.Seconds >= v.start.Unix() &&
			alert.Timestamp.Seconds <= v.stop.Unix() {
			return true, nil
		}
		return false, nil
	case matchResourceIDFilter:
		v, ok := f.value.(string)
		if !ok {
			return false, typeAssertionError.
				Tag("matchResourceIDFilter").
				Tag("func Match")
		}
		if alert.ResourceId == v {
			return true, nil
		}
		return false, nil
	case matchAlertTypeFilter:
		v, ok := f.value.(int64)
		if !ok {
			return false, typeAssertionError.
				Tag("matchAlertTypeFilter").
				Tag("func Match")
		}
		if alert.AlertType == v {
			return true, nil
		}
		return false, nil
	case countSpanFilter:
		v, ok := f.value.([]int64)
		if !ok {
			return false, typeAssertionError.
				Tag("countSpanFilter").
				Tag("func Match")
		}
		if len(v) != 2 {
			return false, incorrectFilterValue.
				Tag("countSpanFilter").
				Tag("func Match")
		}
		if alert.Count >= v[0] && alert.Count <= v[1] {
			return true, nil
		}
		return false, nil
	case minSeverityFilter:
		v, ok := f.value.(api.SeverityType)
		if !ok {
			return false, typeAssertionError.
				Tag("minSeverityFilter").
				Tag("func Match")
		}
		switch v {
		case api.SeverityType_SEVERITY_TYPE_NONE:
			return true, nil
		case api.SeverityType_SEVERITY_TYPE_NOTIFY,
			api.SeverityType_SEVERITY_TYPE_WARNING,
			api.SeverityType_SEVERITY_TYPE_ALARM:
			if alert.Severity <= v &&
				alert.Severity != api.SeverityType_SEVERITY_TYPE_NONE {
				return true, nil
			}
		}
		return false, nil
	case flagCheckFilter:
		v, ok := f.value.(bool)
		if !ok {
			return false, typeAssertionError.
				Tag("flagCheckFilter").
				Tag("func Match")
		}
		if alert.Cleared == v {
			return true, nil
		}
		return false, nil
	// Cases below are for efficient filters
	// -------------------------------------
	case resourceTypeFilter:
		v, ok := f.value.(api.ResourceType)
		if !ok {
			return false, typeAssertionError.
				Tag("resourceTypeFilter").
				Tag("func Match")
		}
		if alert.Resource == v {
			// iterate through options and match alert
			for _, opt := range f.options {
				if w, ok := opt.GetValue().(Filter); !ok {
					return false, typeAssertionError.
						Tag("invalid option").
						Tag("resourceTypeFilter").
						Tag("func Match")
				} else {
					if matched, err := w.Match(alert); err != nil {
						return false, err
					} else {
						if !matched {
							return false, nil
						}
					}
				}
			}
			return true, nil
		}
		return false, nil
	case alertTypeFilter:
		v, ok := f.value.(alertInfo)
		if !ok {
			return false, typeAssertionError.
				Tag("alertTypeFilter").
				Tag("func Match")
		}
		if alert.AlertType == v.alertType &&
			alert.Resource == v.resourceType {
			// iterate through options and match alert
			for _, opt := range f.options {
				if w, ok := opt.GetValue().(Filter); !ok {
					return false, typeAssertionError.
						Tag("alertTypeFilter").
						Tag("func Match")
				} else {
					if matched, err := w.Match(alert); err != nil {
						return false, err
					} else {
						if !matched {
							return false, nil
						}
					}
				}
			}
			return true, nil
		}
		return false, nil
	case resourceIDFilter:
		v, ok := f.value.(alertInfo)
		if !ok {
			return false, typeAssertionError.
				Tag("resourceIDFilter").
				Tag("func Match")
		}
		if alert.AlertType == v.alertType &&
			alert.Resource == v.resourceType &&
			alert.ResourceId == v.resourceID {
			// iterate through options and match alert
			for _, opt := range f.options {
				if w, ok := opt.GetValue().(Filter); !ok {
					return false, typeAssertionError.
						Tag("resourceIDFilter").
						Tag("func Match")
				} else {
					if matched, err := w.Match(alert); err != nil {
						return false, err
					} else {
						if !matched {
							return false, nil
						}
					}
				}
			}
			return true, nil
		}
		return false, nil
	default:
		return false, invalidFilterType.Tag("func Match")
	}
}

// getUniqueKeysFromFilters analyzes filters and outputs a map of unique keys such that
// these keys do not point to sub trees of one another.
func getUniqueKeysFromFilters(filters ...Filter) (map[string]bool, error) {
	keys := make(map[string]bool)

	// sort filters so we know how to query
	if len(filters) > 0 {
		sort.Sort(Filters(filters))

		for _, filter := range filters {
			key := kvdbKey
			switch filter.GetFilterType() {
			// only these filter types benefit from efficient kvdb querying.
			// for everything else we enumerate and then filter.
			case resourceTypeFilter:
				v, ok := filter.GetValue().(api.ResourceType)
				if !ok {
					return nil, typeAssertionError
				}
				key = filepath.Join(key, v.String())
			case alertTypeFilter:
				v, ok := filter.GetValue().(alertInfo)
				if !ok {
					return nil, typeAssertionError
				}
				key = filepath.Join(key,
					v.resourceType.String(), strconv.FormatInt(v.alertType, 16))
			case resourceIDFilter:
				v, ok := filter.GetValue().(alertInfo)
				if !ok {
					return nil, typeAssertionError
				}
				key = filepath.Join(key,
					v.resourceType.String(), strconv.FormatInt(v.alertType, 16), v.resourceID)
			}

			keyPath, _ := filepath.Split(key)
			keyPath = strings.Trim(keyPath, "/")
			if !keys[keyPath] {
				keys[key] = true
			}

		}
	} else {
		keys[kvdbKey] = true
	}

	// remove all keys that access sub tree of another key
	var keysToDelete []string
	for key := range keys {
		keyPath := key
		for len(keyPath) > 0 {
			keyPath, _ = filepath.Split(keyPath)
			keyPath = strings.Trim(keyPath, "/")
			if keys[keyPath] {
				keysToDelete = append(keysToDelete, key)
				break
			}
		}
	}

	for _, key := range keysToDelete {
		delete(keys, key)
	}

	return keys, nil
}
