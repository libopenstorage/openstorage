package nodedrain

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
)

// Provider implements the node drain functionality.
type Provider interface {
	// Drain drains the provided node in the cluster. The options provided
	// in the drain request govern the drain operation.
	Drain(ctx context.Context, in *api.SdkNodeDrainRequest) (*api.SdkNodeDrainResponse, error)
	// UpdateNodeDrainJobState updates an existing node drain task
	// Only acceptable values are
	// NodeDrainJobState_PAUSED - acceptable only from running state
	// NodeDrainJobState_CANCELLED - acceptable only from running/pause state
	// NodeDrainJobState_RUNNING - acceptable only from pause state
	UpdateNodeDrainJobState(context.Context, *api.SdkUpdateNodeDrainJobRequest) (*api.SdkUpdateNodeDrainJobResponse, error)
	// GetDrainStatus drains the provided node in the cluster. The options provided
	// in the drain request govern the drain operation.
	GetDrainStatus(context.Context, *api.SdkGetNodeDrainJobStatusRequest) (*api.SdkGetNodeDrainJobStatusResponse, error)
	// EnumerateNodeDrainJobs returns all the node drain jobs currently known to the system
	EnumerateNodeDrainJobs(context.Context, *api.SdkEnumerateNodeDrainJobsRequest) (*api.SdkEnumerateNodeDrainJobsResponse, error)
}

// NewDefaultNodeDrainProvider does not support any node drain operations.
func NewDefaultNodeDrainProvider() Provider {
	return &UnsupportedNodeDrainProvider{}
}

// UnsupportedNodeDrainProvider unsupported implementation of drain.
type UnsupportedNodeDrainProvider struct {
}

func (u *UnsupportedNodeDrainProvider) Drain(ctx context.Context, in *api.SdkNodeDrainRequest) (*api.SdkNodeDrainResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedNodeDrainProvider) UpdateNodeDrainJobState(ctx context.Context, in *api.SdkUpdateNodeDrainJobRequest) (*api.SdkUpdateNodeDrainJobResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedNodeDrainProvider) GetDrainStatus(ctx context.Context, in *api.SdkGetNodeDrainJobStatusRequest) (*api.SdkGetNodeDrainJobStatusResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedNodeDrainProvider) EnumerateNodeDrainJobs(ctx context.Context, in *api.SdkEnumerateNodeDrainJobsRequest) (*api.SdkEnumerateNodeDrainJobsResponse, error) {
	return nil, &errors.ErrNotSupported{}
}
