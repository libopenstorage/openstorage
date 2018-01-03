package testing

import (
	"github.com/golang/mock/gomock"

	client "github.com/libopenstorage/openstorage/api/client"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
)

// testServer is a simple struct used abstract
// the creation and setup of mock server
type testServer struct {
	client *client.Client
	m      *mockdriver.MockVolumeDriver
	c      *mockcluster.MockCluster
	mc     *gomock.Controller
}

// MockDriver helper method.
func (s *testServer) MockDriver() *mockdriver.MockVolumeDriver {
	return s.m
}

// MockCluster helper method.
func (s *testServer) MockCluster() *mockcluster.MockCluster {
	return s.c
}

// Stop method to to remove the driver and check mocks.
func (s *testServer) Stop() {
	// Remove from registry
	volumedrivers.Remove("mock")
	// Check mocks
	s.mc.Finish()
}
