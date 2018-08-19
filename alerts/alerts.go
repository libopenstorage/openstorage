package alerts

import (
	"path/filepath"
	"strconv"

	"encoding/json"

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
	KvdbKey                  = "alerts"
	typeAssertionError Error = "type assertion error"
	invalidFilterType  Error = "invalid filter type"
)

// Manager manages alerts.
type Manager interface {
	// Raise raises an alert.
	Raise(alert *api.Alert) error
	// Enumerate lists all alerts filtered by a chain of filters.
	Enumerate(filters ...Filter) ([]*api.Alert, error)
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

	key := filepath.Join(KvdbKey,
		alert.Resource.String(),
		alert.ResourceId,
		strconv.FormatInt(alert.GetAlertType(), 16))
	if alert.Timestamp == nil {
		alert.Timestamp = &timestamp.Timestamp{Seconds: time.Now().Unix()}
	}
	_, err := m.kv.Put(key, alert, 0)
	return err
}

// Enumerate takes a variadic list of filters that are first analyzed to see if one filter
// is inclusive of other. Only the filters that are unique supersets are retained and their contents
// is fetched using kvdb enumerate.
func (m *manager) Enumerate(filters ...Filter) ([]*api.Alert, error) {
	myAlerts := make([]*api.Alert, 0, 0)
	keys, err := getKeysFromFilters(filters...)
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
		case CustomFilter, TimeFilter:
			allFiltersIndexBased = false
			break Loop
		}
	}

	if allFiltersIndexBased {
		keys, err := getKeysFromFilters(filters...)
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
			key := filepath.Join(KvdbKey,
				alert.Resource.String(),
				alert.ResourceId,
				strconv.FormatInt(alert.GetAlertType(), 16))
			if _, err := m.kv.Delete(key); err != nil {
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
