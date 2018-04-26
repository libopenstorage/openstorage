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

	"github.com/golang/protobuf/ptypes"

	"github.com/libopenstorage/openstorage/api"
)

// Enumerate returns information about the cluster
func (s *Server) Enumerate(ctx context.Context, req *api.ClusterEnumerateRequest) (*api.ClusterEnumerateResponse, error) {
	c, err := s.cluster.Enumerate()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.ClusterEnumerateResponse{
		Cluster: c.ToStorageCluster(),
	}, nil
}

// Inspect returns information about a specific node
func (s *Server) Inspect(ctx context.Context, req *api.ClusterInspectRequest) (*api.ClusterInspectResponse, error) {
	if len(req.GetNodeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Node id must be provided")
	}

	node, err := s.cluster.Inspect(req.GetNodeId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.ClusterInspectResponse{
		Node: node.ToStorageNode(),
	}, nil
}

// AlertEnumerate returns a list of alerts from the storage cluster
func (s *Server) AlertEnumerate(
	ctx context.Context,
	req *api.ClusterAlertEnumerateRequest,
) (*api.ClusterAlertEnumerateResponse, error) {

	ts, err := ptypes.Timestamp(req.GetTimeStart())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Unable to get start time from request: %v",
			err.Error())
	}

	te, err := ptypes.Timestamp(req.GetTimeEnd())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Unable to get start time from request: %v",
			err.Error())
	}

	alerts, err := s.cluster.EnumerateAlerts(ts, te, req.GetResource())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to enumerate alerts for type %v: %v",
			req.GetResource(),
			err.Error())
	}

	return &api.ClusterAlertEnumerateResponse{
		Alerts: alerts,
	}, nil
}

// AlertClear clears the alert for a given resource
func (s *Server) AlertClear(
	ctx context.Context,
	req *api.ClusterAlertClearRequest,
) (*api.ClusterAlertClearResponse, error) {

	err := s.cluster.ClearAlert(req.GetResource(), req.GetAlertId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to clear alert %d for type %v: %v",
			req.GetAlertId(),
			req.GetResource(),
			err.Error())
	}

	return &api.ClusterAlertClearResponse{}, nil
}

// AlertErase erases an alert for a given resource
func (s *Server) AlertErase(
	ctx context.Context,
	req *api.ClusterAlertEraseRequest,
) (*api.ClusterAlertEraseResponse, error) {

	err := s.cluster.EraseAlert(req.GetResource(), req.GetAlertId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to erase alert %d for type %v: %v",
			req.GetAlertId(),
			req.GetResource(),
			err.Error())
	}

	return &api.ClusterAlertEraseResponse{}, nil
}
