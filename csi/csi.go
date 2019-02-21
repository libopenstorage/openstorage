/*
Package csi is CSI driver interface for OSD
Copyright 2017 Portworx

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
package csi

import (
	"fmt"
	"sync"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/libopenstorage/openstorage/api/spec"
	"github.com/libopenstorage/openstorage/cluster"
	authsecrets "github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
)

// OsdCsiServerConfig provides the configuration to the
// the gRPC CSI server created by NewOsdCsiServer()
type OsdCsiServerConfig struct {
	Net        string
	Address    string
	DriverName string
	Cluster    cluster.Cluster
	SdkUds     string
}

// OsdCsiServer is a OSD CSI compliant server which
// proxies CSI requests for a single specific driver
type OsdCsiServer struct {
	csi.ControllerServer
	csi.NodeServer
	csi.IdentityServer

	*grpcserver.GrpcServer
	specHandler spec.SpecHandler
	driver      volume.VolumeDriver
	cluster     cluster.Cluster
	sdkUds      string
	conn        *grpc.ClientConn
	mu          sync.Mutex
}

// NewOsdCsiServer creates a gRPC CSI complient server on the
// specified port and transport.
func NewOsdCsiServer(config *OsdCsiServerConfig) (grpcserver.Server, error) {
	if nil == config {
		return nil, fmt.Errorf("Must supply configuration")
	}
	if len(config.DriverName) == 0 {
		return nil, fmt.Errorf("OSD Driver name must be provided")
	}
	// Save the driver for future calls
	d, err := volumedrivers.Get(config.DriverName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get driver %s info: %s", config.DriverName, err.Error())
	}

	gServer, err := grpcserver.New(&grpcserver.GrpcServerConfig{
		Name:    "CSI",
		Net:     config.Net,
		Address: config.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create CSI server: %v", err)
	}

	return &OsdCsiServer{
		specHandler: spec.NewSpecHandler(),
		GrpcServer:  gServer,
		driver:      d,
		cluster:     config.Cluster,
		sdkUds:      config.SdkUds,
	}, nil
}

func (s *OsdCsiServer) getConn() (*grpc.ClientConn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		var err error
		fmt.Println("Connecting to", s.sdkUds)
		s.conn, err = grpcserver.Connect(
			s.sdkUds,
			[]grpc.DialOption{grpc.WithInsecure()})
		if err != nil {
			return nil, fmt.Errorf("Failed to connect CSI to SDK uds %s: %v", s.sdkUds, err)
		}
	}
	return s.conn, nil
}

// Gets token from the secrets. In Kubernetes, the side car containers copy
// the contents of a K8S Secret map into the Secrets section of the CSI call.
func (s *OsdCsiServer) setupContextWithToken(ctx context.Context, csiSecrets map[string]string) context.Context {
	if token, ok := csiSecrets[authsecrets.SecretTokenKey]; ok {
		md := metadata.New(map[string]string{
			"authorization": "bearer " + token,
		})
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// Start is used to start the server.
// It will return an error if the server is already running.
func (s *OsdCsiServer) Start() error {
	return s.GrpcServer.Start(func(grpcServer *grpc.Server) {
		csi.RegisterIdentityServer(grpcServer, s)
		csi.RegisterControllerServer(grpcServer, s)
		csi.RegisterNodeServer(grpcServer, s)
	})
}
