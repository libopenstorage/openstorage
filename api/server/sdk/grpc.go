package sdk

import "github.com/libopenstorage/openstorage/pkg/grpcserver"

// NewGrpcServer is a provider of GrpcServer
func NewGrpcServer(net Net, address Address) (*grpcserver.GrpcServer, error) {
	return grpcserver.New(&grpcserver.GrpcServerConfig{
		Name:    "SDK",
		Net:     string(net),
		Address: string(address),
	})
}
