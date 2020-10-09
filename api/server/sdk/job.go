package sdk

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
	"github.com/libopenstorage/openstorage/cluster"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JobServer struct {
	server serverAccessor
}

func (j *JobServer) cluster() cluster.Cluster {
	return j.server.cluster()
}

func (j *JobServer) UpdateJobState(
	ctx context.Context,
	req *api.SdkUpdateJobRequest,
) (*api.SdkUpdateJobResponse, error) {
	if j.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return j.cluster().UpdateJobState(ctx, req)
}

func (j *JobServer) GetJobStatus(
	ctx context.Context,
	req *api.SdkGetJobStatusRequest,
) (*api.SdkGetJobStatusResponse, error) {
	if j.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return j.cluster().GetJobStatus(ctx, req)
}

func (j *JobServer) EnumerateJobs(
	ctx context.Context,
	req *api.SdkEnumerateJobsRequest,
) (*api.SdkEnumerateJobsResponse, error) {
	if j.cluster() == nil {
		return nil, status.Error(codes.Unavailable, errors.ErrResourceNotInitialized.Error())
	}

	return j.cluster().EnumerateJobs(ctx, req)
}
