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
	"github.com/libopenstorage/openstorage/pkg/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClusterPairServer is an implementation of the gRPC OpenStorageClusterServer interface
type ClusterPairServer struct {
	server serverAccessor
}

func (s *ClusterPairServer) cluster() cluster.Cluster {
	return s.server.cluster()
}

// Create a new cluster with remote pair
func (s *ClusterPairServer) Create(
	ctx context.Context,
	req *api.SdkClusterPairCreateRequest,
) (*api.SdkClusterPairCreateResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if req.GetRequest() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Must supply valid request")
	}

	resp, err := s.cluster().CreatePair(req.GetRequest())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot create cluster with remote pair %s : %v",
			req.GetRequest().GetRemoteClusterIp(), err)
	}

	return &api.SdkClusterPairCreateResponse{
		Result: resp,
	}, nil
}

// Inspect information about a cluster pair
func (s *ClusterPairServer) Inspect(
	ctx context.Context,
	req *api.SdkClusterPairInspectRequest,
) (*api.SdkClusterPairInspectResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply cluster ID")
	}
	resp, err := s.cluster().GetPair(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot Get cluster information for %s : %v", req.GetId(), err)
	}
	return &api.SdkClusterPairInspectResponse{
		Result: resp,
	}, nil
}

// Enumerate returns list of cluster pairs
func (s *ClusterPairServer) Enumerate(
	ctx context.Context,
	req *api.SdkClusterPairEnumerateRequest,
) (*api.SdkClusterPairEnumerateResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	resp, err := s.cluster().EnumeratePairs()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot list cluster pairs : %v", err)
	}
	return &api.SdkClusterPairEnumerateResponse{
		Result: resp,
	}, nil
}

// GetToken gets the authentication token for this cluster
func (s *ClusterPairServer) GetToken(
	ctx context.Context,
	req *api.SdkClusterPairGetTokenRequest,
) (*api.SdkClusterPairGetTokenResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	// Check if admin user - only system.admin can get cluster pair tokens
	if userInfo, ok := auth.NewUserInfoFromContext(ctx); ok {
		o := api.Ownership{}
		if !o.IsAdminByUser(userInfo) {
			return nil, status.Error(codes.Unauthenticated, "Must be system admin to get pair token")
		}
	}

	resp, err := s.cluster().GetPairToken(false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot generate token: %v", err)
	}
	return &api.SdkClusterPairGetTokenResponse{
		Result: resp,
	}, nil
}

// ResetToken gets the authentication token for this cluster
func (s *ClusterPairServer) ResetToken(
	ctx context.Context,
	req *api.SdkClusterPairResetTokenRequest,
) (*api.SdkClusterPairResetTokenResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	resp, err := s.cluster().GetPairToken(true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot generate token: %v", err)
	}
	return &api.SdkClusterPairResetTokenResponse{
		Result: resp,
	}, nil
}

// Delete removes the cluster pairing
func (s *ClusterPairServer) Delete(
	ctx context.Context,
	req *api.SdkClusterPairDeleteRequest,
) (*api.SdkClusterPairDeleteResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetClusterId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply valid cluster ID")
	}
	err := s.cluster().DeletePair(req.GetClusterId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot delete the cluster pair %s : %v", req.GetClusterId(), err)
	}
	return &api.SdkClusterPairDeleteResponse{}, nil
}
