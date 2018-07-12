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
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/portworx/kvdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Objectstoreserver is an implementation of the gRPC OpenStorageObjectstore interface
type ObjectstoreServer struct {
	api.OpenStorageObjectstoreServer
	cluster cluster.Cluster
}

// Inspect Objectstore return status of provided objectstore
func (s *ObjectstoreServer) Inspect(
	ctx context.Context,
	req *api.SdkObjectstoreInspectRequest,
) (*api.SdkObjectstoreInspectResponse, error) {

	objResp, err := s.cluster.ObjectStoreInspect(req.GetObjectstoreId())
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(
				codes.NotFound,
				"Id %s not found",
				req.GetObjectstoreId())
		}
		return nil, status.Errorf(
			codes.Internal,
			"Failed to inspect objectstore: %v",
			err.Error())
	}

	return &api.SdkObjectstoreInspectResponse{ObjectstoreStatus: objResp}, nil
}

// CreateObjectstore creates objectstore for given volume
func (s *ObjectstoreServer) Create(
	ctx context.Context,
	req *api.SdkObjectstoreCreateRequest,
) (*api.SdkObjectstoreCreateResponse, error) {

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide volume ID")
	}
	objResp, err := s.cluster.ObjectStoreCreate(req.GetVolumeId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to create objectstore: %v",
			err.Error())
	}

	return &api.SdkObjectstoreCreateResponse{ObjectstoreStatus: objResp}, nil
}

// UpdateObjectstore updates given objectstore state
func (s *ObjectstoreServer) Update(
	ctx context.Context,
	req *api.SdkObjectstoreUpdateRequest,
) (*api.SdkObjectstoreUpdateResponse, error) {

	err := s.cluster.ObjectStoreUpdate(req.GetObjectstoreId(), req.GetEnable())
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(
				codes.NotFound,
				"Id %s not found",
				req.GetObjectstoreId())
		}
		return nil, status.Errorf(
			codes.Internal,
			"Failed to update objectstore: %v",
			err.Error())
	}

	return &api.SdkObjectstoreUpdateResponse{}, nil
}

// DeleteObjectstore delete objectstore from cluster
func (s *ObjectstoreServer) Delete(
	ctx context.Context,
	req *api.SdkObjectstoreDeleteRequest,
) (*api.SdkObjectstoreDeleteResponse, error) {

	err := s.cluster.ObjectStoreDelete(req.GetObjectstoreId())
	if err != nil && err != kvdb.ErrNotFound {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to delete objectstore: %v",
			err.Error())
	}

	return &api.SdkObjectstoreDeleteResponse{}, nil
}
