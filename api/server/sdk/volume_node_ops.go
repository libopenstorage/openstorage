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

// Attach volume to given node
func (s *VolumeServer) Attach(
	ctx context.Context,
	req *api.SdkVolumeAttachRequest,
) (*api.SdkVolumeAttachResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	devPath, err := s.driver.Attach(req.GetVolumeId(), req.GetOptions())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed  to attach volume: %v",
			err.Error())
	}

	return &api.SdkVolumeAttachResponse{DevicePath: devPath}, nil
}

// Detach function for volume node detach
func (s *VolumeServer) Detach(
	ctx context.Context,
	req *api.SdkVolumeDetachRequest,
) (*api.SdkVolumeDetachResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	err := s.driver.Detach(req.GetVolumeId(), nil)

	return &api.SdkVolumeDetachResponse{}, err
}

// Mount function for volume node detach
func (s *VolumeServer) Mount(
	ctx context.Context,
	req *api.SdkVolumeMountRequest,
) (*api.SdkVolumeMountResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Mount Path")
	}

	err := s.driver.Mount(req.GetVolumeId(), req.GetMountPath(), req.GetOptions())

	return &api.SdkVolumeMountResponse{}, err
}

// Unmount volume from given node
func (s *VolumeServer) Unmount(
	ctx context.Context,
	req *api.SdkVolumeUnmountRequest,
) (*api.SdkVolumeUnmountResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply volume id")
	}

	if len(req.GetMountPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid Mount Path")
	}

	err := s.driver.Unmount(req.GetVolumeId(), req.GetMountPath(), req.GetOptions())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed  to attach volume: %v",
			err.Error())
	}

	return &api.SdkVolumeUnmountResponse{}, nil
}
