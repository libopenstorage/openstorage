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

// ClusterPairServer is an implementation of the gRPC OpenStorageClusterServer interface
type ClusterPairServer struct {
	cluster cluster.Cluster
}

// Create a new cluster with remote pair
func (s *ClusterPairServer) Create(
	ctx context.Context,
	req *api.ClusterPairCreateRequest,
) (*api.ClusterPairCreateResponse, error) {

	if len(req.GetRemoteClusterIp()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Remote cluster IP")
	} else if len(req.GetRemoteClusterToken()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Authentication Token")
	} else if req.GetRemoteClusterPort() == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Remote cluster Port")
	}

	resp, err := s.cluster.CreatePair(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot create cluster with remote pair %v : %v",
			req.GetRemoteClusterIp(), err.Error())
	}

	return &api.ClusterPairCreateResponse{
		RemoteClusterId:   resp.GetRemoteClusterId(),
		RemoteClusterName: resp.GetRemoteClusterName(),
	}, nil
}
