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
	"github.com/libopenstorage/openstorage/volume"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FilesystemDefragServer is an implementation of the gRPC OpenStorageFilesystemDefrag interface
type FilesystemDefragServer struct {
	server serverAccessor
}

func (s *FilesystemDefragServer) driver(ctx context.Context) volume.VolumeDriver {
	return s.server.driver(ctx)
}

// Create a schedule to run defragmentation tasks periodically
func (s *FilesystemDefragServer) CreateSchedule(
	ctx context.Context,
	req *api.SdkCreateDefragScheduleRequest,
) (*api.SdkCreateDefragScheduleResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	r, err := s.server.cluster().CreateDefragSchedule(ctx, req)

	return r, err
}

// Clean up defrag schedules and stop all defrag operations
func (s *FilesystemDefragServer) CleanUpSchedules(
	ctx context.Context,
	req *api.SdkCleanUpDefragSchedulesRequest,
) (*api.SdkCleanUpDefragSchedulesResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	r, err := s.server.cluster().CleanUpDefragSchedules(ctx, req)

	return r, err
}

// Get defrag status of a node
func (s *FilesystemDefragServer) GetNodeStatus(
	ctx context.Context,
	req *api.SdkGetDefragNodeStatusRequest,
) (*api.SdkGetDefragNodeStatusResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	r, err := s.server.cluster().GetDefragNodeStatus(ctx, req)

	return r, err
}

// Enumerate all nodes, returning defrag status of the entire cluster
func (s *FilesystemDefragServer) EnumerateNodeStatus(
	ctx context.Context,
	req *api.SdkEnumerateDefragStatusRequest,
) (*api.SdkEnumerateDefragStatusResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	r, err := s.server.cluster().EnumerateDefragStatus(ctx, req)

	return r, err
}
