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
	"github.com/libopenstorage/openstorage/api/errors"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/portworx/kvdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NodeServer is an implementation of the gRPC OpenStorageNodeServer interface
type NodeServer struct {
	server serverAccessor
}

func (s *NodeServer) cluster() cluster.Cluster {
	return s.server.cluster()
}

// Enumerate returns the ids of all the nodes in the cluster
func (s *NodeServer) Enumerate(
	ctx context.Context,
	req *api.SdkNodeEnumerateRequest,
) (*api.SdkNodeEnumerateResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	c, err := s.cluster().Enumerate()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	nodeIds := make([]string, len(c.Nodes))
	for i, v := range c.Nodes {
		nodeIds[i] = v.Id
	}

	return &api.SdkNodeEnumerateResponse{
		NodeIds: nodeIds,
	}, nil
}

// EnumerateWithFilters returns all the nodes in the cluster
func (s *NodeServer) EnumerateWithFilters(
	ctx context.Context,
	req *api.SdkNodeEnumerateWithFiltersRequest,
) (*api.SdkNodeEnumerateWithFiltersResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	c, err := s.cluster().Enumerate()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	nodes := make([]*api.StorageNode, len(c.Nodes))
	for i, node := range c.Nodes {
		nodes[i] = node.ToStorageNode()
	}

	return &api.SdkNodeEnumerateWithFiltersResponse{
		Nodes: nodes,
	}, nil
}

// Inspect returns information about a specific node
func (s *NodeServer) Inspect(
	ctx context.Context,
	req *api.SdkNodeInspectRequest,
) (*api.SdkNodeInspectResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetNodeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Node id must be provided")
	}

	node, err := s.cluster().Inspect(req.GetNodeId())
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(
				codes.NotFound,
				"Id %s not found",
				req.GetNodeId())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.SdkNodeInspectResponse{
		Node: node.ToStorageNode(),
	}, nil
}

func (s *NodeServer) InspectCurrent(
	ctx context.Context,
	req *api.SdkNodeInspectCurrentRequest,
) (*api.SdkNodeInspectCurrentResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	c, err := s.cluster().Enumerate()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to determine node id: %v", err.Error())
	}

	resp, err := s.Inspect(ctx, &api.SdkNodeInspectRequest{
		NodeId: c.NodeId,
	})
	if err != nil {
		return nil, err
	}

	return &api.SdkNodeInspectCurrentResponse{
		Node: resp.GetNode(),
	}, nil
}

func (s *NodeServer) Drain(
	ctx context.Context,
	req *api.SdkNodeDrainRequest,
) (*api.SdkNodeDrainResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().Drain(ctx, req)

}

func (s *NodeServer) UpdateNodeDrainJobState(
	ctx context.Context,
	req *api.SdkUpdateNodeDrainJobRequest,
) (*api.SdkUpdateNodeDrainJobResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().UpdateNodeDrainJobState(ctx, req)
}

func (s *NodeServer) GetDrainStatus(
	ctx context.Context,
	req *api.SdkGetNodeDrainJobStatusRequest,
) (*api.SdkGetNodeDrainJobStatusResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().GetDrainStatus(ctx, req)
}

func (s *NodeServer) EnumerateNodeDrainJobs(
	ctx context.Context,
	req *api.SdkEnumerateNodeDrainJobsRequest,
) (*api.SdkEnumerateNodeDrainJobsResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().EnumerateNodeDrainJobs(ctx, req)
}

func (s *NodeServer) VolumeUsageByNode(
	ctx context.Context,
	req *api.SdkVolumeUsageByNodeRequest,
) (*api.SdkVolumeUsageByNodeResponse, error) {

	resp, err := s.server.driver(ctx).VolumeUsageByNode(req.GetNodeId())
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"Failed obtain usage stats for volumes in node:%v",
			req.GetNodeId(),
			err.Error())
	}
	//TODO check secruity ( only admin must be allowed to do this
	sdkResp := &api.SdkVolumeUsageByNodeResponse{
		VolumeUsageInfo: resp,
	}
	return sdkResp, nil
}
