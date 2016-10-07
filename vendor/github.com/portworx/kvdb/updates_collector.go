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
}

// GetUpdatesCollector creates new Kvdb collector that collects updates
// starting at startIndex + 1 index.
func GetUpdatesCollector(
	db Kvdb,
	prefix string,
	startIndex uint64,
) (UpdatesCollector, error) {
	collector := &updatesCollectorImpl{updates: make([]*kvdbUpdate, 0)}
	logrus.Infof("Starting watch at %v", startIndex)
	if err := db.WatchTree(prefix, startIndex, nil, collector.watchCb); err != nil {
		return nil, err
	}
	return collector, nil
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
	update := &kvdbUpdate{prefix: prefix, kvp: kvp, err: err}
	c.updates = append(c.updates, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *updatesCollectorImpl) Stop() {
	c.stopped = true
}

func (c *updatesCollectorImpl) ReplayUpdates(cbList []ReplayCb) error {
	updates := c.updates
	for _, update := range updates {
		for _, cbInfo := range cbList {
			if update.kvp == nil ||
				strings.HasPrefix(update.kvp.Key, cbInfo.Prefix) {
				err := cbInfo.WatchCB(update.prefix, cbInfo.Opaque, update.kvp,
					update.err)
				if err != nil {
					logrus.Infof("collect error: watchCB returned error: %v",
						err)
					return err
				}
			}
		}
	}
	return nil
}
