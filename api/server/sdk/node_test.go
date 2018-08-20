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

	"github.com/libopenstorage/openstorage/api"
)

func TestSdkNodeEnumerateNoNodes(t *testing.T) {

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
	c := api.NewOpenStorageNodeClient(s.Conn())

	// Get info
	r, err := c.Enumerate(context.Background(), &api.SdkNodeEnumerateRequest{})
	assert.NoError(t, err)
	assert.Nil(t, r.GetNodeIds())
}

func TestSdkNodeEnumerate(t *testing.T) {

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
	c := api.NewOpenStorageNodeClient(s.Conn())

	// Get info
	r, err := c.Enumerate(context.Background(), &api.SdkNodeEnumerateRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetNodeIds())
	assert.Len(t, r.GetNodeIds(), 1)
	assert.Equal(t, r.GetNodeIds()[0], "nodeid")
}

func TestSdkNodeEnumerateFail(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	mockerr := fmt.Errorf("MOCK")
	s.MockCluster().EXPECT().Enumerate().Return(api.Cluster{}, mockerr).Times(1)

	// Setup client
	c := api.NewOpenStorageNodeClient(s.Conn())

	// Get info
	_, err := c.Enumerate(context.Background(), &api.SdkNodeEnumerateRequest{})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Equal(t, serverError.Message(), mockerr.Error())
}

func TestSdkNodeInspect(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	nodeid := "nodeid"
	node := api.Node{
		Id:                nodeid,
		SchedulerNodeName: "nodename",
		Cpu:               1.414,
		MemTotal:          112,
		MemUsed:           41,
		MemFree:           93,
		Avgload:           834,
		Status:            api.Status_STATUS_MAX,
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
	c := api.NewOpenStorageNodeClient(s.Conn())

	// Get info
	r, err := c.Inspect(context.Background(), &api.SdkNodeInspectRequest{
		NodeId: nodeid,
	})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetNode())

	// Verify
	rn := r.GetNode()
	assert.Equal(t, rn.GetId(), node.Id)
	assert.Equal(t, rn.GetSchedulerNodeName(), node.SchedulerNodeName)
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
	c := api.NewOpenStorageNodeClient(s.Conn())

	// Get info
	_, err := c.Inspect(context.Background(), &api.SdkNodeInspectRequest{
		NodeId: "mynode",
	})
	assert.Error(t, err)
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
	c := api.NewOpenStorageNodeClient(s.Conn())

	// Get info
	_, err := c.Inspect(context.Background(), &api.SdkNodeInspectRequest{})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Node")
}

func TestSdkNodeInspectCurrent(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	nodeid := "nodeid"

	// Create response
	node := api.Node{
		Id:                nodeid,
		SchedulerNodeName: "nodename",
		Cpu:               1.414,
		MemTotal:          112,
		MemUsed:           41,
		MemFree:           93,
		Avgload:           834,
		Status:            api.Status_STATUS_MAX,
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

	cluster := api.Cluster{
		Id:     "someid",
		NodeId: nodeid,
		Status: api.Status_STATUS_NOT_IN_QUORUM,
		Nodes:  []api.Node{node},
	}

	s.MockCluster().EXPECT().Enumerate().Return(cluster, nil).Times(1)
	s.MockCluster().EXPECT().Inspect(nodeid).Return(node, nil).Times(1)

	// Setup client
	c := api.NewOpenStorageNodeClient(s.Conn())

	// Get info
	r, err := c.InspectCurrent(context.Background(), &api.SdkNodeInspectCurrentRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetNode())

	// Verify
	rn := r.GetNode()
	assert.Equal(t, rn.GetId(), node.Id)
	assert.Equal(t, rn.GetSchedulerNodeName(), node.SchedulerNodeName)
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
