package osdconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestWatch(t *testing.T) {
	// create in memory kvdb
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	// create new manager
	manager, err := NewManager(context.Background(), kv)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// register cluster watcher callback
	if err := manager.WatchCluster("clusterWatcher", clusterWatcher); err != nil {
		t.Fatal(err)
	}

	// register node watcher callback
	if err := manager.WatchNode("nodeWatcher", nodeWatcher); err != nil {
		t.Fatal(err)
	}

	// update a few values
	if err := setSomeClusterValues(manager); err != nil {
		t.Fatal(err)
	}

	// update more values... each of these updates will trigger callback execution
	if err := setSomeNodeValues(manager); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)

}

// clusterWatcher is an example callback function to watch on cluster config changes
func clusterWatcher(config *ClusterConfig) error {
	if jb, err := json.MarshalIndent(config, "", "  "); err != nil {
		return err
	} else {
		fmt.Println(string(jb))
		if config.ClusterId != "myClusterID" {
			return errors.New("expected myClusterID, received " + config.ClusterId)
			//panic(DataErr)
		}
		if config.Driver != "myDriver" {
			return errors.New("expected myDriver, receive " + config.Driver)
			//panic(DataErr)
		}
	}
	return nil
}

// nodeWatcher is an example callback function to watch on node config changes
func nodeWatcher(config *NodeConfig) error {
	if jb, err := json.MarshalIndent(config, "", "  "); err != nil {
		return err
	} else {
		fmt.Println(string(jb))
		if config.Network.DataIface != "dataIface" {
			return errors.New("expected dataIface, received " + config.Network.DataIface)
			//panic(DataErr)
		}
	}
	return nil
}

// setSomeClusterValues is a helper function to set cluster config values in kvdb
func setSomeClusterValues(manager ConfigManager) error {
	// prepare expected cluster config
	conf := new(ClusterConfig)
	conf.ClusterId = "myClusterID"
	conf.Driver = "myDriver"

	if err := manager.SetClusterConf(conf); err != nil {
		return err
	}

	return nil
}

// setSomeNodeValues is a helper function to set some node config values in kvdb
func setSomeNodeValues(manager ConfigManager) error {
	// prepare expected cluster config
	conf := new(NodesConfig)
	conf.NodeConf = make(map[string]*NodeConfig)
	conf.NodeConf["node1"] = new(NodeConfig)
	conf.NodeConf["node2"] = new(NodeConfig)
	conf.NodeConf["node3"] = new(NodeConfig)

	for key, val := range conf.NodeConf {
		key, val := key, val
		val.NodeId = key
		val.Network = new(NetworkConfig)
		val.Network.DataIface = "dataIface"
		if err := manager.SetNodeConf(val); err != nil {
			return err
		}
	}

	return nil
}
