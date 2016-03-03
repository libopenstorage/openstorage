package alerts

import (
	"encoding/json"
	"fmt"
	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
	"go.pedge.io/proto/time"
	"strconv"
	"strings"
	"time"
)

type watcherStatus int
type watcher struct {
	kvcb   kvdb.WatchCB
	status watcherStatus
	cb     AlertsWatcherFunc
}
type KvAlerts struct {
}

const (
	alertsKey       = "alerts/"
	nextAlertsIdKey = "nextAlertsId"
	clusterKey      = "cluster/"
	volumeKey       = "volume/"
	nodeKey         = "node/"
	bootstrap       = "bootstrap"
	// Name of this alerts client implementation
	Name            = "alerts_kvdb"
)

const (
	watchBootstrap = watcherStatus(iota)
	watchReady
	watchError
)

var (
	alertsWatcher    watcher
	alertsWatchIndex uint64
	watchErrors      int
)


// Init initializes a AlertsClient interface implementation
func Init() (AlertsClient, error) {
	return &KvAlerts{}, nil
}

// Raise raises an Alert
func (kva *KvAlerts) Raise(a api.Alerts) (api.Alerts, error) {
	return raiseAnAlerts(a, getNextIdFromKVDB)
}

// RaiseWithGenerateId raises an Alerts with a custom generateId function for alertIds (currently used only for unit testing)
func (kva *KvAlerts) RaiseWithGenerateId(a api.Alerts, generateId func() (int64, error)) (api.Alerts, error) {
	return raiseAnAlerts(a, generateId)
}


// Erase erases an alert
func (kva *KvAlerts) Erase(resourceType api.ResourceType, alertId int64) error {
	kv := kvdb.Instance()

	if resourceType == api.ResourceType_UNKNOWN_RESOURCE {
		return ErrResourceNotFound
	}
	var err error

	_, err = kv.Delete(getResourceKey(resourceType) + strconv.FormatInt(alertId, 10))
	return err
}

// Clear clears an alert
func (kva *KvAlerts) Clear(resourceType api.ResourceType, alertId int64) error {
	kv := kvdb.Instance()
	var (
		err   error
		alert api.Alerts
	)
	if resourceType == api.ResourceType_UNKNOWN_RESOURCE {
		return ErrResourceNotFound
	}

	_, err = kv.GetVal(getResourceKey(resourceType)+strconv.FormatInt(alertId, 10), &alert)
	if err != nil {
		return err
	}
	alert.Cleared = true

	_, err = kv.Update(getResourceKey(resourceType)+strconv.FormatInt(alertId, 10), &alert, 0)
	return err
}

// Retrieve retrieves a specific alert
func (kva *KvAlerts) Retrieve(resourceType api.ResourceType, alertId int64) (api.Alerts, error) {
	var (
		alert api.Alerts
		err   error
	)
	if resourceType == api.ResourceType_UNKNOWN_RESOURCE {
		return api.Alerts{}, ErrResourceNotFound
	}
	kv := kvdb.Instance()

	_, err = kv.GetVal(getResourceKey(resourceType)+strconv.FormatInt(alertId, 10), &alert)

	return alert, err
}

// Enumerate enumerates alerts
func (kva *KvAlerts) Enumerate(filter api.Alerts) ([]*api.Alerts, error) {
	allAlerts := []*api.Alerts{}
	resourceAlerts := []*api.Alerts{}
	var err error

	if filter.Resource != api.ResourceType_UNKNOWN_RESOURCE {
		resourceAlerts, err = getResourceSpecificAlerts(filter.Resource)
		if err != nil {
			return nil, err
		}
	} else {
		resourceAlerts, err = getAllAlerts()
	}

	if filter.Severity != 0 {
		for _, v := range resourceAlerts {
			if v.Severity <= filter.Severity {
				allAlerts = append(allAlerts, v)
			}
		}
	} else {
		allAlerts = append(allAlerts, resourceAlerts...)
	}

	return allAlerts, err
}

// EnumerateWithinTimeRange enumerates alerts between timeStart and timeEnd
func (kva *KvAlerts) EnumerateWithinTimeRange(
	timeStart time.Time,
	timeEnd time.Time,
	resourceType api.ResourceType,
) ([]*api.Alerts, error) {
	allAlerts := []*api.Alerts{}
	resourceAlerts := []*api.Alerts{}
	var err error

	if resourceType != 0 {
		resourceAlerts, err = getResourceSpecificAlerts(resourceType)
		if err != nil {
			return nil, err
		}
	} else {
		resourceAlerts, err = getAllAlerts()
		if err != nil {
			return nil, err
		}
	}
	for _, v := range resourceAlerts {
		alertTime := prototime.TimestampToTime(v.Timestamp)
		if alertTime.Before(timeEnd) && alertTime.After(timeStart) {
			allAlerts = append(allAlerts, v)
		}
	}
	return allAlerts, nil
}

// Watch on all alerts
func (kva *KvAlerts) Watch(alertsWatcherFunc AlertsWatcherFunc) error {

	prefix := alertsKey
	alertsWatcher = watcher{status: watchBootstrap, cb: alertsWatcherFunc, kvcb: kvdbWatch}

	if err := subscribeWatch(); err != nil {
		return err
	}

	// Subscribe for a watch is in a goroutine. Bootsrap by writing to the key and waiting for an update
	retries := 0
	kv := kvdb.Instance()
	for alertsWatcher.status == watchBootstrap {
		_, err := kv.Put(prefix+bootstrap, time.Now(), 1)
		if err != nil {
			return err
		}
		if alertsWatcher.status == watchBootstrap {
			retries++
			time.Sleep(time.Millisecond * 100)
		}
	}
	if alertsWatcher.status != watchReady {
		return fmt.Errorf("Failed to watch on %v", prefix)
	}

	return nil
}

// Shutdown
func (kva *KvAlerts) Shutdown() {
}

// String
func (kva *KvAlerts) String() string {
	return Name
}

func getResourceKey(resourceType api.ResourceType) string {
	if resourceType == api.ResourceType_VOLUMES {
		return alertsKey + volumeKey
	}
	if resourceType == api.ResourceType_NODE {
		return alertsKey + nodeKey
	}
	return alertsKey + clusterKey
}

func getNextAlertsIdKey() string {
	return alertsKey + nextAlertsIdKey
}

func getNextIdFromKVDB() (int64, error) {

	kv := kvdb.Instance()

	nextAlertsId := 0
	kvp, err := kv.Create(getNextAlertsIdKey(), strconv.FormatInt(int64(nextAlertsId+1), 10), 0)

	for err != nil {
		kvp, err = kv.GetVal(getNextAlertsIdKey(), &nextAlertsId)
		if err != nil {
			err = ErrNotInitialized
			return int64(nextAlertsId), err
		}
		prevValue := kvp.Value
		kvp.Value = []byte(strconv.FormatInt(int64(nextAlertsId+1), 10))
		kvp, err = kv.CompareAndSet(kvp, kvdb.KVFlags(0), prevValue)
	}

	return int64(nextAlertsId), err
}

func raiseAnAlerts(a api.Alerts, generateId func() (int64, error)) (api.Alerts, error) {
	kv := kvdb.Instance()

	if a.Resource == api.ResourceType_UNKNOWN_RESOURCE {
		return api.Alerts{}, ErrResourceNotFound
	}
	alertId, err := generateId()
	if err != nil {
		return a, err
	}
	a.Id = alertId
	a.Timestamp = prototime.Now()
	a.Cleared = false
	_, err = kv.Create(getResourceKey(a.Resource)+strconv.FormatInt(a.Id, 10), &a, 0)
	return a, err

}


func getResourceSpecificAlerts(resourceType api.ResourceType) ([]*api.Alerts, error) {
	kv := kvdb.Instance()
	allAlerts := []*api.Alerts{}
	kvp, err := kv.Enumerate(getResourceKey(resourceType))
	if err != nil {
		return nil, err
	}

	for _, v := range kvp {
		var elem *api.Alerts
		err = json.Unmarshal(v.Value, &elem)
		if err != nil {
			return nil, err
		}
		allAlerts = append(allAlerts, elem)
	}
	return allAlerts, nil
}

func getAllAlerts() ([]*api.Alerts, error) {
	allAlerts := []*api.Alerts{}
	clusterAlerts := []*api.Alerts{}
	nodeAlerts := []*api.Alerts{}
	volumeAlerts := []*api.Alerts{}
	var err error

	nodeAlerts, err = getResourceSpecificAlerts(api.ResourceType_NODE)
	if err == nil {
		allAlerts = append(allAlerts, nodeAlerts...)
	}
	volumeAlerts, err = getResourceSpecificAlerts(api.ResourceType_VOLUMES)
	if err == nil {
		allAlerts = append(allAlerts, volumeAlerts...)
	}
	clusterAlerts, err = getResourceSpecificAlerts(api.ResourceType_CLUSTER)
	if err == nil {
		allAlerts = append(allAlerts, clusterAlerts...)
	}

	if len(allAlerts) > 0 {
		return allAlerts, nil
	} 
	return allAlerts, err
}

func kvdbWatch(prefix string, opaque interface{}, kvp *kvdb.KVPair, err error) error {
	mutex.Lock()
	defer mutex.Unlock()

	if err == nil && strings.HasSuffix(kvp.Key, bootstrap) {
		alertsWatcher.status = watchReady
		return nil
	}

	if err != nil {
		if alertsWatcher.status == watchBootstrap {
			alertsWatcher.status = watchError
			return err
		}
		if watchErrors == 5 {
			return fmt.Errorf("Too many watch errors (%v)", watchErrors)
		}
		watchErrors++
		if err := subscribeWatch(); err != nil {
			return fmt.Errorf("Failed to resubscribe")
		}
	}

	watchErrors = 0

	if kvp.ModifiedIndex > alertsWatchIndex {
		alertsWatchIndex = kvp.ModifiedIndex
	}

	if kvp.Action == kvdb.KVDelete {
		err = alertsWatcher.cb(nil, AlertDeleteAction, prefix, kvp.Key)
		return err
	}

	var alert api.Alerts
	err = json.Unmarshal(kvp.Value, &alert)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal Alert")
	}

	switch kvp.Action {
	case kvdb.KVCreate:
		err = alertsWatcher.cb(&alert, AlertCreateAction, prefix, kvp.Key)
	case kvdb.KVSet:
		err = alertsWatcher.cb(&alert, AlertUpdateAction, prefix, kvp.Key)
	default:
		err = fmt.Errorf("Unhandled KV Action")
	}
	return err
}

func subscribeWatch() error {
	watchIndex := alertsWatchIndex
	if watchIndex != 0 {
		watchIndex = alertsWatchIndex + 1
	}

	kv := kvdb.Instance()
	if err := kv.WatchTree(alertsKey, watchIndex, nil, alertsWatcher.kvcb); err != nil {
		return err
	}
	return nil
}

func init() {
	Register(Name, Init)
}
