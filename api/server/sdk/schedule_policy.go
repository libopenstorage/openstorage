/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2018 Portworx

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

	"github.com/libopenstorage/openstorage/cluster"
	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SchedulePolicyServer is an implementation of the gRPC OpenStorageSchedulePolicy interface
type SchedulePolicyServer struct {
	api.OpenStorageSchedulePolicyServer
	cluster cluster.Cluster
}

// Create method creates schedule policy
func (s *SchedulePolicyServer) Create(
	ctx context.Context,
	req *api.SdkSchedulePolicyCreateRequest,
) (*api.SdkSchedulePolicyCreateResponse, error) {

	if req.GetSchedulePolicy() == nil {
		return nil, status.Error(codes.InvalidArgument, "SchedulePolicy object cannot be nil")
	} else if len(req.GetSchedulePolicy().GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Schedule name")
	} else if req.GetSchedulePolicy().GetSchedules() == nil ||
		len(req.GetSchedulePolicy().GetSchedules()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must a supply Schedule")
	}

	out, err := sdkSchedToRetainInternalSpecYamlByte(req.GetSchedulePolicy().GetSchedules())
	if err != nil {
		return nil, err
	}

	err = s.cluster.SchedPolicyCreate(req.GetSchedulePolicy().GetName(), string(out))
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create schedule policy: %v",
			err.Error())
	}

	return &api.SdkSchedulePolicyCreateResponse{}, err
}

// Update method updates schedule policy
func (s *SchedulePolicyServer) Update(
	ctx context.Context,
	req *api.SdkSchedulePolicyUpdateRequest,
) (*api.SdkSchedulePolicyUpdateResponse, error) {

	if req.GetSchedulePolicy() == nil {
		return nil, status.Error(codes.InvalidArgument, "SchedulePolicy object cannot be nil")
	} else if len(req.GetSchedulePolicy().GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Schedule name")
	} else if req.GetSchedulePolicy().GetSchedules() == nil ||
		len(req.GetSchedulePolicy().GetSchedules()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Schedule")
	}

	out, err := sdkSchedToRetainInternalSpecYamlByte(req.GetSchedulePolicy().GetSchedules())
	if err != nil {
		return nil, err
	}
	err = s.cluster.SchedPolicyUpdate(req.GetSchedulePolicy().GetName(), string(out))
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(
				codes.NotFound,
				"Schedule policy %s not found",
				req.GetSchedulePolicy().GetName())
		}
		return nil, status.Errorf(
			codes.Internal,
			"Failed to update schedule policy: %v",
			err.Error())
	}

	return &api.SdkSchedulePolicyUpdateResponse{}, err
}

// Delete method deletes schedule policy
func (s *SchedulePolicyServer) Delete(
	ctx context.Context,
	req *api.SdkSchedulePolicyDeleteRequest,
) (*api.SdkSchedulePolicyDeleteResponse, error) {

	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Schedule name")
	}

	err := s.cluster.SchedPolicyDelete(req.GetName())
	if err != nil && err != kvdb.ErrNotFound {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to delete schedule policy: %v",
			err.Error())
	}

	return &api.SdkSchedulePolicyDeleteResponse{}, nil
}

// Enumerate method enumerates schedule policies
func (s *SchedulePolicyServer) Enumerate(
	ctx context.Context,
	req *api.SdkSchedulePolicyEnumerateRequest,
) (*api.SdkSchedulePolicyEnumerateResponse, error) {

	policies, err := s.cluster.SchedPolicyEnumerate()

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Failed to enumerate schedule policy: %v",
			err.Error())
	}
	sdkpolicies := []*api.SdkSchedulePolicy{}
	for _, policy := range policies {

		schedules, err := retainInternalSpecYamlByteToSdkSched([]byte(policy.Schedule))
		if err != nil {
			return nil, err
		}

		p := &api.SdkSchedulePolicy{
			Name:      policy.Name,
			Schedules: schedules,
		}
		sdkpolicies = append(sdkpolicies, p)
	}

	return &api.SdkSchedulePolicyEnumerateResponse{Policies: sdkpolicies}, err
}

// Inspect method inspects schedule policy
func (s *SchedulePolicyServer) Inspect(
	ctx context.Context,
	req *api.SdkSchedulePolicyInspectRequest,
) (*api.SdkSchedulePolicyInspectResponse, error) {

	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Schedule name")
	}

	policy, err := s.cluster.SchedPolicyGet(req.GetName())
	if err != nil {
		if err == kvdb.ErrNotFound {
			return nil, status.Errorf(
				codes.NotFound,
				"Schedule policy %s not found",
				req.GetName())
		}
		return nil, status.Errorf(
			codes.Internal,
			"Failed to inspect schedule policy: %v",
			err.Error())
	}

	schedules, err := retainInternalSpecYamlByteToSdkSched([]byte(policy.Schedule))
	if err != nil {
		return nil, err
	}

	sdkpolicy := &api.SdkSchedulePolicy{
		Name:      policy.Name,
		Schedules: schedules,
	}

	return &api.SdkSchedulePolicyInspectResponse{Policy: sdkpolicy}, err
}
