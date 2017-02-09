package discovery

import (
	"fmt"
	"strings"

	"github.com/portworx/kvdb"
	"go.pedge.io/dlog"
)

const (
	// ClusterDiscoveryKey is the key at which discovery info is stored in kvdb
	ClusterDiscoveryKey = "cluster/discoverydb"
	// ClusterDiscoveryLockKey is the discovery lock key
	ClusterDiscoveryLockKey = "cluster/discovery/lock"
)

type discoveryKvdb struct {
	kv       kvdb.Kvdb
	wcb      WatchClusterCB
}

func NewDiscoveryKvdb(kv kvdb.Kvdb) Cluster {
	return &discoveryKvdb{
		kv: kv,
	}
}

func (b *discoveryKvdb) AddNode(ne NodeEntry) (*ClusterInfo, error) {
	kvlock, err := b.kv.LockWithID(ClusterDiscoveryLockKey, ne.Id)
	if err != nil {
		dlog.Warnf("Unable to obtain discoveryDB cluster lock")
		return nil, err
	}
	defer b.kv.Unlock(kvlock)

	ci, err := b.Enumerate()
	if err != nil {
		return nil, err
	}
	if ci.Size == 0 {
		ci.Nodes = make(map[string]NodeEntry)
	}

	if _, exists := ci.Nodes[ne.Id]; !exists {
		// We do not exist in the map.
		ci.Size = ci.Size + 1
	}
	ci.Nodes[ne.Id] = ne
	kvp, err := b.kv.Put(ClusterDiscoveryKey, &ci, 0)
	if err != nil {
		dlog.Warnf("Unable to add ourselves in discovery db")
		return nil, err
	}
	ci.Version = kvp.ModifiedIndex
	return ci, nil
}

func (b *discoveryKvdb) RemoveNode(ne NodeEntry) (*ClusterInfo, error) {
	kvlock, err := b.kv.LockWithID(ClusterDiscoveryLockKey, ne.Id)
	if err != nil {
		dlog.Warnf("Unable to obtain discoveryDB cluster lock")
		return nil, err
	}
	defer b.kv.Unlock(kvlock)

	ci, err := b.Enumerate()
	if err != nil {
		return nil, err
	}
	_, exists := ci.Nodes[ne.Id]
	if !exists {
		return nil, fmt.Errorf("Unable to find node %v in discovery db", ne.Id)
	}
	delete(ci.Nodes, ne.Id)
	kvp, err := b.kv.Put(ClusterDiscoveryKey, &ci, 0)
	if err != nil {
		dlog.Warnf("Unable to add ourselves in discovery db")
		return nil, err
	}
	ci.Version = kvp.ModifiedIndex
	return ci, nil
}

func (b *discoveryKvdb) Enumerate() (*ClusterInfo, error) {
	ci := ClusterInfo{}
	kvp, err := b.kv.GetVal(ClusterDiscoveryKey, &ci)
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		dlog.Warnln("Warning, could not read discovery kv database")
		return nil, err
	}
	ci.Version = kvp.ModifiedIndex
	return &ci, nil
}

func (b *discoveryKvdb) watchCluster(key string, opaque interface{}, kvp *kvdb.KVPair, watchErr error) error {
	wcb := opaque.(WatchClusterCB)
	ci, err  := b.Enumerate()
	// If we fail in enumerate let the callback function decide how to handle this watch
	werr := wcb(ci, err)
	return werr
}

func (b *discoveryKvdb) WatchCluster(wcb WatchClusterCB, lastIndex uint64) error {
	if b.wcb != nil {
		return fmt.Errorf("Watch Cluster already started...")
	}
	dlog.Infof("Cluster discovery starting watch at version %d", lastIndex)
	go b.kv.WatchKey(ClusterDiscoveryKey, lastIndex, wcb, b.watchCluster)
	return nil
}
