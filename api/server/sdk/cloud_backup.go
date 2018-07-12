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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

// CloudBackupServer is an implementation of the gRPC OpenStorageCloudBackup interface
type CloudBackupServer struct {
	driver volume.VolumeDriver
}

// Create creates a backup for a volume
func (s *CloudBackupServer) Create(
	ctx context.Context,
	req *api.SdkCloudBackupCreateRequest,
) (*api.SdkCloudBackupCreateResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	} else if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply credential uuid")
	}

	// Create the backup
	if err := s.driver.CloudBackupCreate(&api.CloudBackupCreateRequest{
		VolumeID:       req.GetVolumeId(),
		CredentialUUID: req.GetCredentialId(),
		Full:           req.GetFull(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create backup: %v", err)
	}

	return &api.SdkCloudBackupCreateResponse{}, nil
}

// Restore a backup
func (s *CloudBackupServer) Restore(
	ctx context.Context,
	req *api.SdkCloudBackupRestoreRequest,
) (*api.SdkCloudBackupRestoreResponse, error) {

	if len(req.GetBackupId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide backup id")
	} else if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credential uuid")
	}

	r, err := s.driver.CloudBackupRestore(&api.CloudBackupRestoreRequest{
		ID:                req.GetBackupId(),
		RestoreVolumeName: req.GetRestoreVolumeName(),
		CredentialUUID:    req.GetCredentialId(),
		NodeID:            req.GetNodeId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to restore backup: %v", err)
	}

	return &api.SdkCloudBackupRestoreResponse{
		RestoreVolumeId: r.RestoreVolumeID,
	}, nil

}

// Delete deletes a backup
func (s *CloudBackupServer) Delete(
	ctx context.Context,
	req *api.SdkCloudBackupDeleteRequest,
) (*api.SdkCloudBackupDeleteResponse, error) {

	if len(req.GetBackupId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide backup id")
	} else if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credential uuid")
	}

	if err := s.driver.CloudBackupDelete(&api.CloudBackupDeleteRequest{
		ID:             req.GetBackupId(),
		CredentialUUID: req.GetCredentialId(),
		Force:          req.GetForce(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete backup: %v", err)
	}

	return &api.SdkCloudBackupDeleteResponse{}, nil
}

// DeleteAll deletes all backups for a certain volume
func (s *CloudBackupServer) DeleteAll(
	ctx context.Context,
	req *api.SdkCloudBackupDeleteAllRequest,
) (*api.SdkCloudBackupDeleteAllResponse, error) {

	if len(req.GetSrcVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide source volume id")
	} else if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credential uuid")
	}

	if err := s.driver.CloudBackupDeleteAll(&api.CloudBackupDeleteAllRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			SrcVolumeID:    req.GetSrcVolumeId(),
			CredentialUUID: req.GetCredentialId(),
		},
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete backup: %v", err)
	}

	return &api.SdkCloudBackupDeleteAllResponse{}, nil
}

// Enumerate returns information about the backups
func (s *CloudBackupServer) Enumerate(
	ctx context.Context,
	req *api.SdkCloudBackupEnumerateRequest,
) (*api.SdkCloudBackupEnumerateResponse, error) {

	if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credential uuid")
	}

	r, err := s.driver.CloudBackupEnumerate(&api.CloudBackupEnumerateRequest{
		CloudBackupGenericRequest: api.CloudBackupGenericRequest{
			SrcVolumeID:    req.GetSrcVolumeId(),
			ClusterID:      req.GetClusterId(),
			CredentialUUID: req.GetCredentialId(),
			All:            req.GetAll(),
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to enumerate backups: %v", err)
	}

	return r.ToSdkCloudBackupEnumerateResponse(), nil
}

// Status provides status on a backup
func (s *CloudBackupServer) Status(
	ctx context.Context,
	req *api.SdkCloudBackupStatusRequest,
) (*api.SdkCloudBackupStatusResponse, error) {

	r, err := s.driver.CloudBackupStatus(&api.CloudBackupStatusRequest{
		SrcVolumeID: req.GetVolumeId(),
		Local:       req.GetLocal(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get status of backup: %v", err)
	}

	return r.ToSdkCloudBackupStatusResponse(), nil
}

// Catalog
func (s *CloudBackupServer) Catalog(
	ctx context.Context,
	req *api.SdkCloudBackupCatalogRequest,
) (*api.SdkCloudBackupCatalogResponse, error) {

	if len(req.GetBackupId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide backup id")
	} else if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credential uuid")
	}

	r, err := s.driver.CloudBackupCatalog(&api.CloudBackupCatalogRequest{
		ID:             req.GetBackupId(),
		CredentialUUID: req.GetCredentialId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get catalog: %v", err)
	}

	return &api.SdkCloudBackupCatalogResponse{
		Contents: r.Contents,
	}, nil
}

// History returns ??
func (s *CloudBackupServer) History(
	ctx context.Context,
	req *api.SdkCloudBackupHistoryRequest,
) (*api.SdkCloudBackupHistoryResponse, error) {

	if len(req.GetSrcVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide volume id")
	}
	r, err := s.driver.CloudBackupHistory(&api.CloudBackupHistoryRequest{
		SrcVolumeID: req.GetSrcVolumeId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get history: %v", err)
	}

	return r.ToSdkCloudBackupHistoryResponse(), nil
}

// StateChange pauses and resumes backups
func (s *CloudBackupServer) StateChange(
	ctx context.Context,
	req *api.SdkCloudBackupStateChangeRequest,
) (*api.SdkCloudBackupStateChangeResponse, error) {

	if len(req.GetSrcVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide volume id")
	} else if req.GetRequestedState() == api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateUnknown {
		return nil, status.Error(codes.InvalidArgument, "Must provide requested state")
	}

	var rs string
	switch req.GetRequestedState() {
	case api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStatePause:
		rs = api.CloudBackupRequestedStatePause
	case api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateResume:
		rs = api.CloudBackupRequestedStateResume
	case api.SdkCloudBackupRequestedState_SdkCloudBackupRequestedStateStop:
		rs = api.CloudBackupRequestedStateStop
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Invalid requested state: %v", req.GetRequestedState())
	}

	err := s.driver.CloudBackupStateChange(&api.CloudBackupStateChangeRequest{
		SrcVolumeID:    req.GetSrcVolumeId(),
		RequestedState: rs,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to change state: %v", err)
	}

	return &api.SdkCloudBackupStateChangeResponse{}, nil
}

// SchedCreate new schedule for cloud backup
func (s *CloudBackupServer) SchedCreate(
	ctx context.Context,
	req *api.SdkCloudBackupSchedCreateRequest,
) (*api.SdkCloudBackupSchedCreateResponse, error) {

	if req.GetCloudSchedInfo() == nil {
		return nil, status.Error(codes.InvalidArgument, "BackupSchedule object cannot be nil")
	} else if len(req.GetCloudSchedInfo().GetSrcVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply source volume id")
	} else if len(req.GetCloudSchedInfo().GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply credential uuid")
	} else if req.GetCloudSchedInfo().GetSchedule() == nil ||
		req.GetCloudSchedInfo().GetSchedule().GetPeriodType() == nil {
		return nil, status.Error(codes.InvalidArgument, "Must supply Schedule")
	}

	sched, err := sdkSchedToRetainInternalSpecYamlByte(req.GetCloudSchedInfo().GetSchedule())
	if err != nil {
		return nil, err
	}

	bkpRequest := api.CloudBackupSchedCreateRequest{}
	bkpRequest.SrcVolumeID = req.GetCloudSchedInfo().GetSrcVolumeId()
	bkpRequest.CredentialUUID = req.GetCloudSchedInfo().GetCredentialId()
	bkpRequest.Schedule = string(sched)
	bkpRequest.MaxBackups = uint(req.GetCloudSchedInfo().GetMaxBackups())

	// Create the backup
	schedResp, err := s.driver.CloudBackupSchedCreate(&bkpRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create backup: %v", err)
	}

	return &api.SdkCloudBackupSchedCreateResponse{
		BackupScheduleId: schedResp.UUID,
	}, nil

}

// SchedDelete cloud backup schedule
func (s *CloudBackupServer) SchedDelete(
	ctx context.Context,
	req *api.SdkCloudBackupSchedDeleteRequest,
) (*api.SdkCloudBackupSchedDeleteResponse, error) {

	if len(req.GetBackupScheduleId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credential uuid")
	}

	// Call cloud backup driver function to delete cloud schedule
	if err := s.driver.CloudBackupSchedDelete(&api.CloudBackupSchedDeleteRequest{
		UUID: req.GetBackupScheduleId(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete cloud backup schedule: %v", err)
	}

	return &api.SdkCloudBackupSchedDeleteResponse{}, nil
}

// SchedEnumerate cloud backup schedule
func (s *CloudBackupServer) SchedEnumerate(
	ctx context.Context,
	req *api.SdkCloudBackupSchedEnumerateRequest,
) (*api.SdkCloudBackupSchedEnumerateResponse, error) {
	r, err := s.driver.CloudBackupSchedEnumerate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to enumerate backups: %v", err)
	}
	// since can't import sdk/utils to api because of cyclic import, converting
	// api.CloudBackupScheduleInfo to api.SdkCloudBackupScheduleInfo
	return ToSdkCloudBackupSchedEnumerateResponse(r), nil
}

func ToSdkCloudBackupSchedEnumerateResponse(r *api.CloudBackupSchedEnumerateResponse) *api.SdkCloudBackupSchedEnumerateResponse {
	resp := &api.SdkCloudBackupSchedEnumerateResponse{
		CloudSchedList: make(map[string]*api.SdkCloudBackupScheduleInfo),
	}

	for k, v := range r.Schedules {
		resp.CloudSchedList[k] = ToSdkCloudBackupdScheduleInfo(v)
	}

	return resp
}

func ToSdkCloudBackupdScheduleInfo(s api.CloudBackupScheduleInfo) *api.SdkCloudBackupScheduleInfo {

	schedule, err := retainInternalSpecYamlByteToSdkSched([]byte(s.Schedule))
	if err != nil {
		return nil
	}
	cloudSched := &api.SdkCloudBackupScheduleInfo{
		SrcVolumeId:  s.SrcVolumeID,
		CredentialId: s.CredentialUUID,
		Schedule:     schedule,
		// Not sure about go and protobuf type conversion, converting to higher type
		// converting uint to uint64
		MaxBackups: uint64(s.MaxBackups),
	}
	return cloudSched
}
