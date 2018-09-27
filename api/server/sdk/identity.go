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
	"fmt"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IdentityServer is an implementation of the gRPC OpenStorageIdentityServer interface
type IdentityServer struct {
	driver volume.VolumeDriver
}

// Capabilities returns the capabilities of the SDK server
func (s *IdentityServer) Capabilities(
	ctx context.Context,
	req *api.SdkIdentityCapabilitiesRequest,
) (*api.SdkIdentityCapabilitiesResponse, error) {

	capCluster := &api.SdkServiceCapability{
		Type: &api.SdkServiceCapability_Service{
			Service: &api.SdkServiceCapability_OpenStorageService{
				Type: api.SdkServiceCapability_OpenStorageService_CLUSTER,
			},
		},
	}
	capCloudBackup := &api.SdkServiceCapability{
		Type: &api.SdkServiceCapability_Service{
			Service: &api.SdkServiceCapability_OpenStorageService{
				Type: api.SdkServiceCapability_OpenStorageService_CLOUD_BACKUP,
			},
		},
	}
	capCredentials := &api.SdkServiceCapability{
		Type: &api.SdkServiceCapability_Service{
			Service: &api.SdkServiceCapability_OpenStorageService{
				Type: api.SdkServiceCapability_OpenStorageService_CREDENTIALS,
			},
		},
	}
	capNode := &api.SdkServiceCapability{
		Type: &api.SdkServiceCapability_Service{
			Service: &api.SdkServiceCapability_OpenStorageService{
				Type: api.SdkServiceCapability_OpenStorageService_NODE,
			},
		},
	}
	capObjectStorage := &api.SdkServiceCapability{
		Type: &api.SdkServiceCapability_Service{
			Service: &api.SdkServiceCapability_OpenStorageService{
				Type: api.SdkServiceCapability_OpenStorageService_OBJECT_STORAGE,
			},
		},
	}
	capSchedulePolicy := &api.SdkServiceCapability{
		Type: &api.SdkServiceCapability_Service{
			Service: &api.SdkServiceCapability_OpenStorageService{
				Type: api.SdkServiceCapability_OpenStorageService_SCHEDULE_POLICY,
			},
		},
	}
	capVolume := &api.SdkServiceCapability{
		Type: &api.SdkServiceCapability_Service{
			Service: &api.SdkServiceCapability_OpenStorageService{
				Type: api.SdkServiceCapability_OpenStorageService_VOLUME,
			},
		},
	}
	capAlerts := &api.SdkServiceCapability{
		Type: &api.SdkServiceCapability_Service{
			Service: &api.SdkServiceCapability_OpenStorageService{
				Type: api.SdkServiceCapability_OpenStorageService_ALERTS,
			},
		},
	}
	capMountAttach := &api.SdkServiceCapability{
		Type: &api.SdkServiceCapability_Service{
			Service: &api.SdkServiceCapability_OpenStorageService{
				Type: api.SdkServiceCapability_OpenStorageService_MOUNT_ATTACH,
			},
		},
	}

	return &api.SdkIdentityCapabilitiesResponse{
		Capabilities: []*api.SdkServiceCapability{
			capCluster,
			capCloudBackup,
			capCredentials,
			capNode,
			capObjectStorage,
			capSchedulePolicy,
			capVolume,
			capAlerts,
			capMountAttach,
		},
	}, nil
}

// Version returns version of the storage system
func (s *IdentityServer) Version(
	ctx context.Context,
	req *api.SdkIdentityVersionRequest,
) (*api.SdkIdentityVersionResponse, error) {

	version, err := s.driver.Version()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to get version information: %v", err,
		)
	}

	sdkVersion := &api.SdkVersion{
		Major: int32(api.SdkVersion_Major),
		Minor: int32(api.SdkVersion_Minor),
		Patch: int32(api.SdkVersion_Patch),
		Version: fmt.Sprintf("%d.%d.%d",
			api.SdkVersion_Major,
			api.SdkVersion_Minor,
			api.SdkVersion_Patch,
		),
	}

	return &api.SdkIdentityVersionResponse{
		SdkVersion: sdkVersion,
		Version:    version,
	}, nil
}
