package proto

import (
	"io"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"golang.org/x/net/context"
)

type ClusterSpecClientIO interface {
	Get(ctx context.Context, in *Empty) (*Config, error)
	Set(ctx context.Context, in *Config) (*Ack, error)
}

type clusterSpecClientIO struct {
	cc io.ReadWriter
}

func NewClusterSpecClientIO(cc io.ReadWriter) ClusterSpecClientIO {
	return &clusterSpecClientIO{cc}
}

func (c *clusterSpecClientIO) Get(ctx context.Context, in *Empty) (*Config, error) {
	out := new(Config)
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

func (c *clusterSpecClientIO) Set(ctx context.Context, in *Config) (*Ack, error) {
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
