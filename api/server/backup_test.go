package server

import (
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	client "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientBackup(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupCreate(&api.CloudBackupCreateRequest{
		VolumeID:       "goodvol",
		CredentialUUID: "",
		Full:           false}).Return(&api.CloudBackupCreateResponse{Name: "good-backup-Name"}, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupCreate(&api.CloudBackupCreateRequest{
		VolumeID:       "badvol",
		CredentialUUID: "",
		Full:           false}).Return(nil, fmt.Errorf("Volume not found")).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupCreate(&api.CloudBackupCreateRequest{
		VolumeID:       "goodvol",
		CredentialUUID: "",
		Full:           false,
		Name:           "unique-id-from-app"}).Return(&api.CloudBackupCreateResponse{Name: "unique-id-from-app"}, nil).Times(1)

	// Create Backup
	createResponse, err := client.VolumeDriver(cl).
		CloudBackupCreate(&api.CloudBackupCreateRequest{
			VolumeID:       "goodvol",
			CredentialUUID: "",
			Full:           false})
	require.NoError(t, err)
	require.NotEqual(t, createResponse.Name, "")
	createResponse, err = client.VolumeDriver(cl).
		CloudBackupCreate(&api.CloudBackupCreateRequest{
			VolumeID:       "badvol",
			CredentialUUID: "",
			Full:           false})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Volume not found")
	createResponse, err = client.VolumeDriver(cl).
		CloudBackupCreate(&api.CloudBackupCreateRequest{
			VolumeID:       "goodvol",
			CredentialUUID: "",
			Full:           false,
			Name:           "unique-id-from-app"})
	require.NoError(t, err)
	require.Equal(t, createResponse.Name, "unique-id-from-app")

}

func TestClientGroupBackup(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupGroupCreate(&api.CloudBackupGroupCreateRequest{
		GroupID:        "goodvolgroup",
		Labels:         nil,
		CredentialUUID: "",
		Full:           false}).Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupGroupCreate(&api.CloudBackupGroupCreateRequest{
		GroupID:        "badvolgroup",
		Labels:         nil,
		CredentialUUID: "",
		Full:           false}).Return(fmt.Errorf("Volume group not found")).Times(1)

	// Create Backup
	err = client.VolumeDriver(cl).
		CloudBackupGroupCreate(&api.CloudBackupGroupCreateRequest{
			GroupID:        "goodvolgroup",
			Labels:         nil,
			CredentialUUID: "",
			Full:           false})
	require.NoError(t, err)
	err = client.VolumeDriver(cl).
		CloudBackupGroupCreate(&api.CloudBackupGroupCreateRequest{
			GroupID:        "badvolgroup",
			Labels:         nil,
			CredentialUUID: "",
			Full:           false})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Volume group not found")
}

func TestClientBackupRestore(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupRestore(&api.CloudBackupRestoreRequest{
		ID:             "goodBackupid",
		CredentialUUID: ""}).
		Return(&api.CloudBackupRestoreResponse{}, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupRestore(&api.CloudBackupRestoreRequest{
		ID:             "badbackupid",
		CredentialUUID: ""}).
		Return(nil, fmt.Errorf("Backup not found")).Times(1)

	// Invoke restore
	_, err = client.VolumeDriver(cl).
		CloudBackupRestore(&api.CloudBackupRestoreRequest{
			ID:             "goodBackupid",
			CredentialUUID: ""})
	require.NoError(t, err)
	_, err = client.VolumeDriver(cl).
		CloudBackupRestore(&api.CloudBackupRestoreRequest{
			ID:             "badbackupid",
			CredentialUUID: ""})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Backup not found")
}

func TestClientBackupRestoreGetNodeIdFromIp(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	// Create a new global test cluster
	tc := newTestCluster(t)
	defer tc.Finish()

	// Mock cluster
	nodeIp := "192.168.1.1"
	nodeId := "nodeid"
	tc.MockCluster().
		EXPECT().
		GetNodeIdFromIp(nodeIp).
		Return(nodeId, nil).Times(1)

	testVolDriver.MockDriver().EXPECT().CloudBackupRestore(&api.CloudBackupRestoreRequest{
		ID:             "goodBackupid",
		CredentialUUID: "",
		NodeID:         "nodeid"}).
		Return(&api.CloudBackupRestoreResponse{Name: "good-back-taskid"}, nil)

	// Invoke restore with IP Success
	_, err = client.VolumeDriver(cl).
		CloudBackupRestore(&api.CloudBackupRestoreRequest{
			ID:             "goodBackupid",
			CredentialUUID: "",
			NodeID:         nodeIp})
	require.NoError(t, err)

	// Mock cluster
	tc.MockCluster().
		EXPECT().
		GetNodeIdFromIp(nodeIp).
		Return(nodeIp, fmt.Errorf("Failed to locate IP in this cluster."))

	// Invoke restore with IP Failure
	_, err = client.VolumeDriver(cl).
		CloudBackupRestore(&api.CloudBackupRestoreRequest{
			ID:             "goodBackupid",
			CredentialUUID: "",
			NodeID:         nodeIp})

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to locate IP in this cluster.")
}

func TestClientBackupDelete(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	goodInput := &api.CloudBackupDeleteRequest{}
	goodInput.ID = "goodID"
	goodInput.CredentialUUID = ""

	badInput := &api.CloudBackupDeleteRequest{}
	badInput.ID = "badID"
	badInput.CredentialUUID = ""
	testVolDriver.MockDriver().EXPECT().CloudBackupDelete(goodInput).
		Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupDelete(badInput).
		Return(fmt.Errorf("BackupID not found")).Times(1)

	// Invoke Delete
	err = client.VolumeDriver(cl).CloudBackupDelete(goodInput)
	require.NoError(t, err)
	err = client.VolumeDriver(cl).CloudBackupDelete(badInput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "BackupID not found")
}

func TestClientBackupDeleteAll(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	goodInput := &api.CloudBackupDeleteAllRequest{}
	goodInput.SrcVolumeID = "goodsrc"
	goodInput.CredentialUUID = ""

	badInput := &api.CloudBackupDeleteAllRequest{}
	badInput.SrcVolumeID = "badsrc"
	badInput.CredentialUUID = ""
	testVolDriver.MockDriver().EXPECT().CloudBackupDeleteAll(goodInput).
		Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupDeleteAll(badInput).
		Return(fmt.Errorf("Src volume not found")).Times(1)

	// Invoke DeleteAll
	err = client.VolumeDriver(cl).CloudBackupDeleteAll(goodInput)
	require.NoError(t, err)
	err = client.VolumeDriver(cl).CloudBackupDeleteAll(badInput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Src volume not found")
}

func TestClientBackupEnumerate(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	goodInput := &api.CloudBackupEnumerateRequest{}
	goodInput.CredentialUUID = "goodCred"
	badInput := &api.CloudBackupEnumerateRequest{}
	badInput.CredentialUUID = "badCred"

	testVolDriver.MockDriver().EXPECT().CloudBackupEnumerate(goodInput).
		Return(&api.CloudBackupEnumerateResponse{}, nil).Times(1)

	testVolDriver.MockDriver().EXPECT().CloudBackupEnumerate(badInput).
		Return(nil, fmt.Errorf("Credential invalid")).Times(1)

	// Invoke Enumerate
	response, err := client.VolumeDriver(cl).CloudBackupEnumerate(goodInput)
	require.NoError(t, err)
	require.NotNil(t, response)

	response, err = client.VolumeDriver(cl).CloudBackupEnumerate(badInput)
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "Credential invalid")
}

func TestClientBackupStatus(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: "goodsrc"}).
		Return(&api.CloudBackupStatusResponse{}, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: "badsrc"}).
		Return(&api.CloudBackupStatusResponse{}, fmt.Errorf("Invalid source volume")).Times(1)

	// Invoke Status
	_, err = client.VolumeDriver(cl).
		CloudBackupStatus(&api.CloudBackupStatusRequest{
			SrcVolumeID: "goodsrc"})
	require.NoError(t, err)
	_, err = client.VolumeDriver(cl).
		CloudBackupStatus(&api.CloudBackupStatusRequest{
			SrcVolumeID: "badsrc"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid source volume")
}

func TestClientBackupCatalog(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupCatalog(&api.CloudBackupCatalogRequest{
		ID:             "goodcloudbackup",
		CredentialUUID: ""}).
		Return(&api.CloudBackupCatalogResponse{}, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupCatalog(&api.CloudBackupCatalogRequest{
		ID:             "badcloudbackup",
		CredentialUUID: ""}).
		Return(&api.CloudBackupCatalogResponse{}, fmt.Errorf("Failed to get catalog")).Times(1)

	// Invoke Catalog
	_, err = client.VolumeDriver(cl).
		CloudBackupCatalog(&api.CloudBackupCatalogRequest{
			ID:             "goodcloudbackup",
			CredentialUUID: ""})
	require.NoError(t, err)
	_, err = client.VolumeDriver(cl).
		CloudBackupCatalog(&api.CloudBackupCatalogRequest{
			ID:             "badcloudbackup",
			CredentialUUID: ""})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Failed to get catalog")
}

func TestClientBackupHistory(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupHistory(&api.CloudBackupHistoryRequest{
		SrcVolumeID: "goodsrc"}).
		Return(&api.CloudBackupHistoryResponse{}, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupHistory(&api.CloudBackupHistoryRequest{
		SrcVolumeID: "badsrc"}).
		Return(&api.CloudBackupHistoryResponse{}, fmt.Errorf("Failed to get history")).Times(1)

	// Invoke History
	_, err = client.VolumeDriver(cl).
		CloudBackupHistory(&api.CloudBackupHistoryRequest{
			SrcVolumeID: "goodsrc"})
	require.NoError(t, err)
	_, err = client.VolumeDriver(cl).
		CloudBackupHistory(&api.CloudBackupHistoryRequest{
			SrcVolumeID: "badsrc"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Failed to get history")
}

func TestClientBackupStateChange(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
		Name:           "good-task",
		RequestedState: "pause"}).
		Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
		Name:           "",
		RequestedState: ""}).
		Return(fmt.Errorf("Failed to change state")).Times(1)

	//Invoke StateChange
	err = client.VolumeDriver(cl).
		CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
			Name:           "good-task",
			RequestedState: "pause"})
	require.NoError(t, err)
	err = client.VolumeDriver(cl).
		CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
			Name:           "",
			RequestedState: ""})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Failed to change state")
}

func TestClientBackupSchedCreate(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	goodRequest := api.CloudBackupSchedCreateRequest{}
	goodRequest.SrcVolumeID = "goodsrc"
	goodRequest.CredentialUUID = ""
	goodRequest.Full = false
	goodRequest.Schedule = "daily@10:00"
	testVolDriver.MockDriver().EXPECT().CloudBackupSchedCreate(&goodRequest).
		Return(&api.CloudBackupSchedCreateResponse{}, nil).Times(1)
	badRequest := api.CloudBackupSchedCreateRequest{}
	badRequest.SrcVolumeID = "badsrc"
	badRequest.CredentialUUID = ""
	badRequest.Schedule = ""
	testVolDriver.MockDriver().EXPECT().CloudBackupSchedCreate(&badRequest).
		Return(&api.CloudBackupSchedCreateResponse{}, fmt.Errorf("Invalid src volume or schedule")).Times(1)

	// Invoke Schedule Create
	_, err = client.VolumeDriver(cl).CloudBackupSchedCreate(&goodRequest)
	require.NoError(t, err)
	_, err = client.VolumeDriver(cl).CloudBackupSchedCreate(&badRequest)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid src volume or schedule")
}

func TestClientBackupGroupSchedCreate(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	goodRequest := api.CloudBackupGroupSchedCreateRequest{}
	goodRequest.GroupID = "goodvolgroup"
	goodRequest.CredentialUUID = ""
	goodRequest.Schedule = "daily@10:00"
	testVolDriver.MockDriver().EXPECT().CloudBackupGroupSchedCreate(&goodRequest).
		Return(&api.CloudBackupSchedCreateResponse{}, nil).Times(1)
	badRequest := api.CloudBackupGroupSchedCreateRequest{}
	badRequest.GroupID = "badvolgroup"
	badRequest.CredentialUUID = ""
	badRequest.Schedule = ""
	testVolDriver.MockDriver().EXPECT().CloudBackupGroupSchedCreate(&badRequest).
		Return(&api.CloudBackupSchedCreateResponse{}, fmt.Errorf("Invalid volume group or schedule")).Times(1)

	// Invoke Schedule Create
	_, err = client.VolumeDriver(cl).CloudBackupGroupSchedCreate(&goodRequest)
	require.NoError(t, err)
	_, err = client.VolumeDriver(cl).CloudBackupGroupSchedCreate(&badRequest)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid volume group or schedule")
}

func TestClientBackupSchedDelete(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupSchedDelete(&api.CloudBackupSchedDeleteRequest{
		UUID: "goodscheduuid"}).
		Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupSchedDelete(&api.CloudBackupSchedDeleteRequest{
		UUID: "badscheduuid"}).
		Return(fmt.Errorf("Invalid Schedule UUID")).Times(1)

	// Invoke Schedule Delete
	err = client.VolumeDriver(cl).
		CloudBackupSchedDelete(&api.CloudBackupSchedDeleteRequest{
			UUID: "goodscheduuid"})
	require.NoError(t, err)
	err = client.VolumeDriver(cl).
		CloudBackupSchedDelete(&api.CloudBackupSchedDeleteRequest{
			UUID: "badscheduuid"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid Schedule UUID")
}

func TestClientBackupSchedEnumerate(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupSchedEnumerate().
		Return(&api.CloudBackupSchedEnumerateResponse{}, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupSchedEnumerate().
		Return(&api.CloudBackupSchedEnumerateResponse{}, fmt.Errorf("Failed to Enumerate cloudsnap Schedules")).Times(1)

	// Invoke Schedule Enumerate
	_, err = client.VolumeDriver(cl).CloudBackupSchedEnumerate()
	require.NoError(t, err)
	_, err = client.VolumeDriver(cl).CloudBackupSchedEnumerate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Failed to Enumerate")
}

func TestCloudBackupWait(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	id := "testId"
	notStartedStatus := &api.CloudBackupStatusResponse{
		Statuses: map[string]api.CloudBackupStatus{
			id: api.CloudBackupStatus{
				ID:     id,
				OpType: api.CloudBackupOp,
				Status: api.CloudBackupStatusNotStarted,
			},
		},
	}
	activeStatus := &api.CloudBackupStatusResponse{
		Statuses: map[string]api.CloudBackupStatus{
			id: api.CloudBackupStatus{
				ID:     id,
				OpType: api.CloudBackupOp,
				Status: api.CloudBackupStatusActive,
			},
		},
	}
	doneStatus := &api.CloudBackupStatusResponse{
		Statuses: map[string]api.CloudBackupStatus{
			id: api.CloudBackupStatus{
				ID:     id,
				OpType: api.CloudBackupOp,
				Status: api.CloudBackupStatusDone,
			},
		},
	}
	failedStatus := &api.CloudBackupStatusResponse{
		Statuses: map[string]api.CloudBackupStatus{
			id: api.CloudBackupStatus{
				ID:     id,
				OpType: api.CloudBackupOp,
				Status: api.CloudBackupStatusFailed,
			},
		},
	}

	testVolDriver.MockDriver().EXPECT().CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: id}).
		Return(notStartedStatus, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: id}).
		Return(activeStatus, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: id}).
		Return(doneStatus, nil).Times(1)

	err = volume.CloudBackupWaitForCompletion(client.VolumeDriver(cl), id, api.CloudBackupOp)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: id}).
		Return(notStartedStatus, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: id}).
		Return(activeStatus, nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: id}).
		Return(failedStatus, nil).Times(1)

	err = volume.CloudBackupWaitForCompletion(client.VolumeDriver(cl), id, api.CloudBackupOp)
	require.Error(t, err)
}
