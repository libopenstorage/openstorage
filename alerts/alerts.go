package alerts

import (
	"path/filepath"
	"sort"
	"strconv"

	"crypto/md5"
	"encoding/json"

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
	_, err := m.kv.Put(key, alert, 0)
	return err
}

func (m *manager) Enumerate(filters ...Filter) ([]*api.Alert, error) {
	myAlerts := make([]*api.Alert, 0, 0)
	alertsMap := make(map[string]*api.Alert)
	var keys []string

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

			keys = append(keys, key)
		}
	} else {
		keys = []string{KvdbKey}
	}

	for _, key := range keys {
		kvps, err := m.kv.Enumerate(key)
		if err != nil {
			return nil, err
		}

		for _, kvp := range kvps {
			alert := new(api.Alert)
			if err := json.Unmarshal(kvp.Value, alert); err != nil {
				return nil, err
			}
			md5sum := md5.Sum(kvp.Value)
			alertsMap[string(md5sum[:])] = alert
		}
	}

	for _, alert := range alertsMap {
		alert := alert
		myAlerts = append(myAlerts, alert)
	}

	return myAlerts, nil
}

func (m *manager) Delete(filters ...Filter) error {
	/*for _, rule := range m.rules {
		if rule.GetEvent() == DeleteEvent {
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
	}*/

	_, err := m.kv.Delete(KvdbKey)
	if err != nil {
		return err
	}
	return nil
}

func (m *manager) SetRules(rules ...Rule) {
	m.rules = append(m.rules, rules...)
}
