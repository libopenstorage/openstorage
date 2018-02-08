package osdconfig

import (
	"context"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/libopenstorage/openstorage/pkg/dbg"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

// NewManager can be used to instantiate ConfigManager
// Users of this function are expected to manage the execution via context
// github.com/Sirupsen/logrus package is used for logging internally
func NewManager(ctx context.Context, kv kvdb.Kvdb) (ConfigManager, error) {
	manager := new(configManager)
	manager.cb = make(map[string]*callbackData)
	manager.status = make(map[string]*Status)
	manager.dataToCallback = make(chan *DataWrite)

	// placeholder for external pointers
	manager.cc = kv
	manager.parentContext = ctx

	// derive local contexts from parent context
	// patent context -> manager.ctx -> manager.runCtx
	// if parent context is cancelled everything is cancelled
	// if manager.ctx is cancelled, then local context is also cancelled
	manager.ctx, manager.cancel = context.WithCancel(ctx)
	manager.runCtx, manager.runCancel = context.WithCancel(manager.ctx)

	// register function with kvdb to watch cluster level changes
	watcherTypes := []Watcher{ClusterWatcher, NodeWatcher}
	for _, watcherType := range watcherTypes {
		dataToKvdb := new(DataToKvdb)
		dataToKvdb.ctx = manager.ctx
		dataToKvdb.wd = manager.dataToCallback
		dataToKvdb.Type = watcherType

		// register function with different metadata but same channel to watch on
		switch watcherType {
		case ClusterWatcher:
			if err := kv.WatchTree(filepath.Join(baseKey, clusterKey), 0, dataToKvdb, cb); err != nil {
				logrus.Error(err)
				return nil, err
			}
		case NodeWatcher:
			if err := kv.WatchTree(filepath.Join(baseKey, nodesKey), 0, dataToKvdb, cb); err != nil {
				logrus.Error(err)
				return nil, err
			}
		}
	}

	// start a watch on kvdb
	go manager.watch(manager.dataToCallback) // start watching

	return manager, nil
}

// newManager can be used to instantiate configManager
// Users of this function are expected to manage the execution via context
// github.com/Sirupsen/logrus package is used for logging internally
func newManager(ctx context.Context, kv kvdb.Kvdb) (*configManager, error) {
	manager := new(configManager)
	manager.cb = make(map[string]*callbackData)
	manager.status = make(map[string]*Status)
	manager.dataToCallback = make(chan *DataWrite)

	// placeholder for external pointers
	manager.cc = kv
	manager.parentContext = ctx

	// derive local contexts from parent context
	// patent context -> manager.ctx -> manager.runCtx
	// if parent context is cancelled everything is cancelled
	// if manager.ctx is cancelled, then local context is also cancelled
	manager.ctx, manager.cancel = context.WithCancel(ctx)
	manager.runCtx, manager.runCancel = context.WithCancel(manager.ctx)

	// register function with kvdb to watch cluster level changes
	watcherTypes := []Watcher{ClusterWatcher, NodeWatcher}
	for _, watcherType := range watcherTypes {
		dataToKvdb := new(DataToKvdb)
		dataToKvdb.ctx = manager.ctx
		dataToKvdb.wd = manager.dataToCallback
		dataToKvdb.Type = watcherType

		// register function with different metadata but same channel to watch on
		switch watcherType {
		case ClusterWatcher:
			if err := kv.WatchTree(filepath.Join(baseKey, clusterKey), 0, dataToKvdb, cb); err != nil {
				logrus.Error(err)
				return nil, err
			}
		case NodeWatcher:
			if err := kv.WatchTree(filepath.Join(baseKey, nodesKey), 0, dataToKvdb, cb); err != nil {
				logrus.Error(err)
				return nil, err
			}
		}
	}

	// start a watch on kvdb
	go manager.watch(manager.dataToCallback) // start watching

	return manager, nil
}

// helper function to get a new callback function that can be registered
func newCallback(name string, minSleep, maxSleep int) func(ctx context.Context,
	opt interface{}) (chan<- *DataWrite, <-chan *DataRead) {
	f := func(ctx context.Context, opt interface{}) (chan<- *DataWrite, <-chan *DataRead) {
		send := make(chan *DataWrite)
		recv := make(chan *DataRead)
		go func() {
			select {
			case msg := <-send:
				if msg.Err != nil {
					logrus.Error(msg.Err)
				}
			case <-ctx.Done():
				logrus.Info("context cancellation received in: ", name)
				return
			}

			dur := time.Millisecond * time.Duration(rand.Intn(maxSleep)+minSleep)
			logrus.Info(name, " sleeping for ", dur)
			select {
			case <-time.After(dur):
			case <-ctx.Done():
				logrus.Info("context cancellation received in: ", name)
				return
			}

			d := new(DataRead)
			d.Name = name

			select {
			case recv <- d:
				return
			case <-ctx.Done():
				logrus.Info("context cancellation received in: ", name)
				return
			}
		}()
		return send, recv
	}
	return f
}

// helper function go get a new kvdb instance
func newInMemKvdb() (kvdb.Kvdb, error) {
	// create in memory kvdb
	if kv, err := kvdb.New(mem.Name, "", []string{}, nil, nil); err != nil {
		return nil, err
	} else {
		return kv, nil
	}
}

// helper function to obtain kvdb key for node based on nodeID
// the check for empty nodeID needs to be done elsewhere
func getNodeKeyFromNodeID(nodeID string) string {
	dbg.Assert(len(nodeID) > 0, "%s", "nodeID string can not be empty")
	return filepath.Join(baseKey, nodesKey, nodeID)
}
