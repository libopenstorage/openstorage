package sdk

import (
	"context"
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestObjectstoreCreateFailedBadArgument(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	volName := ""

	// Setup client
	c := api.NewOpenStorageObjectstoreClient(s.Conn())

	// Get info
	_, err := c.CreateObjectstore(context.Background(), &api.SdkObjectstoreCreateRequest{VolumeName: volName})
	assert.Error(t, err)

	// Verify
	assert.Contains(t, err.Error(), "Must provide volume name")
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must provide volume name")
}

func TestObjectstoreUpdateSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	req := &api.SdkObjectstoreUpdateRequest{
		ObjectstoreId: "test-obj-uuid",
		Enable:        true,
	}
	s.MockCluster().
		EXPECT().
		ObjectStoreUpdate(req.ObjectstoreId, req.Enable).
		Return(nil)

	// Setup client
	c := api.NewOpenStorageObjectstoreClient(s.Conn())

	// Update objectstore state
	_, err := c.UpdateObjectstore(context.Background(), req)

	// Check result
	assert.NoError(t, err)
	assert.Nil(t, err)
}

func TestObjectstoreUpdateFailed(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkObjectstoreUpdateRequest{
		ObjectstoreId: "test-obj-uuid",
	}
	s.MockCluster().
		EXPECT().
		ObjectStoreUpdate(req.ObjectstoreId, req.GetEnable()).
		Return(fmt.Errorf("update error"))

	// Setup client
	c := api.NewOpenStorageObjectstoreClient(s.Conn())

	// Update ObjectstoreState
	resp, err := c.UpdateObjectstore(context.Background(), req)

	// Check response
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "update error")
}

func TestObjectstoreDeleteSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	objID := "test-obj-uuid1"

	s.MockCluster().
		EXPECT().
		ObjectStoreDelete(objID).
		Return(nil)

	// Setup client
	c := api.NewOpenStorageObjectstoreClient(s.Conn())

	// Delete object store
	_, err := c.DeleteObjectstore(
		context.Background(),
		&api.SdkObjectstoreDeleteRequest{ObjectstoreId: objID})

	assert.NoError(t, err)
}

func TestObjectstoreDeleteFailed(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	objID := "test-obj-uuid1"
	s.MockCluster().
		EXPECT().
		ObjectStoreDelete(objID).
		Return(fmt.Errorf("delete error"))

	// Setup objectstore client
	c := api.NewOpenStorageObjectstoreClient(s.Conn())

	// Delete Object store
	resp, err := c.DeleteObjectstore(
		context.Background(),
		&api.SdkObjectstoreDeleteRequest{ObjectstoreId: objID})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "delete error")
}
