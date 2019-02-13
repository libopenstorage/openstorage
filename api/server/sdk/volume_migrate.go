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

	"github.com/libopenstorage/openstorage/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Start a volume migration
func (s *VolumeServer) Start(
	ctx context.Context,
	req *api.SdkCloudMigrateStartRequest,
) (*api.SdkCloudMigrateStartResponse, error) {

	if volume := req.GetVolume(); volume != nil {
		// Check ownership
		if err := checkAccessFromDriverForVolumeId(ctx, s.driver(), volume.GetVolumeId(), api.Ownership_Read); err != nil {
			return nil, err
		}

		//migrate volume
		return s.volumeMigrate(ctx, req, volume)
	} else if volumeGroup := req.GetVolumeGroup(); volumeGroup != nil {

		labels := make(map[string]string, 0)
		labels["group"] = volumeGroup.GetGroupId()
		if !s.haveOwnership(ctx, labels) {
			return nil, status.Error(codes.InvalidArgument, "Volume Operation not Permitted.")
		}
		//migrate volume groups
		return s.volumeGroupMigrate(ctx, req, volumeGroup)
	} else if allVolumes := req.GetAllVolumes(); allVolumes != nil {
		// migrate all volumes
		if !s.haveOwnership(ctx, nil) {
			return nil, status.Error(codes.InvalidArgument, "Volume Operation not Permitted.")
		}
		return s.allVolumesMigrate(ctx, req, allVolumes)
	}
	return nil, status.Error(codes.InvalidArgument, "Unknown operation request")
}

func (s *VolumeServer) haveOwnership(ctx context.Context, labels map[string]string) bool {
	vols, err := s.driver().Enumerate(nil, labels)
	if err != nil {
		return false
	}
	for _, vol := range vols {
		// Check ownership
		if err := checkAccessFromDriverForVolumeId(ctx, s.driver(), vol.Id, api.Ownership_Read); err != nil {
			return false
		}
	}

	return true
}

func (s *VolumeServer) volumeGroupMigrate(
	ctx context.Context,
	req *api.SdkCloudMigrateStartRequest,
	volumeGroup *api.SdkCloudMigrateStartRequest_MigrateVolumeGroup,
) (*api.SdkCloudMigrateStartResponse, error) {
	//Create a request object with operation as Migrate volume
	request := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateVolumeGroup,
		ClusterId: req.GetClusterId(),
		TargetId:  volumeGroup.GetGroupId(),
		TaskId:    req.GetTaskId(), // optional will be "" if not passed
	}
	resp, err := s.driver().CloudMigrateStart(request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot start migration for %s : %v", req.GetClusterId(), err)
	}
	return &api.SdkCloudMigrateStartResponse{
		Result: resp,
	}, nil
}

func (s *VolumeServer) allVolumesMigrate(
	ctx context.Context,
	req *api.SdkCloudMigrateStartRequest,
	allVolume *api.SdkCloudMigrateStartRequest_MigrateAllVolumes,
) (*api.SdkCloudMigrateStartResponse, error) {
	//Create a request object with operation as Migrate volume
	request := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateCluster,
		ClusterId: req.GetClusterId(),
		TaskId:    req.GetTaskId(),
	}
	resp, err := s.driver().CloudMigrateStart(request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot start migration for %s : %v", req.GetClusterId(), err)
	}
	return &api.SdkCloudMigrateStartResponse{
		Result: resp,
	}, nil
}

func (s *VolumeServer) volumeMigrate(
	ctx context.Context,
	req *api.SdkCloudMigrateStartRequest,
	volume *api.SdkCloudMigrateStartRequest_MigrateVolume,
) (*api.SdkCloudMigrateStartResponse, error) {
	//Create a request object with operation as Migrate volume
	request := &api.CloudMigrateStartRequest{
		Operation: api.CloudMigrate_MigrateVolume,
		ClusterId: req.GetClusterId(),
		TargetId:  volume.GetVolumeId(),
		TaskId:    req.GetTaskId(),
	}
	resp, err := s.driver().CloudMigrateStart(request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot start migration for %s : %v", req.GetClusterId(), err)
	}
	return &api.SdkCloudMigrateStartResponse{
		Result: resp,
	}, nil
}

// Cancel or stop a ongoing migration
func (s *VolumeServer) Cancel(
	ctx context.Context,
	req *api.SdkCloudMigrateCancelRequest,
) (*api.SdkCloudMigrateCancelResponse, error) {

	if req.GetRequest() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid request")
	} else if len(req.GetRequest().GetTaskId()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid Task ID")
	}
	err := s.driver().CloudMigrateCancel(req.GetRequest())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot stop migration for %s : %v",
			req.GetRequest().GetTaskId(), err)
	}
	return &api.SdkCloudMigrateCancelResponse{}, nil
}

// Status of ongoing migration
func (s *VolumeServer) Status(
	ctx context.Context,
	req *api.SdkCloudMigrateStatusRequest,
) (*api.SdkCloudMigrateStatusResponse, error) {

	resp, err := s.driver().CloudMigrateStatus(req.GetRequest())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot get status of migration : %v", err)
	}
	return &api.SdkCloudMigrateStatusResponse{
		Result: resp,
	}, nil

}
