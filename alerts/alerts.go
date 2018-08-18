package alerts

import (
	"path/filepath"
	"sort"
	"strconv"

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

type manager struct {
	kv    kvdb.Kvdb
	rules []Rule
}

func NewManager() Manager {
	m := new(manager)
	m.kv = kvdb.Instance()
	return m
}

func (m *manager) Raise(alert *api.Alert) error {
	for _, rule := range m.rules {
		if rule.GetEvent() == RaiseEvent {
			match, err := rule.GetFilter().Match(alert)
			if err != nil {
				return err
			}
			if match {
				rule.GetAction().Run(m)
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

func (m *manager) Enumerate(filters ...Filter) ([]*api.Alert, error) {
	key := KvdbKey

	// sort filters so we know how to query
	if len(filters) > 0 {
		sort.Sort(Filters(filters))

		// determine how to query kvdb tree based on first filter
		filter := filters[0]
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
	}

	kvp, err := m.kv.Get(key)
	if err != nil {
		return nil, err
	}

	_ = kvp

	return nil, nil
}

func (m *manager) Delete(filters ...Filter) error {
	_, err := m.kv.Delete(KvdbKey)
	if err != nil {
		return err
	}
	return nil
}

func (m *manager) SetRules(rules ...Rule) {
	m.rules = append(m.rules, rules...)
}
