//go:generate mockgen -package=mockalerts -destination=mock/alerts.mock.go github.com/libopenstorage/openstorage/alerts FilterDeleter
package alerts

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
	"github.com/sirupsen/logrus"
)

// Error type for defining error constants.
type Error string

// Error satisfies Error interface from std lib.
func (e Error) Error() string {
	return string(e)
}

// Tag tags an error with a tag string.
// Helpful for providing error contexts.
func (e Error) Tag(tag Error) Error {
	return Error(string(tag) + ":" + string(e))
}

const (
	kvdbKey                    = "alerts"
	typeAssertionError   Error = "type assertion error"
	invalidFilterType    Error = "invalid filter type"
	invalidOptionType    Error = "invalid option type"
	incorrectFilterValue Error = "incorrectly set filter value"
)

const (
	HalfDay  = Day / 2
	Day      = 60 * 60 * 24
	FiveDays = Day * 5
)

// Manager manages alerts.
type Manager interface {
	// FilterDeleter allows read only operation on alerts
	FilterDeleter
	// Raise raises an alert.
	Raise(alert *api.Alert) error
	// SetRules sets a set of rules to be performed on alert events.
	SetRules(rules ...Rule)
	// DeleteRules deletes rules
	DeleteRules(rules ...Rule)
}

// FilterDeleter defines a list and delete interface on alerts.
// This interface is used in SDK.
type FilterDeleter interface {
	// Enumerate lists all alerts filtered by a variadic list of filters.
	// It will fetch a superset such that every alert is matched by at least one filter.
	Enumerate(filters ...Filter) ([]*api.Alert, error)
	// Filter filters given list of alerts successively through each filter.
	Filter(alerts []*api.Alert, filters ...Filter) ([]*api.Alert, error)
	// Delete deletes alerts filtered by a chain of filters.
	Delete(filters ...Filter) error
}

func newManager(kv kvdb.Kvdb, options ...Option) (*manager, error) {
	m := &manager{kv: kv, rules: make(map[string]Rule), ttl: HalfDay}
	for _, option := range options {
		switch option.GetType() {
		case ttlOption:
			v, ok := option.GetValue().(uint64)
			if !ok {
				return nil, typeAssertionError
			}
			m.ttl = v
		}
	}
	return m, nil
}

// manager implements Manager interface.
type manager struct {
	kv    kvdb.Kvdb
	rules map[string]Rule
	ttl   uint64
	sync.Mutex
}

// getKey is a util func that constructs kvdb key.
// kvdb tree structure is setup as follows:
// <baseKey>/<resourceType>/<dec2hex(alertType)>/<resourceID>/<alertObject>
func getKey(resourceType string, alertType int64, resourceID string) string {
	return filepath.Join(kvdbKey, resourceType, strconv.FormatInt(alertType, 16), resourceID)
}

func (m *manager) Raise(alert *api.Alert) error {
	for _, rule := range m.rules {
		if rule.GetEvent() == raiseEvent {
			match, err := rule.GetFilter().Match(alert)
			if err != nil {
				return err
			}
			if match {
				if err := rule.GetAction().Run(m); err != nil {
					return err
				}
			}
		}
	}

	if alert.Timestamp == nil {
		alert.Timestamp = &timestamp.Timestamp{Seconds: time.Now().Unix()}
	}

	key := getKey(alert.Resource.String(), alert.GetAlertType(), alert.ResourceId)
	if _, err := m.kv.Delete(key); err != nil && err != kvdb.ErrNotFound {
		logrus.WithField("pkg", "openstorage/alerts").WithField("func", "Raise").Error(err)
	}

	// ttl is time to live. it indicates how long (in seconds) the object should live inside kvdb backend.
	// kvdb will delete the object once ttl elapses.
	if alert.Cleared {
		// if the alert is marked Cleared, it is pushed to kvdb with a ttlOption of half day
		_, err := m.kv.Put(key, alert, m.ttl)
		return err
	} else {
		// otherwise use the ttl value embedded in the alert object
		_, err := m.kv.Put(key, alert, alert.Ttl)
		return err
	}
}

// Enumerate takes a variadic list of filters that are first analyzed to see if one filter
// is inclusive of other. Only the filters that are unique supersets are retained and their contents
// is fetched using kvdb enumerate.
func (m *manager) Enumerate(filters ...Filter) ([]*api.Alert, error) {
	myAlerts := make([]*api.Alert, 0, 0)
	keys, err := getUniqueKeysFromFilters(filters...)
	if err != nil {
		return nil, err
	}

	// enumerate for unique keys
	for key := range keys {
		kvps, err := enumerate(m.kv, key)
		if err != nil {
			return nil, err
		}

		for _, kvp := range kvps {
			alert := new(api.Alert)
			if err := json.Unmarshal(kvp.Value, alert); err != nil {
				return nil, err
			}

			if len(filters) == 0 {
				myAlerts = append(myAlerts, alert)
				continue
			}

			for _, filter := range filters {
				if match, err := filter.Match(alert); err != nil {
					return nil, err
				} else {
					// if alert is matched by at least one filter,
					// include it and break out of loop to avoid further checks.
					if match {
						myAlerts = append(myAlerts, alert)
						break
					}
				}
			}
		}
	}

	return myAlerts, nil
}

// enumerate recursively fetches kvpairs.
// Recursive call is required since, unlike mem kv, an etcd or consul based kv will not return
// leaf objects if the key is a prefix referencing only higher level paths. For instance if the
// kvdb structure is as follows:
// a/b/c/<data>
// a/B/C/<data>
// then enumerating for keys using "a" will only return "b" and "B".
func enumerate(kv kvdb.Kvdb, key string) (kvdb.KVPairs, error) {
	kvps, err := kv.Enumerate(key)
	if err != nil {
		return nil, err
	}

	var keys []string
	var out kvdb.KVPairs
	for _, kvp := range kvps {
		kvp := kvp
		if len(kvp.Value) == 0 {
			keys = append(keys, kvp.Key)
			continue
		}
		out = append(out, kvp)
	}

	for _, key := range keys {
		kvps, err := enumerate(kv, key)
		if err != nil {
			return nil, err
		}
		out = append(out, kvps...)
	}
	return out, nil
}

func (m *manager) Filter(alerts []*api.Alert, filters ...Filter) ([]*api.Alert, error) {
	for _, filter := range filters {
		i := 0
		for j, alert := range alerts {
			match, err := filter.Match(alert)
			if err != nil {
				return nil, err
			}

			if match {
				alerts[i] = alerts[j]
				i += 1
			}
		}

		// shrink the list
		alerts = alerts[:i]

		if len(alerts) == 0 {
			return alerts, nil
		}
	}

	return alerts, nil
}

func (m *manager) Delete(filters ...Filter) error {
	for _, rule := range m.rules {
		if rule.GetEvent() == deleteEvent {
			if err := rule.GetAction().Run(m); err != nil {
				return err
			}
		}
	}

	allFiltersIndexBased := true
Loop:
	for _, filter := range filters {
		switch filter.GetFilterType() {
		case CustomFilter,
			timeSpanFilter,
			alertTypeFilter,
			countSpanFilter,
			minSeverityFilter,
			flagCheckFilter,
			matchAlertTypeFilter,
			matchResourceIDFilter:
			allFiltersIndexBased = false
			break Loop
		}
	}

	if allFiltersIndexBased {
		keys, err := getUniqueKeysFromFilters(filters...)
		if err != nil {
			return err
		}

		for key := range keys {
			if err := m.kv.DeleteTree(key); err != nil {
				return err
			}
		}
	} else {
		myAlerts, err := m.Enumerate(filters...)
		if err != nil {
			return err
		}

		for _, alert := range myAlerts {
			if _, err := m.kv.Delete(getKey(alert.Resource.String(), alert.GetAlertType(), alert.ResourceId)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *manager) SetRules(rules ...Rule) {
	m.Lock()
	defer m.Unlock()
	for _, rule := range rules {
		m.rules[rule.GetName()] = rule
	}
}

func (m *manager) DeleteRules(rules ...Rule) {
	m.Lock()
	defer m.Unlock()
	for _, rule := range rules {
		delete(m.rules, rule.GetName())
	}
}
