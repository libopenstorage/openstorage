package proto

import (
	"io"
	"io/ioutil"

	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type SpecClientIO interface {
	GetGlobalSpec(ctx context.Context, in *Empty) (*GlobalConfig, error)
	SetGlobalSpec(ctx context.Context, in *GlobalConfig) (*Ack, error)
	GetClusterSpec(ctx context.Context, in *Empty) (*ClusterConfig, error)
	SetClusterSpec(ctx context.Context, in *ClusterConfig) (*Ack, error)
	GetNodeSpec(ctx context.Context, in *NodeID) (*NodeConfig, error)
	SetNodeSpec(ctx context.Context, in *NodeConfig) (*Ack, error)
}

type specClientIO struct {
	cc io.ReadWriter
	mu sync.Mutex
}

func NewSpecClientIO(cc io.ReadWriter) SpecClientIO {
	var mu sync.Mutex
	return &specClientIO{cc, mu}
}

func (c *specClientIO) GetGlobalSpec(ctx context.Context, in *Empty) (*GlobalConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	globalConf := new(GlobalConfig)

	b, err := ioutil.ReadAll(c.cc)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(b, globalConf); err != nil {
		return nil, err
	}

	return globalConf, nil
}

func (c *specClientIO) SetGlobalSpec(ctx context.Context, in *GlobalConfig) (*Ack, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

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

func (c *specClientIO) GetClusterSpec(ctx context.Context, in *Empty) (*ClusterConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	globalConf := new(GlobalConfig)

	b, err := ioutil.ReadAll(c.cc)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(b, globalConf); err != nil {
		return nil, err
	}

	if globalConf == nil || globalConf.ClusterConf == nil {
		return nil, errors.New("data does not contain configuration parameters")
	}

	return globalConf.ClusterConf, nil
}

func (c *specClientIO) SetClusterSpec(ctx context.Context, in *ClusterConfig) (*Ack, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	globalConf := new(GlobalConfig)

	b, err := ioutil.ReadAll(c.cc)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(b, globalConf); err != nil {
		return nil, err
	}

	if globalConf == nil {
		globalConf = new(GlobalConfig)
	}

	globalConf.ClusterConf = in

	b, err = proto.Marshal(globalConf)
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
	c.mu.Lock()
	defer c.mu.Unlock()

	if in == nil {
		return nil, errors.New("input is nil")
	}

	globalConf := new(GlobalConfig)
	b, err := ioutil.ReadAll(c.cc)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(b, globalConf); err != nil {
		return nil, err
	}

	if globalConf == nil {
		return nil, errors.New("data does not contain configuration parameters")
	}

	if globalConf.NodesConf == nil || globalConf.NodesConf.NodeConf == nil {
		return nil, errors.New("node configuration data not found")
	}

	nodeConfig, present := globalConf.NodesConf.NodeConf[in.ID]
	if !present {
		return nil, errors.New("node-id not found")
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

func (c *specClientIO) SetNodeSpec(ctx context.Context, in *NodeConfig) (*Ack, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if in == nil {
		return nil, errors.New("input is nil")
	}

	globalConf := new(GlobalConfig)
	b, err := ioutil.ReadAll(c.cc)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(b, globalConf); err != nil {
		return nil, err
	}

	if globalConf == nil {
		globalConf = new(GlobalConfig)
	}

	if globalConf.NodesConf == nil {
		globalConf.NodesConf = new(NodesConfig)
		globalConf.NodesConf.NodeConf = make(map[string]*NodeConfig)
	}

	if globalConf.NodesConf.NodeConf == nil {
		globalConf.NodesConf.NodeConf = make(map[string]*NodeConfig)
	}

	// set node spec
	globalConf.NodesConf.NodeConf[in.NodeId] = in

	// serialize it using proto buff
	b, err = proto.Marshal(globalConf)
	if err != nil {
		return nil, err
	}

	if n, err := c.cc.Write(b); err != nil {
		return nil, err
	} else {
		return &Ack{N: int64(n)}, nil
	}
}
