package sdk

import (
	"context"
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
)

func TestObjectstoreInspectSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	objResp := &api.ObjectstoreInfo{
		Uuid:     "test-uuid",
		VolumeId: "test-vol-id",
		Enabled:  false,
	}

	s.MockCluster().
		EXPECT().
		ObjectStoreInspect(objResp.Uuid).
		Return(objResp, nil)

	// Setup client
	c := api.NewOpenStorageObjectstoreClient(s.Conn())

	// Get info
	resp, err := c.InspectObjectstore(context.Background(), &api.SdkObjectstoreInspectRequest{ObjectstoreId: objResp.Uuid})
	assert.NoError(t, err)
	assert.NotNil(t, resp.GetObjectstoreStatus())

	// Verify
	assert.Equal(t, resp.GetObjectstoreStatus().GetVolumeId(), objResp.VolumeId)
}

func TestObjectstoreInspectFailed(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	objResp := &api.ObjectstoreInfo{
		Uuid:     "test-uuid",
		VolumeId: "test-vol-id",
		Enabled:  false,
	}

	s.MockCluster().
		EXPECT().
		ObjectStoreInspect(objResp.Uuid).
		Return(nil, fmt.Errorf("some error"))

	// Setup client
	c := api.NewOpenStorageObjectstoreClient(s.Conn())

	// Get info
	resp, err := c.InspectObjectstore(context.Background(), &api.SdkObjectstoreInspectRequest{ObjectstoreId: objResp.Uuid})
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Verify
	assert.Contains(t, err.Error(), "some error")
}

func TestObjectstoreCreateSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	objResp := &api.ObjectstoreInfo{
		Uuid:     "test-uuid",
		VolumeId: "test-vol-id",
		Enabled:  false,
	}

	volName := "test-vol"

	s.MockCluster().
		EXPECT().
		ObjectStoreCreate(volName).
		Return(objResp, nil)

	// Setup client
	c := api.NewOpenStorageObjectstoreClient(s.Conn())

	// Get info
	resp, err := c.CreateObjectstore(context.Background(), &api.SdkObjectstoreCreateRequest{VolumeName: volName})
	assert.NoError(t, err)
	assert.NotNil(t, resp.GetObjectstoreStatus())

	// Verify
	assert.Equal(t, resp.GetObjectstoreStatus().GetVolumeId(), objResp.VolumeId)
}

func TestObjectstoreCreateFailed(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	volName := "test-vol"
	s.MockCluster().
		EXPECT().
		ObjectStoreCreate(volName).
		Return(nil, fmt.Errorf("some error"))

	// Setup client
	c := api.NewOpenStorageObjectstoreClient(s.Conn())

	// Get info
	resp, err := c.CreateObjectstore(context.Background(), &api.SdkObjectstoreCreateRequest{VolumeName: volName})
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Verify
	assert.Contains(t, err.Error(), "some error")
}
