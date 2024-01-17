package loadbalancer

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// Balancer provides APIs to load balance a gRPC connection over a given
// cluster.
type Balancer interface {
	// GetRemoteNodeConnection returns a gRPC client connection to a node
	// in the cluster using a round-robin algorithm. The API will return
	// an error if it fails to create a connection to a node in the cluster.
	// The boolean return argument is set to false if the connection is created
	// to the local node.
	GetRemoteNodeConnection(ctx context.Context) (*grpc.ClientConn, bool, error)
	// GetRemoteNode returns the node ID of the node to which the next remote
	// connection will be created. The boolean return argument is set to false
	// if the connection is created to the local node.
	GetRemoteNode() (string, bool, error)
}

type nullBalancer struct{}

// NewNullBalancer is the no-op implementation of the Balancer interface
func NewNullBalancer() Balancer {
	return &nullBalancer{}
}

func (n *nullBalancer) GetRemoteNodeConnection(ctx context.Context) (*grpc.ClientConn, bool, error) {
	return nil, false, fmt.Errorf("remote connections not supported")
}

func (n *nullBalancer) GetRemoteNode() (string, bool, error) {
	return "", false, fmt.Errorf("remote connections not supported")
}
