package alerts

import (
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

var (
	kva    KvAlerts = KvAlerts{}
	nextId int64    = 0
)

func TestSetup(t *testing.T) {
	kv := kvdb.Instance()
	if kv == nil {
		kv, err := kvdb.New(mem.Name, "openstorage/1", []string{}, nil)
		if err != nil {
			t.Fatalf("Failed to set default KV store : (%v): %v", mem.Name, err)
		}
		err = kvdb.SetInstance(kv)
		if err != nil {
			t.Fatalf("Failed to set default KV store: (%v): %v", mem.Name, err)
		}
	}
}

func mockGenerateId() (int64, error) {
	nextId = nextId + 1
	return nextId, nil
}

func TestRaiseWithGenerateIdAndErase(t *testing.T) {
	// RaiseWithGenerateId Alert Id : 1

	raiseAlert, err := kva.RaiseWithGenerateId(Alert{Resource: Volumes, Severity: AlertsNotify}, mockGenerateId)
	require.NoError(t, err, "Failed in raising an alert")

	kv := kvdb.Instance()
	var alert Alert

	_, err = kv.GetVal(getResourceKey(Volumes)+strconv.FormatInt(raiseAlert.Id, 10), &alert)
	require.NoError(t, err, "Failed to retrieve alert from kvdb")
	require.NotNil(t, alert, "Alert object null in kvdb")
	require.Equal(t, raiseAlert.Id, alert.Id, "Alert Id mismatch")
	require.Equal(t, Volumes, alert.Resource, "Alert Resource mismatch")
	require.Equal(t, AlertsNotify, alert.Severity, "Alert Severity mismatch")

	// RaiseWithGenerateId Alert with no Resource
	_, err = kva.RaiseWithGenerateId(Alert{Severity: 3}, mockGenerateId)
	require.Error(t, err, "An error was expected")
	require.Equal(t, ErrResourceNotFound, err, "Error mismatch")

	// Erase Alert Id : 1
	err = kva.Erase(Volumes, raiseAlert.Id)
	require.NoError(t, err, "Failed to erase an alert")

	_, err = kv.GetVal(getResourceKey(Volumes)+"1", &alert)
	require.Error(t, err, "Alert not erased from kvdb")

}

func TestRetrieve(t *testing.T) {
	var alert Alert

	// RaiseWithGenerateId a Node specific Alert
	raiseAlert, err := kva.RaiseWithGenerateId(Alert{Resource: Node, Severity: AlertsAlarm}, mockGenerateId)

	alert, err = kva.Retrieve(Node, raiseAlert.Id)
	require.NoError(t, err, "Failed to retrieve alert")
	require.NotNil(t, alert, "Alert object null")
	require.Equal(t, raiseAlert.Id, alert.Id, "Alert Id mismatch")
	require.Equal(t, Node, alert.Resource, "Alert Node Id mismatch")
	require.Equal(t, AlertsAlarm, alert.Severity, "Alert severity mismatch")

	// Retrieve non existing alert
	alert, err = kva.Retrieve(Volumes, 5)
	require.Error(t, err, "Expected an error")

	// Cleanup
	err = kva.Erase(Node, raiseAlert.Id)
}

func TestClear(t *testing.T) {
	// RaiseWithGenerateId an alert
	var alert Alert
	kv := kvdb.Instance()
	raiseAlert, err := kva.RaiseWithGenerateId(Alert{Resource: Node, Severity: AlertsAlarm}, mockGenerateId)

	err = kva.Clear(Node, raiseAlert.Id)
	require.NoError(t, err, "Failed to clear alert")

	_, err = kv.GetVal(getResourceKey(Node)+strconv.FormatInt(raiseAlert.Id, 10), &alert)
	require.Equal(t, true, alert.Cleared, "Failed to clear alert")

	err = kva.Erase(Node, raiseAlert.Id)
}

func TestEnumerateAlerts(t *testing.T) {
	// RaiseWithGenerateId a few alerts
	_, err := kva.RaiseWithGenerateId(Alert{Resource: Volumes, Severity: AlertsNotify}, mockGenerateId)
	_, err = kva.RaiseWithGenerateId(Alert{Resource: Volumes, Severity: AlertsNotify}, mockGenerateId)
	_, err = kva.RaiseWithGenerateId(Alert{Resource: Volumes, Severity: AlertsWarning}, mockGenerateId)
	_, err = kva.RaiseWithGenerateId(Alert{Resource: Node, Severity: AlertsWarning}, mockGenerateId)

	enAlerts, err := kva.Enumerate(Alert{Resource: Volumes})
	require.NoError(t, err, "Failed to enumerate alerts")
	require.Equal(t, 3, len(enAlerts), "Enumerated incorrect number of alerts")

	enAlerts, err = kva.Enumerate(Alert{Resource: Volumes, Severity: AlertsWarning})
	require.NoError(t, err, "Failed to enumerate alerts")
	require.Equal(t, 1, len(enAlerts), "Enumerated incorrect number of alerts")
	require.Equal(t, AlertsWarning, enAlerts[0].Severity, "Severity mismatch")

	enAlerts, err = kva.Enumerate(Alert{})
	require.NoError(t, err, "Failed to enumerate alerts")
	require.Equal(t, 4, len(enAlerts), "Enumerated incorrect number of alerts")

	enAlerts, err = kva.Enumerate(Alert{Severity: AlertsWarning})
	require.NoError(t, err, "Failed to enumerate alerts")
	require.Equal(t, 2, len(enAlerts), "Enumerated incorrect number of alerts")
	require.Equal(t, AlertsWarning, enAlerts[0].Severity, "Severity mismatch")

	// Add a dummy event into kvdb two hours ago
	kv := kvdb.Instance()
	nowT := time.Now()

	alert := Alert{Timestamp: nowT.Add(-1 * time.Duration(2) * time.Hour), Id: 10, Resource: Volumes}

	_, err = kv.Put(getResourceKey(Volumes)+strconv.FormatInt(10, 10), &alert, 0)
	enAlerts, err = kva.EnumerateWithinTimeRange(nowT.Add(-1*time.Duration(10)*time.Second), nowT, Volumes)
	require.NoError(t, err, "Failed to enumerate results")
	require.Equal(t, 3, len(enAlerts), "Enumerated incorrect number of alerts")
}
