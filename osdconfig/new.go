package osdconfig

import (
	"container/heap"
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

// helper function to get a new callback function that can be registered
func newCallback(name string, minSleep, maxSleep int) func(ctx context.Context,
	opt interface{}) (chan<- *dataWrite, <-chan *dataRead) {
	f := func(ctx context.Context, opt interface{}) (chan<- *dataWrite, <-chan *dataRead) {
		send := make(chan *dataWrite)
		recv := make(chan *dataRead)
		go func() {
			select {
			case msg := <-send:
				if msg.Err != nil {
					logrus.Error(msg.Err)
				}
			case <-ctx.Done():
				logrus.WithField("source", sourceCallback).
					WithField("callback", name).
					Warn(msgCtxCancelled)
				return
			}

			dur := time.Millisecond * time.Duration(rand.Intn(maxSleep)+minSleep)
			logrus.WithField("source", sourceCallback).
				WithField("callback", name).
				WithField("duration", dur).
				Debug("sleeping")
			select {
			case <-time.After(dur):
			case <-ctx.Done():
				logrus.WithField("source", sourceCallback).
					WithField("callback", name).
					Warn(msgCtxCancelled)
				return
			}

			d := new(dataRead)
			d.Name = name

			select {
			case recv <- d:
				return
			case <-ctx.Done():
				logrus.WithField("source", sourceCallback).
					WithField("callback", name).
					Warn(msgCtxCancelled)
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
	return filepath.Join(rootKey, nodeKey, nodeID)
}
