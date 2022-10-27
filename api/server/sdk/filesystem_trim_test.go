package sdk

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
)

func TestSdkFilesystemTrimStartSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testMountPath := "/var/lib/osd/test"
	testStatus := api.FilesystemTrim_FS_TRIM_STARTED
	testMessage := "Test Message"
	req := &api.SdkFilesystemTrimStartRequest{
		VolumeId:  testVolumeId,
		MountPath: testMountPath,
	}

	testMockResp := &api.SdkFilesystemTrimStartResponse{
		Status:  testStatus,
		Message: testMessage,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemTrimStart(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemTrimClient(s.Conn())

	// Get info
	resp, err := c.Start(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Message, testMockResp.Message)
}

func TestSdkFilesystemTrimGetStatus(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testMountPath := "/var/lib/osd/test"
	testStatus := api.FilesystemTrim_FS_TRIM_INPROGRESS
	testMessage := "Test Message : FStrim in progress"
	req := &api.SdkFilesystemTrimStatusRequest{
		VolumeId:  testVolumeId,
		MountPath: testMountPath,
	}

	testMockResp := &api.SdkFilesystemTrimStatusResponse{
		Status:  testStatus,
		Message: testMessage,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemTrimStatus(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemTrimClient(s.Conn())

	// Get info
	resp, err := c.Status(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Message, testMockResp.Message)
}

func TestSdkFilesystemTrimStop(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testMountPath := "/var/lib/osd/test"
	req := &api.SdkFilesystemTrimStopRequest{
		VolumeId:  testVolumeId,
		MountPath: testMountPath,
	}

	testMockResp := &api.SdkFilesystemTrimStopResponse{}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemTrimStop(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemTrimClient(s.Conn())

	// Get info
	_, err := c.Stop(context.Background(), req)
	assert.NoError(t, err)
}
