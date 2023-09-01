package storagepool

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
)

// NewDefaultStoragePoolProvider returns an implementation of the storage pool provider that returns a not supported error
func NewDefaultStoragePoolProvider() api.OpenStoragePoolServer {
	return &UnsupportedPoolProvider{}
}

// UnsupportedPoolProvider does not support any storage pool APIs
type UnsupportedPoolProvider struct {
}

func (n *UnsupportedPoolProvider) EnumerateRebalanceJobs(
	c context.Context, request *api.SdkEnumerateRebalanceJobsRequest) (*api.SdkEnumerateRebalanceJobsResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (n *UnsupportedPoolProvider) Resize(
	c context.Context, request *api.SdkStoragePoolResizeRequest) (*api.SdkStoragePoolResizeResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (n *UnsupportedPoolProvider) Rebalance(
	c context.Context, request *api.SdkStorageRebalanceRequest) (*api.SdkStorageRebalanceResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (n *UnsupportedPoolProvider) UpdateRebalanceJobState(
	c context.Context, request *api.SdkUpdateRebalanceJobRequest) (*api.SdkUpdateRebalanceJobResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (n *UnsupportedPoolProvider) GetRebalanceJobStatus(
	c context.Context, request *api.SdkGetRebalanceJobStatusRequest) (*api.SdkGetRebalanceJobStatusResponse, error) {
	panic("implement me")
}

func (n *UnsupportedPoolProvider) CreateRebalanceSchedule(
	c context.Context, request *api.SdkCreateRebalanceScheduleRequest) (*api.SdkCreateRebalanceScheduleResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (n *UnsupportedPoolProvider) GetRebalanceSchedule(
	c context.Context, request *api.SdkGetRebalanceScheduleRequest) (*api.SdkGetRebalanceScheduleResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (n *UnsupportedPoolProvider) DeleteRebalanceSchedule(
	c context.Context, request *api.SdkDeleteRebalanceScheduleRequest) (*api.SdkDeleteRebalanceScheduleResponse, error) {
	return nil, &errors.ErrNotSupported{}
}