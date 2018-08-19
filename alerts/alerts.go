package alerts

import (
	"path/filepath"
	"strconv"

	"encoding/json"

	"sync"

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
}

func newManager(kv kvdb.Kvdb) *manager {
	return &manager{kv: kv}
}

type manager struct {
	kv    kvdb.Kvdb
	rules []Rule
	sync.Mutex
}

func (m *manager) Raise(alert *api.Alert) error {
	m.Lock()
	defer m.Unlock()

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
			myAlerts = append(myAlerts, alert)
		}
	}

	return myAlerts, nil
}

func (m *manager) Delete(filters ...Filter) error {
	m.Lock()
	defer m.Unlock()

	for _, rule := range m.rules {
		if rule.GetEvent() == DeleteEvent {
			if err := rule.GetAction().Run(m); err != nil {
				return err
			}
		}
	}

	keys, err := getKeysFromFilters(filters...)
	if err != nil {
		return err
	}

	for key := range keys {
		if err := m.kv.DeleteTree(key); err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) SetRules(rules ...Rule) {
	m.Lock()
	defer m.Unlock()
	m.rules = append(m.rules, rules...)
}
