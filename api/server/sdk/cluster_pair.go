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
<<<<<<< HEAD
<<<<<<< HEAD
		return nil, status.Errorf(codes.Internal, "Cannot create cluster with remote pair %s : %v",
=======
		return nil, status.Errorf(codes.Internal, "Cannot create cluster with remote pair %v : %v",
>>>>>>> CreatePair API implementation with Success test
=======
		return nil, status.Errorf(codes.Internal, "Cannot create cluster with remote pair %s : %v",
>>>>>>> 64acf85c16fd9c2e293c6dcbbcaabf2675131885
			req.GetRemoteClusterIp(), err.Error())
	}

	return &api.ClusterPairCreateResponse{
		RemoteClusterId:   resp.GetRemoteClusterId(),
		RemoteClusterName: resp.GetRemoteClusterName(),
	}, nil
}
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> 64acf85c16fd9c2e293c6dcbbcaabf2675131885

// ProcessRequest handles a remote cluster's pair request
func (s *ClusterPairServer) ProcessRequest(
	ctx context.Context,
	request *api.ClusterPairProcessRequest,
) (*api.ClusterPairProcessResponse, error) {

	if len(request.GetSourceClusterId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Source cluster ID")
	} else if len(request.GetRemoteClusterToken()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Remote cluster Token")
	}
	resp, err := s.cluster.ProcessPairRequest(request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot process cluster pair request with remote pair %s : %v",
			request.GetSourceClusterId(), err.Error())
	}
	return resp, nil
}

// Get information about a cluster pair
func (s *ClusterPairServer) Get(
	ctx context.Context,
	req *api.ClusterPairGetRequest,
) (*api.ClusterPairGetResponse, error) {
	name := req.GetId()
	if len(name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply cluster ID")
	}
	resp, err := s.cluster.GetPair(name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot Get cluster information for %s : %v", name, err.Error())
	}
	return resp, nil
}

// Enumerate returns list of cluster pairs
func (s *ClusterPairServer) Enumerate(
	ctx context.Context,
	req *api.SdkClusterPairsEnumerateRequest,
) (*api.ClusterPairsEnumerateResponse, error) {
	resp, err := s.cluster.EnumeratePairs()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot list cluster pairs : %v", err.Error())
	}
	return resp, nil
}

// GetToken gets the authentication token for this cluster
func (s *ClusterPairServer) GetToken(
	ctx context.Context,
	req *api.ClusterPairTokenGetRequest,
) (*api.ClusterPairTokenGetResponse, error) {
	reset := req.GetResetToken()
	resp, err := s.cluster.GetPairToken(reset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot generate token : %v", err.Error())
	}
	return resp, nil
}

// Delete removes the cluster pairing
func (s *ClusterPairServer) Delete(
	ctx context.Context,
	req *api.ClusterPairDeleteRequest,
) (*api.SdkClusterPairDeleteResponse, error) {
	id := req.GetClusterId()
	if len(id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply valid cluster ID")
	}
	err := s.cluster.DeletePair(id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot delete the cluster pair %s : %v", id, err.Error())
	}
	return &api.SdkClusterPairDeleteResponse{}, nil
}
<<<<<<< HEAD
=======
>>>>>>> CreatePair API implementation with Success test
=======
>>>>>>> 64acf85c16fd9c2e293c6dcbbcaabf2675131885
