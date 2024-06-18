package schedule

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
)

// Provider implements the APIs for creating and querying schedules
type Provider interface {
	// InspectSchedule queries a schedule by type and ID
	InspectSchedule(ctx context.Context, in *api.SdkInspectScheduleRequest) (*api.SdkInspectScheduleResponse, error)
	// EnumerateSchedules returns all the schedules of a type
	EnumerateSchedules(ctx context.Context, in *api.SdkEnumerateSchedulesRequest) (*api.SdkEnumerateSchedulesResponse, error)
	// DeleteSchedule deletes a schedule by type and ID
	DeleteSchedule(ctx context.Context, in *api.SdkDeleteScheduleRequest) (*api.SdkDeleteScheduleResponse, error)
}

// NewDefaultScheduleProvider does not support schedules
func NewDefaultScheduleProvider() Provider {
	return &UnsupportedScheduleProvider{}
}

// UnsupportedJobProvider unsupported implementation of schedule APIs
type UnsupportedScheduleProvider struct {
}

func (u *UnsupportedScheduleProvider) InspectSchedule(
	ctx context.Context, in *api.SdkInspectScheduleRequest,
) (*api.SdkInspectScheduleResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedScheduleProvider) EnumerateSchedules(
	ctx context.Context, in *api.SdkEnumerateSchedulesRequest,
) (*api.SdkEnumerateSchedulesResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedScheduleProvider) DeleteSchedule(
	ctx context.Context, in *api.SdkDeleteScheduleRequest,
) (*api.SdkDeleteScheduleResponse, error) {
	return nil, &errors.ErrNotSupported{}
}
