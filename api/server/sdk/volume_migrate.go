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
	req *api.CloudMigrateStartRequest,
) (*api.SdkCloudMigrateStartResponse, error) {

	if req.GetOperation() == api.CloudMigrate_InvalidType {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid Operation")
	} else if len(req.GetClusterId()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid Cluster ID")
	} else if len(req.GetTargetId()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid Target cluster ID")
	}
	err := s.driver.CloudMigrateStart(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot start migration for %s : %v", req.GetClusterId(), err.Error())
	}
	return &api.SdkCloudMigrateStartResponse{}, nil
}

// Cancel or stop a ongoing migration
func (s *VolumeServer) Cancel(
	ctx context.Context,
	req *api.CloudMigrateCancelRequest,
) (*api.SdkCloudMigrateCancelResponse, error) {

	if req.GetOperation() == api.CloudMigrate_InvalidType {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid Operation")
	} else if len(req.GetClusterId()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid Cluster ID")
	} else if len(req.GetTargetId()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid Target cluster ID")
	}
	err := s.driver.CloudMigrateCancel(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot stop migration for %s : %v", req.GetClusterId(), err.Error())
	}
	return &api.SdkCloudMigrateCancelResponse{}, nil
}

// Status of ongoing migration
func (s *VolumeServer) Status(
	ctx context.Context,
	req *api.SdkCloudMigrateStatusRequest,
) (*api.CloudMigrateStatusResponse, error) {

	resp, err := s.driver.CloudMigrateStatus()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot get status of migration : %v", err.Error())
	}
	return resp, nil
}
