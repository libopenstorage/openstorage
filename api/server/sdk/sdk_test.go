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
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"

	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
)

const (
	mockDriverName = "mock"
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC CSI service
type testServer struct {
	conn   *grpc.ClientConn
	server grpcserver.Server
	m      *mockdriver.MockVolumeDriver
	c      *mockcluster.MockCluster
	mc     *gomock.Controller
}

func setupMockDriver(tester *testServer, t *testing.T) {
	volumedrivers.Add(mockDriverName, func(map[string]string) (volume.VolumeDriver, error) {
		return tester.m, nil
	})

	var err error

	// Register mock driver
	err = volumedrivers.Register(mockDriverName, nil)
	assert.Nil(t, err)
}

func newTestServer(t *testing.T) *testServer {
	tester := &testServer{}

	// Add driver to registry
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.m = mockdriver.NewMockVolumeDriver(tester.mc)
	tester.c = mockcluster.NewMockCluster(tester.mc)

	setupMockDriver(tester, t)

	var err error
	// Setup simple driver
	tester.server, err = New(&ServerConfig{
		DriverName: mockDriverName,
		Net:        "tcp",
		Address:    "127.0.0.1:0",
		Cluster:    tester.c,
	})
	assert.Nil(t, err)
	err = tester.server.Start()
	assert.Nil(t, err)

	// Setup a connection to the driver
	tester.conn, err = grpc.Dial(tester.server.Address(), grpc.WithInsecure())
	assert.Nil(t, err)

	return tester
}

func (s *testServer) MockDriver() *mockdriver.MockVolumeDriver {
	return s.m
}

func (s *testServer) MockCluster() *mockcluster.MockCluster {
	return s.c
}

func (s *testServer) Stop() {
	// Remove from registry
	volumedrivers.Remove("mock")

	// Shutdown servers
	s.conn.Close()
	s.server.Stop()

	// Check mocks
	s.mc.Finish()
}

func (s *testServer) Conn() *grpc.ClientConn {
	return s.conn
}

func (s *testServer) Server() grpcserver.Server {
	return s.server
}
