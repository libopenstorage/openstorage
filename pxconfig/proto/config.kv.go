package proto

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/portworx/kvdb"
	"golang.org/x/net/context"
)

const (
	// todo
	PXConfigKey = "someSecurePXConfigKey"
)

type ClusterSpecClientKV interface {
	Get(ctx context.Context, in *Empty) (*Config, error)
	Set(ctx context.Context, in *Config) (*Ack, error)
}

type clusterSpecClientKV struct {
	cc kvdb.Kvdb
}

func NewClusterSpecClientKV(cc kvdb.Kvdb) ClusterSpecClientKV {
	return &clusterSpecClientKV{cc}
}

func (c *clusterSpecClientKV) Get(ctx context.Context, in *Empty) (*Config, error) {
	out := new(Config)
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

func (c *clusterSpecClientKV) Set(ctx context.Context, in *Config) (*Ack, error) {
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
