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
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupExpectedCredentialsPassing(s *testServer, credid string) {
	enumAzure := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account",
		api.OptCredAzureAccountKey:  "test-azure-account",
		api.OptCredProxy:            "false",
	}
	creds := map[string]interface{}{
		credid: enumAzure,
	}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(creds, nil)
}

func setupExpectedCredentialsNotPassing(s *testServer) {
	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(nil, nil)
}

func setupExpectedCredentialsNotPassingMoreThanOne(s *testServer) {
	enumAzure1 := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account-1",
		api.OptCredAzureAccountKey:  "test-azure-account-1",
	}
	enumAzure2 := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account-2",
		api.OptCredAzureAccountKey:  "test-azure-account-2",
	}
	creds := map[string]interface{}{
		"uuid-1": enumAzure1,
		"uuid-2": enumAzure2,
	}
	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(creds, nil)
}

func TestSdkCloudBackupCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	uuid := "uuid"
	taskId := "backup-task"
	full := false
	labels := map[string]string{"foo": "bar"}
	req := &api.SdkCloudBackupCreateRequest{
		VolumeId:     id,
		CredentialId: uuid,
		Full:         full,
		TaskId:       taskId,
		Labels:       labels,
		DeleteLocal:  true,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupCreate(&api.CloudBackupCreateRequest{
			VolumeID:       id,
			CredentialUUID: uuid,
			Full:           false,
			Name:           taskId,
			Labels:         labels,
			DeleteLocal:    true,
		}).
		Return(&api.CloudBackupCreateResponse{Name: "good-backup-name"}, nil).
		Times(2)
	setupExpectedCredentialsPassing(s, uuid)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	_, err := c.Create(context.Background(), req)
	assert.NoError(t, err)

	setupExpectedCredentialsPassing(s, uuid)
	// default credentials
	req.CredentialId = ""
	_, err = c.Create(context.Background(), req)
	assert.NoError(t, err)

}

func TestSdkCloudBackupCreateBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupCreateRequest{}
	req.TaskId = "backup-task"

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// volume id missing
	_, err := c.Create(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")

	// cred id missing
	req.VolumeId = "myvol"
	setupExpectedCredentialsNotPassing(s)

	_, err = c.Create(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "No configured credentials found")

	// more than 1 default creds
	setupExpectedCredentialsNotPassingMoreThanOne(s)
	_, err = c.Create(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "credential name or uuid to use")
}

func TestSdkCloudRestoreCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	backupid := "backupid"
	id := "myvol"
	taskId := "restore-task"
	uuid := "uuid"
	req := &api.SdkCloudBackupRestoreRequest{
		BackupId:     backupid,
		CredentialId: uuid,
		TaskId:       taskId,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupRestore(&api.CloudBackupRestoreRequest{
			ID:             backupid,
			CredentialUUID: uuid,
			Name:           taskId,
		}).
		Return(&api.CloudBackupRestoreResponse{
			RestoreVolumeID: id,
			Name:            taskId,
		}, nil).
		Times(1)
	setupExpectedCredentialsPassing(s, uuid)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.Restore(context.Background(), req)
	assert.Equal(t, r.GetRestoreVolumeId(), id)
	assert.NoError(t, err)
}

func TestSdkCloudRestoreCreateErrorCheck(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	backupid := "backupid"
	taskId := "restore-task"
	uuid := "uuid"
	req := &api.SdkCloudBackupRestoreRequest{
		BackupId:     backupid,
		CredentialId: uuid,
		TaskId:       taskId,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupRestore(&api.CloudBackupRestoreRequest{
			ID:             backupid,
			CredentialUUID: uuid,
			Name:           taskId,
		}).
		Return(nil, volume.ErrExist).
		Times(1)
	setupExpectedCredentialsPassing(s, uuid)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	_, err := c.Restore(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.AlreadyExists)
	assert.Contains(t, serverError.Message(), "Restore task with this name already exists")
}

func TestSdkCloudBackupRestoreBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupRestoreRequest{}
	req.TaskId = "restore-task"
	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// backup id missing
	_, err := c.Restore(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "backup id")

	// Missing credential uuid
	setupExpectedCredentialsNotPassing(s)
	req.BackupId = "id"
	_, err = c.Restore(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "No configured credentials found")
}

func TestSdkCloudDeleteCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	backupid := "backupid"
	uuid := "uuid"
	req := &api.SdkCloudBackupDeleteRequest{
		BackupId:     backupid,
		CredentialId: uuid,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupDelete(&api.CloudBackupDeleteRequest{
			ID:             backupid,
			CredentialUUID: uuid,
			Force:          false,
		}).
		Return(nil).
		Times(1)
	setupExpectedCredentialsPassing(s, uuid)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkCloudBackupDeleteBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupDeleteRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// backup id missing
	_, err := c.Delete(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "backup id")

	// Missing credential uuid
	req.BackupId = "id"
	setupExpectedCredentialsNotPassing(s)
	_, err = c.Delete(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "No configured credentials found")
}

func TestSdkCloudDeleteAllCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	uuid := "uuid"
	req := &api.SdkCloudBackupDeleteAllRequest{
		SrcVolumeId:  id,
		CredentialId: uuid,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupDeleteAll(&api.CloudBackupDeleteAllRequest{
			CloudBackupGenericRequest: api.CloudBackupGenericRequest{
				SrcVolumeID:    id,
				CredentialUUID: uuid,
			},
		}).
		Return(nil).
		Times(1)
	setupExpectedCredentialsPassing(s, uuid)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	_, err := c.DeleteAll(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkCloudBackupDeleteAllBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupDeleteAllRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// volume id missing
	_, err := c.DeleteAll(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")

	// Missing credential uuid
	setupExpectedCredentialsNotPassing(s)
	req.SrcVolumeId = "id"
	_, err = c.DeleteAll(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "No configured credentials found")
}

func TestSdkCloudBackupEnumerateWithFilters(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	uuid := "uuid"
	namespace := "test-ns"
	req := &api.SdkCloudBackupEnumerateWithFiltersRequest{
		SrcVolumeId:  id,
		CredentialId: uuid,
	}

	list := &api.CloudBackupEnumerateResponse{
		Backups: []api.CloudBackupInfo{
			{
				ID:            "one",
				SrcVolumeID:   "one:vol",
				SrcVolumeName: "one:volname",
				Timestamp:     time.Now(),
				Metadata: map[string]string{
					"hello": "world",
				},
				Status:      "Done",
				ClusterType: api.SdkCloudBackupClusterType_SdkCloudBackupClusterCurrent,
				Namespace:   namespace,
			},
			{
				ID:            "two",
				SrcVolumeID:   "two:vol",
				SrcVolumeName: "two:volname",
				Timestamp:     time.Now(),
				Metadata: map[string]string{
					"what a": "world",
				},
				Status:      "Failed",
				ClusterType: api.SdkCloudBackupClusterType_SdkCloudBackupClusterCurrent,
				Namespace:   namespace,
			},
		},
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
			CloudBackupGenericRequest: api.CloudBackupGenericRequest{
				SrcVolumeID:    id,
				CredentialUUID: uuid,
			},
		}).
		Return(list, nil).
		Times(1)
	setupExpectedCredentialsPassing(s, uuid)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.EnumerateWithFilters(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetBackups())
	assert.Len(t, r.GetBackups(), 2)

	for i, v := range list.Backups {
		assert.Equal(t, r.Backups[i].GetId(), v.ID)
		assert.Equal(t, r.Backups[i].GetSrcVolumeId(), v.SrcVolumeID)
		assert.Equal(t, r.Backups[i].GetSrcVolumeName(), v.SrcVolumeName)
		assert.Equal(t,
			r.Backups[i].GetStatus(),
			api.StringToSdkCloudBackupStatusType(v.Status))
		assert.True(t, reflect.DeepEqual(r.Backups[i].GetMetadata(), v.Metadata))
		ts, err := ptypes.TimestampProto(v.Timestamp)
		assert.NoError(t, err)
		assert.Equal(t, r.Backups[i].Timestamp, ts)
		assert.Equal(t, r.Backups[i].ClusterType, api.SdkCloudBackupClusterType_SdkCloudBackupClusterCurrent)
		assert.Equal(t, r.Backups[i].Namespace, namespace)
	}
}

func TestSdkCloudBackupEnumerateWithFiltersBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupEnumerateWithFiltersRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Missing credential uuid
	req.SrcVolumeId = "id"
	setupExpectedCredentialsNotPassing(s)
	_, err := c.EnumerateWithFilters(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "No configured credentials found")
}

func TestSdkCloudBackupEnumerateWithFiltersSingle(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	uuid := "uuid"
	backupId := "one"
	req := &api.SdkCloudBackupEnumerateWithFiltersRequest{
		CloudBackupId: backupId,
		CredentialId:  uuid,
	}
	list := &api.CloudBackupEnumerateResponse{
		Backups: []api.CloudBackupInfo{
			{
				ID:            backupId,
				SrcVolumeID:   "one:vol",
				SrcVolumeName: "one:volname",
				Timestamp:     time.Now(),
				Metadata: map[string]string{
					"hello": "world",
				},
				Status:      "Done",
				ClusterType: api.SdkCloudBackupClusterType_SdkCloudBackupClusterUnknown,
			},
		},
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
			CloudBackupGenericRequest: api.CloudBackupGenericRequest{
				CloudBackupID:  backupId,
				CredentialUUID: uuid,
			},
		}).
		Return(list, nil).
		Times(1)
	setupExpectedCredentialsPassing(s, uuid)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.EnumerateWithFilters(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetBackups())
	assert.Len(t, r.GetBackups(), 1)

	for i, v := range list.Backups {
		assert.Equal(t, r.Backups[i].GetId(), v.ID)
		assert.Equal(t, r.Backups[i].GetSrcVolumeId(), v.SrcVolumeID)
		assert.Equal(t, r.Backups[i].GetSrcVolumeName(), v.SrcVolumeName)
		assert.Equal(t,
			r.Backups[i].GetStatus(),
			api.StringToSdkCloudBackupStatusType(v.Status))
		assert.True(t, reflect.DeepEqual(r.Backups[i].GetMetadata(), v.Metadata))
		ts, err := ptypes.TimestampProto(v.Timestamp)
		assert.NoError(t, err)
		assert.Equal(t, r.Backups[i].Timestamp, ts)
	}
}

func TestSdkCloudBackupStatus(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	req := &api.SdkCloudBackupStatusRequest{
		VolumeId: id,
	}
	statuses := &api.CloudBackupStatusResponse{
		Statuses: map[string]api.CloudBackupStatus{
			"hello": api.CloudBackupStatus{
				ID:             "myid",
				OpType:         api.CloudBackupOp,
				Status:         api.CloudBackupStatusPaused,
				BytesDone:      123456,
				BytesTotal:     123456,
				EtaSeconds:     0,
				SrcVolumeID:    id,
				StartTime:      time.Now(),
				CompletedTime:  time.Now(),
				NodeID:         "mynode",
				CredentialUUID: "uuid",
			},
			"world": api.CloudBackupStatus{
				ID:             "another",
				OpType:         api.CloudRestoreOp,
				Status:         api.CloudBackupStatusDone,
				BytesDone:      97324,
				BytesTotal:     123456,
				EtaSeconds:     37,
				SrcVolumeID:    id,
				StartTime:      time.Now(),
				CompletedTime:  time.Now(),
				NodeID:         "myothernode",
				CredentialUUID: "uuid",
			},
		},
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupStatus(&api.CloudBackupStatusRequest{
			SrcVolumeID: id,
			Local:       false,
		}).
		Return(statuses, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.Status(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetStatuses())
	assert.Len(t, r.GetStatuses(), 2)

	// Verify
	for k, v := range r.GetStatuses() {
		status := statuses.Statuses[k]
		assert.Equal(t, v.GetBackupId(), status.ID)
		assert.Equal(t, v.GetBytesDone(), status.BytesDone)
		assert.Equal(t, v.GetBytesTotal(), status.BytesTotal)
		assert.Equal(t, v.GetEtaSeconds(), status.EtaSeconds)
		assert.Equal(t, v.GetNodeId(), status.NodeID)
		assert.Equal(t, v.GetOptype(), api.CloudBackupOpTypeToSdkCloudBackupOpType(status.OpType))
		assert.Equal(t, v.GetStatus(), api.CloudBackupStatusTypeToSdkCloudBackupStatusType(status.Status))
		assert.Equal(t, v.GetCredentialId(), status.CredentialUUID)

		ts, err := ptypes.TimestampProto(status.StartTime)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(v.GetStartTime(), ts))
		ts, err = ptypes.TimestampProto(status.CompletedTime)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(v.GetCompletedTime(), ts))

	}
}

func TestSdkCloudBackupCatalog(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "mybackup"
	creds := "creds"
	req := &api.SdkCloudBackupCatalogRequest{
		BackupId:     id,
		CredentialId: creds,
	}
	catalog := &api.CloudBackupCatalogResponse{
		Contents: []string{
			"one",
			"two",
			"three",
		},
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupCatalog(&api.CloudBackupCatalogRequest{
			ID:             id,
			CredentialUUID: creds,
		}).
		Return(catalog, nil).
		Times(1)
	setupExpectedCredentialsPassing(s, creds)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.Catalog(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetContents())
	assert.Len(t, r.GetContents(), 3)

	// Verify
	for i, v := range catalog.Contents {
		assert.Equal(t, r.GetContents()[i], v)
	}
}

func TestSdkCloudBackupCatalogBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupCatalogRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// backup id missing
	_, err := c.Catalog(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "backup id")

	// Missing credential uuid
	req.BackupId = "id"
	setupExpectedCredentialsNotPassing(s)
	_, err = c.Catalog(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "No configured credentials found")
}

func TestSdkCloudBackupHistory(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	req := &api.SdkCloudBackupHistoryRequest{
		SrcVolumeId: id,
	}
	history := &api.CloudBackupHistoryResponse{
		HistoryList: []api.CloudBackupHistoryItem{
			{
				Timestamp:   time.Now(),
				Status:      "Done",
				SrcVolumeID: id,
			},
			{
				Timestamp:   time.Now(),
				Status:      "Failed",
				SrcVolumeID: id,
			},
		},
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupHistory(&api.CloudBackupHistoryRequest{
			SrcVolumeID: id,
		}).
		Return(history, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.History(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetHistoryList())
	assert.Len(t, r.GetHistoryList(), 2)

	// Verify
	for i, v := range history.HistoryList {
		assert.Equal(t, r.GetHistoryList()[i].GetStatus(), api.StringToSdkCloudBackupStatusType(v.Status))
		assert.Equal(t, r.GetHistoryList()[i].GetSrcVolumeId(), v.SrcVolumeID)

		ts, err := ptypes.TimestampProto(v.Timestamp)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(r.GetHistoryList()[i].GetTimestamp(), ts))
	}
}

func TestSdkCloudBackupHistoryBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupHistoryRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// src volume id missing
	_, err := c.History(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")
}

func TestSdkCloudBackupStateChange(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	tests := []struct {
		internalrs string
		sdkrs      api.SdkCloudBackupRequestedState
	}{
		/*
			{
				api.CloudBackupRequestedStatePause,
				api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStatePause,
			},
			{
				api.CloudBackupRequestedStateResume,
				api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateResume,
			},
		*/
		{
			api.CloudBackupRequestedStateStop,
			api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateStop,
		},
	}
	id := "myvol"
	taskId := "myid"
	statuses := &api.CloudBackupStatusResponse{
		Statuses: map[string]api.CloudBackupStatus{
			"hello": api.CloudBackupStatus{
				ID:             taskId,
				OpType:         api.CloudBackupOp,
				Status:         api.CloudBackupStatusPaused,
				BytesDone:      123456,
				BytesTotal:     123456,
				EtaSeconds:     0,
				SrcVolumeID:    id,
				StartTime:      time.Now(),
				CompletedTime:  time.Now(),
				NodeID:         "mynode",
				CredentialUUID: "uuid",
			},
		},
	}

	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	for _, test := range tests {
		// Create response
		s.MockDriver().
			EXPECT().
			CloudBackupStatus(&api.CloudBackupStatusRequest{
				ID: taskId,
			}).
			Return(statuses, nil).
			Times(1)
		s.MockDriver().
			EXPECT().
			CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
				Name:           taskId,
				RequestedState: test.internalrs,
			}).
			Return(nil).
			Times(1)

		// Get info
		_, err := c.StateChange(context.Background(), &api.SdkCloudBackupStateChangeRequest{
			TaskId:         taskId,
			RequestedState: test.sdkrs,
		})
		assert.NoError(t, err)
	}
}

func TestSdkCloudSchedDelete(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	uuid := "uuid-test-1"
	req := &api.SdkCloudBackupSchedDeleteRequest{
		BackupScheduleId: uuid,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupSchedDelete(&api.CloudBackupSchedDeleteRequest{
			UUID: uuid,
		}).
		Return(nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	_, err := c.SchedDelete(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkCloudBackupSchedDeleteBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupSchedDeleteRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// backup id missing
	_, err := c.SchedDelete(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must provide schedule uuid")
}

func TestSdkCloudBackupSchedCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	id := "test-id"
	testSched := []*api.SdkSchedulePolicyInterval{
		&api.SdkSchedulePolicyInterval{
			Retain: 1,
			PeriodType: &api.SdkSchedulePolicyInterval_Daily{
				Daily: &api.SdkSchedulePolicyIntervalDaily{
					Hour:   0,
					Minute: 30,
				},
			},
		},
	}
	req := &api.SdkCloudBackupSchedCreateRequest{
		CloudSchedInfo: &api.SdkCloudBackupScheduleInfo{
			SrcVolumeId:   id,
			CredentialId:  "uuid",
			Schedules:     testSched,
			Full:          true,
			RetentionDays: 3,
		},
	}

	mockReq := api.CloudBackupSchedCreateRequest{}
	mockReq.SrcVolumeID = req.GetCloudSchedInfo().GetSrcVolumeId()
	mockReq.CredentialUUID = req.GetCloudSchedInfo().GetCredentialId()
	mockReq.Schedule = "- freq: daily\n  minute: 30\n  retain: 1\n"
	mockReq.Full = true
	mockReq.RetentionDays = 3

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupSchedCreate(&mockReq).
		Return(&api.CloudBackupSchedCreateResponse{UUID: "uuid"}, nil).
		Times(1)
	setupExpectedCredentialsPassing(s, "uuid")

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	_, err := c.SchedCreate(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkCloudBackupSchedCreateBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupCreateRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// volume id missing
	req.TaskId = "backup-task"
	_, err := c.Create(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")

	// Missing credential uuid
	req.VolumeId = "id"
	setupExpectedCredentialsNotPassing(s)
	req.TaskId = "backup-task"
	_, err = c.Create(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "No configured credentials found")
}

func TestSdkCloudBackupSchedUpdate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	id := "test-id"
	testSched := []*api.SdkSchedulePolicyInterval{
		&api.SdkSchedulePolicyInterval{
			Retain: 1,
			PeriodType: &api.SdkSchedulePolicyInterval_Daily{
				Daily: &api.SdkSchedulePolicyIntervalDaily{
					Hour:   0,
					Minute: 30,
				},
			},
		},
	}
	updateReq := &api.SdkCloudBackupSchedUpdateRequest{
		CloudSchedInfo: &api.SdkCloudBackupScheduleInfo{
			SrcVolumeId:  id,
			CredentialId: "uuid",
			Schedules:    testSched,
		},
		SchedUuid: "uuid-1",
	}

	mockUpdateReq := api.CloudBackupSchedUpdateRequest{}
	mockUpdateReq.SrcVolumeID = updateReq.GetCloudSchedInfo().GetSrcVolumeId()
	mockUpdateReq.CredentialUUID = updateReq.GetCloudSchedInfo().GetCredentialId()
	mockUpdateReq.Schedule = "- freq: daily\n  minute: 30\n  retain: 1\n"
	mockUpdateReq.SchedUUID = updateReq.GetSchedUuid()
	// Create response

	schedList := &api.CloudBackupSchedEnumerateResponse{
		Schedules: map[string]api.CloudBackupScheduleInfo{
			"uuid-1": api.CloudBackupScheduleInfo{
				SrcVolumeID:    "test-id",
				CredentialUUID: "uuid",
				Schedule:       "- freq: daily\n  minute: 30\n  retain: 1\n",
			},
		},
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupSchedEnumerate().
		Return(schedList, nil).
		Times(1)
	s.MockDriver().
		EXPECT().
		CloudBackupSchedUpdate(&mockUpdateReq).
		Return(nil).
		Times(1)
	setupExpectedCredentialsPassing(s, "uuid")

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Update info
	_, err := c.SchedUpdate(context.Background(), updateReq)
	assert.NoError(t, err)
}

func TestSdkCloudBackupSchedEnumerate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupSchedEnumerateRequest{}
	schedList := &api.CloudBackupSchedEnumerateResponse{
		Schedules: map[string]api.CloudBackupScheduleInfo{
			"test-uuid-1": api.CloudBackupScheduleInfo{
				SrcVolumeID:    "myid",
				CredentialUUID: "test-uuid-1",
				Schedule:       "- freq: daily\n  minute: 30\n",
				MaxBackups:     4,
			},
			"test-uuid-2": api.CloudBackupScheduleInfo{
				SrcVolumeID:    "myid2",
				CredentialUUID: "test-uuid-1",
				Schedule:       "- freq: daily\n  minute: 30\n",
				MaxBackups:     3,
			},
		},
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupSchedEnumerate().
		Return(schedList, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.SchedEnumerate(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetCloudSchedList())
	assert.Len(t, r.GetCloudSchedList(), 2)

	// Verify
	for k, v := range r.GetCloudSchedList() {
		sched := schedList.Schedules[k]
		assert.Equal(t, v.GetSrcVolumeId(), sched.SrcVolumeID)
		assert.Equal(t, v.GetCredentialId(), sched.CredentialUUID)
		assert.Equal(t, v.GetMaxBackups(), uint64(sched.MaxBackups))
		assert.Len(t, v.GetSchedules(), 1)
		assert.Equal(t, v.GetSchedules()[0].GetDaily().GetMinute(), int32(30))
	}
}

func TestSdkCloudBackupSize(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "pxbackup"
	creds := "creds"
	backupSize := uint64(176433)
	resp := &api.SdkCloudBackupSizeResponse{
		Size: backupSize,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupSize(&api.SdkCloudBackupSizeRequest{BackupId: id, CredentialId: creds}).
		Return(resp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.Size(context.Background(), &api.SdkCloudBackupSizeRequest{BackupId: id, CredentialId: creds})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetSize())
	assert.Equal(t, r.GetSize(), backupSize)
}
