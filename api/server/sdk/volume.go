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
	"github.com/libopenstorage/openstorage/api/spec"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/volume"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VolumeServer is an implementation of the gRPC OpenStorageVolume interface
type VolumeServer struct {
	specHandler spec.SpecHandler
	server      serverAccessor
}

func (s *VolumeServer) cluster() cluster.Cluster {
	return s.server.cluster()
}

func (s *VolumeServer) driver() volume.VolumeDriver {
	return s.server.driver()
}

func (s *VolumeServer) checkAccessForVolumeId(
	ctx context.Context,
	volumeId string,
	accessType api.Ownership_AccessType,
) error {
	// Inspect will check access for us
	resp, err := s.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: volumeId,
	})
	if err != nil {
		return err
	}
	if !resp.GetVolume().IsPermitted(ctx, accessType) {
		return status.Errorf(codes.PermissionDenied, "Access denied to volume %v", resp.GetVolume().GetId())
	}
	return nil
}
