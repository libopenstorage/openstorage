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

// FilesystemCheckServer is an implementation of the gRPC OpenStorageFilesystemCheck interface
type FilesystemCheckServer struct {
	server serverAccessor
}

func (s *FilesystemCheckServer) driver(ctx context.Context) volume.VolumeDriver {
	return s.server.driver(ctx)
}

// Start a filesystem check operation on a unmounted volume
func (s *FilesystemCheckServer) Start(
	ctx context.Context,
	req *api.SdkFilesystemCheckStartRequest,
) (*api.SdkFilesystemCheckStartResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}
	if len(req.GetMode()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a mode parameter")
	}

	r, err := s.driver(ctx).FilesystemCheckStart(req)

	return r, err
}

// GetStatus of a filesystem check operation
func (s *FilesystemCheckServer) Status(
	ctx context.Context,
	req *api.SdkFilesystemCheckStatusRequest,
) (*api.SdkFilesystemCheckStatusResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}

	r, err := s.driver(ctx).FilesystemCheckStatus(req)

	return r, err
}

// Stop the background filesystem check (CheckHealth or FixSafe or Fixall) operation on a volume, if any
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
