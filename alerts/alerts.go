package alerts

import (
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
	KvdbKey = "alerts"
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

type manager struct {
	kv    kvdb.Kvdb
	key   string
	rules []Rule
}

func NewManager() Manager {
	m := new(manager)
	m.kv = kvdb.Instance()
	m.key = KvdbKey
	return m
}

func (m *manager) Raise(alert *api.Alert) error {
	for _, rule := range m.rules {
		if rule.GetEvent() == RaiseEvent {
			if rule.GetFilter().Match(alert) {
				rule.GetAction().Run(m)
			}
		}
	}
	_, err := m.kv.Put(m.key, alert, 0)
	return err
}

func (m *manager) Enumerate(filters ...Filter) ([]*api.Alert, error) {
	_, err := m.kv.Get(m.key)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (m *manager) Delete(filters ...Filter) error {
	_, err := m.kv.Delete(m.key)
	if err != nil {
		return err
	}
	return nil
}

func (m *manager) SetRules(rules ...Rule) {
	m.rules = append(m.rules, rules...)
}
