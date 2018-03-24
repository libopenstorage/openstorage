package osdconfig

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

// GetClusterConf retrieves cluster level data from kvdb
func (manager *configManager) GetClusterConf() (*ClusterConfig, error) {
	// get json from kvdb and unmarshal into config
	kvPair, err := manager.kv.Get(filepath.Join(baseKey, clusterKey))
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
	if config == nil {
		return fmt.Errorf("input cannot be nil")
	}

	manager.Lock()
	defer manager.Unlock()

	// push into kvdb
	if _, err := manager.kv.Put(filepath.Join(baseKey, clusterKey), config, 0); err != nil {
		return err
	}

	return nil
}

// GetNodeConf retrieves node config data using nodeID
func (manager *configManager) GetNodeConf(nodeID string) (*NodeConfig, error) {
	if len(nodeID) == 0 {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// get json from kvdb and unmarshal into config
	kvPair, err := manager.kv.Get(getNodeKeyFromNodeID(nodeID))
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
	if config == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if len(config.NodeId) == 0 {
		return fmt.Errorf("node id cannot be nil")
	}

	manager.Lock()
	defer manager.Unlock()

	// push node data into kvdb
	if _, err := manager.kv.Put(getNodeKeyFromNodeID(config.NodeId), config, 0); err != nil {
		return err
	}

	return nil
}

// DeleteNodeConf deletes node config data in kvdb
func (manager *configManager) DeleteNodeConf(nodeID string) error {
	if len(nodeID) == 0 {
		return fmt.Errorf("node id cannot be nil")
	}

	manager.Lock()
	defer manager.Unlock()

	// remove dode data from kvdb
	if _, err := manager.kv.Delete(getNodeKeyFromNodeID(nodeID)); err != nil {
		return err
	}

	return nil
}

// EnumerateNodeConf fetches data for all nodes
func (manager *configManager) EnumerateNodeConf() (*NodesConfig, error) {
	keys, err := manager.kv.Keys(filepath.Join(baseKey, nodeKey), "/")
	if err != nil {
		return nil, fmt.Errorf("kvdb.Keys() returned error: " + err.Error())
	}

	nodesConfig := new(NodesConfig)
	*nodesConfig = make([]*NodeConfig, len(keys))
	for i, key := range keys {
		key := key
		nodeConfig, err := manager.GetNodeConf(key)
		if err != nil {
			return nil, err
		}
		(*nodesConfig)[i] = nodeConfig
	}

	return nodesConfig, nil
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
