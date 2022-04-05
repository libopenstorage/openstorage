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
	"net"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/pkg/correlation"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/portworx/kvdb"
	"google.golang.org/grpc"
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

func (s *NodeServer) DrainAttachments(
	ctx context.Context,
	req *api.SdkNodeDrainAttachmentsRequest,
) (*api.SdkJobResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().DrainAttachments(ctx, req)
}

func (s *NodeServer) CordonAttachments(
	ctx context.Context,
	req *api.SdkNodeCordonAttachmentsRequest,
) (*api.SdkNodeCordonAttachmentsResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().CordonAttachments(ctx, req)
}

func (s *NodeServer) UncordonAttachments(
	ctx context.Context,
	req *api.SdkNodeUncordonAttachmentsRequest,
) (*api.SdkNodeUncordonAttachmentsResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().UncordonAttachments(ctx, req)
}

func (s *NodeServer) VolumeUsageByNode(
	ctx context.Context,
	req *api.SdkNodeVolumeUsageByNodeRequest,
) (*api.SdkNodeVolumeUsageByNodeResponse, error) {

	useProxy, host, err := s.needsProxyRequest(ctx, req.GetNodeId())
	if err != nil {
		return nil, err
	}
	if useProxy {
		return s.proxyVolumeUsageByNode(ctx, req, host)
	}
	// Get the info locally
	if s.server.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	resp, err := s.server.driver(ctx).VolumeUsageByNode(req.GetNodeId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, " Failed to get VolumeUsageByNode :%v", err.Error())
	}
	sdkResp := &api.SdkNodeVolumeUsageByNodeResponse{
		VolumeUsageInfo: resp,
	}
	return sdkResp, nil
}

func (s *NodeServer) proxyVolumeUsageByNode(
	ctx context.Context,
	req *api.SdkNodeVolumeUsageByNodeRequest,
	host string,
) (*api.SdkNodeVolumeUsageByNodeResponse, error) {

	proxyClient, err := s.getProxyClient(host)
	if err != nil {
		return nil, err
	}
	return proxyClient.VolumeUsageByNode(ctx, req)
}

func (s *NodeServer) needsProxyRequest(
	ctx context.Context,
	NodeID string,
) (bool, string, error) {
	// If not a local request, proxy to the approriate node
	nodeInspectData, err := s.Inspect(ctx, &api.SdkNodeInspectRequest{NodeId: NodeID})
	if err != nil {
		return false, "", err
	}
	curNodedata, err := s.InspectCurrent(ctx, &api.SdkNodeInspectCurrentRequest{})
	if err != nil {
		return false, "", err
	}
	return (curNodedata.Node.Id != nodeInspectData.Node.Id), nodeInspectData.Node.MgmtIp, nil
}

func (s *NodeServer) getProxyClient(
	host string,
) (api.OpenStorageNodeClient, error) {
	endpoint := net.JoinHostPort(host, s.server.port())
	// TODO TLS
	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(correlation.ContextUnaryClientInterceptor),
	}
	conn, err := grpcserver.Connect(endpoint, dialOpts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Node usage from remote node failed with :%v", err.Error())
	}
	return api.NewOpenStorageNodeClient(conn), nil
}

func (s *NodeServer) RelaxedReclaimPurge(
	ctx context.Context,
	req *api.SdkNodeRelaxedReclaimPurgeRequest,
) (*api.SdkNodeRelaxedReclaimPurgeResponse, error) {

	useProxy, host, err := s.needsProxyRequest(ctx, req.GetNodeId())
	if err != nil {
		return nil, err
	}
	if useProxy {
		return s.proxyRelaxedReclaimPurge(ctx, req, host)
	}
	// Get the info locally
	if s.server.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}
	resp, err := s.server.driver(ctx).RelaxedReclaimPurge(req.GetNodeId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, " Failed to trigger RelaxedReclaimPurge :%v", err.Error())
	}
	sdkResp := &api.SdkNodeRelaxedReclaimPurgeResponse{
		Status: resp,
	}
	return sdkResp, nil
}

func (s *NodeServer) proxyRelaxedReclaimPurge(
	ctx context.Context,
	req *api.SdkNodeRelaxedReclaimPurgeRequest,
	host string,
) (*api.SdkNodeRelaxedReclaimPurgeResponse, error) {
	proxyClient, err := s.getProxyClient(host)
	if err != nil {
		return nil, err
	}
	return proxyClient.RelaxedReclaimPurge(ctx, req)
}
