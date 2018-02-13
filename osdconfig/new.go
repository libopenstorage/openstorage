package osdconfig

import (
	"container/heap"
	"context"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/portworx/kvdb"
)

// NewManager can be used to instantiate ConfigManager
// Users of this function are expected to manage the execution via context
// github.com/Sirupsen/logrus package is used for logging internally
func NewManager(ctx context.Context, kv kvdb.Kvdb) (ConfigManager, error) {
	return newManager(ctx, kv)
}

// newManager can be used to instantiate configManager
// Users of this function are expected to manage the execution via context
// github.com/Sirupsen/logrus package is used for logging internally
func newManager(ctx context.Context, kv kvdb.Kvdb) (*configManager, error) {
	manager := new(configManager)
	manager.cb = make(map[string]*callbackData)
	manager.status = make(map[string]*Status)
	manager.trigger = make(chan struct{})
	manager.lu = new(lastUpdate)
	manager.lu.Time = make(map[string]int64)
	manager.lu.Time["start"] = time.Now().UnixNano()

	// init heap
	manager.jobs = new(callbackHeap)
	heap.Init(manager.jobs)

	// placeholder for external pointers
	manager.cc = kv

	// derive local contexts from parent context
	// patent context -> manager.ctx -> manager.runCtx
	// if parent context is cancelled everything is cancelled
	// if manager.ctx is cancelled, then local context is also cancelled
	manager.ctx, manager.cancel = context.WithCancel(ctx)
	manager.runCtx, manager.runCancel = context.WithCancel(manager.ctx)

	// register function with kvdb to watch cluster level changes
	watcherTypes := []Band{TuneCluster, TuneNode}
	for _, watcherType := range watcherTypes {
		data := new(dataToKvdb)
		data.ctx = manager.ctx
		data.Type = watcherType
		data.cbh = manager.jobs
		data.trigger = manager.trigger

		// register function with different metadata but same channel to watch on
		switch watcherType {
		case TuneCluster:
			if err := kv.WatchTree(filepath.Join(rootKey, clusterKey), 0, data, cb); err != nil {
				logrus.Error(err)
				return nil, err
			}
		case TuneNode:
			if err := kv.WatchTree(filepath.Join(rootKey, nodeKey), 0, data, cb); err != nil {
				logrus.Error(err)
				return nil, err
			}
		}
	}

	// start a watch on kvdb
	go manager.watch(manager.trigger) // start watching

	return manager, nil
}
