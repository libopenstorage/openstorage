package server

import (
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	client "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/stretchr/testify/require"
)

func TestMigrateStart(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	goodRequest := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateCluster,
		ClusterId: "clusterID",
		TargetId:  "goodVolumeID",
	}
	badRequest := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateCluster,
		ClusterId: "clusterID",
		TargetId:  "badVolumeID",
	}
	goodResponse := &api.CloudMigrateStartResponse{
		TaskId: "random-id",
	}
	testVolDriver.MockDriver().EXPECT().CloudMigrateStart(badRequest).Return(nil, fmt.Errorf("Volume not found")).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudMigrateStart(goodRequest).Return(goodResponse, nil).Times(1)

	// Start Migrate
	resp, err := client.VolumeDriver(cl).CloudMigrateStart(badRequest)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "Volume not found")
	resp, err = client.VolumeDriver(cl).CloudMigrateStart(goodRequest)
	require.NoError(t, err)
	require.Equal(t, goodResponse.TaskId, resp.TaskId, "Unexpected taskId in response")
}

func TestMigrateCancel(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	goodRequest := &api.CloudMigrateCancelRequest{
		TaskId: "goodTaskID",
	}
	badRequest := &api.CloudMigrateCancelRequest{
		TaskId: "badTaskID",
	}
	testVolDriver.MockDriver().EXPECT().CloudMigrateCancel(badRequest).Return(fmt.Errorf("TaskId not found")).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudMigrateCancel(goodRequest).Return(nil).Times(1)

	// Cancel Migrate
	err = client.VolumeDriver(cl).CloudMigrateCancel(badRequest)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TaskId not found")
	err = client.VolumeDriver(cl).CloudMigrateCancel(goodRequest)
	require.NoError(t, err)
}

func TestMigrateiStatus(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	emptyStatus := &api.CloudMigrateStatusResponse{}
	statusResponse := &api.CloudMigrateStatusResponse{
		Info: map[string]*api.CloudMigrateInfoList{
			"clusterId": &api.CloudMigrateInfoList{
				List: []*api.CloudMigrateInfo{
					&api.CloudMigrateInfo{
						TaskId:          "taskId",
						ClusterId:       "clusterId",
						LocalVolumeId:   "localVolumeId",
						LocalVolumeName: "localVolumeName",
						RemoteVolumeId:  "remoteVolumeName",
						CloudbackupId:   "cloudbackupId",
						CurrentStage:    api.CloudMigrate_Done,
						Status:          api.CloudMigrate_Complete,
					}}}},
	}

	testVolDriver.MockDriver().EXPECT().CloudMigrateStatus().Return(emptyStatus, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudMigrateStatus().Return(statusResponse, nil).Times(1)

	// Get Migrate status
	status, err := client.VolumeDriver(cl).CloudMigrateStatus()
	require.NoError(t, err)
	require.Equal(t, 0, len(status.Info))
	status, err = client.VolumeDriver(cl).CloudMigrateStatus()
	require.NoError(t, err)
	require.Equal(t, 1, len(status.Info))
	require.Equal(t, statusResponse, status)

}
