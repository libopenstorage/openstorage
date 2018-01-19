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
	INVALID_INPUT_TYPE_ERROR_MESG = "invalid input type"
)

type OsdConfigInterface interface {
	GetClusterSpec(ctx context.Context, in *proto.Empty, options ...interface{}) (*proto.ClusterConfig, error)
	SetClusterSpec(ctx context.Context, in *proto.ClusterConfig, options ...interface{}) (*proto.Ack, error)
	GetNodeSpec(ctx context.Context, in *proto.NodeID, options ...interface{}) (*proto.NodeConfig, error)
	SetNodeSpec(ctx context.Context, in *proto.NodeConfig, options ...interface{}) (*proto.Ack, error)
}

type osdConfig struct {
	cc interface{}
}

func NewInterface(c interface{}) (OsdConfigInterface, error) {
	switch v := c.(type) {
	case *grpc.ClientConn, io.ReadWriter, kvdb.Kvdb:
		return &osdConfig{v}, nil
	default:
		return nil, errors.Errorf("OSD interface not implemented. %s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}

func (c *osdConfig) GetClusterSpec(ctx context.Context, in *proto.Empty, options ...interface{}) (*proto.ClusterConfig, error) {
	switch v := c.cc.(type) {
	case *grpc.ClientConn:
		callOptions := make([]grpc.CallOption, len(options))
		for i, option := range options {
			switch v := option.(type) {
			case grpc.CallOption:
				callOptions[i] = v.(grpc.CallOption)
			default:
				return nil, errors.Errorf("GRPC call option error, %s %T:", INVALID_INPUT_TYPE_ERROR_MESG, option)
			}
		}
		return proto.NewSpecClient(v).GetClusterSpec(ctx, in, callOptions...)
	case io.ReadWriter:
		return proto.NewSpecClientIO(v).GetClusterSpec(ctx, in)
	case kvdb.Kvdb:
		return proto.NewSpecClientKV(v).GetClusterSpec(ctx, in)
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}

func (c *osdConfig) SetClusterSpec(ctx context.Context, in *proto.ClusterConfig, options ...interface{}) (*proto.Ack, error) {
	switch v := c.cc.(type) {
	case *grpc.ClientConn:
		callOptions := make([]grpc.CallOption, len(options))
		for i, option := range options {
			switch v := option.(type) {
			case grpc.CallOption:
				callOptions[i] = v.(grpc.CallOption)
			default:
				return nil, errors.Errorf("GRPC call option error, %s %T:", INVALID_INPUT_TYPE_ERROR_MESG, option)
			}
		}
		return proto.NewSpecClient(v).SetClusterSpec(ctx, in, callOptions...)
	case io.ReadWriter:
		return proto.NewSpecClientIO(v).SetClusterSpec(ctx, in)
	case kvdb.Kvdb:
		return proto.NewSpecClientKV(v).SetClusterSpec(ctx, in)
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}

func (c *osdConfig) GetNodeSpec(ctx context.Context, in *proto.NodeID, options ...interface{}) (*proto.NodeConfig, error) {
	switch v := c.cc.(type) {
	case *grpc.ClientConn:
		callOptions := make([]grpc.CallOption, len(options))
		for i, option := range options {
			switch v := option.(type) {
			case grpc.CallOption:
				callOptions[i] = v.(grpc.CallOption)
			default:
				return nil, errors.Errorf("GRPC call option error, %s %T:", INVALID_INPUT_TYPE_ERROR_MESG, option)
			}
		}
		return proto.NewSpecClient(v).GetNodeSpec(ctx, in, callOptions...)
	case io.ReadWriter:
		return proto.NewSpecClientIO(v).GetNodeSpec(ctx, in)
	case kvdb.Kvdb:
		return proto.NewSpecClientKV(v).GetNodeSpec(ctx, in)
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}

func (c *osdConfig) SetNodeSpec(ctx context.Context, in *proto.NodeConfig, options ...interface{}) (*proto.Ack, error) {
	switch v := c.cc.(type) {
	case *grpc.ClientConn:
		callOptions := make([]grpc.CallOption, len(options))
		for i, option := range options {
			switch v := option.(type) {
			case grpc.CallOption:
				callOptions[i] = v.(grpc.CallOption)
			default:
				return nil, errors.Errorf("GRPC call option error, %s %T:", INVALID_INPUT_TYPE_ERROR_MESG, option)
			}
		}
		return proto.NewSpecClient(v).SetNodeSpec(ctx, in, callOptions...)
	case io.ReadWriter:
		return proto.NewSpecClientIO(v).SetNodeSpec(ctx, in)
	case kvdb.Kvdb:
		return proto.NewSpecClientKV(v).SetNodeSpec(ctx, in)
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}
