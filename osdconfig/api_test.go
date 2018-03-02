package osdconfig

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSetGetCluster(t *testing.T) {
	// create in memory kvdb
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	// get new config manager using handle to kvdb
	manager, err := NewManager(kv)
	if err != nil {
		t.Fatal(err)
	}

	// prepare expected cluster config
	expectedConf := new(ClusterConfig)
	expectedConf.ClusterId = "myClusterID"
	expectedConf.Driver = "myDriver"

	// set the expected cluster config value
	if err := manager.SetClusterConf(expectedConf); err != nil {
		t.Fatal(err)
	}

	// get the cluster config value
	receivedConf, err := manager.GetClusterConf()
	if err != nil {
		t.Fatal(err)
	}

	// compare expected and received
	if !reflect.DeepEqual(expectedConf, receivedConf) {
		t.Fatal("expected and received values are not deep equal")
	}
}

func TestSetGetNode(t *testing.T) {
	// create in memory kvdb
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	// get new config manager using handle to kvdb
	manager, err := NewManager(kv)
	if err != nil {
		t.Fatal(err)
	}

	// prepare expected cluster config
	expectedConf := new(NodeConfig)
	expectedConf.NodeId = "myNodeID"
	expectedConf.Storage = new(StorageConfig)
	expectedConf.Storage.Devices = []string{"dev1", "dev2"}

	// set the expected cluster config value
	if err := manager.SetNodeConf(expectedConf); err != nil {
		t.Fatal(err)
	}

	// get the cluster config value
	receivedConf, err := manager.GetNodeConf(expectedConf.NodeId)
	if err != nil {
		t.Fatal(err)
	}

	// compare expected and received
	if !reflect.DeepEqual(expectedConf, receivedConf) {
		t.Fatal("expected and received values are not deep equal")
	}

	// now delete the node
	if err := manager.UnsetNodeConf(expectedConf.NodeId); err != nil {
		t.Fatal("error in deleting node config")
	}

	// get the cluster config value
	_, err = manager.GetNodeConf(expectedConf.NodeId)
	if err == nil {
		t.Fatal("node does not exist, so this should error out")
	}
}

func TestCallback(t *testing.T) {
	// create in memory kvdb
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	// create new manager
	manager, err := NewManager(kv)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan error)
	clusterWatcher := func(config *ClusterConfig) error {
		var err error
		if config.ClusterId != "myClusterID" {
			err = fmt.Errorf("data not as expected")

		}
		if config.Driver != "myDriver" {
			err = fmt.Errorf("data not as expected")
		}
		ch <- err
		return nil
	}

	nodeWatcher := func(config *NodeConfig) error {
		var err error
		if config.Network.DataIface != "dataIface" {
			err = fmt.Errorf("data not as expected")
		}
		ch <- err
		return nil
	}

	// register cluster watcher callback
	if err := manager.WatchCluster("clusterWatcher", clusterWatcher); err != nil {
		t.Fatal(err)
	}

	// register node watcher callback
	if err := manager.WatchNode("nodeWatcher", nodeWatcher); err != nil {
		t.Fatal(err)
	}

	// update a few values
	if err := setSomeClusterValues(ch, manager); err != nil {
		t.Fatal(err)
	}

	// update more values... each of these updates will trigger callback execution
	if err := setSomeNodeValues(ch, manager); err != nil {
		t.Fatal(err)
	}
}

// setSomeClusterValues is a helper function to set cluster config values in kvdb
func setSomeClusterValues(ch chan error, manager ConfigManager) error {
	// prepare expected cluster config
	conf := new(ClusterConfig)
	conf.ClusterId = "myClusterID"
	conf.Driver = "myDriver"

	if err := manager.SetClusterConf(conf); err != nil {
		return err
	}

	return <-ch
}

// setSomeNodeValues is a helper function to set some node config values in kvdb
func setSomeNodeValues(ch chan error, manager ConfigManager) error {
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

		if err := <-ch; err != nil {
			return err
		}
	}

	return nil
}
