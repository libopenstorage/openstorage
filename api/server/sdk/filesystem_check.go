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
	"github.com/libopenstorage/openstorage/volume"
)

// FilesystemCheckServer is an implementation of the gRPC OpenStorageFilesystemCheck interface
type FilesystemCheckServer struct {
	server      serverAccessor
}

func (s *FilesystemCheckServer) driver(ctx context.Context) volume.VolumeDriver {
	return s.server.driver(ctx)
}

// Start a filesystem trim operation on a mounted volume
func (s *FilesystemCheckServer) CheckHealth(
	ctx context.Context,
	req *api.SdkFilesystemCheckCheckHealthRequest,
) (*api.SdkFilesystemCheckCheckHealthResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}

	r, err := s.driver(ctx).FilesystemCheckCheckHealth(req)

	return r, err
}

// Get Status of a filesystem trim operation
func (s *FilesystemCheckServer) CheckHealthGetStatus(
	ctx context.Context,
	req *api.SdkFilesystemCheckCheckHealthGetStatusRequest,
) (*api.SdkFilesystemCheckCheckHealthGetStatusResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}

	r, err := s.driver(ctx).FilesystemCheckCheckHealthGetStatus(req)

	return r, err
}
// Start a filesystem trim operation on a mounted volume
func (s *FilesystemCheckServer) FixAll(
	ctx context.Context,
	req *api.SdkFilesystemCheckFixAllRequest,
) (*api.SdkFilesystemCheckFixAllResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}

	r, err := s.driver(ctx).FilesystemCheckFixAll(req)

	return r, err
}

// Get Status of a filesystem trim operation
func (s *FilesystemCheckServer) FixAllGetStatus(
	ctx context.Context,
	req *api.SdkFilesystemCheckFixAllGetStatusRequest,
) (*api.SdkFilesystemCheckFixAllGetStatusResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}
	r, err := s.driver(ctx).FilesystemCheckFixAllGetStatus(req)

	return r, err
}

// Stop the background filesystem check (CheckHealth or FixAll) operation on a volume, if any
func (s *FilesystemCheckServer) Stop(
	ctx context.Context,
	req *api.SdkFilesystemCheckStopRequest,
) (*api.SdkFilesystemCheckStopResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}

	r, err := s.driver(ctx).FilesystemCheckStop(req)

	return r, err
}
