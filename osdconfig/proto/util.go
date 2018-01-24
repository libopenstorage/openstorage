package proto

import "github.com/pkg/errors"

func NewGlobalConfig() *GlobalConfig {
	gc := new(GlobalConfig)
	gc.ClusterConf = new(ClusterConfig)
	gc.NodesConf = new(NodesConfig)
	gc.NodesConf.NodeConf = make(map[string]*NodeConfig)
	return gc
}

func NewNodeConfig() *NodeConfig {
	nc := new(NodeConfig)
	return nc
}

func (gc *GlobalConfig) SetNode(node *NodeConfig) error {
	if gc == nil {
		return errors.New("receiver is nil")
	}

	if gc.NodesConf == nil {
		gc.NodesConf = new(NodesConfig)
		gc.NodesConf.NodeConf = make(map[string]*NodeConfig)
	}

	if gc.NodesConf.NodeConf == nil {
		gc.NodesConf.NodeConf = make(map[string]*NodeConfig)
	}

	gc.NodesConf.NodeConf[node.NodeId] = node
	return nil
}
