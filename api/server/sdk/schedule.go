package sdk

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
	"github.com/libopenstorage/openstorage/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ScheduleServer struct {
	server serverAccessor
}

func (s *ScheduleServer) cluster() cluster.Cluster {
	return s.server.cluster()
}

func (s *ScheduleServer) Inspect(
	ctx context.Context,
	req *api.SdkInspectScheduleRequest,
) (*api.SdkInspectScheduleResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().InspectSchedule(ctx, req)
}

func (s *ScheduleServer) Enumerate(
	ctx context.Context,
	req *api.SdkEnumerateSchedulesRequest,
) (*api.SdkEnumerateSchedulesResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().EnumerateSchedules(ctx, req)
}

func (s *ScheduleServer) Delete(
	ctx context.Context,
	req *api.SdkDeleteScheduleRequest,
) (*api.SdkDeleteScheduleResponse, error) {
	if s.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return s.cluster().DeleteSchedule(ctx, req)
}
