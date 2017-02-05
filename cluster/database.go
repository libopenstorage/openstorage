package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"go.pedge.io/dlog"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
)

const (
	// ClusterDBKey is the key at which cluster info is store in kvdb
	ClusterDBKey = "cluster/database"
	// ClusterBootstrapKey is the key at which bootstrap info is stored in kvdb
	ClusterBootstrapKey = "cluster/bootstrapdb"
)

func snapAndReadClusterInfo() (*ClusterInitState, error) {
	kv := kvdb.Instance()

	snap, version, err := kv.Snapshot("")
	if err != nil {
		dlog.Errorf("Snapshot failed for cluster db: %v", err)
		return nil, err
	}
	dlog.Infof("Cluster db snapshot at: %v", version)
	collector, err := kvdb.NewUpdatesCollector(kv, "", version)
	if err != nil {
		dlog.Errorf("Failed to start collector for cluster db: %v", err)
		return nil, err
	}

	clusterDB, err := snap.Get(ClusterDBKey)
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		dlog.Warnln("Warning, could not read cluster database")
		return nil, err
	}

	db := ClusterInfo{
		Status:      api.Status_STATUS_INIT,
		NodeEntries: make(map[string]NodeEntry),
	}
	state := &ClusterInitState{
		ClusterInfo: &db,
		InitDb:      snap,
		Version:     version,
		Collector:   collector,
	}

	if clusterDB == nil || bytes.Compare(clusterDB.Value, []byte("{}")) == 0 {
		dlog.Infoln("Cluster is uninitialized...")
		return state, nil
	}
	if err := json.Unmarshal(clusterDB.Value, &db); err != nil {
		dlog.Warnln("Fatal, Could not parse cluster database ", kv)
		return state, err
	}

	return state, nil
}

func readClusterInfo() (ClusterInfo, error) {
	kvdb := kvdb.Instance()

	db := ClusterInfo{
		Status:      api.Status_STATUS_INIT,
		NodeEntries: make(map[string]NodeEntry),
	}

	kv, err := kvdb.Get(ClusterDBKey)
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		dlog.Warnln("Warning, could not read cluster database")
		return db, err
	}

	if kv == nil || bytes.Compare(kv.Value, []byte("{}")) == 0 {
		dlog.Infoln("Cluster is uninitialized...")
		return db, nil
	}
	if err := json.Unmarshal(kv.Value, &db); err != nil {
		dlog.Warnln("Fatal, Could not parse cluster database ", kv)
		return db, err
	}

	return db, nil
}

func writeClusterInfo(db *ClusterInfo) (*kvdb.KVPair, error) {
	kvdb := kvdb.Instance()
	b, err := json.Marshal(db)
	if err != nil {
		dlog.Warnf("Fatal, Could not marshal cluster database to JSON: %v", err)
		return nil, err
	}

	kvp, err := kvdb.Put(ClusterDBKey, b, 0)
	if err != nil {
		dlog.Warnf("Fatal, Could not marshal cluster database to JSON: %v", err)
		return nil, err
	}
	return kvp, nil
}

func readBootstrapDB(bootstrapKvdb kvdb.Kvdb) (*kvdb.KVPair, *BootstrapClusterInfo, error) {
	bci := BootstrapClusterInfo{}
	kvp, err := bootstrapKvdb.GetVal(ClusterBootstrapKey, &bci)
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		dlog.Warnln("Warning, could not read cluster database")
		return nil, nil, err
	}
	return kvp, &bci, nil
}

func addOurselvesInBootstrapDB(bootstrapKvdb kvdb.Kvdb, bne BootstrapNodeEntry) (*kvdb.KVPair, *BootstrapClusterInfo, error) {
	kvlock, err := bootstrapKvdb.LockWithID(clusterLockKey, bne.Id)
	if err != nil {
		dlog.Warnf("Unable to obtain bootstrapDB cluster lock")
		return nil, nil, err
	}
	defer bootstrapKvdb.Unlock(kvlock)

	kvp, bci, err := readBootstrapDB(bootstrapKvdb)
	if err != nil {
		return nil, nil, err
	}
	if bci.Size == 0 {
		bci.Nodes = make(map[string]BootstrapNodeEntry)
	}

	if _, exists := bci.Nodes[bne.Id]; !exists {
		// We do not exist in the map.
		bci.Size = bci.Size + 1
	}
	bci.Nodes[bne.Id] = bne
	kvp, err = bootstrapKvdb.Put(ClusterBootstrapKey, &bci, 0)
	if err != nil {
		dlog.Warnf("Unable to add ourselves in bootstrap db")
		return nil, nil, err
	}
	return kvp, bci, nil
}

func removeNodeFromBootstrapDB(bootstrapKvdb kvdb.Kvdb, nodeId string) error {
	kvlock, err := bootstrapKvdb.LockWithID(clusterLockKey, nodeId)
	if err != nil {
		dlog.Warnf("Unable to obtain bootstrapDB cluster lock")
		return err
	}
	defer bootstrapKvdb.Unlock(kvlock)

	_, bci, err := readBootstrapDB(bootstrapKvdb)
	if err != nil {
		return err
	}
	_, exists := bci.Nodes[nodeId]
	if !exists {
		return fmt.Errorf("Unable to find node %v in bootstrap db", nodeId)
	}
	delete(bci.Nodes, nodeId)
	_, err = bootstrapKvdb.Put(ClusterBootstrapKey, &bci, 0)
	if err != nil {
		dlog.Warnf("Unable to add ourselves in bootstrap db")
		return err
	}
	return nil
}
