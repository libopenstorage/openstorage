package proto

import (
	"io"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type SpecClientIO interface {
	GetClusterSpec(ctx context.Context, in *Empty) (*ClusterConfig, error)
	SetClusterSpec(ctx context.Context, in *ClusterConfig) (*Ack, error)
	GetNodeSpec(ctx context.Context, in *NodeID) (*NodeConfig, error)
	SetNodeSpec(ctx context.Context, in *NodeConfig) (*Ack, error)
}

type specClientIO struct {
	cc io.ReadWriter
}

func NewSpecClientIO(cc io.ReadWriter) SpecClientIO {
	return &specClientIO{cc}
}

func (c *specClientIO) GetClusterSpec(ctx context.Context, in *Empty) (*ClusterConfig, error) {
	out := new(ClusterConfig)
	b, err := ioutil.ReadAll(c.cc)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(b, out); err != nil {
		return nil, err
	} else {
		return out, nil
	}
}

func (c *specClientIO) SetClusterSpec(ctx context.Context, in *ClusterConfig) (*Ack, error) {
	b, err := proto.Marshal(in)
	if err != nil {
		return nil, err
	}

	if n, err := c.cc.Write(b); err != nil {
		return nil, err
	} else {
		return &Ack{N: int64(n)}, nil
	}
}

func (c *specClientIO) GetNodeSpec(ctx context.Context, in *NodeID) (*NodeConfig, error) {
	if in == nil {
		return nil, errors.New("NodeID was a nil pointer")
	}

	clusterConfig := new(ClusterConfig)
	b, err := ioutil.ReadAll(c.cc)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(b, clusterConfig); err != nil {
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

func (c *specClientIO) SetNodeSpec(ctx context.Context, in *NodeConfig) (*Ack, error) {
	if in == nil {
		return nil, errors.New("NodeConfig was a nil pointer")
	}

	clusterConfig := new(ClusterConfig)
	b, err := ioutil.ReadAll(c.cc)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(b, clusterConfig); err != nil {
		return nil, err
	} else {
		clusterConfig.Nodes[in.NodeId] = in
	}

	b, err = proto.Marshal(clusterConfig)
	if err != nil {
		return nil, err
	}

	if n, err := c.cc.Write(b); err != nil {
		return nil, err
	} else {
		return &Ack{N: int64(n)}, nil
	}
}
