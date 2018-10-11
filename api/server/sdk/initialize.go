package sdk

import (
	"github.com/google/go-cloud/wire"
	"github.com/libopenstorage/openstorage/alerts"
	"github.com/libopenstorage/openstorage/cluster"
)

// NewServer initializes a new SDK server.
func NewServer(
	Net,
	Address,
	RestPort,
	Driver,
	cluster.Cluster,
	alerts.FilterDeleter,
) (*Server, error) {
	wire.Build(ProviderSet)
	return &Server{}, nil
}

// NewMockServer initializes a new mock SDK server.
func NewMockServer() (*Server, error) {
	wire.Build(MockProviderSet)
	return &Server{}, nil
}
