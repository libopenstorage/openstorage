/*
CSI Interface for OSD
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
	"net"
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC CSI service
type testServer struct {
	listener     net.Listener
	conn         *grpc.ClientConn
	osdCsiServer *OsdCsiServer
}

func newTestServer(t *testing.T) *testServer {
	tester := &testServer{}

	// Setup simple driver
	var err error
	tester.osdCsiServer, err = NewOsdCsiServer(&OsdCsiServerConfig{
		Net:     "tcp",
		Address: "127.0.0.1:0",
	})
	assert.Nil(t, err)
	err = tester.osdCsiServer.Start()
	assert.Nil(t, err)

	// Setup a connection to the driver
	tester.conn, err = grpc.Dial(tester.osdCsiServer.Address(), grpc.WithInsecure())
	assert.Nil(t, err)

	return tester
}

func (s *testServer) Stop() {
	s.conn.Close()
	s.osdCsiServer.Stop()
}

func (s *testServer) Conn() *grpc.ClientConn {
	return s.conn
}

func (s *testServer) OsdCsiServer() *OsdCsiServer {
	return s.osdCsiServer
}

func TestCSIServerStart(t *testing.T) {
	s := newTestServer(t)
	assert.True(t, s.OsdCsiServer().running)
	defer s.Stop()

	// Check if we can still talk to the server
	// after starting multiple times.
	err := s.OsdCsiServer().Start()
	assert.True(t, s.OsdCsiServer().running)
	assert.NotNil(t, err)
	err = s.OsdCsiServer().Start()
	assert.True(t, s.OsdCsiServer().running)
	assert.NotNil(t, err)
	err = s.OsdCsiServer().Start()
	assert.True(t, s.OsdCsiServer().running)
	assert.NotNil(t, err)

	// Make a call
	c := csi.NewIdentityClient(s.Conn())
	r, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
	assert.Nil(t, err)

	// Verify
	name := r.GetResult().GetName()
	version := r.GetResult().GetVendorVersion()
	assert.Equal(t, name, csiDriverName)
	assert.Equal(t, version, csiDriverVersion)
}

func TestCSIServerStop(t *testing.T) {
	s := newTestServer(t)
	assert.True(t, s.OsdCsiServer().running)
	s.Stop()
	assert.False(t, s.OsdCsiServer().running)

	assert.NotPanics(t, s.Stop)
	assert.False(t, s.OsdCsiServer().running)
	assert.NotPanics(t, s.Stop)
	assert.False(t, s.OsdCsiServer().running)
	assert.NotPanics(t, s.Stop)
	assert.False(t, s.OsdCsiServer().running)
	assert.NotPanics(t, s.Stop)
	assert.False(t, s.OsdCsiServer().running)
}

func TestNewCSIServerGetPluginInfo(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewIdentityClient(s.Conn())
	r, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
	assert.Nil(t, err)

	// Verify
	name := r.GetResult().GetName()
	version := r.GetResult().GetVendorVersion()
	assert.Equal(t, name, csiDriverName)
	assert.Equal(t, version, csiDriverVersion)
}

func TestNewCSIServerGetSupportedVersions(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewIdentityClient(s.Conn())
	r, err := c.GetSupportedVersions(context.Background(), &csi.GetSupportedVersionsRequest{})
	assert.Nil(t, err)

	// Verify
	versions := r.GetResult().GetSupportedVersions()
	assert.Equal(t, len(versions), 1)
	assert.True(t, reflect.DeepEqual(versions[0], csiVersion))
}
