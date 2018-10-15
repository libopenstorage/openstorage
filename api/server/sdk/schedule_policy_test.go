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
	"time"

	"github.com/golang/mock/gomock"

	"github.com/libopenstorage/openstorage/pkg/sched"
	"github.com/libopenstorage/openstorage/schedpolicy"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
)

func TestSdkSchedulePolicyCreateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	seconds := int64(1234567)
	req := &api.SdkSchedulePolicyCreateRequest{
		SchedulePolicy: &api.SdkSchedulePolicy{
			Name: "dummy-schedule-name",
			Schedules: []*api.SdkSchedulePolicyInterval{
				&api.SdkSchedulePolicyInterval{
					Retain: 10,
					PeriodType: &api.SdkSchedulePolicyInterval_Daily{
						Daily: &api.SdkSchedulePolicyIntervalDaily{
							Hour:   0,
							Minute: 30,
						},
					},
				},
				&api.SdkSchedulePolicyInterval{
					Retain: 20,
					PeriodType: &api.SdkSchedulePolicyInterval_Monthly{
						Monthly: &api.SdkSchedulePolicyIntervalMonthly{
							Day:    20,
							Hour:   11,
							Minute: 22,
						},
					},
				},
				&api.SdkSchedulePolicyInterval{
					Retain: 30,
					PeriodType: &api.SdkSchedulePolicyInterval_Periodic{
						Periodic: &api.SdkSchedulePolicyIntervalPeriodic{
							Seconds: seconds,
						},
					},
				},
				&api.SdkSchedulePolicyInterval{
					Retain: 40,
					PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
						Weekly: &api.SdkSchedulePolicyIntervalWeekly{
							Day:    api.SdkTimeWeekday_SdkTimeWeekdayTuesday,
							Hour:   10,
							Minute: 10,
						},
					},
				},
			},
		},
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyCreate(req.GetSchedulePolicy().GetName(), gomock.Any()).
		Do(func(name, schedule string) {
			// Verify that the yaml string created can be parsed
			intervals, _, err := sched.ParseScheduleAndPolicies(schedule)
			assert.NoError(t, err)
			assert.Len(t, intervals, len(req.GetSchedulePolicy().GetSchedules()))

			// Verify name is correct
			assert.Equal(t, req.GetSchedulePolicy().GetName(), name)

			// Verify data is correct
			policies, err := retainInternalSpecYamlByteToSdkSched([]byte(schedule))
			assert.NoError(t, err)
			assert.Len(t, policies, len(req.GetSchedulePolicy().GetSchedules()))

			// Check Daily data
			actualDaily := policies[0]
			expectedDaily := req.GetSchedulePolicy().GetSchedules()[0]
			assert.Equal(t, expectedDaily.GetRetain(), actualDaily.GetRetain())
			assert.Equal(t, expectedDaily.GetDaily().GetHour(), actualDaily.GetDaily().GetHour())
			assert.Equal(t, expectedDaily.GetDaily().GetMinute(), actualDaily.GetDaily().GetMinute())

			// Check Monthly data
			actualMonthly := policies[1]
			expectedMonthly := req.GetSchedulePolicy().GetSchedules()[1]
			assert.Equal(t, expectedMonthly.GetRetain(), actualMonthly.GetRetain())
			assert.Equal(t, expectedMonthly.GetMonthly().GetDay(), actualMonthly.GetMonthly().GetDay())
			assert.Equal(t, expectedMonthly.GetMonthly().GetHour(), actualMonthly.GetMonthly().GetHour())
			assert.Equal(t, expectedMonthly.GetMonthly().GetMinute(), actualMonthly.GetMonthly().GetMinute())

			// Check Periodic data
			actualPeriodic := policies[2]
			expectedPeriodic := req.GetSchedulePolicy().GetSchedules()[2]
			assert.Equal(t, expectedPeriodic.GetRetain(), actualPeriodic.GetRetain())
			assert.Equal(t, expectedPeriodic.GetPeriodic().GetSeconds(), actualPeriodic.GetPeriodic().GetSeconds())

			// Check weekly data
			actualWeekly := policies[3]
			expectedWeekly := req.GetSchedulePolicy().GetSchedules()[3]
			assert.Equal(t, expectedWeekly.GetRetain(), actualWeekly.GetRetain())
			assert.Equal(t, expectedWeekly.GetWeekly().GetDay(), actualWeekly.GetWeekly().GetDay())
			assert.Equal(t, expectedWeekly.GetWeekly().GetHour(), actualWeekly.GetWeekly().GetHour())
			assert.Equal(t, expectedWeekly.GetWeekly().GetMinute(), actualWeekly.GetWeekly().GetMinute())
		}).
		Return(nil).
		Times(1)

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
			Schedules: []*api.SdkSchedulePolicyInterval{
				&api.SdkSchedulePolicyInterval{
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
			Schedules: []*api.SdkSchedulePolicyInterval{
				&api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Monthly{
						Monthly: &api.SdkSchedulePolicyIntervalMonthly{
							Day:    1,
							Hour:   0,
							Minute: 30,
						},
					},
				},
				&api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Daily{
						Daily: &api.SdkSchedulePolicyIntervalDaily{
							Minute: 1,
							Hour:   0,
						},
					},
				},
			},
		},
	}

	s.MockCluster().
		EXPECT().
		SchedPolicyUpdate(req.GetSchedulePolicy().GetName(),
			gomock.Any()).
		Do(func(name, schedule string) {
			_, _, err := sched.ParseScheduleAndPolicies(schedule)
			assert.NoError(t, err)
		}).
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
			Schedules: []*api.SdkSchedulePolicyInterval{
				&api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Monthly{
						Monthly: &api.SdkSchedulePolicyIntervalMonthly{
							Day:    1,
							Hour:   0,
							Minute: 30,
						},
					},
				},
				&api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Daily{
						Daily: &api.SdkSchedulePolicyIntervalDaily{
							Minute: 1,
							Hour:   0,
						},
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

	seconds := int64(123456)
	enumerateData := []*schedpolicy.SchedPolicy{
		&schedpolicy.SchedPolicy{
			Name:     "sched-1",
			Schedule: "- freq: daily\n  minute: 30\n",
		},
		&schedpolicy.SchedPolicy{
			Name:     "sched-2",
			Schedule: "- freq: weekly\n  weekday: 5\n  hour: 11\n  minute: 30\n",
		},
		&schedpolicy.SchedPolicy{
			Name:     "sched-3",
			Schedule: fmt.Sprintf("- freq: periodic\n  period: %d\n", int64(time.Duration(seconds)*time.Second)),
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
	assert.Len(t, resp.GetPolicies(), 3)
	assert.Equal(t, resp.GetPolicies()[0].GetName(), "sched-1")
	assert.Equal(t, resp.GetPolicies()[2].GetName(), "sched-3")

	s3policy := resp.GetPolicies()[2].GetSchedules()
	assert.Len(t, s3policy, 1)
	assert.Equal(t, s3policy[0].GetPeriodic().GetSeconds(), seconds)
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
		Schedule: "- freq: monthly\n  day: 5\n  hour: 10\n  minute: 30\n",
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
