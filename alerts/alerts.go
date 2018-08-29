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
)

// Error type for defining error constants.
type Error string

// Error satisfies Error interface from std lib.
func (e Error) Error() string {
	return string(e)
}

const (
	KvdbKey                    = "alerts"
	typeAssertionError   Error = "type assertion error"
	invalidFilterType    Error = "invalid filter type"
	incorrectFilterValue Error = "incorrectly set filter value"
)

const (
	HalfDay  = 60 * 60 * 12
	Day      = 60 * 60 * 24
	FiveDays = 60 * 60 * 24 * 5
)

// Manager manages alerts.
type Manager interface {
	// Raise raises an alert.
	Raise(alert *api.Alert) error
	// Enumerate lists all alerts filtered by a variadic list of filters.
	// It will fetch a superset such that every alert is matched by at least one filter.
	Enumerate(filters ...Filter) ([]*api.Alert, error)
	// Filter filters given list of alerts successively through each filter.
	Filter(alerts []*api.Alert, filters ...Filter) ([]*api.Alert, error)
	// Delete deletes alerts filtered by a chain of filters.
	Delete(filters ...Filter) error
	// SetRules sets a set of rules to be performed on alert events.
	SetRules(rules ...Rule)
	// DeleteRules deletes rules
	DeleteRules(rules ...Rule)
}

func newManager(kv kvdb.Kvdb) *manager {
	return &manager{kv: kv, rules: make(map[string]Rule)}
}

type manager struct {
	kv    kvdb.Kvdb
	rules map[string]Rule
	sync.Mutex
}

// getKey is a util func that constructs kvdb key.
// kvdb tree structure is setup as follows:
// <baseKey>/<resourceType>/<dec2hex(alertType)>/<resourceID>/<alertObject>
func getKey(resourceType string, alertType int64, resourceID string) string {
	return filepath.Join(KvdbKey, resourceType, strconv.FormatInt(alertType, 16), resourceID)
}

func (m *manager) Raise(alert *api.Alert) error {
	for _, rule := range m.rules {
		if rule.GetEvent() == RaiseEvent {
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

	if alert.Cleared {
		// if the alert is marked Cleared, it is pushed to kvdb with a TTL of half day
		_, err := m.kv.Put(getKey(alert.Resource.String(), alert.GetAlertType(), alert.ResourceId), alert, HalfDay)
		return err
	} else {
		_, err := m.kv.Put(getKey(alert.Resource.String(), alert.GetAlertType(), alert.ResourceId), alert, 0)
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
		kvps, err := m.kv.Enumerate(key)
		if err != nil {
			return nil, err
		}

		for _, kvp := range kvps {
			alert := new(api.Alert)
			if err := json.Unmarshal(kvp.Value, alert); err != nil {
				return nil, err
			}

			atLeastOneMatch := false
			for _, filter := range filters {
				if match, err := filter.Match(alert); err != nil {
					return nil, err
				} else {
					if match {
						atLeastOneMatch = match
					}
				}
			}

			if atLeastOneMatch || len(filters) == 0 {
				myAlerts = append(myAlerts, alert)
			}
		}
	}

	return myAlerts, nil
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
		if rule.GetEvent() == DeleteEvent {
			if err := rule.GetAction().Run(m); err != nil {
				return err
			}
		}
	}

	allFiltersIndexBased := true
Loop:
	for _, filter := range filters {
		switch filter.GetFilterType() {
		case CustomFilter, TimeFilter, AlertTypeFilter, ResourceIDFilter, CountFilter:
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
