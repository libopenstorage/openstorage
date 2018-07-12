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
	"testing"

	"github.com/golang/protobuf/ptypes"

	"github.com/stretchr/testify/assert"

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

func TestSdkClusterInspectCurrent(t *testing.T) {

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
	r, err := c.InspectCurrent(context.Background(), &api.SdkClusterInspectCurrentRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetCluster())
	assert.Equal(t, cluster.Id, r.GetCluster().GetId())
	assert.Equal(t, cluster.Status, r.GetCluster().GetStatus())
}

func TestSdkAlertEnumerate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup client
	c := api.NewOpenStorageClusterClient(s.Conn())

	// Create request
	req := &api.SdkClusterAlertEnumerateRequest{
		TimeStart: ptypes.TimestampNow(),
		TimeEnd:   ptypes.TimestampNow(),
		Resource:  api.ResourceType_RESOURCE_TYPE_DRIVE,
	}

	// Mock output
	out := &api.Alerts{
		Alert: []*api.Alert{
			&api.Alert{
				Id: 1234,
			},
			&api.Alert{
				Id: 6789,
			},
		},
	}

	// Mock
	ts, err := ptypes.Timestamp(req.TimeStart)
	assert.NoError(t, err)
	te, err := ptypes.Timestamp(req.TimeEnd)
	assert.NoError(t, err)
	s.MockCluster().
		EXPECT().
		EnumerateAlerts(ts, te, api.ResourceType_RESOURCE_TYPE_DRIVE).
		Return(out, nil).
		Times(1)

	// Get info
	r, err := c.AlertEnumerate(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetAlerts())
	assert.Len(t, r.GetAlerts(), 2)
	assert.Equal(t, r.GetAlerts()[0].Id, out.Alert[0].Id)
	assert.Equal(t, r.GetAlerts()[1].Id, out.Alert[1].Id)
}

func TestSdkAlertClear(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup client
	c := api.NewOpenStorageClusterClient(s.Conn())

	// Create request
	req := &api.SdkClusterAlertClearRequest{
		AlertId:  1234,
		Resource: api.ResourceType_RESOURCE_TYPE_DRIVE,
	}

	// Mock
	s.MockCluster().
		EXPECT().
		ClearAlert(req.Resource, req.AlertId).
		Return(nil).
		Times(1)

	// Get info
	_, err := c.AlertClear(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkAlertDelete(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup client
	c := api.NewOpenStorageClusterClient(s.Conn())

	// Create request
	req := &api.SdkClusterAlertDeleteRequest{
		AlertId:  1234,
		Resource: api.ResourceType_RESOURCE_TYPE_DRIVE,
	}

	// Mock
	s.MockCluster().
		EXPECT().
		EraseAlert(req.Resource, req.AlertId).
		Return(nil).
		Times(1)

	// Get info
	_, err := c.AlertDelete(context.Background(), req)
	assert.NoError(t, err)
}
