/*
Package grpcserver is a generic gRPC server manager
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
package grpcserver

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC service
type testServer struct {
	conn   *grpc.ClientConn
	server *GrpcServer
}

func newTestServer(t *testing.T) *testServer {
	tester := &testServer{}

	var err error
	// Setup simple driver
	tester.server, err = New(&GrpcServerConfig{
		Name:    "unit-test",
		Net:     "tcp",
		Address: "127.0.0.1:0",
	})
	assert.Nil(t, err)

	// Test Start
	var testGrpcServer *grpc.Server
	startCalled := false
	err = tester.server.Start(func(grpcServer *grpc.Server) {
		testGrpcServer = grpcServer
		startCalled = true
	})
	assert.Nil(t, err)
	assert.NotNil(t, testGrpcServer)
	assert.True(t, startCalled)

	// Setup a connection to the driver
	tester.conn, err = grpc.Dial(tester.server.Address(), grpc.WithInsecure())
	assert.Nil(t, err)

	return tester
}

func (s *testServer) Stop() {
	// Shutdown servers
	s.conn.Close()
	s.server.Stop()
}

func (s *testServer) Conn() *grpc.ClientConn {
	return s.conn
}

func (s *testServer) Server() *GrpcServer {
	return s.server
}

func TestGrpcServerStartConfig(t *testing.T) {
	_, err := New(nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Configuration")

	_, err = New(&GrpcServerConfig{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "must be provided")

	_, err = New(&GrpcServerConfig{
		Name: "name",
		Net:  "net",
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Address")

	_, err = New(&GrpcServerConfig{
		Address: "address",
		Net:     "net",
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Name")

	_, err = New(&GrpcServerConfig{
		Name:    "name",
		Address: "address",
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Net")

}

func TestGrpcServerStart(t *testing.T) {
	s := newTestServer(t)
	assert.True(t, s.Server().IsRunning())
	defer s.Stop()

	// Check if we can still talk to the server
	// after starting multiple times.
	called := false
	err := s.Server().Start(func(grpcserver *grpc.Server) {
		called = true
	})
	assert.False(t, called)
	assert.True(t, s.Server().IsRunning())
	assert.NotNil(t, err)
	err = s.Server().Start(func(grpcserver *grpc.Server) {
		called = true
	})
	assert.False(t, called)
	assert.True(t, s.Server().IsRunning())
	assert.NotNil(t, err)
	err = s.Server().Start(func(grpcserver *grpc.Server) {
		called = true
	})
	assert.False(t, called)
	assert.True(t, s.Server().IsRunning())
	assert.NotNil(t, err)

}

func TestServerStop(t *testing.T) {
	s := newTestServer(t)
	assert.True(t, s.Server().IsRunning())
	s.Stop()
	assert.False(t, s.Server().IsRunning())

	assert.NotPanics(t, s.Stop)
	assert.False(t, s.Server().IsRunning())
	assert.NotPanics(t, s.Stop)
	assert.False(t, s.Server().IsRunning())
	assert.NotPanics(t, s.Stop)
	assert.False(t, s.Server().IsRunning())
	assert.NotPanics(t, s.Stop)
	assert.False(t, s.Server().IsRunning())
}

func TestContextMetadata(t *testing.T) {

	// setup context
	ctx := AddMetadataToContext(context.Background(), "hello", "world")
	ctx = AddMetadataToContext(ctx, "jay", "kay")
	ctx = AddMetadataToContext(ctx, "one", "two")

	// TODO: Replace this manual conversion to an actual grpc call
	outgoingMd := metautils.ExtractOutgoing(ctx)
	incomingCtx := outgoingMd.ToIncoming(context.Background())

	assert.Equal(t, GetMetadataValueFromKey(incomingCtx, "hello"), "world")
	assert.Equal(t, GetMetadataValueFromKey(incomingCtx, "jay"), "kay")
	assert.Equal(t, GetMetadataValueFromKey(incomingCtx, "one"), "two")
}
