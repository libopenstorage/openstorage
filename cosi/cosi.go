package cosi

import (
	"fmt"

	"github.com/libopenstorage/openstorage/bucket"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"google.golang.org/grpc"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

// Server represents a COSI server
type Server struct {
	cosi.IdentityServer
	cosi.ProvisionerServer
	*grpcserver.GrpcServer

	driver bucket.BucketDriver
}

// Config for setting up a COSI server
type Config struct {
	Driver  bucket.BucketDriver
	Net     string
	Address string
}

// NewServer creates a new COSI gRPC server
func NewServer(cfg *Config) (grpcserver.Server, error) {
	// Create server
	gServer, err := grpcserver.New(&grpcserver.GrpcServerConfig{
		Name:    "COSI Alpha Server",
		Net:     cfg.Net,
		Address: cfg.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create CSI server: %v", err)
	}

	return &Server{
		GrpcServer: gServer,
		driver:     cfg.Driver,
	}, nil
}

// Start registers COSI services and starts the gRPC server
func (s *Server) Start() error {
	return s.GrpcServer.Start(func(grpcServer *grpc.Server) {
		cosi.RegisterIdentityServer(grpcServer, s)
		cosi.RegisterProvisionerServer(grpcServer, s)
	})
}
