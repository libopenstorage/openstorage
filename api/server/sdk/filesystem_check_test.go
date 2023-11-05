package sdk

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
)

func TestSdkFilesystemCheckCheckHealth(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_STARTED
	testMode := "check_health"
	testMessage := "Test Message"
	req := &api.SdkFilesystemCheckStartRequest{
		VolumeId: testVolumeId,
		Mode:     testMode,
	}

	testMockResp := &api.SdkFilesystemCheckStartResponse{
		Status:  testStatus,
		Message: testMessage,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckStart(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.Start(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Message, testMockResp.Message)
}

func TestSdkFilesystemCheckCheckHealthStatus(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_INPROGRESS
	testMessage := "Test Message : FSCheck in progress"
	testMode := "check_health"
	req := &api.SdkFilesystemCheckStatusRequest{
		VolumeId: testVolumeId,
	}

	testMockResp := &api.SdkFilesystemCheckStatusResponse{
		Status:  testStatus,
		Mode:    testMode,
		Message: testMessage,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckStatus(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.Status(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Mode, testMockResp.Mode)
	assert.Equal(t, resp.Message, testMockResp.Message)
}

func TestSdkFilesystemCheckFixAll(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_STARTED
	testMode := "fix_all"
	testMessage := "Test Message"
	req := &api.SdkFilesystemCheckStartRequest{
		VolumeId: testVolumeId,
		Mode:     testMode,
	}

	testMockResp := &api.SdkFilesystemCheckStartResponse{
		Status:  testStatus,
		Message: testMessage,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckStart(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.Start(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Message, testMockResp.Message)
}

func TestSdkFilesystemCheckFixAllStatus(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_INPROGRESS
	testMessage := "Test Message : FSCheck in progress"
	testMode := "fix_all"
	req := &api.SdkFilesystemCheckStatusRequest{
		VolumeId: testVolumeId,
	}

	testMockResp := &api.SdkFilesystemCheckStatusResponse{
		Status:  testStatus,
		Mode:    testMode,
		Message: testMessage,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckStatus(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.Status(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Mode, testMockResp.Mode)
	assert.Equal(t, resp.Message, testMockResp.Message)
}

func TestSdkFilesystemCheckFixSafe(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_STARTED
	testMode := "fix_safe"
	testMessage := "Test Message"
	req := &api.SdkFilesystemCheckStartRequest{
		VolumeId: testVolumeId,
		Mode:     testMode,
	}

	testMockResp := &api.SdkFilesystemCheckStartResponse{
		Status:  testStatus,
		Message: testMessage,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckStart(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.Start(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Message, testMockResp.Message)
}

func TestSdkFilesystemCheckFixSafeStatus(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testStatus := api.FilesystemCheck_FS_CHECK_INPROGRESS
	testMessage := "Test Message : FSCheck in progress"
	testMode := "fix_safe"
	req := &api.SdkFilesystemCheckStatusRequest{
		VolumeId: testVolumeId,
	}

	testMockResp := &api.SdkFilesystemCheckStatusResponse{
		Status:  testStatus,
		Mode:    testMode,
		Message: testMessage,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckStatus(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	resp, err := c.Status(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Status, testMockResp.Status)
	assert.Equal(t, resp.Mode, testMockResp.Mode)
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

	testMockResp := &api.SdkFilesystemCheckStopResponse{}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckStop(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)
	// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	_, err := c.Stop(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkFilesystemCheckListSnapshots(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testNodeId := "mynode"
	var testMap = make(map[string]*api.FilesystemCheckSnapInfo)
	testMap["1"] = &api.FilesystemCheckSnapInfo{
		VolumeSnapshotName: "s1",
	}
	req := &api.SdkFilesystemCheckListSnapshotsRequest{
		VolumeId: testVolumeId,
		NodeId:   testNodeId,
	}

	testMockResp := &api.SdkFilesystemCheckListSnapshotsResponse{
		Snapshots: testMap,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckListSnapshots(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

		// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	_, err := c.ListSnapshots(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkFilesystemCheckDeleteSnapshots(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	testVolumeId := "myvol"
	testNodeId := "mynode"
	req := &api.SdkFilesystemCheckDeleteSnapshotsRequest{
		VolumeId: testVolumeId,
		NodeId:   testNodeId,
	}

	testMockResp := &api.SdkFilesystemCheckDeleteSnapshotsResponse{}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckDeleteSnapshots(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

		// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	_, err := c.DeleteSnapshots(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkFilesystemCheckListVolumes(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	var testMap = make(map[string]*api.FilesystemCheckVolInfo)
	testMap["1"] = &api.FilesystemCheckVolInfo{
		VolumeName:   "myvol",
		HealthStatus: 1,
		FsStatusMsg:  "test",
	}

	testMockResp := &api.SdkFilesystemCheckListVolumesResponse{
		Volumes: testMap,
	}

	testNodeId := "mynode"
	req := &api.SdkFilesystemCheckListVolumesRequest{
		NodeId: testNodeId,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		FilesystemCheckListVolumes(gomock.Any()).
		Return(testMockResp, nil).
		Times(1)

		// Setup client
	c := api.NewOpenStorageFilesystemCheckClient(s.Conn())

	// Get info
	_, err := c.ListVolumes(context.Background(), req)
	assert.NoError(t, err)
}
