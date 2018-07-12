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
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/libopenstorage/openstorage/schedpolicy"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
)

/*
//
// Schedules in yaml format
//
// Day: "freq: daily\nminute: 30\n"
// Weekly: "freq: weekly\nweekday: 5\nhour: 11\nminute: 30\n"
// Monthly: "freq: monthly\nday: 5\nhour: 10\nminute: 30\n"
//
package main

import (
        "fmt"
        "github.com/libopenstorage/openstorage/pkg/sched"
        "gopkg.in/yaml.v2"
        "time"
)

func main() {
        daily := sched.Daily(0, 30).Spec()
        weekly := sched.Weekly(time.Friday, 11, 30).Spec()
        monthly := sched.Monthly(5, 10, 30).Spec()

        do, _ := yaml.Marshal(daily)
        wo, _ := yaml.Marshal(weekly)
        mo, _ := yaml.Marshal(monthly)

        fmt.Printf("D: %q\n", do)
        fmt.Printf("W: %q\n", wo)
        fmt.Printf("M: %q\n", mo)
}
*/

func TestSdkSchedulePolicyCreateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyCreateRequest{
		SchedulePolicy: &api.SdkSchedulePolicy{
			Name: "dummy-schedule-name",
			Schedule: &api.SdkSchedulePolicyInterval{
				Retain: 1,
				PeriodType: &api.SdkSchedulePolicyInterval_Daily{
					Daily: &api.SdkSchedulePolicyIntervalDaily{
						Hour:   0,
						Minute: 30,
					},
				},
			},
		},
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyCreate(req.GetSchedulePolicy().GetName(),
			gomock.Any()).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Create Schedule Policy
	_, err := c.Create(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkSchedulePolicyCreateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyCreateRequest{
		SchedulePolicy: &api.SdkSchedulePolicy{
			Name: "dummy-schedule-name",
			Schedule: &api.SdkSchedulePolicyInterval{
				Retain: 2,
				PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
					Weekly: &api.SdkSchedulePolicyIntervalWeekly{
						Day:    api.SdkTimeWeekday_SdkTimeWeekdayFriday,
						Hour:   0,
						Minute: 30,
					},
				},
			},
		},
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyCreate(req.GetSchedulePolicy().GetName(),
			gomock.Any()).
		Return(fmt.Errorf("Failed to create schedule policy"))

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Create Schedule Policy
	_, err := c.Create(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Failed to create schedule policy")
}

func TestSdkSchedulePolicyCreateNilSchedulePolicyBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyCreateRequest{
		SchedulePolicy: nil,
	}
	// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Create Schedule Policy
	_, err := c.Create(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "SchedulePolicy object cannot be nil")
}

func TestSdkSchedulePolicyCreateBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyCreateRequest{
		SchedulePolicy: &api.SdkSchedulePolicy{},
	}

	// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Create Schedule Policy
	_, err := c.Create(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Schedule name")
}

func TestSdkSchedulePolicyUpdateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyUpdateRequest{
		SchedulePolicy: &api.SdkSchedulePolicy{
			Name: "dummy-schedule-name",
			Schedule: &api.SdkSchedulePolicyInterval{
				Retain: 1,
				PeriodType: &api.SdkSchedulePolicyInterval_Monthly{
					Monthly: &api.SdkSchedulePolicyIntervalMonthly{
						Day:    1,
						Hour:   0,
						Minute: 30,
					},
				},
			},
		},
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyUpdate(req.GetSchedulePolicy().GetName(),
			gomock.Any()).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Update Schedule Policy
	_, err := c.Update(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkSchedulePolicyUpdateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyUpdateRequest{
		SchedulePolicy: &api.SdkSchedulePolicy{
			Name: "dummy-schedule-name",
			Schedule: &api.SdkSchedulePolicyInterval{
				Retain: 10,
				PeriodType: &api.SdkSchedulePolicyInterval_Monthly{
					Monthly: &api.SdkSchedulePolicyIntervalMonthly{
						Day:    1,
						Hour:   0,
						Minute: 30,
					},
				},
			},
		},
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyUpdate(req.GetSchedulePolicy().GetName(),
			gomock.Any()).
		Return(fmt.Errorf("Failed to update schedule policy"))

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Update Schedule Policy
	_, err := c.Update(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Failed to update schedule policy")
}

func TestSdkSchedulePolicyUpdateNilSchedulePolicyBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyUpdateRequest{
		SchedulePolicy: nil,
	}
	// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Update Schedule Policy
	_, err := c.Update(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "SchedulePolicy object cannot be nil")
}

func TestSdkSchedulePolicyUpdateBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyUpdateRequest{
		SchedulePolicy: &api.SdkSchedulePolicy{},
	}

	// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Update Schedule Policy
	_, err := c.Update(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Schedule name")
}

func TestSdkSchedulePolicyDeleteSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyDeleteRequest{
		Name: "dummy-schedule-name",
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyDelete(req.GetName()).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Delete Schedule Policy
	_, err := c.Delete(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkSchedulePolicyDeleteFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyDeleteRequest{
		Name: "dummy-schedule-name",
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyDelete(req.GetName()).
		Return(fmt.Errorf("Failed to delete schedule policy"))

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Delete Schedule Policy
	_, err := c.Delete(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Failed to delete schedule policy")
}

func TestSdkSchedulePolicyDeleteBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyDeleteRequest{}

	// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Delete Schedule Policy
	_, err := c.Delete(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Schedule name")
}

func TestSdkSchedulePolicyEnumerateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyEnumerateRequest{}

	enumerateData := []*schedpolicy.SchedPolicy{
		&schedpolicy.SchedPolicy{
			Name:     "sched-1",
			Schedule: "freq: daily\nminute: 30\n",
		},
		&schedpolicy.SchedPolicy{
			Name:     "sched-2",
			Schedule: "freq: weekly\nweekday: 5\nhour: 11\nminute: 30\n",
		},
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyEnumerate().
		Return(enumerateData, nil)

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Enumerate Schedule Policy
	resp, err := c.Enumerate(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, len(resp.GetPolicies()), 2)
	assert.Equal(t, resp.GetPolicies()[0].GetName(), "sched-1")
}
func TestSdkSchedulePolicyEnumerateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyEnumerateRequest{}

	s.MockCluster().
		EXPECT().
		SchedPolicyEnumerate().
		Return(nil, fmt.Errorf("Failed to enumerate schedule policies"))

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Enumerate Schedule Policy
	_, err := c.Enumerate(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Failed to enumerate schedule policies")
}

func TestSdkSchedulePolicyInspectSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyInspectRequest{
		Name: "sched-2",
	}

	policy := &schedpolicy.SchedPolicy{
		Name:     "sched-2",
		Schedule: "freq: monthly\nday: 5\nhour: 10\nminute: 30\n",
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyGet(req.GetName()).
		Return(policy, nil)

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Inspect Schedule Policy
	resp, err := c.Inspect(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.GetPolicy().GetName(), "sched-2")
}

func TestSdkSchedulePolicyInspectFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyInspectRequest{
		Name: "sched-x",
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyGet(req.GetName()).
		Return(nil, fmt.Errorf("Failed to inspect schedule policy"))

		// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Inspect Schedule Policy
	_, err := c.Inspect(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Failed to inspect schedule policy")
}

func TestSdkSchedulePolicyInspectBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkSchedulePolicyInspectRequest{}

	// Setup client
	c := api.NewOpenStorageSchedulePolicyClient(s.Conn())

	// Inspect Schedule Policy
	_, err := c.Inspect(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Schedule name")
}
