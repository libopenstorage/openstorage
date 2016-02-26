package alerts

import (
	"encoding/json"
	"github.com/portworx/kvdb"
	"strconv"
	"time"
)

const (
	alertsKey      = "alerts/"
	nextAlertIdKey = "nextAlertId"
	clusterKey     = "cluster/"
	volumeKey      = "volume/"
	nodeKey        = "node/"
)

type KvAlerts struct {
}

func getResourceKey(resourceType ResourceType) string {
	if resourceType == Volumes {
		return alertsKey + volumeKey
	} else if resourceType == Node {
		return alertsKey + nodeKey
	} else {
		return alertsKey + clusterKey
	}
}

func getNextAlertIdKey() string {
	return alertsKey + nextAlertIdKey
}

func getNextIdFromKVDB() (int64, error) {

	kv := kvdb.Instance()

	nextAlertId := 0
	kvp, err := kv.Create(getNextAlertIdKey(), strconv.FormatInt(int64(nextAlertId+1), 10), 0)

	for err != nil {
		kvp, err = kv.GetVal(getNextAlertIdKey(), &nextAlertId)
		if err != nil {
			err = ErrNotInitialized
			return int64(nextAlertId), err
		}
		prevValue := kvp.Value
		kvp.Value = []byte(strconv.FormatInt(int64(nextAlertId+1), 10))
		kvp, err = kv.CompareAndSet(kvp, kvdb.KVFlags(0), prevValue)
	}

	return int64(nextAlertId), err
}

func raiseAnAlert(a Alert, generateId func() (int64, error)) (Alert, error) {
	kv := kvdb.Instance()

	if a.Resource == 0 {
		return Alert{}, ErrResourceNotFound
	}
	alertId, err := generateId()
	if err != nil {
		return a, err
	}
	a.Id = alertId
	a.Timestamp = time.Now()
	a.Cleared = false
	_, err = kv.Create(getResourceKey(a.Resource)+strconv.FormatInt(a.Id, 10), &a, 0)
	return a, err

}

func (kva *KvAlerts) Raise(a Alert) (Alert, error) {
	return raiseAnAlert(a, getNextIdFromKVDB)
}

func (kva *KvAlerts) RaiseWithGenerateId(a Alert, generateId func() (int64, error)) (Alert, error) {
	return raiseAnAlert(a, generateId)
}

func (kva *KvAlerts) Erase(resourceType ResourceType, alertId int64) error {
	kv := kvdb.Instance()

	if resourceType == 0 {
		return ErrResourceNotFound
	}
	var err error

	_, err = kv.Delete(getResourceKey(resourceType) + strconv.FormatInt(alertId, 10))
	return err
}

func (kva *KvAlerts) Clear(resourceType ResourceType, alertId int64) error {
	kv := kvdb.Instance()
	var (
		err   error
		alert Alert
	)
	if resourceType == 0 {
		return ErrResourceNotFound
	}

	_, err = kv.GetVal(getResourceKey(resourceType)+strconv.FormatInt(alertId, 10), &alert)
	if err != nil {
		return err
	}
	alert.Cleared = true

	_, err = kv.Put(getResourceKey(resourceType)+strconv.FormatInt(alertId, 10), &alert, 0)
	return err
}

func (kva *KvAlerts) Retrieve(resourceType ResourceType, alertId int64) (Alert, error) {
	var (
		alert Alert
		err   error
	)
	if resourceType == 0 {
		return Alert{}, ErrResourceNotFound
	}
	kv := kvdb.Instance()

	_, err = kv.GetVal(getResourceKey(resourceType)+strconv.FormatInt(alertId, 10), &alert)

	return alert, err
}

func getResourceSpecificAlerts(resourceType ResourceType) ([]*Alert, error) {
	kv := kvdb.Instance()
	allAlerts := []*Alert{}
	kvp, err := kv.Enumerate(getResourceKey(resourceType))
	if err != nil {
		return nil, err
	}

	for _, v := range kvp {
		var elem *Alert
		err = json.Unmarshal(v.Value, &elem)
		if err != nil {
			return nil, err
		}
		allAlerts = append(allAlerts, elem)
	}
	return allAlerts, nil
}

func getAllAlerts() ([]*Alert, error) {
	allAlerts := []*Alert{}
	clusterAlerts := []*Alert{}
	nodeAlerts := []*Alert{}
	volumeAlerts := []*Alert{}
	var err error

	nodeAlerts, _ = getResourceSpecificAlerts(Node)
	if err == nil {
		allAlerts = append(allAlerts, nodeAlerts...)
	}
	volumeAlerts, _ = getResourceSpecificAlerts(Volumes)
	if err == nil {
		allAlerts = append(allAlerts, volumeAlerts...)
	}
	clusterAlerts, _ = getResourceSpecificAlerts(Cluster)
	if err == nil {
		allAlerts = append(allAlerts, clusterAlerts...)
	}

	if len(allAlerts) > 0 {
		return allAlerts, nil
	} else {
		return allAlerts, err
	}
}

func (kva *KvAlerts) Enumerate(filter Alert) ([]*Alert, error) {
	allAlerts := []*Alert{}
	resourceAlerts := []*Alert{}
	var err error

	if filter.Resource != 0 {
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

func (kva *KvAlerts) EnumerateWithinTimeRange(timeStart time.Time, timeEnd time.Time, resourceType ResourceType) ([]*Alert, error) {
	allAlerts := []*Alert{}
	resourceAlerts := []*Alert{}
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
		if v.Timestamp.Before(timeEnd) && v.Timestamp.After(timeStart) {
			allAlerts = append(allAlerts, v)
		}
	}
	return allAlerts, nil
}

func (kva *KvAlerts) String() string {
	return "alerts_kvdb"
}
