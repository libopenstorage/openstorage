package defrag

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
)

// Provider is a collection of APIs for performing different kinds of defrag
// operations on a node
type Provider interface {
	// Create a schedule to run defragmentation tasks periodically
	CreateDefragSchedule(ctx context.Context, req *api.SdkCreateDefragScheduleRequest) (
		*api.SdkCreateDefragScheduleResponse, error,
	)
	// Clean up defrag schedules and stop all defrag operations
	CleanUpDefragSchedules(ctx context.Context, req *api.SdkCleanUpDefragSchedulesRequest) (
		*api.SdkCleanUpDefragSchedulesResponse, error,
	)
	// Get defrag status of a node
	GetDefragNodeStatus(ctx context.Context, req *api.SdkGetDefragNodeStatusRequest) (
		*api.SdkGetDefragNodeStatusResponse, error,
	)
	// Enumerate all nodes, returning defrag status of the entire cluster
	EnumerateDefragStatus(ctx context.Context, req *api.SdkEnumerateDefragStatusRequest) (
		*api.SdkEnumerateDefragStatusResponse, error,
	)
}

// NewDefaultNodeDrainProvider does not any defrag related operations
func NewDefaultDefragProvider() Provider {
	return &UnsupportedDefragProvider{}
}

// UnsupportedDefragProvider unsupported implementation of defrag.
type UnsupportedDefragProvider struct {
}

func (u *UnsupportedDefragProvider) CreateDefragSchedule(
	ctx context.Context, req *api.SdkCreateDefragScheduleRequest,
) (*api.SdkCreateDefragScheduleResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedDefragProvider) CleanUpDefragSchedules(
	ctx context.Context, req *api.SdkCleanUpDefragSchedulesRequest,
) (*api.SdkCleanUpDefragSchedulesResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedDefragProvider) GetDefragNodeStatus(
	ctx context.Context, req *api.SdkGetDefragNodeStatusRequest,
) (*api.SdkGetDefragNodeStatusResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedDefragProvider) EnumerateDefragStatus(
	ctx context.Context, req *api.SdkEnumerateDefragStatusRequest,
) (*api.SdkEnumerateDefragStatusResponse, error) {
	return nil, &errors.ErrNotSupported{}
}
