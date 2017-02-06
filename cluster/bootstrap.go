package cluster

import (
	"fmt"
	"strings"

	"go.pedge.io/dlog"
	"github.com/portworx/kvdb"
)

const (
	// ClusterBootstrapKey is the key at which bootstrap info is stored in kvdb
	ClusterBootstrapKey = "cluster/bootstrapdb"
)


func readBootstrapDB(bootstrapKvdb kvdb.Kvdb) (*kvdb.KVPair, *BootstrapClusterInfo, error) {
	bci := BootstrapClusterInfo{}
	kvp, err := bootstrapKvdb.GetVal(ClusterBootstrapKey, &bci)
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		dlog.Warnln("Warning, could not read cluster database")
		return nil, nil, err
	}
	return kvp, &bci, nil
}

func addNodeInBootstrapDB(bootstrapKvdb kvdb.Kvdb, bne BootstrapNodeEntry) (*kvdb.KVPair, *BootstrapClusterInfo, error) {
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
