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
		if alert.Timestamp.Seconds >= v.start.Unix() &&
			alert.Timestamp.Seconds <= v.stop.Unix() {
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

// getKeysFromFilters analyzes filters and outputs a map of unique keys such that
// these keys do not point to sub trees of one another.
func getKeysFromFilters(filters ...Filter) (map[string]bool, error) {
	keys := make(map[string]bool)

	// sort filters so we know how to query
	if len(filters) > 0 {
		sort.Sort(Filters(filters))

		for _, filter := range filters {
			key := KvdbKey
			switch filter.GetFilterType() {
			case CustomFilter, TimeFilter:
			case ResourceTypeFilter:
				v, ok := filter.GetValue().(api.ResourceType)
				if !ok {
					return nil, typeAssertionError
				}
				key = filepath.Join(key, v.String())
			case ResourceIDFilter:
				v, ok := filter.GetValue().(resourceInfo)
				if !ok {
					return nil, typeAssertionError
				}
				key = filepath.Join(key, v.resourceType.String(), v.resourceID)
			case AlertTypeFilter:
				v, ok := filter.GetValue().(alertInfo)
				if !ok {
					return nil, typeAssertionError
				}
				key = filepath.Join(key,
					v.resourceType.String(), v.resourceID, strconv.FormatInt(v.alertType, 16))
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
