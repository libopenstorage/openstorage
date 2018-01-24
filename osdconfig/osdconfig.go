// osdconfig allows access to osd config params
package osdconfig

import (
	"io"

	"github.com/portworx/kvdb"
	"github.com/sdeoras/openstorage/osdconfig/api"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// top level obj to perform I/O on this package
type OsdConfig struct {
	cc interface{}
}

// constructors returning a handle to do I/O on osd config data
// that exist in KVDB
func NewKVDBConnection(conn kvdb.Kvdb) *OsdConfig {
	return &OsdConfig{conn}
}

// constructors returning a handle to do I/O on osd config data
// that exist in, say, a local file
func NewIOConnection(conn io.ReadWriter) *OsdConfig {
	return &OsdConfig{conn}
}

// constructors returning a handle to do I/O on osd config data
// over grpc connection
func NewGrpcConnection(conn *grpc.ClientConn) *OsdConfig {
	return &OsdConfig{conn}
}

func (c *OsdConfig) GetGlobalSpec(ctx context.Context, options ...interface{}) (*proto.GlobalConfig, error) {
	if osd, err := api.NewInterface(c.cc); err != nil {
		return nil, err
	} else {
		return osd.GetGlobalSpec(ctx, &proto.Empty{})
	}
}

func (c *OsdConfig) SetGlobalSpec(ctx context.Context, in *proto.GlobalConfig, options ...interface{}) (*proto.Ack, error) {
	if osd, err := api.NewInterface(c.cc); err != nil {
		return nil, err
	} else {
		return osd.SetGlobalSpec(ctx, in)
	}
}

func (c *OsdConfig) GetClusterSpec(ctx context.Context, options ...interface{}) (*proto.ClusterConfig, error) {
	if osd, err := api.NewInterface(c.cc); err != nil {
		return nil, err
	} else {
		return osd.GetClusterSpec(ctx, &proto.Empty{})
	}
}

func (c *OsdConfig) SetClusterSpec(ctx context.Context, in *proto.ClusterConfig, options ...interface{}) (*proto.Ack, error) {
	if osd, err := api.NewInterface(c.cc); err != nil {
		return nil, err
	} else {
		return osd.SetClusterSpec(ctx, in)
	}
}

func (c *OsdConfig) GetNodeSpec(ctx context.Context, in *proto.NodeID, options ...interface{}) (*proto.NodeConfig, error) {
	if osd, err := api.NewInterface(c.cc); err != nil {
		return nil, err
	} else {
		return osd.GetNodeSpec(ctx, &proto.NodeID{ID: in.ID})
	}
}

func (c *OsdConfig) SetNodeSpec(ctx context.Context, in *proto.NodeConfig, options ...interface{}) (*proto.Ack, error) {
	if osd, err := api.NewInterface(c.cc); err != nil {
		return nil, err
	} else {
		return osd.SetNodeSpec(ctx, in)
	}
}
