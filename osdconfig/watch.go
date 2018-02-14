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
		// executes on the queue
		manager.jobs.Lock()
		defer manager.jobs.Unlock()
		for manager.jobs.Len() > 0 {
			wd := heap.Pop(manager.jobs).(*dataWrite)

			logrus.WithField("source", sourceManagerWatch).
				WithField("execType", wd.Type).
				Info(msgTrigArrived)

			if wd.Err != nil {
				logrus.WithField("source", sourceManagerWatch).
					Error(msgExecError)
			} else {
				// execute callbacks
				manager.run(wd)
			}
		}
	}
}
