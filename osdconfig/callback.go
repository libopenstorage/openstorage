package osdconfig

import (
	"encoding/json"
	"fmt"

	"github.com/portworx/kvdb"
	"github.com/sirupsen/logrus"
)

// execClusterCallbacks executes a registered cluster watcher
func (manager *configManager) execClusterCallbacks(f CallbackClusterConfigFunc, data *data) {
	config := new(ClusterConfig)
	if err := json.Unmarshal(data.Value, config); err != nil {
		logrus.Error(err)
		return
	}

	f(config)
}

// execNodeCallbacks executes a registered node watcher
func (manager *configManager) execNodeCallbacks(f CallbackNodeConfigFunc, data *data) {
	config := new(NodeConfig)
	if err := json.Unmarshal(data.Value, config); err != nil {
		logrus.Error(err)
		return
	}

	f(config)
}

// kvdbCallback is a callback to be registered with kvdb.
// this callback simply receives data from kvdb and reflects it on a channel it receives in opaque
func (manager *configManager) kvdbCallback(prefix string,
	opaque interface{}, kvp *kvdb.KVPair, err error) error {
	manager.Lock()
	defer manager.Unlock()

	c, ok := opaque.(*dataToKvdb)
	if !ok {
		return fmt.Errorf("opaque value type is incorrect")
	}

	x := new(data)
	if kvp != nil {
		x.Key = kvp.Key
		x.Value = kvp.Value
	}
	x.Type = c.Type
	switch c.Type {
	case clusterWatcher:
		for _, f := range manager.cbCluster {
			go func(f1 CallbackClusterConfigFunc, wd *data) {
				manager.execClusterCallbacks(f1, wd)
			}(f, copyData(x))
		}
	case nodeWatcher:
		for _, f := range manager.cbNode {
			go func(f1 CallbackNodeConfigFunc, wd *data) {
				manager.execNodeCallbacks(f1, wd)
			}(f, copyData(x))
		}
	}

	return nil
}
