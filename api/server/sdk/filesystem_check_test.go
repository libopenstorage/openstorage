package sdk

import (
	"context"
	"testing"


	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"

)

func TestSdkFilesystemCheckCheckHealth(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_STARTED
	testMessage := "Test Message"
	req := &api.SdkFilesystemCheckCheckHealthRequest{
		VolumeId: testVolumeId,
	}

	testMockResp := &api.SdkFilesystemCheckCheckHealthResponse {
		Status: testStatus,
		Message : testMessage,
	}

	// Create response
		s.MockDriver().
			EXPECT().
			FilesystemCheckCheckHealth(&api.SdkFilesystemCheckCheckHealthRequest{
				VolumeId: testVolumeId,
			}).
			Return(testMockResp, nil).
			Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.CheckHealth(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Message, testMockResp.Message)
}



func TestSdkFilesystemCheckCheckHealthGetStatus(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_CHECK_HEALTH_INPROGRESS
	testMessage := "Test Message : FStrim in progress"
	req := &api.SdkFilesystemCheckCheckHealthGetStatusRequest{
		VolumeId: testVolumeId,
	}

	testMockResp := &api.SdkFilesystemCheckCheckHealthGetStatusResponse {
		Status: testStatus,
		Message : testMessage,
	}

	// Create response
		s.MockDriver().
			EXPECT().
			FilesystemCheckCheckHealthGetStatus(&api.SdkFilesystemCheckCheckHealthGetStatusRequest{
				VolumeId: testVolumeId,
			}).
			Return(testMockResp, nil).
			Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.CheckHealthGetStatus(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Message, testMockResp.Message)
}


func TestSdkFilesystemCheckFixAll(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_STARTED
	testMessage := "Test Message"
	req := &api.SdkFilesystemCheckFixAllRequest{
		VolumeId: testVolumeId,
	}

	testMockResp := &api.SdkFilesystemCheckFixAllResponse {
		Status: testStatus,
		Message : testMessage,
	}

	// Create response
		s.MockDriver().
			EXPECT().
			FilesystemCheckFixAll(&api.SdkFilesystemCheckFixAllRequest{
				VolumeId: testVolumeId,
			}).
			Return(testMockResp, nil).
			Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.FixAll(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Message, testMockResp.Message)
}



func TestSdkFilesystemCheckFixAllGetStatus(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_CHECK_HEALTH_INPROGRESS
	testMessage := "Test Message : FStrim in progress"
	req := &api.SdkFilesystemCheckFixAllGetStatusRequest{
		VolumeId: testVolumeId,
	}

	testMockResp := &api.SdkFilesystemCheckFixAllGetStatusResponse {
		Status: testStatus,
		Message : testMessage,
	}

	// Create response
		s.MockDriver().
			EXPECT().
			FilesystemCheckFixAllGetStatus(&api.SdkFilesystemCheckFixAllGetStatusRequest{
				VolumeId: testVolumeId,
			}).
			Return(testMockResp, nil).
			Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.FixAllGetStatus(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Message, testMockResp.Message)
}

func TestSdkFilesystemCheckStop(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	req := &api.SdkFilesystemCheckStopRequest{
		VolumeId: testVolumeId,
	}

	testMockResp := &api.SdkFilesystemCheckStopResponse {
	}

	// Create response
		s.MockDriver().
			EXPECT().
			FilesystemCheckStop(&api.SdkFilesystemCheckStopRequest{
				VolumeId: testVolumeId,
			}).
			Return(testMockResp, nil).
			Times(1)
	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	_, err := c.Stop(context.Background(), req)
	assert.NoError(t, err)
}

