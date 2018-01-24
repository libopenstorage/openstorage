package proto

import (
	"encoding/json"

	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/portworx/kvdb"
	"golang.org/x/net/context"
)

const (
	// todo
	kvGlobalKey  = "someSecurePXConfigKey"
	kvClusterKey = kvGlobalKey + "/" + "cluster_conf"
	kvNodeKey    = kvGlobalKey + "/" + "node_conf"
)

type clusterSpecClientKV struct {
	cc kvdb.Kvdb
	mu sync.Mutex
}

func NewSpecClientKV(cc kvdb.Kvdb) SpecClientIO {
	var mu sync.Mutex
	return &clusterSpecClientKV{cc, mu}
}

func (c *clusterSpecClientKV) GetGlobalSpec(ctx context.Context, in *Empty) (*GlobalConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := new(GlobalConfig)

	kvpair, err := c.cc.Get(kvGlobalKey)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(kvpair.Value, out); err != nil {
		return nil, err
	} else {
		return out, nil
	}
}

func (c *clusterSpecClientKV) SetGlobalSpec(ctx context.Context, in *GlobalConfig) (*Ack, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := proto.Marshal(in); err != nil {
		return nil, err
	}

	kvpair, err := c.cc.Put(kvGlobalKey, in, 0)
	if err != nil {
		return nil, err
	} else {
		return &Ack{N: int64(len(kvpair.Value))}, nil
	}
}

func (c *clusterSpecClientKV) GetClusterSpec(ctx context.Context, in *Empty) (*ClusterConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := new(ClusterConfig)

	kvpair, err := c.cc.Get(kvClusterKey)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(kvpair.Value, out); err != nil {
		return nil, err
	} else {
		return out, nil
	}
}

func (c *clusterSpecClientKV) SetClusterSpec(ctx context.Context, in *ClusterConfig) (*Ack, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// schema check
	globalConf := new(GlobalConfig)
	globalConf.ClusterConf = in
	if _, err := proto.Marshal(globalConf); err != nil {
		return nil, err
	}

	kvpair, err := c.cc.Put(kvClusterKey, in, 0)
	if err != nil {
		return nil, err
	} else {
		return &Ack{N: int64(len(kvpair.Value))}, nil
	}
}

func (c *clusterSpecClientKV) GetNodeSpec(ctx context.Context, in *NodeID) (*NodeConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if in == nil {
		return nil, errors.New("input is nil")
	}

	nodeConfig := new(NodeConfig)
	kvpair, err := c.cc.Get(kvNodeKey + "/" + in.ID)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(kvpair.Value, nodeConfig); err != nil {
		return nil, err
	} else {
		if nodeConfig == nil {
			return nil, errors.New("node not found")
		}

		if nodeConfig.NodeId != in.ID {
			return nil, errors.New("node not found")
		}
		return nodeConfig, nil
	}
}

func (c *clusterSpecClientKV) SetNodeSpec(ctx context.Context, in *NodeConfig) (*Ack, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if in == nil {
		return nil, errors.New("input is nil")
	}

	// schema check
	globalConf := new(GlobalConfig)
	globalConf.NodesConf = new(NodesConfig)
	globalConf.NodesConf.NodeConf = make(map[string]*NodeConfig)
	globalConf.NodesConf.NodeConf[in.NodeId] = in
	if _, err := proto.Marshal(globalConf); err != nil {
		return nil, err
	}

	kvpair, err := c.cc.Put(kvNodeKey+"/"+in.NodeId, in, 0)
	if err != nil {
		return nil, err
	} else {
		return &Ack{N: int64(len(kvpair.Value))}, nil
	}
}
