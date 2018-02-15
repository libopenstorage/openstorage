package osdconfig

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

// GetClusterConf retrieves cluster level data from kvdb
func (manager *configManager) GetClusterConf() (*ClusterConfig, error) {
	// get json from kvdb and unmarshal into config
	kvPair, err := manager.cc.Get(filepath.Join(baseKey, clusterKey))
	if err != nil {
		return nil, err
	}

	config := new(ClusterConfig)
	if err := json.Unmarshal(kvPair.Value, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SetClusterConf sets cluster config in kvdb
func (manager *configManager) SetClusterConf(config *ClusterConfig) error {
	manager.Lock()
	defer manager.Unlock()

	if config == nil {
		return fmt.Errorf("input cannot be nil")
	}

	// push into kvdb
	_, err := manager.cc.Put(filepath.Join(baseKey, clusterKey), config, 0)
	return err
}

// GetNodeConf retrieves node config data using nodeID
func (manager *configManager) GetNodeConf(nodeID string) (*NodeConfig, error) {
	if len(nodeID) == 0 {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// get json from kvdb and unmarshal into config
	kvPair, err := manager.cc.Get(getNodeKeyFromNodeID(nodeID))
	if err != nil {
		return nil, err
	}

	config := new(NodeConfig)
	if err = json.Unmarshal(kvPair.Value, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SetNodeConf sets node config data in kvdb
func (manager *configManager) SetNodeConf(config *NodeConfig) error {
	manager.Lock()
	defer manager.Unlock()

	if config == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if len(config.NodeId) == 0 {
		return fmt.Errorf("node id cannot be nil")
	}

	// push node data into kvdb
	_, err := manager.cc.Put(getNodeKeyFromNodeID(config.NodeId), config, 0)
	return err
}

// WatchCluster registers user defined function as callback and sets a watch for changes
// to cluster configuration
func (manager *configManager) WatchCluster(name string, cb func(config *ClusterConfig) error) error {
	manager.Lock()
	defer manager.Unlock()

	if _, present := manager.cbCluster[name]; present {
		return fmt.Errorf("%s already present", name)
	}
	manager.cbCluster[name] = cb
	return nil
}

// WatchNode registers user defined function as callback and sets a watch for changes
// to node configuration
func (manager *configManager) WatchNode(name string, cb func(config *NodeConfig) error) error {
	manager.Lock()
	defer manager.Unlock()

	if _, present := manager.cbNode[name]; present {
		return fmt.Errorf("%s already present", name)
	}
	manager.cbNode[name] = cb
	return nil
}
