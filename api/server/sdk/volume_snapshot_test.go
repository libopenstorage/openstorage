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

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
)

func TestSdkVolumeSnapshotCreateBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkVolumeSnapshotCreateRequest{}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.SnapshotCreate(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")
}

func TestSdkVolumeSnapshotCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	volid := "volid"
	snapid := "snapid"
	req := &api.SdkVolumeSnapshotCreateRequest{
		VolumeId: volid,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		Snapshot(req.GetVolumeId(), true, &api.VolumeLocator{}).
		Return(snapid, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.SnapshotCreate(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, r.GetSnapshotId(), snapid)
}

func TestSdkVolumeSnapshotRestoreBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkVolumeSnapshotRestoreRequest{}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.SnapshotRestore(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")

	// Now only provide the volume id
	req = &api.SdkVolumeSnapshotRestoreRequest{
		VolumeId: "volid",
	}

	// Setup client
	c = api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err = c.SnapshotRestore(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "snapshot id")
}

func TestSdkVolumeSnapshotRestore(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	volid := "volid"
	snapid := "snapid"
	req := &api.SdkVolumeSnapshotRestoreRequest{
		VolumeId:   volid,
		SnapshotId: snapid,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		Restore(volid, snapid).
		Return(nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	_, err := c.SnapshotRestore(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkVolumeSnapshotEnumerateBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkVolumeSnapshotEnumerateRequest{}

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.SnapshotEnumerate(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, r)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")
}

func TestSdkVolumeSnapshotEnumerate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	volid := "volid"
	snapid := "snapid"
	req := &api.SdkVolumeSnapshotEnumerateRequest{
		VolumeId: volid,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		SnapEnumerate([]string{volid}, nil).
		Return([]*api.Volume{
			&api.Volume{
				Id: snapid,
			},
		}, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageVolumeClient(s.Conn())

	// Get info
	r, err := c.SnapshotEnumerate(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetVolumeSnapshotIds())
	assert.Len(t, r.GetVolumeSnapshotIds(), 1)
	assert.Equal(t, r.GetVolumeSnapshotIds()[0], snapid)
}
