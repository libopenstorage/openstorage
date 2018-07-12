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

	return &api.SdkClusterInspectCurrentResponse{
		Cluster: c.ToStorageCluster(),
	}, nil
}

// AlertEnumerate returns a list of alerts from the storage cluster
func (s *ClusterServer) AlertEnumerate(
	ctx context.Context,
	req *api.SdkClusterAlertEnumerateRequest,
) (*api.SdkClusterAlertEnumerateResponse, error) {

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

	return &api.SdkClusterAlertEnumerateResponse{
		Alerts: alerts.Alert,
	}, nil
}

// AlertClear clears the alert for a given resource
func (s *ClusterServer) AlertClear(
	ctx context.Context,
	req *api.SdkClusterAlertClearRequest,
) (*api.SdkClusterAlertClearResponse, error) {

	err := s.cluster.ClearAlert(req.GetResource(), req.GetAlertId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to clear alert %d for type %v: %v",
			req.GetAlertId(),
			req.GetResource(),
			err.Error())
	}

	return &api.SdkClusterAlertClearResponse{}, nil
}

// AlertDelete deletes an alert for a given resource
func (s *ClusterServer) AlertDelete(
	ctx context.Context,
	req *api.SdkClusterAlertDeleteRequest,
) (*api.SdkClusterAlertDeleteResponse, error) {

	err := s.cluster.EraseAlert(req.GetResource(), req.GetAlertId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to delete alert %d for type %v: %v",
			req.GetAlertId(),
			req.GetResource(),
			err.Error())
	}

	return &api.SdkClusterAlertDeleteResponse{}, nil
}
