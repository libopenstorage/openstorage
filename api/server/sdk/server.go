/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2018 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package sdk

import (
	"fmt"

	"google.golang.org/grpc"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/spec"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
)

// ServerConfig provides the configuration to the SDK server
type ServerConfig struct {
	Net        string
	Address    string
	DriverName string
	Cluster    cluster.Cluster
}

// Server is an implementation of the gRPC SDK interface
type Server struct {
	*grpcserver.GrpcServer

	specHandler spec.SpecHandler
	driver      volume.VolumeDriver
	cluster     cluster.Cluster
}

// Interface check
var _ grpcserver.Server = &Server{}

// New creates a new SDK gRPC server
func New(config *ServerConfig) (*Server, error) {
	if nil == config {
		return nil, fmt.Errorf("Configuration must be provided")
	}
	if len(config.DriverName) == 0 {
		return nil, fmt.Errorf("OpenStorage Driver name must be provided")
	}

	// Save the driver for future calls
	d, err := volumedrivers.Get(config.DriverName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get driver %s info: %s", config.DriverName, err.Error())
	}

	// Create gRPC server
	gServer, err := grpcserver.New(&grpcserver.GrpcServerConfig{
		Name:    "SDK",
		Net:     config.Net,
		Address: config.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to setup server: %v", err)
	}

	return &Server{
		GrpcServer:  gServer,
		driver:      d,
		cluster:     config.Cluster,
		specHandler: spec.NewSpecHandler(),
	}, nil
}

// Start is used to start the server.
// It will return an error if the server is already running.
func (s *Server) Start() error {
	return s.GrpcServer.Start(func(grpcServer *grpc.Server) {
		api.RegisterOpenStorageClusterServer(grpcServer, s)
	})
}
