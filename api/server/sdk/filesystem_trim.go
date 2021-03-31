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

// FilesystemTrimServer is an implementation of the gRPC OpenStorageFilesystemTrim interface
type FilesystemTrimServer struct {
	server serverAccessor
}

func (s *FilesystemTrimServer) driver(ctx context.Context) volume.VolumeDriver {
	return s.server.driver(ctx)
}

// Start a filesystem trim operation on a mounted volume
func (s *FilesystemTrimServer) Start(
	ctx context.Context,
	req *api.SdkFilesystemTrimStartRequest,
) (*api.SdkFilesystemTrimStartResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}
	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume mount path")
	}

	r, err := s.driver(ctx).FilesystemTrimStart(req)

	return r, err
}

// Status of a filesystem trim operation
func (s *FilesystemTrimServer) Status(
	ctx context.Context,
	req *api.SdkFilesystemTrimStatusRequest,
) (*api.SdkFilesystemTrimStatusResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}
	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume mount path")
	}

	r, err := s.driver(ctx).FilesystemTrimStatus(req)

	return r, err
}

// Status of auto fs trim operation
func (s *FilesystemTrimServer) AutoFSTrimStatus(
	ctx context.Context,
	req *api.SdkAutoFSTrimStatusRequest,
) (*api.SdkAutoFSTrimStatusResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	r, err := s.driver(ctx).AutoFilesystemTrimStatus(req)

	return r, err
}

// Usage of auto fs trim
func (s *FilesystemTrimServer) AutoFSTrimUsage(
	ctx context.Context,
	req *api.SdkAutoFSTrimUsageRequest,
) (*api.SdkAutoFSTrimUsageResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	r, err := s.driver(ctx).AutoFilesystemTrimUsage(req)

	return r, err
}

// Stop the background filesystem trim operation on a volume, if any
func (s *FilesystemTrimServer) Stop(
	ctx context.Context,
	req *api.SdkFilesystemTrimStopRequest,
) (*api.SdkFilesystemTrimStopResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}
	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume mount path")
	}

	r, err := s.driver(ctx).FilesystemTrimStop(req)

	return r, err
}
