package alerts

import (
	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/stretchr/testify/require"
	"go.pedge.io/proto/time"
	"strconv"
	"testing"
	"time"
)

var (
	kva             AlertsClient
	nextId          int64
	isWatcherCalled int
	watcherAction   AlertAction
	watcherAlert    api.Alerts
	watcherPrefix   string
	watcherKey      string
)

const (
	kvdbDomain = "openstorage"
	clusterName = "1"
)

func TestSetup(t *testing.T) {
	kv := kvdb.Instance()
	if kv == nil {
		kv, err := kvdb.New(mem.Name, kvdbDomain+"/"+clusterName, []string{}, nil)
		if err != nil {
			t.Fatalf("Failed to set default KV store : (%v): %v", mem.Name, err)
		}
		err = kvdb.SetInstance(kv)
		if err != nil {
			t.Fatalf("Failed to set default KV store: (%v): %v", mem.Name, err)
		}
	}

	var err error
	kva, err = New("alerts_kvdb", mem.Name, kvdbDomain, []string{}, clusterName)
	if err != nil {
		t.Fatalf("Failed to create new Kvapi.Alerts object")
	}
}

func TestRaiseAndErase(t *testing.T) {
	// Raise api.Alerts Id : 1

	raiseAlerts, err := kva.Raise(api.Alerts{Resource: api.ResourceType_VOLUMES, Severity: api.SeverityType_NOTIFY, Message: "Test Message"})
	require.NoError(t, err, "Failed in raising an alert")

	kv := kva.GetKvdbInstance()
	var alert api.Alerts

	_, err = kv.GetVal(getResourceKey(api.ResourceType_VOLUMES)+strconv.FormatInt(raiseAlerts.Id, 10), &alert)
	require.NoError(t, err, "Failed to retrieve alert from kvdb")
	require.NotNil(t, alert, "api.Alerts object null in kvdb")
	require.Equal(t, raiseAlerts.Id, alert.Id, "api.Alerts Id mismatch")
	require.Equal(t, api.ResourceType_VOLUMES, alert.Resource, "api.Alerts Resource mismatch")
	require.Equal(t, api.SeverityType_NOTIFY, alert.Severity, "api.Alerts Severity mismatch")

	// Raise api.Alerts with no Resource
	_, err = kva.Raise(api.Alerts{Severity: api.SeverityType_NOTIFY})
	require.Error(t, err, "An error was expected")
	require.Equal(t, ErrResourceNotFound, err, "Error mismatch")

	// Erase api.Alerts Id : 1
	err = kva.Erase(api.ResourceType_VOLUMES, raiseAlerts.Id)
	require.NoError(t, err, "Failed to erase an alert")

	_, err = kv.GetVal(getResourceKey(api.ResourceType_VOLUMES)+"1", &alert)
	require.Error(t, err, "api.Alerts not erased from kvdb")

}

func TestRetrieve(t *testing.T) {
	var alert api.Alerts

	// Raise a ResourceType_NODE specific api.Alerts
	raiseAlerts, err := kva.Raise(api.Alerts{Resource: api.ResourceType_NODE, Severity: api.SeverityType_ALARM})

	alert, err = kva.Retrieve(api.ResourceType_NODE, raiseAlerts.Id)
	require.NoError(t, err, "Failed to retrieve alert")
	require.NotNil(t, alert, "api.Alerts object null")
	require.Equal(t, raiseAlerts.Id, alert.Id, "api.Alerts Id mismatch")
	require.Equal(t, api.ResourceType_NODE, alert.Resource, "api.Alerts ResourceType_NODE Id mismatch")
	require.Equal(t, api.SeverityType_ALARM, alert.Severity, "api.Alerts severity mismatch")

	// Retrieve non existing alert
	alert, err = kva.Retrieve(api.ResourceType_VOLUMES, 5)
	require.Error(t, err, "Expected an error")

	// Cleanup
	err = kva.Erase(api.ResourceType_NODE, raiseAlerts.Id)
}

func TestClear(t *testing.T) {
	// Raise an alert
	var alert api.Alerts
	kv := kva.GetKvdbInstance()
	raiseAlerts, err := kva.Raise(api.Alerts{Resource: api.ResourceType_NODE, Severity: api.SeverityType_ALARM})

	err = kva.Clear(api.ResourceType_NODE, raiseAlerts.Id)
	require.NoError(t, err, "Failed to clear alert")

	_, err = kv.GetVal(getResourceKey(api.ResourceType_NODE)+strconv.FormatInt(raiseAlerts.Id, 10), &alert)
	require.Equal(t, true, alert.Cleared, "Failed to clear alert")

	err = kva.Erase(api.ResourceType_NODE, raiseAlerts.Id)
}

func TestEnumerateAlerts(t *testing.T) {
	// Raise a few alerts
	raiseAlert1, err := kva.Raise(api.Alerts{Resource: api.ResourceType_VOLUMES, Severity: api.SeverityType_NOTIFY})
	raiseAlert2, err := kva.Raise(api.Alerts{Resource: api.ResourceType_VOLUMES, Severity: api.SeverityType_NOTIFY})
	raiseAlert3, err := kva.Raise(api.Alerts{Resource: api.ResourceType_VOLUMES, Severity: api.SeverityType_WARNING})
	raiseAlert4, err := kva.Raise(api.Alerts{Resource: api.ResourceType_NODE, Severity: api.SeverityType_WARNING})

	enAlerts, err := kva.Enumerate(api.Alerts{Resource: api.ResourceType_VOLUMES})
	require.NoError(t, err, "Failed to enumerate alerts")
	require.Equal(t, 3, len(enAlerts), "Enumerated incorrect number of alerts")

	enAlerts, err = kva.Enumerate(api.Alerts{Resource: api.ResourceType_VOLUMES, Severity: api.SeverityType_WARNING})
	require.NoError(t, err, "Failed to enumerate alerts")
	require.Equal(t, 1, len(enAlerts), "Enumerated incorrect number of alerts")
	require.Equal(t, api.SeverityType_WARNING, enAlerts[0].Severity, "Severity mismatch")

	enAlerts, err = kva.Enumerate(api.Alerts{})
	require.NoError(t, err, "Failed to enumerate alerts")
	require.Equal(t, 4, len(enAlerts), "Enumerated incorrect number of alerts")

	enAlerts, err = kva.Enumerate(api.Alerts{Severity: api.SeverityType_WARNING})
	require.NoError(t, err, "Failed to enumerate alerts")
	require.Equal(t, 2, len(enAlerts), "Enumerated incorrect number of alerts")
	require.Equal(t, api.SeverityType_WARNING, enAlerts[0].Severity, "Severity mismatch")

	// Add a dummy event into kvdb two hours ago
	kv := kva.GetKvdbInstance()
	currentTime := time.Now()
	delayedTime := currentTime.Add(-1 * time.Duration(2) * time.Hour)

	alert := api.Alerts{Timestamp: prototime.TimeToTimestamp(delayedTime), Id: 10, Resource: api.ResourceType_VOLUMES}

	_, err = kv.Put(getResourceKey(api.ResourceType_VOLUMES)+strconv.FormatInt(10, 10), &alert, 0)
	enAlerts, err = kva.EnumerateWithinTimeRange(currentTime.Add(-1*time.Duration(10)*time.Second), currentTime, api.ResourceType_VOLUMES)
	require.NoError(t, err, "Failed to enumerate results")
	require.Equal(t, 3, len(enAlerts), "Enumerated incorrect number of alerts")

	err = kva.Erase(api.ResourceType_VOLUMES, raiseAlert1.Id)
	err = kva.Erase(api.ResourceType_VOLUMES, raiseAlert2.Id)
	err = kva.Erase(api.ResourceType_VOLUMES, raiseAlert3.Id)
	err = kva.Erase(api.ResourceType_NODE, raiseAlert4.Id)
}

func testAlertsWatcher(alert *api.Alerts, action AlertAction, prefix string, key string) error {
	// A dummy callback function
	// Setting the global variables so that we can check them in our unit tests
	isWatcherCalled = 1
	if action != AlertDeleteAction {
		watcherAlert = *alert
	} else {
		watcherAlert = api.Alerts{}
	}
	watcherAction = action
	watcherPrefix = prefix
	watcherKey = key
	return nil

}

func TestWatch(t *testing.T) {
	isWatcherCalled = 0

	err := kva.Watch(clusterName, testAlertsWatcher)
	require.NoError(t, err, "Failed to subscribe a watch function")

	raiseAlert1, err := kva.Raise(api.Alerts{Resource: api.ResourceType_CLUSTER, Severity: api.SeverityType_NOTIFY})

	// Sleep for sometime so that we pass on some previous watch callbacks
	time.Sleep(time.Millisecond * 100)

	require.Equal(t, 1, isWatcherCalled, "Callback function not called")
	require.Equal(t, AlertCreateAction, watcherAction, "action mismatch for create")
	require.Equal(t, raiseAlert1.Id, watcherAlert.Id, "alert id mismatch")
	require.Equal(t, "alerts/cluster/"+strconv.FormatInt(raiseAlert1.Id, 10), watcherKey, "key mismatch")

	err = kva.Clear(api.ResourceType_CLUSTER, raiseAlert1.Id)

	// Sleep for sometime so that we pass on some previous watch callbacks
	time.Sleep(time.Millisecond * 100)

	require.Equal(t, AlertUpdateAction, watcherAction, "action mismatch for update")
	require.Equal(t, "alerts/cluster/"+strconv.FormatInt(raiseAlert1.Id, 10), watcherKey, "key mismatch")

	err = kva.Erase(api.ResourceType_CLUSTER, raiseAlert1.Id)

	// Sleep for sometime so that we pass on some previous watch callbacks
	time.Sleep(time.Millisecond * 100)

	require.Equal(t, AlertDeleteAction, watcherAction, "action mismatch for delete")
	require.Equal(t, "alerts/cluster/"+strconv.FormatInt(raiseAlert1.Id, 10), watcherKey, "key mismatch")

	// Watch on a new clusterID
	newClusterId := "2"
	isWatcherCalled = 0
	err = kva.Watch(newClusterId, testAlertsWatcher)

	// Create a new alerts instance for raising an alert in this new cluster id
	kvaNew, err := New("alerts_kvdb_test", mem.Name, kvdbDomain, []string{}, newClusterId)
	if err != nil {
		t.Fatalf("Failed to create new Kvapi.Alerts object %s", err.Error())
	}

	raiseAlertNew, err := kvaNew.Raise(api.Alerts{Resource: api.ResourceType_NODE, Severity: api.SeverityType_ALARM})
	// Sleep for sometime so that we pass on some previous watch callbacks
	time.Sleep(time.Millisecond * 100)

	require.Equal(t, 1, isWatcherCalled, "Callback function not called")
	require.Equal(t, AlertCreateAction, watcherAction, "action mismatch for create")
	require.Equal(t, raiseAlertNew.Id, watcherAlert.Id, "alert id mismatch")
	require.Equal(t, "alerts/node/"+strconv.FormatInt(raiseAlertNew.Id, 10), watcherKey, "key mismatch")
	
}


func mockGenerateId(clusterId string) (int64, error) {
	nextId = nextId + 1
	return nextId, nil
}
