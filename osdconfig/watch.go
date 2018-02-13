package osdconfig

import (
	"container/heap"
	"time"

	"github.com/Sirupsen/logrus"
)

// watch starts a watch on kvdb.
// it listens on a channel forever as long as context is alive
func (manager *configManager) watch(c <-chan struct{}) {
	// watch lock guarantees only one watcher is spawned
	manager.watchLock.Lock()
	defer manager.watchLock.Unlock()

	// loop forever until watch context is alive
	for {
		select {
		case <-manager.ctx.Done():
			logrus.WithField("source", sourceManagerWatch).
				Info("context cancelled")
			logrus.WithField("source", sourceManagerWatch).
				Info(msgWatchStopped)
			return
		case <-c:
			logrus.WithField("source", sourceManagerWatch).
				Info(msgTrigOnChan)
			manager.execQueue()
		case <-time.After(time.Second):
			logrus.WithField("source", sourceManagerWatch).
				Debug(msgTrigOnPoll)
			manager.execQueue()
		}
	}
}

func (manager *configManager) execQueue() {
	if manager.jobs.Len() > 0 {
		// lock, dedupe and flush heaps and unlock
		manager.jobs.Lock()
		logrus.WithField("source", sourceManagerWatch).
			WithField("len", manager.jobs.Len()).
			Info(msgHeapStatus)

		// build maps for deduping
		clusterJobs := make(map[string]*dataWrite)
		nodeJobs := make(map[string]*dataWrite)
		for manager.jobs.Len() > 0 {
			wd := heap.Pop(manager.jobs).(*dataWrite)
			switch wd.Type {
			case TuneCluster:
				clusterJobs[string(wd.Value)] = wd
			case TuneNode:
				nodeJobs[string(wd.Value)] = wd
			}
		}
		manager.jobs.Unlock()

		// build two new heaps and populate them
		clusterHeap := new(callbackHeap)
		nodeHeap := new(callbackHeap)
		for _, v := range clusterJobs {
			heap.Push(clusterHeap, v)
		}
		for _, v := range nodeJobs {
			heap.Push(nodeHeap, v)
		}

		// check update is newer and worth applying
		// timestamp check is done based on key
		manager.lu.Lock()
		defer manager.lu.Unlock()

		heaps := []*callbackHeap{clusterHeap, nodeHeap}
		for _, h := range heaps {
			for h.Len() > 0 {
				wd := heap.Pop(h).(*dataWrite)
				if wd.Time > manager.lu.Time[wd.Key] {
					manager.lu.Time[wd.Key] = wd.Time
					logrus.WithField("source", sourceManagerWatch).
						WithField("execType", wd.Type).
						Info(msgTrigArrived)

					if wd.Err != nil {
						logrus.WithField("source", sourceManagerWatch).
							Error(msgExecError)
					} else {
						// execute callbacks
						manager.run(wd)

						// wait for callbacks to complete
						manager.wait()

						// print execution status
						manager.printStatus()
					}
				}
			}
		}
	}
}
