package sdk

import "github.com/libopenstorage/openstorage/pkg/grpcserver"

type GrpcNet string
type GrpcAddress string

// NewGrpcServer is a provider of GrpcServer
func NewGrpcServer(net GrpcNet, address GrpcAddress) (*grpcserver.GrpcServer, error) {
	return grpcserver.New(&grpcserver.GrpcServerConfig{
		Name:    "SDK",
		Net:     string(net),
		Address: string(address),
	})
}
