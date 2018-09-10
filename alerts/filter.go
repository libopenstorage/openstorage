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
	// Filter types listed below do not query kvdb directly. Instead they enumerate and then filter.

	// CustomFilter is based on a user defined function (UDF). All alert entries are fetched from kvdb
	// and then passed through UDF. Entries that match are returned. Therefore, this filter does not query
	// kvdb efficiently.
	CustomFilter FilterType = iota
	// TimeSpanFilter is based on a start and end timestamp. All alert entries are fetched from kvdb
	// and then parsed to see if the timestamp for each entry falls within the start and end timestamp
	// of this filter. Matching entries are returned.
	// This filter is not an efficient filter.
	TimeSpanFilter
	// CounterFilter parses on the count value of alert entries. This filter requires pulling all entries
	// and is, therefore, not an efficient filter.
	CountSpanFilter
	// InefficientResourceIDFilter takes only one argument, i.e., resource id. It fetches all entries from kvdb
	// then parses them to see resource id's are matching. Matching entries are returned.
	// This filter is not an efficient filter since it requires pulling all entries.
	// Recommend to use ResourceIDFilter is you resource type and alert type info is also known for
	// efficient querying.
	InefficientResourceIDFilter

	// Filter types listed below provide more efficient querying into kvdb by directly querying kvdb sub tree.

	// ResourceTypeFilter takes resource type and fetches all alert entries under that resource type prefix.
	// Since resource type is a top level indexing of data, it always performs querying efficiently without
	// requiring any further filtering.
	ResourceTypeFilter
	// AlertTypeFilter draws alert entries under a prefix based on resourceType/alertType.
	// In other words, to use this filter you would need to provide both inputs.
	// This filter also fetches only the required contents which do not require further filtering.
	AlertTypeFilter
	// ResourceIDFilter, similarly, requires three inputs, i.e., resourceID, alert type and
	// resource type. This filter also fetches efficiently based on prefix defined by these three
	// inputs and requires no further filtering on the fetched contents.
	ResourceIDFilter
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
			return false, typeAssertionError
		}
		return v(alert)
	case TimeSpanFilter:
		v, ok := f.value.(timeZone)
		if !ok {
			return false, typeAssertionError
		}
		if alert.Timestamp.Seconds >= v.start.Unix() &&
			alert.Timestamp.Seconds <= v.stop.Unix() {
			return true, nil
		} else {
			return false, nil
		}
	case InefficientResourceIDFilter:
		v, ok := f.value.(string)
		if !ok {
			return false, typeAssertionError
		}
		if alert.ResourceId == v {
			return true, nil
		} else {
			return false, nil
		}
	case CountSpanFilter:
		v, ok := f.value.([]int64)
		if !ok {
			return false, typeAssertionError
		}
		if len(v) != 2 {
			return false, incorrectFilterValue
		}
		if alert.Count >= v[0] && alert.Count <= v[1] {
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
	case AlertTypeFilter:
		v, ok := f.value.(alertInfo)
		if !ok {
			return false, typeAssertionError
		}
		if alert.AlertType == v.alertType &&
			alert.Resource == v.resourceType {
			return true, nil
		} else {
			return false, nil
		}
	case ResourceIDFilter:
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

// getUniqueKeysFromFilters analyzes filters and outputs a map of unique keys such that
// these keys do not point to sub trees of one another.
func getUniqueKeysFromFilters(filters ...Filter) (map[string]bool, error) {
	keys := make(map[string]bool)

	// sort filters so we know how to query
	if len(filters) > 0 {
		sort.Sort(Filters(filters))

		for _, filter := range filters {
			key := KvdbKey
			switch filter.GetFilterType() {
			// only these filter types benefit from efficient kvdb querying.
			// for everything else we enumerate and then filter.
			case ResourceTypeFilter:
				v, ok := filter.GetValue().(api.ResourceType)
				if !ok {
					return nil, typeAssertionError
				}
				key = filepath.Join(key, v.String())
			case AlertTypeFilter:
				v, ok := filter.GetValue().(alertInfo)
				if !ok {
					return nil, typeAssertionError
				}
				key = filepath.Join(key,
					v.resourceType.String(), strconv.FormatInt(v.alertType, 16))
			case ResourceIDFilter:
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
		keys[KvdbKey] = true
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
