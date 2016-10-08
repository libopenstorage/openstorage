package kvdb

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"strings"
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
	// updates stores the updates in order
	updates []*kvdbUpdate
	// stopped is true if collection is stopped
	stopped bool
	// start index
	startIndex uint64
}

func (c *updatesCollectorImpl) watchCb(
	prefix string,
	opaque interface{},
	kvp *KVPair,
	err error,
) error {
	if c.stopped {
		return fmt.Errorf("Stop watch")
	}
	if err != nil {
		c.stopped = true
		return err
	}
	update := &kvdbUpdate{prefix: prefix, kvp: kvp, err: err}
	c.updates = append(c.updates, update)
	return nil
}

func (c *updatesCollectorImpl) Stop() {
	c.stopped = true
}

func (c *updatesCollectorImpl) ReplayUpdates(cbList []ReplayCb) (uint64, error) {
	updates := c.updates
	index := c.startIndex
	for _, update := range updates {
		if update.kvp != nil {
			index = update.kvp.KVDBIndex
		}
		for _, cbInfo := range cbList {
			if update.kvp == nil ||
				strings.HasPrefix(update.kvp.Key, cbInfo.Prefix) {
				err := cbInfo.WatchCB(update.prefix, cbInfo.Opaque, update.kvp,
					update.err)
				if err != nil {
					logrus.Infof("collect error: watchCB returned error: %v",
						err)
					return index, err
				}
			}
		}
	}
	return index, nil
}
