package client

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sdeoras/openstorage/pxconfig/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	INVALID_INPUT_TYPE_ERROR_MESG = "client I/O not defined for input type"
)

type PxConfigClient interface {
	Get(ctx context.Context, in *proto.Empty) (*proto.Config, error)
	Set(ctx context.Context, in *proto.Config) (*proto.Ack, error)
}

type pxConfigClient struct {
	cc interface{}
}

func New(c interface{}) (PxConfigClient, error) {
	switch v := c.(type) {
	case *grpc.ClientConn:
		return &pxConfigClient{v}, nil
	case *http.Client:
		return &pxConfigClient{v}, nil
	case io.ReadWriter:
		return &pxConfigClient{v}, nil
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}

func (c *pxConfigClient) Get(ctx context.Context, in *proto.Empty) (*proto.Config, error) {
	switch v := c.cc.(type) {
	case *grpc.ClientConn:
		return proto.NewClusterSpecClient(v).Get(ctx, in)
	case io.ReadWriter:
		return proto.NewClusterSpecClientIO(v).Get(ctx, in)
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}

func (c *pxConfigClient) Set(ctx context.Context, in *proto.Config) (*proto.Ack, error) {
	switch v := c.cc.(type) {
	case *grpc.ClientConn:
		return proto.NewClusterSpecClient(v).Set(ctx, in)
	case io.ReadWriter:
		return proto.NewClusterSpecClientIO(v).Set(ctx, in)
	default:
		return nil, errors.Errorf("%s %T:", INVALID_INPUT_TYPE_ERROR_MESG, c)
	}
}
