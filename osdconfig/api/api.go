package api

import (
	"io"

	"github.com/pkg/errors"
	"github.com/portworx/kvdb"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	INVALID_INPUT_TYPE_ERROR_MESG = "client I/O not defined for input type"
)

type OsdConfigInterface interface {
	Get(ctx context.Context, in *proto.Empty, options ...interface{}) (*proto.Config, error)
	Set(ctx context.Context, in *proto.Config, options ...interface{}) (*proto.Ack, error)
}

type pxConfigClient struct {
	cc interface{}
}

func NewInterface(c interface{}) (OsdConfigInterface, error) {
	switch v := c.(type) {
	case *grpc.ClientConn:
		return &pxConfigClient{v}, nil
	case io.ReadWriter:
		return &pxConfigClient{v}, nil
	case kvdb.Kvdb:
		return &pxConfigClient{v}, nil
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}

func (c *pxConfigClient) Get(ctx context.Context, in *proto.Empty, options ...interface{}) (*proto.Config, error) {
	switch v := c.cc.(type) {
	case *grpc.ClientConn:
		callOptions := make([]grpc.CallOption, len(options))
		for i, v := range options {
			callOptions[i] = v.(grpc.CallOption)
		}
		return proto.NewClusterSpecClient(v).Get(ctx, in, callOptions...)
	case io.ReadWriter:
		return proto.NewClusterSpecClientIO(v).Get(ctx, in)
	case kvdb.Kvdb:
		return proto.NewClusterSpecClientKV(v).Get(ctx, in)
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}

func (c *pxConfigClient) Set(ctx context.Context, in *proto.Config, options ...interface{}) (*proto.Ack, error) {
	switch v := c.cc.(type) {
	case *grpc.ClientConn:
		callOptions := make([]grpc.CallOption, len(options))
		for i, v := range options {
			callOptions[i] = v.(grpc.CallOption)
		}
		return proto.NewClusterSpecClient(v).Set(ctx, in, callOptions...)
	case io.ReadWriter:
		return proto.NewClusterSpecClientIO(v).Set(ctx, in)
	case kvdb.Kvdb:
		return proto.NewClusterSpecClientKV(v).Set(ctx, in)
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}
