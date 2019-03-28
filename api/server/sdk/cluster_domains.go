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

	clusterDomainInfos, err := m.cluster().EnumerateDomains()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot enumerate cluster cluster domains: %v", err)
	}
	clusterResp := &api.SdkClusterDomainsEnumerateResponse{}

	for _, clusterDomainInfo := range clusterDomainInfos {
		clusterResp.ClusterDomainNames = append(clusterResp.ClusterDomainNames, clusterDomainInfo.Name)
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

	resp, err := m.cluster().InspectDomain(req.GetClusterDomainName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot enumerate cluster cluster domains: %v", err)
	}

	var isActive bool
	if resp.State == types.CLUSTER_DOMAIN_STATE_ACTIVE {
		isActive = true
	}
	return &api.SdkClusterDomainInspectResponse{
		ClusterDomainName: resp.Name,
		IsActive:          isActive,
	}, nil

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

	if err := m.cluster().UpdateDomainState(req.GetClusterDomainName(), types.CLUSTER_DOMAIN_STATE_ACTIVE); err != nil {
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

	if err := m.cluster().UpdateDomainState(req.GetClusterDomainName(), types.CLUSTER_DOMAIN_STATE_INACTIVE); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to activate cluster domain %v: %v", req.GetClusterDomainName(), err)
	}

	return &api.SdkClusterDomainDeactivateResponse{}, nil
}
