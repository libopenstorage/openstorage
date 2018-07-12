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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"

	"github.com/libopenstorage/openstorage/api"
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
	server *Server
	m      *mockdriver.MockVolumeDriver
	c      *mockcluster.MockCluster
	mc     *gomock.Controller
	gw     *httptest.Server
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

	// Setup REST gateway
	mux, err := tester.server.restServerSetupHandlers()
	assert.NoError(t, err)
	assert.NotNil(t, mux)
	tester.gw = httptest.NewServer(mux)

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
	s.gw.Close()

	// Check mocks
	s.mc.Finish()
}

func (s *testServer) Conn() *grpc.ClientConn {
	return s.conn
}

func (s *testServer) Server() grpcserver.Server {
	return s.server
}

func (s *testServer) GatewayURL() string {
	return s.gw.URL
}

func TestSdkGateway(t *testing.T) {
	s := newTestServer(t)
	defer s.Stop()

	// Check we can get the swagger.json file
	res, err := http.Get(s.GatewayURL() + "/swagger.json")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Check we get the swagger-ui
	res, err = http.Get(s.GatewayURL() + "/swagger-ui")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Check unhandled address
	res, err = http.Get(s.GatewayURL() + "/this-should-not-work")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	// Check the gateway works
	// First setup the mock
	cluster := api.Cluster{
		Id:     "someid",
		NodeId: "somenodeid",
		Status: api.Status_STATUS_NOT_IN_QUORUM,
	}
	s.MockCluster().EXPECT().Enumerate().Return(cluster, nil).Times(1)

	// Then send the request
	res, err = http.Get(s.GatewayURL() + "/v1/clusters/current")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
