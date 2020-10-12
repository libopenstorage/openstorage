package job

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
)

// Provider implements the APIs for executing and querying asynchronous jobs
type Provider interface {
	// UpdateJobState updates an existing job
	// Only acceptable values are
	// JobState_PAUSED - acceptable only from running state
	// JobState_CANCELLED - acceptable only from running/pause state
	// JobState_RUNNING - acceptable only from pause state
	UpdateJobState(ctx context.Context, in *api.SdkUpdateJobRequest) (*api.SdkUpdateJobResponse, error)
	// GetJobStatus gets the status of a job
	GetJobStatus(ctx context.Context, in *api.SdkGetJobStatusRequest) (*api.SdkGetJobStatusResponse, error)
	// EnumerateJobs returns all the jobs currently known to the system
	EnumerateJobs(ctx context.Context, in *api.SdkEnumerateJobsRequest) (*api.SdkEnumerateJobsResponse, error)
}

// NewDefaultJobProvider does not support asynchronous jobs
func NewDefaultJobProvider() Provider {
	return &UnsupportedJobProvider{}
}

// UnsupportedJobProvider unsupported implementation of jobs APIs
type UnsupportedJobProvider struct {
}

func (u *UnsupportedJobProvider) UpdateJobState(ctx context.Context, in *api.SdkUpdateJobRequest) (*api.SdkUpdateJobResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedJobProvider) GetJobStatus(ctx context.Context, in *api.SdkGetJobStatusRequest) (*api.SdkGetJobStatusResponse, error) {
	return nil, &errors.ErrNotSupported{}
}

func (u *UnsupportedJobProvider) EnumerateJobs(ctx context.Context, in *api.SdkEnumerateJobsRequest) (*api.SdkEnumerateJobsResponse, error) {
	return nil, &errors.ErrNotSupported{}
}
