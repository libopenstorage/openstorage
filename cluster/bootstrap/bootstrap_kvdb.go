package bootstrap

import (
	"fmt"
	"strings"

	"github.com/portworx/kvdb"
	"go.pedge.io/dlog"
)

const (
	// ClusterBootstrapKey is the key at which bootstrap info is stored in kvdb
	ClusterBootstrapKey = "cluster/bootstrapdb"
	// ClusterBootstrapLockKey is the bootstrap lock key
	ClusterBootstrapLockKey = "cluster/bootstrap/lock"
)

type bootstrapKvdb struct {
	kv       kvdb.Kvdb
	wcb      WatchClusterCB
}

func NewBootstrapKvdb(kv kvdb.Kvdb) ClusterBootstrap {
	return &bootstrapKvdb{
		kv: kv,
	}
}

func (b *bootstrapKvdb) AddNode(bne BootstrapNodeEntry) (*BootstrapClusterInfo, error) {
	kvlock, err := b.kv.LockWithID(ClusterBootstrapLockKey, bne.Id)
	if err != nil {
		dlog.Warnf("Unable to obtain bootstrapDB cluster lock")
		return nil, err
	}
	defer b.kv.Unlock(kvlock)

	bci, err := b.Enumerate()
	if err != nil {
		return nil, err
	}
	if bci.Size == 0 {
		bci.Nodes = make(map[string]BootstrapNodeEntry)
	}

	if _, exists := bci.Nodes[bne.Id]; !exists {
		// We do not exist in the map.
		bci.Size = bci.Size + 1
	}
	bci.Nodes[bne.Id] = bne
	kvp, err := b.kv.Put(ClusterBootstrapKey, &bci, 0)
	if err != nil {
		dlog.Warnf("Unable to add ourselves in bootstrap db")
		return nil, err
	}
	bci.Version = kvp.ModifiedIndex
	return bci, nil
}

func (b *bootstrapKvdb) RemoveNode(bne BootstrapNodeEntry) (*BootstrapClusterInfo, error) {
	kvlock, err := b.kv.LockWithID(ClusterBootstrapLockKey, bne.Id)
	if err != nil {
		dlog.Warnf("Unable to obtain bootstrapDB cluster lock")
		return nil, err
	}
	defer b.kv.Unlock(kvlock)

	bci, err := b.Enumerate()
	if err != nil {
		return nil, err
	}
	_, exists := bci.Nodes[bne.Id]
	if !exists {
		return nil, fmt.Errorf("Unable to find node %v in bootstrap db", bne.Id)
	}
	delete(bci.Nodes, bne.Id)
	kvp, err := b.kv.Put(ClusterBootstrapKey, &bci, 0)
	if err != nil {
		dlog.Warnf("Unable to add ourselves in bootstrap db")
		return nil, err
	}
	bci.Version = kvp.ModifiedIndex
	return bci, nil
}

func (b *bootstrapKvdb) Enumerate() (*BootstrapClusterInfo, error) {
	bci := BootstrapClusterInfo{}
	kvp, err := b.kv.GetVal(ClusterBootstrapKey, &bci)
	if err != nil && !strings.Contains(err.Error(), "Key not found") {
		dlog.Warnln("Warning, could not read bootstrap kv database")
		return nil, err
	}
	bci.Version = kvp.ModifiedIndex
	return &bci, nil
}

func (b *bootstrapKvdb) watchCluster(key string, opaque interface{}, kvp *kvdb.KVPair, watchErr error) error {
	wcb := opaque.(WatchClusterCB)
	bci, err  := b.Enumerate()
	// If we fail in enumerate let the callback function decide how to handle this watch
	werr := wcb(bci, err)
	return werr
}

func (b *bootstrapKvdb) WatchCluster(wcb WatchClusterCB, lastIndex uint64) error {
	if b.wcb != nil {
		return fmt.Errorf("Watch Cluster already started...")
	}
	dlog.Infof("Cluster bootstrap starting watch at version %d", lastIndex)
	go b.kv.WatchKey(ClusterBootstrapKey, lastIndex, wcb, b.watchCluster)
	return nil
}
