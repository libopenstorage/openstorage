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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster"
)

// ClusterServer is an implementation of the gRPC OpenStorageClusterServer interface
type ClusterServer struct {
	cluster cluster.Cluster
}

// InspectCurrent returns information about the current cluster
func (s *ClusterServer) InspectCurrent(
	ctx context.Context,
	req *api.SdkClusterInspectCurrentRequest,
) (*api.SdkClusterInspectCurrentResponse, error) {
	c, err := s.cluster.Enumerate()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get cluster information
	cluster := c.ToStorageCluster()

	// Get cluster unique id
	cluster.Id = s.cluster.Uuid()

	return &api.SdkClusterInspectCurrentResponse{
		Cluster: cluster,
	}, nil
}

// CreatePair with a remote cluster
func (s *ClusterServer) CreatePair(
	ctx context.Context,
	req *api.ClusterPairCreateRequest,
) (*api.ClusterPairCreateResponse, error) {
	return &api.ClusterPairCreateResponse{}, nil
}

// ProcessPairRequest handles an incoming pair request from a remote cluster
func (s *ClusterServer) ProcessPairRequest(
	ctx context.Context,
	req *api.ClusterPairProcessRequest,
) (*api.ClusterPairProcessResponse, error) {
	return &api.ClusterPairProcessResponse{}, nil
}

// GetPair returns pair information for a cluster
func (s *ClusterServer) GetPair(
	ctx context.Context,
	req *api.ClusterPairGetRequest,
) (*api.ClusterPairGetResponse, error) {
	return &api.ClusterPairGetResponse{}, nil
}

// EnumeratePairs returns list of cluster pairs
func (s *ClusterServer) EnumeratePairs(
	ctx context.Context,
	req *api.ClusterPairsEnumerateRequest,
) (*api.ClusterPairsEnumerateResponse, error) {
	return &api.ClusterPairsEnumerateResponse{}, nil
}

// DeletePair Delete a cluster pairing
func (s *ClusterServer) DeletePair(
	ctx context.Context,
	req *api.ClusterPairDeleteRequest,
) (*api.ClusterPairDeleteResponse, error) {
	return &api.ClusterPairDeleteResponse{}, nil
}

// GetPairToken gets the authentication token for this cluster
func (s *ClusterServer) GetPairToken(
	ctx context.Context,
	req *api.ClusterPairTokenGetRequest,
) (*api.ClusterPairTokenGetResponse, error) {
	return &api.ClusterPairTokenGetResponse{}, nil
}
