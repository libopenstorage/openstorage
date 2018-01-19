package osdconfig

import (
	"io"

	"github.com/portworx/kvdb"
	"github.com/sdeoras/openstorage/osdconfig/osdconfigapi"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type OsdConfigKVConnector interface {
	Handler() kvdb.Kvdb
}

type OsdConfigIOConnector interface {
	Handler() io.ReadWriter
}

type OsdConfigGrpcConnector interface {
	Handler() *grpc.ClientConn
}

type OsdConfig struct {
	cc interface{}
}

func NewKVConnection(conn OsdConfigKVConnector) *OsdConfig {
	return &OsdConfig{conn.Handler()}
}

func NewIOConnection(conn OsdConfigIOConnector) *OsdConfig {
	return &OsdConfig{conn.Handler()}
}

func NewGrpcConnection(conn OsdConfigGrpcConnector) *OsdConfig {
	return &OsdConfig{conn.Handler()}
}

func (c *OsdConfig) Get(ctx context.Context, options ...interface{}) (*proto.Config, error) {
	if pcc, err := osdconfigapi.NewInterface(c.cc); err != nil {
		return nil, err
	} else {
		return pcc.Get(ctx, &proto.Empty{})
	}
}

func (c *OsdConfig) Set(ctx context.Context, config *proto.Config, options ...interface{}) (*proto.Ack, error) {
	if pcc, err := osdconfigapi.NewInterface(c.cc); err != nil {
		return nil, err
	} else {
		return pcc.Set(ctx, config)
	}
}
