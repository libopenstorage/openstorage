package kvdb

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
)

type kvdbUpdate struct {
	// prefix is the path on which update was triggered
	prefix string
	// kvp is the actual key-value pair update
	kvp *KVPair
	// errors on update
	err error
}

type updatesCollectorImpl struct {
	// stopped is true if collection is stopped
	stopped bool
	// start index
	startIndex uint64
	// updatesMutex protects updates and start index
	updatesMutex sync.Mutex
	// updates stores the updates in order
	updates []*kvdbUpdate
}

func (c *updatesCollectorImpl) watchCb(
	prefix string,
	opaque interface{},
	kvp *KVPair,
	err error,
) error {
	if c.stopped {
		return fmt.Errorf("Stopped watch")
	}
	if err != nil {
		c.stopped = true
		return err
	}
	update := &kvdbUpdate{prefix: prefix, kvp: kvp, err: err}
	c.updatesMutex.Lock()
	c.updates = append(c.updates, update)
	c.updatesMutex.Unlock()
	return nil
}

func (c *updatesCollectorImpl) Stop() {
	logrus.Info("Stopping updates collector")
	c.stopped = true
}

func (c *updatesCollectorImpl) ReplayUpdates(
	cbList []ReplayCb,
) (uint64, error) {
	c.updatesMutex.Lock()
	updates := make([]*kvdbUpdate, len(c.updates))
	copy(updates, c.updates)
	c.updatesMutex.Unlock()
	index := c.startIndex
	logrus.Infof("collect: replaying %d update(s) for %d callback(s)",
		len(updates), len(cbList))
	for _, update := range updates {
		if update.kvp == nil {
			continue
		}
		index = update.kvp.ModifiedIndex
		for _, cbInfo := range cbList {
			if strings.HasPrefix(update.kvp.Key, cbInfo.Prefix) &&
				cbInfo.WaitIndex < update.kvp.ModifiedIndex {
				err := cbInfo.WatchCB(update.prefix, cbInfo.Opaque, update.kvp,
					update.err)
				if err != nil {
					logrus.Infof("collect error: watchCB returned error: %v",
						err)
					return index, err
				}
			} // else ignore the update
		}
	}
	return index, nil
}
