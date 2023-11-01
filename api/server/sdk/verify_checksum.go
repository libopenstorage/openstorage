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

// VerifyChecksumServer is an implementation of the gRPC OpenStorageVerifyChecksum interface
type VerifyChecksumServer struct {
	server serverAccessor
}

func (s *VerifyChecksumServer) driver(ctx context.Context) volume.VolumeDriver {
	return s.server.driver(ctx)
}

// Start checksum validation on the volume
func (s *VerifyChecksumServer) Start(
	ctx context.Context,
	req *api.SdkVerifyChecksumStartRequest,
) (*api.SdkVerifyChecksumStartResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}

	// Check ownership
	if err := checkAccessFromDriverForVolumeIds(ctx, s.driver(ctx), []string{req.GetVolumeId()}, api.Ownership_Write); err != nil {
		return nil, err
	}

	r, err := s.driver(ctx).VerifyChecksumStart(req)

	return r, err
}

// GetStatus of checksum validation on the volume
func (s *VerifyChecksumServer) Status(
	ctx context.Context,
	req *api.SdkVerifyChecksumStatusRequest,
) (*api.SdkVerifyChecksumStatusResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}

	// Check ownership
	if err := checkAccessFromDriverForVolumeIds(ctx, s.driver(ctx), []string{req.GetVolumeId()}, api.Ownership_Read); err != nil {
		return nil, err
	}

	r, err := s.driver(ctx).VerifyChecksumStatus(req)

	return r, err
}

// Stop checksum validation on the volume 
func (s *VerifyChecksumServer) Stop(
	ctx context.Context,
	req *api.SdkVerifyChecksumStopRequest,
) (*api.SdkVerifyChecksumStopResponse, error) {

	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	var err error
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a volume id")
	}

	// Check ownership
	if err := checkAccessFromDriverForVolumeIds(ctx, s.driver(ctx), []string{req.GetVolumeId()}, api.Ownership_Write); err != nil {
		return nil, err
	}

	r, err := s.driver(ctx).VerifyChecksumStop(req)

	return r, err
}
