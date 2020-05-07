package sdk

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
	"github.com/libopenstorage/openstorage/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StoragePoolServer struct {
	server serverAccessor
}

func (sp *StoragePoolServer) cluster() cluster.Cluster {
	return sp.server.cluster()
}

func (sp *StoragePoolServer) Rebalance(
	c context.Context, req *api.SdkStorageRebalanceRequest) (*api.SdkStorageRebalanceResponse, error) {
	if sp.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return sp.cluster().Rebalance(c, req)
}

func (sp *StoragePoolServer) UpdateRebalanceJobState(
	c context.Context, req *api.SdkUpdateRebalanceJobRequest) (*api.SdkUpdateRebalanceJobResponse, error) {
	if sp.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}
	return sp.cluster().UpdateRebalanceJobState(c, req)
}

func (sp *StoragePoolServer) GetRebalanceJobStatus(
	c context.Context, req *api.SdkGetRebalanceJobStatusRequest) (*api.SdkGetRebalanceJobStatusResponse, error) {
	if sp.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}
	return sp.cluster().GetRebalanceJobStatus(c, req)
}

func (sp *StoragePoolServer) Resize(
	c context.Context, req *api.SdkStoragePoolResizeRequest) (*api.SdkStoragePoolResizeResponse, error) {
	if sp.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}
	return sp.cluster().Resize(c, req)
}

func (sp *StoragePoolServer) EnumerateRebalanceJobs(
	c context.Context, req *api.SdkEnumerateRebalanceJobsRequest) (*api.SdkEnumerateRebalanceJobsResponse, error) {
	if sp.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}
	return sp.cluster().EnumerateRebalanceJobs(c, req)
}
