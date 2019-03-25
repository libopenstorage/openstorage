/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2019 Portworx

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

	"github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClusterDomainsServer is an implementation of the gRPC OpenStorageClusterDomains interface
type ClusterDomainsServer struct {
	server serverAccessor
}

func (m *ClusterDomainsServer) cluster() cluster.Cluster {
	return m.server.cluster()
}

func (m *ClusterDomainsServer) Enumerate(
	ctx context.Context,
	req *api.SdkClusterDomainsEnumerateRequest,
) (*api.SdkClusterDomainsEnumerateResponse, error) {
	if m.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	resp, err := m.cluster().Enumerate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot enumerate cluster cluster domains: %v", err)
	}
	clusterResp := &api.SdkClusterDomainsEnumerateResponse{}

	for clusterDomainName, _ := range resp.ClusterDomainsActiveMap {
		clusterResp.ClusterDomainNames = append(clusterResp.ClusterDomainNames, clusterDomainName)
	}
	return clusterResp, nil
}

func (m *ClusterDomainsServer) Inspect(
	ctx context.Context,
	req *api.SdkClusterDomainInspectRequest,
) (*api.SdkClusterDomainInspectResponse, error) {
	if m.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetClusterDomainName()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Must provide a valid cluster domain name")
	}

	resp, err := m.cluster().Enumerate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot enumerate cluster cluster domains: %v", err)
	}

	for clusterDomainName, clusterDomainState := range resp.ClusterDomainsActiveMap {
		if clusterDomainName == req.GetClusterDomainName() {
			isActive := true
			if clusterDomainState == types.CLUSTER_DOMAIN_STATE_INACTIVE {
				isActive = false
			}
			return &api.SdkClusterDomainInspectResponse{
				ClusterDomainName: clusterDomainName,
				IsActive:          isActive,
			}, nil
		}
	}
	return nil, status.Errorf(codes.InvalidArgument, "Cannot find a cluster domain with name: %v", req.GetClusterDomainName())
}

func (m *ClusterDomainsServer) Activate(
	ctx context.Context,
	req *api.SdkClusterDomainActivateRequest,
) (*api.SdkClusterDomainActivateResponse, error) {
	if m.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if !api.IsAdminByContext(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "Activating cluster domain %v not permitted."+
			" Only the storage administrator is authorized.", req.GetClusterDomainName())
	}

	if len(req.GetClusterDomainName()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Must provide a valid cluster domain name")
	}

	if err := m.cluster().ActivateClusterDomain(&api.ActivateClusterDomainRequest{
		ClusterDomain: req.GetClusterDomainName(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to activate cluster domain %v: %v", req.GetClusterDomainName(), err)
	}
	return &api.SdkClusterDomainActivateResponse{}, nil
}

func (m *ClusterDomainsServer) Deactivate(
	ctx context.Context,
	req *api.SdkClusterDomainDeactivateRequest,
) (*api.SdkClusterDomainDeactivateResponse, error) {
	if m.cluster() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if !api.IsAdminByContext(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "Activating cluster domain %v not permitted."+
			" Only the storage administrator is authorized.", req.GetClusterDomainName())
	}

	if len(req.GetClusterDomainName()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Must provide a valid cluster domain name")
	}

	if err := m.cluster().DeactivateClusterDomain(&api.DeactivateClusterDomainRequest{
		ClusterDomain: req.GetClusterDomainName(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to deactivate cluster domain %v: %v", req.GetClusterDomainName(), err)
	}
	return &api.SdkClusterDomainDeactivateResponse{}, nil
}
