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
	clusterId string
	kvdb kvdb.Kvdb
}
type KvAlerts struct {
	kvdbName string
	kvdbDomain string
	kvdbMachines []string
	clusterId string
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
	// NameTest :  This alert instance used only for unit tests
	NameTest        = "alerts_kvdb_test"
)

const (
	watchBootstrap = watcherStatus(iota)
	watchReady
	watchError
)

var (
	kvdbMap          map[string]kvdb.Kvdb
	watcherMap       map[string]*watcher
	alertsWatchIndex uint64
	watchErrors      int
)

// GetKvdbInstance - Returns a kvdb instance associated with this alert client and clusterId combination
func (kva *KvAlerts) GetKvdbInstance() kvdb.Kvdb {
	return kvdbMap[kva.clusterId]
}

// Init initializes a AlertsClient interface implementation
func Init(name string, domain string, machines []string, clusterId string) (AlertsClient, error) {
	_, ok := kvdbMap[clusterId]
	if !ok {
		kv, err := kvdb.New(name, domain + "/" + clusterId, machines, nil)
		if err != nil {
			return nil, err
		}
		kvdbMap[clusterId] = kv
	}
	
	return &KvAlerts{kvdbName: name, kvdbDomain: domain, kvdbMachines: machines, clusterId: clusterId}, nil
}

// Raise raises an Alert
func (kva *KvAlerts) Raise(a api.Alerts) (api.Alerts, error) {
	kv := kva.GetKvdbInstance()
	if a.Resource == api.ResourceType_UNKNOWN_RESOURCE {
		return api.Alerts{}, ErrResourceNotFound
	}
	alertId, err := kva.getNextIdFromKVDB()
	if err != nil {
		return a, err
	}
	a.Id = alertId
	a.Timestamp = prototime.Now()
	a.Cleared = false
	_, err = kv.Create(getResourceKey(a.Resource)+strconv.FormatInt(a.Id, 10), &a, 0)
	return a, err
}


// Erase erases an alert
func (kva *KvAlerts) Erase(resourceType api.ResourceType, alertId int64) error {
	kv := kva.GetKvdbInstance()

	if resourceType == api.ResourceType_UNKNOWN_RESOURCE {
		return ErrResourceNotFound
	}
	var err error

	_, err = kv.Delete(getResourceKey(resourceType) + strconv.FormatInt(alertId, 10))
	return err
}

// Clear clears an alert
func (kva *KvAlerts) Clear(resourceType api.ResourceType, alertId int64) error {
	kv := kva.GetKvdbInstance()
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
	kv := kva.GetKvdbInstance()

	_, err = kv.GetVal(getResourceKey(resourceType)+strconv.FormatInt(alertId, 10), &alert)

	return alert, err
}

// Enumerate enumerates alerts
func (kva *KvAlerts) Enumerate(filter api.Alerts) ([]*api.Alerts, error) {
	allAlerts := []*api.Alerts{}
	resourceAlerts := []*api.Alerts{}
	var err error

	if filter.Resource != api.ResourceType_UNKNOWN_RESOURCE {
		resourceAlerts, err = kva.getResourceSpecificAlerts(filter.Resource)
		if err != nil {
			return nil, err
		}
	} else {
		resourceAlerts, err = kva.getAllAlerts()
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
		resourceAlerts, err = kva.getResourceSpecificAlerts(resourceType)
		if err != nil {
			return nil, err
		}
	} else {
		resourceAlerts, err = kva.getAllAlerts()
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
func (kva *KvAlerts) Watch(clusterId string, alertsWatcherFunc AlertsWatcherFunc) error {
	_, ok := kvdbMap[clusterId]
	if !ok {
		kv, err := kvdb.New(kva.kvdbName, kva.kvdbDomain + "/" + clusterId, kva.kvdbMachines, nil)
		if err != nil {
			return err
		}
		kvdbMap[clusterId] = kv
	}
	
	kv := kvdbMap[clusterId]
	alertsWatcher := &watcher{status: watchBootstrap, cb: alertsWatcherFunc, kvcb: kvdbWatch, kvdb: kv}
	watcherKey := clusterId
	watcherMap[watcherKey] = alertsWatcher

	if err := subscribeWatch(watcherKey); err != nil {
		return err
	}

	// Subscribe for a watch can be in a goroutine. Bootstrap by writing to the key and waiting for an update
	retries := 0
	
	for alertsWatcher.status == watchBootstrap {
		_, err := kv.Put(alertsKey+bootstrap, time.Now(), 1)
		if err != nil {
			return err
		}
		if alertsWatcher.status == watchBootstrap {
			retries++
			time.Sleep(time.Millisecond * 100)
		}
		if retries == 5 {
			return fmt.Errorf("Failed to bootstrap watch on %s", clusterId)
		}
	}
	if alertsWatcher.status != watchReady {
		return fmt.Errorf("Failed to watch on %s", clusterId)
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

func (kva *KvAlerts) getNextIdFromKVDB() (int64, error) {
	kv := kva.GetKvdbInstance()
	nextAlertsId := 0
	kvp, err := kv.Create(getNextAlertsIdKey(), strconv.FormatInt(int64(nextAlertsId+1), 10), 0)

	for err != nil {
		kvp, err = kv.GetVal(getNextAlertsIdKey(), &nextAlertsId)
		if err != nil {
			err = ErrNotInitialized
			return -1, err
		} 
		prevValue := kvp.Value
		newKvp := *kvp
		newKvp.Value = []byte(strconv.FormatInt(int64(nextAlertsId+1), 10))
		kvp, err = kv.CompareAndSet(&newKvp, kvdb.KVFlags(0), prevValue)
	}
	return int64(nextAlertsId), err
}

func (kva *KvAlerts) getResourceSpecificAlerts(resourceType api.ResourceType) ([]*api.Alerts, error) {
	kv := kva.GetKvdbInstance()
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

func (kva *KvAlerts) getAllAlerts() ([]*api.Alerts, error) {
	allAlerts := []*api.Alerts{}
	clusterAlerts := []*api.Alerts{}
	nodeAlerts := []*api.Alerts{}
	volumeAlerts := []*api.Alerts{}
	var err error

	nodeAlerts, err = kva.getResourceSpecificAlerts(api.ResourceType_NODE)
	if err == nil {
		allAlerts = append(allAlerts, nodeAlerts...)
	}
	volumeAlerts, err = kva.getResourceSpecificAlerts(api.ResourceType_VOLUMES)
	if err == nil {
		allAlerts = append(allAlerts, volumeAlerts...)
	}
	clusterAlerts, err = kva.getResourceSpecificAlerts(api.ResourceType_CLUSTER)
	if err == nil {
		allAlerts = append(allAlerts, clusterAlerts...)
	}

	if len(allAlerts) > 0 {
		return allAlerts, nil
	} else if len(allAlerts) == 0 {
		return nil, fmt.Errorf("No alerts raised yet")
	}
	return allAlerts, err
}

func kvdbWatch(prefix string, opaque interface{}, kvp *kvdb.KVPair, err error) error {
	mutex.Lock()
	defer mutex.Unlock()

	watcherKey := strings.Split(prefix, "/")[1]
	
	if err == nil && strings.HasSuffix(kvp.Key, bootstrap) {
		w := watcherMap[watcherKey]
		w.status = watchReady
		return nil
	}

	if err != nil {
		if w:= watcherMap[watcherKey]; w.status == watchBootstrap {
			w.status = watchError
			return err
		}
		if watchErrors == 5 {
			return fmt.Errorf("Too many watch errors (%v)", watchErrors)
		}
		watchErrors++
		if err := subscribeWatch(watcherKey); err != nil {
			return fmt.Errorf("Failed to resubscribe")
		}
	}


	if strings.HasSuffix(kvp.Key, nextAlertsIdKey) {
		// Ignore write on this key
		// Todo : Add a map of ignore keys
		return nil
	}
	watchErrors = 0

	if kvp.ModifiedIndex > alertsWatchIndex {
		alertsWatchIndex = kvp.ModifiedIndex
	}

	w := watcherMap[watcherKey]

	if kvp.Action == kvdb.KVDelete {
		err = w.cb(nil, AlertDeleteAction, prefix, kvp.Key)
		return err
	}

	var alert api.Alerts
	err = json.Unmarshal(kvp.Value, &alert)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal Alert")
	}

	switch kvp.Action {
	case kvdb.KVCreate:
		err = w.cb(&alert, AlertCreateAction, prefix, kvp.Key)
	case kvdb.KVSet:
		err = w.cb(&alert, AlertUpdateAction, prefix, kvp.Key)
	default:
		err = fmt.Errorf("Unhandled KV Action")
	}
	return err
}

func subscribeWatch(key string) error {
	watchIndex := alertsWatchIndex
	if watchIndex != 0 {
		watchIndex = alertsWatchIndex + 1
	}

	w, ok := watcherMap[key]
	if !ok {
		return fmt.Errorf("Failed to find a watch on cluster : %v", key)
	}
	
	kv := w.kvdb
	if err := kv.WatchTree(alertsKey, watchIndex, nil, w.kvcb); err != nil {
		return err
	}
	return nil
}

func init() {
	kvdbMap = make(map[string]kvdb.Kvdb)
	watcherMap = make(map[string]*watcher)
	Register(Name, Init)
	Register(NameTest, Init)
}
