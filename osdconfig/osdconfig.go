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

// connector interfaces for clients of this library to implement
type KVDBConnector interface {
	Handler() kvdb.Kvdb
}

type IOConnector interface {
	Handler() io.ReadWriter
}

type GrpcConnector interface {
	Handler() *grpc.ClientConn
}

// top level obj to perform I/O on this package
type OsdConfig struct {
	cc interface{}
}

// constructors returning a handle to do I/O on osd config data
// that exist in KVDB
func NewKVDBConnection(conn KVDBConnector) *OsdConfig {
	return &OsdConfig{conn.Handler()}
}

// constructors returning a handle to do I/O on osd config data
// that exist in, say, a local file
func NewIOConnection(conn IOConnector) *OsdConfig {
	return &OsdConfig{conn.Handler()}
}

// constructors returning a handle to do I/O on osd config data
// over grpc connection
func NewGrpcConnection(conn GrpcConnector) *OsdConfig {
	return &OsdConfig{conn.Handler()}
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
