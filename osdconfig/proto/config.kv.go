package proto

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/portworx/kvdb"
	"golang.org/x/net/context"
)

const (
	// todo
	PXConfigKey = "someSecurePXConfigKey"
)

type clusterSpecClientKV struct {
	cc kvdb.Kvdb
}

func NewSpecClientKV(cc kvdb.Kvdb) SpecClientIO {
	return &clusterSpecClientKV{cc}
}

func (c *clusterSpecClientKV) GetClusterSpec(ctx context.Context, in *Empty) (*ClusterConfig, error) {
	out := new(ClusterConfig)
	kvpair, err := c.cc.Get(PXConfigKey)
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
	// schema check
	if _, err := proto.Marshal(in); err != nil {
		return nil, err
	}

	kvpair, err := c.cc.Put(PXConfigKey, in, 0)
	if err != nil {
		return nil, err
	} else {
		return &Ack{N: int64(len(kvpair.Value))}, nil
	}
}

func (c *clusterSpecClientKV) GetNodeSpec(ctx context.Context, in *NodeID) (*NodeConfig, error) {
	if in == nil {
		return nil, errors.New("NodeID was a nil pointer")
	}

	clusterConfig := new(ClusterConfig)
	kvpair, err := c.cc.Get(PXConfigKey)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(kvpair.Value, clusterConfig); err != nil {
		return nil, err
	} else {
		nodeConfig, present := clusterConfig.Nodes[in.ID]
		if !present {
			return nil, errors.New("NodeID not found")
		} else {
			return nodeConfig, nil
		}
	}
}

func (c *clusterSpecClientKV) SetNodeSpec(ctx context.Context, in *NodeConfig) (*Ack, error) {
	if in == nil {
		return nil, errors.New("NodeConfig was a nil pointer")
	}

	// schema check
	if _, err := proto.Marshal(in); err != nil {
		return nil, err
	}

	clusterConfig := new(ClusterConfig)
	kvpair, err := c.cc.Get(PXConfigKey)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(kvpair.Value, clusterConfig); err != nil {
		return nil, err
	} else {
		clusterConfig.Nodes[in.NodeId] = in
	}

	kvpair, err = c.cc.Put(PXConfigKey, clusterConfig, 0)
	if err != nil {
		return nil, err
	} else {
		return &Ack{N: int64(len(kvpair.Value))}, nil
	}
}
