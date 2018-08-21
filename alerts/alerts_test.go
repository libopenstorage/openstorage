package alerts

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

// helper function go get a new kvdb instance
func newInMemKvdb() (kvdb.Kvdb, error) {
	// create in memory kvdb
	if kv, err := kvdb.New(mem.Name, "", []string{}, nil, nil); err != nil {
		return nil, err
	} else {
		return kv, nil
	}
}

// raiseAlerts is a helper func that raises some alerts for test purposes.
func raiseAlerts(manager Manager) error {
	var alert *api.Alert
	alert = new(api.Alert)
	alert.AlertType = 10
	alert.Resource = api.ResourceType_RESOURCE_TYPE_VOLUME
	alert.ResourceId = "inca"
	alert.Timestamp = &timestamp.Timestamp{Seconds: time.Now().AddDate(0, -1, 0).Unix()}
	if err := manager.Raise(alert); err != nil {
		return err
	}

	alert = new(api.Alert)
	alert.AlertType = 12
	alert.Resource = api.ResourceType_RESOURCE_TYPE_CLUSTER
	alert.ResourceId = "aztec"
	alert.Timestamp = &timestamp.Timestamp{Seconds: time.Now().AddDate(0, -2, 0).Unix()}
	if err := manager.Raise(alert); err != nil {
		return err
	}

	alert = new(api.Alert)
	alert.AlertType = 10
	alert.Resource = api.ResourceType_RESOURCE_TYPE_DRIVE
	alert.ResourceId = "maya"
	alert.Timestamp = &timestamp.Timestamp{Seconds: time.Now().AddDate(0, -3, 0).Unix()}
	if err := manager.Raise(alert); err != nil {
		return err
	}

	alert = new(api.Alert)
	alert.AlertType = 10
	alert.Resource = api.ResourceType_RESOURCE_TYPE_DRIVE
	alert.ResourceId = "inca"
	alert.Timestamp = &timestamp.Timestamp{Seconds: time.Now().AddDate(0, -1, 0).Unix()}
	if err := manager.Raise(alert); err != nil {
		return err
	}

	alert = new(api.Alert)
	alert.AlertType = 14
	alert.Resource = api.ResourceType_RESOURCE_TYPE_DRIVE
	alert.ResourceId = "aztec"
	alert.Timestamp = &timestamp.Timestamp{Seconds: time.Now().AddDate(0, -4, 0).Unix()}
	if err := manager.Raise(alert); err != nil {
		return err
	}

	alert = new(api.Alert)
	alert.AlertType = 12
	alert.Resource = api.ResourceType_RESOURCE_TYPE_DRIVE
	alert.ResourceId = "maya"
	alert.Timestamp = &timestamp.Timestamp{Seconds: time.Now().AddDate(0, -5, 0).Unix()}
	if err := manager.Raise(alert); err != nil {
		return err
	}

	return nil
}

// TestManager_Enumerate tests enumeration based on various filters.
func TestManager_Enumerate(t *testing.T) {
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	manager := NewManager(kv)

	if err := raiseAlerts(manager); err != nil {
		t.Fatal(err)
	}

	// prepare a test configuration table
	configs := []struct {
		name          string
		filters       []Filter
		expectedCount int
	}{
		{
			name:          "by none",
			expectedCount: 6,
		},
		{
			name: "by 1 resource type",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
			},
			expectedCount: 1,
		},
		{
			name: "by 2 resource types",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_CLUSTER),
			},
			expectedCount: 2,
		},
		{
			name: "by 2 resource types",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_CLUSTER),
			},
			expectedCount: 5,
		},
		{
			name: "by 1 resource id",
			filters: []Filter{
				NewResourceIDFilter("inca", api.ResourceType_RESOURCE_TYPE_VOLUME),
			},
			expectedCount: 1,
		},
		{
			name: "by 2 resource ids",
			filters: []Filter{
				NewResourceIDFilter("inca", api.ResourceType_RESOURCE_TYPE_VOLUME),
				NewResourceIDFilter("maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 3,
		},
		{
			name: "by 2 different filter types",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
				NewResourceIDFilter("maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 3,
		},
		{
			name: "two levels of filter",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewResourceIDFilter("maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 4,
		},
		{
			name: "skip a level",
			filters: []Filter{
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 4,
		},
		{
			name: "alert type",
			filters: []Filter{
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 1,
		},
		{
			name: "alert types",
			filters: []Filter{
				NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 2,
		},
		{
			name: "time and other filters",
			filters: []Filter{
				NewTimeFilter(time.Now().AddDate(0, 0, -1), time.Now()),
				NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 2,
		},
		{
			name: "time filter match none",
			filters: []Filter{
				NewTimeFilter(time.Now().AddDate(0, 0, -1), time.Now()),
			},
			expectedCount: 0,
		},
		{
			name: "time filter match all in last 3 months",
			filters: []Filter{
				NewTimeFilter(time.Now().AddDate(0, -3, 0), time.Now()),
			},
			expectedCount: 4,
		},
	}

	// iterate over all configs and test
	for _, config := range configs {
		myAlerts, err := manager.Enumerate(config.filters...)
		if err != nil {
			t.Fatal(err)
		}

		if len(myAlerts) != config.expectedCount {
			t.Fatal("test:", config.name, ", alert count: expected:", config.expectedCount, ", found:", len(myAlerts))
		}
	}
}

// TestManager_Filter tests alert filtering based on various filters.
func TestManager_Filter(t *testing.T) {
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	manager := NewManager(kv)

	if err := raiseAlerts(manager); err != nil {
		t.Fatal(err)
	}

	// prepare a test configuration table
	configs := []struct {
		name          string
		filters       []Filter
		expectedCount int
	}{
		{
			name:          "by none",
			expectedCount: 6,
		},
		{
			name: "by 1 resource type",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
			},
			expectedCount: 1,
		},
		{
			name: "by 2 resource types",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_CLUSTER),
			},
			expectedCount: 0,
		},
		{
			name: "by 2 resource types",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_CLUSTER),
			},
			expectedCount: 0,
		},
		{
			name: "by 1 resource id",
			filters: []Filter{
				NewResourceIDFilter("inca", api.ResourceType_RESOURCE_TYPE_VOLUME),
			},
			expectedCount: 1,
		},
		{
			name: "by 2 resource ids",
			filters: []Filter{
				NewResourceIDFilter("inca", api.ResourceType_RESOURCE_TYPE_VOLUME),
				NewResourceIDFilter("maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 0,
		},
		{
			name: "by 2 different filter types",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
				NewResourceIDFilter("maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 0,
		},
		{
			name: "two levels of filter",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewResourceIDFilter("maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 2,
		},
		{
			name: "skip a level",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 1,
		},
		{
			name: "alert type",
			filters: []Filter{
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 1,
		},
		{
			name: "alert types",
			filters: []Filter{
				NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 0,
		},
		{
			name: "time and other filters",
			filters: []Filter{
				NewTimeFilter(time.Now().AddDate(0, 0, -1), time.Now()),
				NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 0,
		},
		{
			name: "time filter match none",
			filters: []Filter{
				NewTimeFilter(time.Now().AddDate(0, 0, -1), time.Now()),
			},
			expectedCount: 0,
		},
		{
			name: "time filter match all in last 3 months",
			filters: []Filter{
				NewTimeFilter(time.Now().AddDate(0, -3, 0), time.Now()),
			},
			expectedCount: 4,
		},
		{
			name: "time filter match all in last 3 months for drive type",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewTimeFilter(time.Now().AddDate(0, -3, 0), time.Now()),
			},
			expectedCount: 2,
		},
	}

	// iterate over all configs and test
	for _, config := range configs {
		myAlerts, err := manager.Enumerate()
		if err != nil {
			t.Fatal(err)
		}

		myAlerts, err = manager.Filter(myAlerts, config.filters...)
		if err != nil {
			t.Fatal(err)
		}

		if len(myAlerts) != config.expectedCount {
			t.Fatal("test:", config.name, ", alert count: expected:", config.expectedCount, ", found:", len(myAlerts))
		}
	}
}

// TestManager_Delete tests if delete works as governed by filters.
func TestManager_Delete(t *testing.T) {
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	manager := NewManager(kv)

	// prepare a test configuration table
	configs := []struct {
		name          string
		filters       []Filter
		expectedCount int
	}{
		{
			name:          "by none",
			expectedCount: 0,
		},
		{
			name: "by 1 resource type",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
			},
			expectedCount: 5,
		},
		{
			name: "by 2 resource types",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_CLUSTER),
			},
			expectedCount: 4,
		},
		{
			name: "by 2 resource types",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_CLUSTER),
			},
			expectedCount: 1,
		},
		{
			name: "by 1 resource id",
			filters: []Filter{
				NewResourceIDFilter("inca", api.ResourceType_RESOURCE_TYPE_VOLUME),
			},
			expectedCount: 5,
		},
		{
			name: "by 2 resource ids",
			filters: []Filter{
				NewResourceIDFilter("inca", api.ResourceType_RESOURCE_TYPE_VOLUME),
				NewResourceIDFilter("maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 3,
		},
		{
			name: "by 2 different filter types",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
				NewResourceIDFilter("maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 3,
		},
		{
			name: "two levels of filter",
			filters: []Filter{
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewResourceIDFilter("maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 2,
		},
		{
			name: "skip a level",
			filters: []Filter{
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 2,
		},
		{
			name: "alert type",
			filters: []Filter{
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 5,
		},
		{
			name: "alert types",
			filters: []Filter{
				NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 4,
		},
		{
			name: "time filter",
			filters: []Filter{
				NewTimeFilter(time.Now().AddDate(0, 0, -1), time.Now()),
				NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
				NewAlertTypeFilter(12, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
			},
			expectedCount: 4,
		},
	}

	// iterate over all configs and test
	for _, config := range configs {
		if err := raiseAlerts(manager); err != nil {
			t.Fatal(err)
		}

		if err := manager.Delete(config.filters...); err != nil {
			t.Fatal(err)
		}

		myAlerts, err := manager.Enumerate()
		if err != nil {
			t.Fatal(err)
		}

		if len(myAlerts) != config.expectedCount {
			t.Fatal("test:", config.name, ", alert count: expected:", config.expectedCount, ", found:", len(myAlerts))
		}
	}
}

// TestManager_SetRules_Raise tests if set rules activate on raise.
func TestManager_SetRules_Raise(t *testing.T) {
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	manager := NewManager(kv)

	configs := []struct {
		name          string
		rules         []Rule
		myAlert       *api.Alert
		expectedCount int
	}{
		{
			name: "deleteOnRaise",
			rules: []Rule{
				// define a rule that will delete every alert of RESOURCE_TYPE_VOLUME
				// and RESOURCE_TYPE_CLUSTER before
				// raising an alert that matches NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE)
				NewRaiseRule("deleteOnRaise0",
					NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
					NewDeleteAction(
						NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
						NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_CLUSTER),
					)),
			},
			myAlert: &api.Alert{
				AlertType:  10,
				ResourceId: "maya",
				Resource:   api.ResourceType_RESOURCE_TYPE_DRIVE,
			},
			expectedCount: 4,
		},
		{
			name: "deleteAllOnRaise",
			rules: []Rule{
				// define a rule that will delete every alert before
				// raising an alert that matches NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE)
				NewRaiseRule("deleteOnRaise0",
					NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
					NewDeleteAction(),
				),
			},
			myAlert: &api.Alert{
				AlertType:  10,
				ResourceId: "maya",
				Resource:   api.ResourceType_RESOURCE_TYPE_DRIVE,
			},
			expectedCount: 1,
		},
		{
			name: "deleteAllOnRaiseMismatch",
			rules: []Rule{
				// define a rule that will delete every alert of RESOURCE_TYPE_VOLUME
				// and RESOURCE_TYPE_CLUSTER before
				// raising an alert that matches NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE)
				NewRaiseRule("deleteOnRaise0",
					NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
					NewDeleteAction(
						NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_VOLUME),
						NewResourceTypeFilter(api.ResourceType_RESOURCE_TYPE_CLUSTER),
					)),
			},
			myAlert: &api.Alert{
				AlertType:  17, // <<< this would cause mismatch with the filter, so none should be deleted
				ResourceId: "maya",
				Resource:   api.ResourceType_RESOURCE_TYPE_DRIVE,
			},
			expectedCount: 7,
		},
		{
			name: "deleteOldOnRaise",
			rules: []Rule{
				// define a rule that will delete every alert of RESOURCE_TYPE_VOLUME
				// and RESOURCE_TYPE_CLUSTER before
				// raising an alert that matches NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE)
				NewRaiseRule("deleteOnRaise0",
					NewAlertTypeFilter(10, "maya", api.ResourceType_RESOURCE_TYPE_DRIVE),
					NewDeleteAction(
						NewTimeFilter(time.Now().AddDate(0, -3, 0), time.Now()),
					)),
			},
			myAlert: &api.Alert{
				AlertType:  10,
				ResourceId: "maya",
				Resource:   api.ResourceType_RESOURCE_TYPE_DRIVE,
			},
			expectedCount: 3,
		},
	}

	for _, config := range configs {
		// first ensure there are no rules
		manager.DeleteRules(config.rules...)

		// then raise some alerts
		if err := raiseAlerts(manager); err != nil {
			t.Fatal(err)
		}

		// then set rules
		manager.SetRules(config.rules...)

		// then raise an alert
		if err := manager.Raise(config.myAlert); err != nil {
			t.Fatal(err)
		}

		myAlerts, err := manager.Enumerate()
		if err != nil {
			t.Fatal(err)
		}

		if len(myAlerts) != config.expectedCount {
			t.Fatal(config.name, "expected", config.expectedCount, "found", len(myAlerts))
		}

	}
}
