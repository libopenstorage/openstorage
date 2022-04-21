package manager

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/portworx/kvdb"
)

const (
	// ClusterDBKey is the key at which cluster info is store in kvdb
	ClusterDBKey = "cluster/database"
)

type updateCallbackFn func(db *cluster.ClusterInfo) (bool, error)

func snapAndReadClusterInfo(snapshotPrefixes []string) (*cluster.ClusterInitState, error) {
	kv := kvdb.Instance()

	// To work-around a kvdb issue with watches, try snapshot in a loop
	var (
		collector kvdb.UpdatesCollector
		err       error
		version   uint64
		snap      kvdb.Kvdb
	)
	for i := 0; i < 3; i++ {
		if i > 0 {
			logrus.Infof("Retrying snapshot")
		}
		// Start the watch before the snapshot
		collector, err = kvdb.NewUpdatesCollector(kv, "", 0)
		if err != nil {
			logrus.Errorf("Failed to start collector for cluster db: %v", err)
			collector = nil
			continue
		}
		// Create the snapshot
		snap, version, err = kv.Snapshot(snapshotPrefixes, true)
		if err != nil {
			logrus.Errorf("Snapshot failed for cluster db: %v", err)
			collector.Stop()
			collector = nil
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	logrus.Infof("Cluster db snapshot at: %v", version)

	clusterDB, err := snap.Get(ClusterDBKey)
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		logrus.Warnln("Warning, could not read cluster database")
		return nil, err
	}

	db := emptyClusterInfo()
	state := &cluster.ClusterInitState{
		ClusterInfo: &db,
		InitDb:      snap,
		Version:     version,
		Collector:   collector,
	}

	if clusterDB == nil || bytes.Compare(clusterDB.Value, []byte("{}")) == 0 {
		logrus.Infoln("Cluster is uninitialized...")
		return state, nil
	}
	if err := json.Unmarshal(clusterDB.Value, &db); err != nil {
		logrus.Warnln("Fatal, Could not parse cluster database ", kv)
		return state, err
	}

	return state, nil
}

func emptyClusterInfo() cluster.ClusterInfo {
	return cluster.ClusterInfo{
		Status:      api.Status_STATUS_INIT,
		NodeEntries: make(map[string]cluster.NodeEntry),
	}
}

func unmarshalClusterInfo(kv *kvdb.KVPair) (cluster.ClusterInfo, uint64, error) {
	db := emptyClusterInfo()
	version := uint64(0)
	if kv != nil {
		version = kv.ModifiedIndex
	}
	if kv == nil || bytes.Compare(kv.Value, []byte("{}")) == 0 {
		logrus.Infoln("Cluster is uninitialized...")
		return db, version, nil
	}
	if err := json.Unmarshal(kv.Value, &db); err != nil {
		logrus.Warnln("Fatal, Could not parse cluster database ", kv)
		return db, version, err
	}
	return db, version, nil
}

func readClusterInfo() (cluster.ClusterInfo, uint64, error) {
	kv, err := kvdb.Instance().Get(ClusterDBKey)
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		logrus.Warnln("Warning, could not read cluster database")
		return emptyClusterInfo(), 0, err
	}
	return unmarshalClusterInfo(kv)
}

func lockAndUpdateDB(fn, lockID string, cb updateCallbackFn) (*kvdb.KVPair, error) {
	kv := kvdb.Instance()

	kvlock, err := kv.LockWithTimeout(clusterLockKey, lockID, 10 * time.Minute, kv.GetLockTimeout())
	if err != nil {
		logrus.Warnf("Unable to obtain cluster lock for %v op: %v", fn, err)
		return nil, err
	}
	defer kv.Unlock(kvlock)

	db, version, err := readClusterInfo()
	if err != nil {
		return nil, err
	}

	update, err := cb(&db)
	if err != nil {
		return nil, err
	}
	if !update {
		return nil, nil
	}

	if version != 0 {
		b, err := json.Marshal(db)
		if err != nil {
			logrus.Warnf("Fatal, Could not marshal cluster database to JSON: %v", err)
			return nil, err
		}

		kvpInput := &kvdb.KVPair{
			Key:           ClusterDBKey,
			Value:         b,
			ModifiedIndex: version,
		}

		return kv.CompareAndSet(kvpInput, kvdb.KVModifiedIndex, nil)
	} else {
		// key does not exists yet, call create
		kvp, err := kv.Create(ClusterDBKey, db, 0)
		if err == kvdb.ErrExist {
			// key created, we may have lost the lock, retry
			err = kvdb.ErrModified
		}
		return kvp, err
	}
}

func updateDB(fn, lockID string, cb updateCallbackFn) error {
	_, err := updateAndGetDB(fn, lockID, cb, true)
	return err
}

func updateAndGetDB(
	fn string,
	lockID string,
	cb updateCallbackFn,
	retry bool,
) (*kvdb.KVPair, error) {
	for {
		if kvp, err := lockAndUpdateDB(fn, lockID, cb); err == kvdb.ErrModified && retry {
			// compare and set failed, retry
			continue
		} else {
			return kvp, err
		}
	}
}
