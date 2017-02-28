package discovery

import (
	"fmt"
	"strings"
	"encoding/json"

	"github.com/portworx/kvdb"
	"go.pedge.io/dlog"
)

const (
	// clusterDiscoveryKey is the key at which discovery info is stored in kvdb
	clusterDiscoveryKey = "cluster/discoverydb"
	// clusterDiscoveryLockKey is the discovery lock key
	clusterDiscoveryLockKey = "cluster/discovery/lock"
)

type discoveryKvdb struct {
	kv  kvdb.Kvdb
	wcb WatchCB
}

func NewDiscoveryKvdb(kv kvdb.Kvdb) Cluster {
	return &discoveryKvdb{
		kv: kv,
	}
}

func (b *discoveryKvdb) compareAndSetClusterInfo(
	prevCi *ClusterInfo,
	currentCi *ClusterInfo,
) (*kvdb.KVPair, error) {
	prevValue, err := json.Marshal(prevCi)
	if err != nil {
		return nil, err
	}
	currValue, err := json.Marshal(currentCi)
	if err != nil {
		return nil, err
	}
	kvPair := &kvdb.KVPair{
		Key: clusterDiscoveryKey,
		Value: currValue,
	}
	return b.kv.CompareAndSet(kvPair, kvdb.KVFlags(0), prevValue)
}

func (b *discoveryKvdb) AddNode(ne NodeEntry) (*ClusterInfo, error) {
	kvlock, err := b.kv.LockWithID(clusterDiscoveryLockKey, ne.Id)
	if err != nil {
		dlog.Warnf("Unable to obtain discoveryDB cluster lock")
		return nil, err
	}
	defer b.kv.Unlock(kvlock)

	prevCi, err := b.Enumerate()
	if err != nil {
		return nil, err
	}

	ci := prevCi
	var (
		addErr error
		kvp *kvdb.KVPair
	)
	if len(ci.Nodes) == 0 {
		ci.Nodes = make(map[string]NodeEntry)
		ci.Nodes[ne.Id] = ne
		kvp, addErr = b.kv.Put(clusterDiscoveryKey, &ci, 0)
	} else {
		ci.Nodes[ne.Id] = ne
		kvp, addErr = b.compareAndSetClusterInfo(prevCi, ci)
	}
	if addErr != nil {
		dlog.Warnf("Unable to add ourselves in discovery db: %v", addErr)
		return nil, addErr
	}

	ci.Version = kvp.ModifiedIndex
	return ci, nil
}

func (b *discoveryKvdb) RemoveNode(ne NodeEntry) (*ClusterInfo, error) {
	kvlock, err := b.kv.LockWithID(clusterDiscoveryLockKey, ne.Id)
	if err != nil {
		dlog.Warnf("Unable to obtain discoveryDB cluster lock: %v", err)
		return nil, err
	}
	defer b.kv.Unlock(kvlock)

	prevCi, err := b.Enumerate()
	if err != nil {
		return nil, err
	}
	ci := prevCi
	_, exists := ci.Nodes[ne.Id]
	if !exists {
		return nil, ErrNodeDoesNotExist
	}
	delete(ci.Nodes, ne.Id)
	kvp, err := b.compareAndSetClusterInfo(prevCi, ci)
	if err != nil {
		dlog.Warnf("Unable to add ourselves in discovery db: %v", err)
		return nil, err
	}
	ci.Version = kvp.ModifiedIndex
	return ci, nil
}

func (b *discoveryKvdb) Enumerate() (*ClusterInfo, error) {
	ci := ClusterInfo{}
	kvp, err := b.kv.GetVal(clusterDiscoveryKey, &ci)
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		dlog.Warnln("Warning, could not read discovery kv database: %v",err)
		return nil, err
	}
	if kvp != nil {
		ci.Version = kvp.ModifiedIndex
	}
	return &ci, nil
}

func (b *discoveryKvdb) watchCluster(key string, opaque interface{}, kvp *kvdb.KVPair, watchErr error) error {
	wcb := opaque.(WatchCB)
	ci, err := b.Enumerate()
	// If we fail in enumerate let the callback function decide how to handle this watch
	werr := wcb(ci, err)
	return werr
}

func (b *discoveryKvdb) Watch(wcb WatchCB, lastIndex uint64) error {
	if b.wcb != nil {
		return fmt.Errorf("Watch Cluster already started...")
	}
	dlog.Infof("Cluster discovery starting watch at version %d", lastIndex)
	go b.kv.WatchKey(clusterDiscoveryKey, lastIndex, wcb, b.watchCluster)
	return nil
}
