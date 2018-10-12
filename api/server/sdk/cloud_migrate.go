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
)

// CloudMigrateServer is an implementation of the gRPC OpenStorageCloudMigrate interface
type CloudMigrateServer struct {
	driver volume.VolumeDriver
}

// CloudMigrateStart starts a migrate operation
func (s *CloudMigrateServer) CloudMigrateStart(
	ctx context.Context,
	request *api.CloudMigrateStartRequest,
) (*api.CloudMigrateStartResponse, error) {
	return nil, nil
}

// CloudMigrateCancel cancels a migrate operation
func (s *CloudMigrateServer) CloudMigrateCancel(
	ctx context.Context,
	request *api.CloudMigrateCancelRequest,
) (*api.CloudMigrateCancelResponse, error) {
	return nil, nil
}

// CloudMigrateStatus returns status for the migration operations
func (s *CloudMigrateServer) CloudMigrateStatus(
	ctx context.Context,
	req *api.CloudMigrateStatusRequest,
) (*api.CloudMigrateStatusResponse, error) {
	return &api.CloudMigrateStatusResponse{}, nil
}
