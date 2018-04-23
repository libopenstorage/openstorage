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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/mock/gomock"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
)

func TestNewSdkServerBadParameters(t *testing.T) {
	setupMockDriver(&testServer{}, t)
	s, err := New(nil)
	assert.Nil(t, s)
	assert.NotNil(t, err)

	s, err = New(&ServerConfig{})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "must be provided")

	s, err = New(&ServerConfig{
		Net: "test",
	})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "must be provided")

	s, err = New(&ServerConfig{
		Net:     "test",
		Address: "blah",
	})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "must be provided")

	s, err = New(&ServerConfig{
		Net:        "test",
		Address:    "blah",
		DriverName: "name",
	})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unable to get driver")

	// Add driver to registry
	mc := gomock.NewController(t)
	defer mc.Finish()
	m := mockdriver.NewMockVolumeDriver(mc)
	volumedrivers.Add("mock", func(map[string]string) (volume.VolumeDriver, error) {
		return m, nil
	})
	defer volumedrivers.Remove("mock")
	s, err = New(&ServerConfig{
		Net:        "test",
		Address:    "blah",
		DriverName: "mock",
	})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unable to setup server")
}

func TestSdkEnumerateNoNodes(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	cluster := api.Cluster{
		Id:     "someid",
		NodeId: "somenodeid",
		Status: api.Status_STATUS_NOT_IN_QUORUM,
	}
	s.MockCluster().EXPECT().Enumerate().Return(cluster, nil).Times(1)

	// Setup client
	c := api.NewOpenStorageClusterClient(s.Conn())

	// Get info
	r, err := c.Enumerate(context.Background(), &api.ClusterEnumerateRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetCluster())

	// Verify
	assert.Equal(t, r.GetCluster().GetId(), cluster.Id)
	assert.Equal(t, r.GetCluster().GetNodeId(), cluster.NodeId)
	assert.Equal(t, r.GetCluster().GetStatus(), cluster.Status)
	assert.Len(t, r.GetCluster().GetNodes(), 0)
}

func TestSdkEnumerate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	cluster := api.Cluster{
		Id:     "someid",
		NodeId: "somenodeid",
		Status: api.Status_STATUS_NOT_IN_QUORUM,
		Nodes: []api.Node{
			api.Node{
				Id:       "nodeid",
				Cpu:      1.414,
				MemTotal: 112,
				MemUsed:  41,
				MemFree:  93,
			},
		},
	}
	s.MockCluster().EXPECT().Enumerate().Return(cluster, nil).Times(1)

	// Setup client
	c := api.NewOpenStorageClusterClient(s.Conn())

	// Get info
	r, err := c.Enumerate(context.Background(), &api.ClusterEnumerateRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetCluster())

	// Verify
	assert.Equal(t, r.GetCluster().GetId(), cluster.Id)
	assert.Equal(t, r.GetCluster().GetNodeId(), cluster.NodeId)
	assert.Equal(t, r.GetCluster().GetStatus(), cluster.Status)
	assert.Len(t, r.GetCluster().GetNodes(), 1)

	// Verify node
	node := cluster.Nodes[0]
	rn := r.GetCluster().GetNodes()[0]
	assert.Equal(t, rn.GetId(), node.Id)
	assert.Equal(t, rn.GetCpu(), node.Cpu)
	assert.Equal(t, rn.GetMemFree(), node.MemFree)
	assert.Equal(t, rn.GetMemTotal(), node.MemTotal)
	assert.Equal(t, rn.GetMemUsed(), node.MemUsed)
}

func TestSdkEnumerateFail(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	mockerr := fmt.Errorf("MOCK")
	s.MockCluster().EXPECT().Enumerate().Return(api.Cluster{}, mockerr).Times(1)

	// Setup client
	c := api.NewOpenStorageClusterClient(s.Conn())

	// Get info
	r, err := c.Enumerate(context.Background(), &api.ClusterEnumerateRequest{})
	assert.Error(t, err)
	assert.Nil(t, r.GetCluster())

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Equal(t, serverError.Message(), mockerr.Error())
}

func TestSdkInspect(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	nodeid := "nodeid"
	node := api.Node{
		Id:       nodeid,
		Cpu:      1.414,
		MemTotal: 112,
		MemUsed:  41,
		MemFree:  93,
		Avgload:  834,
		Status:   api.Status_STATUS_MAX,
		Disks: map[string]api.StorageResource{
			"disk1": api.StorageResource{
				Id:     "12345",
				Path:   "mymount",
				Medium: api.StorageMedium_STORAGE_MEDIUM_SSD,
				Online: true,
			},
		},
		Timestamp: time.Now(),
		StartTime: time.Now(),
		NodeLabels: map[string]string{
			"hello": "world",
		},
	}
	s.MockCluster().EXPECT().Inspect(nodeid).Return(node, nil).Times(1)

	// Setup client
	c := api.NewOpenStorageClusterClient(s.Conn())

	// Get info
	r, err := c.Inspect(context.Background(), &api.ClusterInspectRequest{
		NodeId: nodeid,
	})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetNode())

	// Verify
	rn := r.GetNode()
	assert.Equal(t, rn.GetId(), node.Id)
	assert.Equal(t, rn.GetCpu(), node.Cpu)
	assert.Equal(t, rn.GetMemTotal(), node.MemTotal)
	assert.Equal(t, rn.GetMemFree(), node.MemFree)
	assert.Equal(t, rn.GetMemUsed(), node.MemUsed)
	assert.Equal(t, rn.GetAvgLoad(), int64(node.Avgload))
	assert.Equal(t, rn.GetStatus(), node.Status)

	// Check Disk
	assert.Len(t, rn.GetDisks(), 1)
	assert.Equal(t, *rn.GetDisks()["disk1"], node.Disks["disk1"])

	// Check Labels
	assert.Len(t, rn.GetNodeLabels(), 1)
	assert.Equal(t, rn.GetNodeLabels(), node.NodeLabels)
}

func TestSdkInspectFail(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	mockerr := fmt.Errorf("MOCK")
	s.MockCluster().EXPECT().Inspect("mynode").Return(api.Node{}, mockerr).Times(1)

	// Setup client
	c := api.NewOpenStorageClusterClient(s.Conn())

	// Get info
	r, err := c.Inspect(context.Background(), &api.ClusterInspectRequest{
		NodeId: "mynode",
	})
	assert.Error(t, err)
	assert.Nil(t, r.GetNode())

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Equal(t, serverError.Message(), mockerr.Error())
}

func TestSdkInspectIdNotPassed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup client
	c := api.NewOpenStorageClusterClient(s.Conn())

	// Get info
	r, err := c.Inspect(context.Background(), &api.ClusterInspectRequest{})
	assert.Error(t, err)
	assert.Nil(t, r.GetNode())

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Node")
}
