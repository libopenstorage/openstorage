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

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
)

func TestSdkCloudBackupCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	uuid := "uuid"
	full := false
	req := &api.SdkCloudBackupCreateRequest{
		VolumeId:       id,
		CredentialUuid: uuid,
		Full:           full,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupCreate(&api.CloudBackupCreateRequest{
			VolumeID:       id,
			CredentialUUID: uuid,
			Full:           false,
		}).
		Return(nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	_, err := c.Create(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkCloudBackupCreateBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupCreateRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// volume id missing
	_, err := c.Create(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")

	// Missing credential uuid
	req.VolumeId = "id"
	_, err = c.Create(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "credential uuid")
}

func TestSdkCloudRestoreCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	backupid := "backupid"
	id := "myvol"
	uuid := "uuid"
	req := &api.SdkCloudBackupRestoreRequest{
		BackupId:       backupid,
		CredentialUuid: uuid,
	}

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupRestore(&api.CloudBackupRestoreRequest{
			ID:             backupid,
			CredentialUUID: uuid,
		}).
		Return(&api.CloudBackupRestoreResponse{
			RestoreVolumeID: id,
		}, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.Restore(context.Background(), req)
	assert.Equal(t, r.GetRestoreVolumeId(), id)
	assert.NoError(t, err)
}

func TestSdkCloudBackupRestoreBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupRestoreRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// backup id missing
	_, err := c.Restore(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "backup id")

	// Missing credential uuid
	req.BackupId = "id"
	_, err = c.Restore(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "credential uuid")
}

func TestSdkCloudDeleteCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	backupid := "backupid"
	uuid := "uuid"
	req := &api.SdkCloudBackupDeleteRequest{
		BackupId:       backupid,
		CredentialUuid: uuid,
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
	_, err = c.Delete(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "credential uuid")
}

func TestSdkCloudDeleteAllCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	uuid := "uuid"
	req := &api.SdkCloudBackupDeleteAllRequest{
		SrcVolumeId:    id,
		CredentialUuid: uuid,
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
	req.SrcVolumeId = "id"
	_, err = c.DeleteAll(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "credential uuid")
}

func TestSdkCloudBackupEnumerate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	uuid := "uuid"
	req := &api.SdkCloudBackupEnumerateRequest{
		SrcVolumeId:    id,
		CredentialUuid: uuid,
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
				Status: "Done",
			},
			{
				ID:            "two",
				SrcVolumeID:   "two:vol",
				SrcVolumeName: "two:volname",
				Timestamp:     time.Now(),
				Metadata: map[string]string{
					"what a": "world",
				},
				Status: "Failed",
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

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Get info
	r, err := c.Enumerate(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, r.GetBackupIds())
	assert.Len(t, r.GetBackupIds(), 2)

	for i, backup := range list.Backups {
		assert.Equal(t, r.BackupIds[i], backup.ID)
	}
}

func TestSdkCloudBackupEnumerateBadArguments(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCloudBackupEnumerateRequest{}

	// Setup client
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	// Missing credential uuid
	req.SrcVolumeId = "id"
	_, err := c.Enumerate(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "credential uuid")
}

func TestSdkCloudBackupStatus(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	id := "myvol"
	req := &api.SdkCloudBackupStatusRequest{
		SrcVolumeId: id,
	}
	statuses := &api.CloudBackupStatusResponse{
		Statuses: map[string]api.CloudBackupStatus{
			"hello": api.CloudBackupStatus{
				ID:            "myid",
				OpType:        api.CloudBackupOp,
				Status:        api.CloudBackupStatusPaused,
				BytesDone:     123456,
				StartTime:     time.Now(),
				CompletedTime: time.Now(),
				NodeID:        "mynode",
			},
			"world": api.CloudBackupStatus{
				ID:            "another",
				OpType:        api.CloudRestoreOp,
				Status:        api.CloudBackupStatusDone,
				BytesDone:     97324,
				StartTime:     time.Now(),
				CompletedTime: time.Now(),
				NodeID:        "myothernode",
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
		assert.Equal(t, v.GetNodeId(), status.NodeID)
		assert.Equal(t, v.GetOptype(), api.CloudBackupOpTypeToSdkCloudBackupOpType(status.OpType))
		assert.Equal(t, v.GetStatus(), api.CloudBackupStatusTypeToSdkCloudBackupStatusType(status.Status))

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
		BackupId:       id,
		CredentialUuid: creds,
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
	_, err = c.Catalog(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "credential uuid")
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
		{
			api.CloudBackupRequestedStatePause,
			api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStatePause,
		},
		{
			api.CloudBackupRequestedStateResume,
			api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateResume,
		},
		{
			api.CloudBackupRequestedStateStop,
			api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateStop,
		},
	}
	id := "myvol"
	c := api.NewOpenStorageCloudBackupClient(s.Conn())

	for _, test := range tests {
		// Create response
		s.MockDriver().
			EXPECT().
			CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
				SrcVolumeID:    id,
				RequestedState: test.internalrs,
			}).
			Return(nil).
			Times(1)

		// Get info
		_, err := c.StateChange(context.Background(), &api.SdkCloudBackupStateChangeRequest{
			SrcVolumeId:    id,
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
		Uuid: uuid,
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
	assert.Contains(t, serverError.Message(), "credential uuid")
}

func TestSdkCloudBackupSchedCreate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	testSched := &api.SdkSchedulePolicyInterval{
		Retain: 1,
		PeriodType: &api.SdkSchedulePolicyInterval_Daily{
			Daily: &api.SdkSchedulePolicyIntervalDaily{
				Hour:   0,
				Minute: 30,
			},
		},
	}
	req := &api.SdkCloudBackupSchedCreateRequest{
		CloudSchedInfo: &api.SdkCloudBackupScheduleInfo{
			SrcVolumeId:    "test-id",
			CredentialUuid: "uuid",
			Schedule:       testSched,
		},
	}

	mockReq := api.CloudBackupSchedCreateRequest{}
	mockReq.SrcVolumeID = req.GetCloudSchedInfo().GetSrcVolumeId()
	mockReq.CredentialUUID = req.GetCloudSchedInfo().GetCredentialUuid()
	mockReq.Schedule = "freq: daily\nminute: 30\nretain: 1\n"

	// Create response
	s.MockDriver().
		EXPECT().
		CloudBackupSchedCreate(&mockReq).
		Return(&api.CloudBackupSchedCreateResponse{}, nil).
		Times(1)

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
	_, err := c.Create(context.Background(), req)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume id")

	// Missing credential uuid
	req.VolumeId = "id"
	_, err = c.Create(context.Background(), req)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "credential uuid")
}
