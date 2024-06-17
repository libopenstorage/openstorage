package sdk

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
)

func TestInspectSchedule(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	resp := &api.SdkInspectScheduleResponse{
		Schedule: &api.Schedule{
			Id:		"12345",
			StartTime: "daily=19:15",
			MaxDurationMinutes: 60,
			Type: api.Job_DEFRAG,
			Tasks: []*api.Job{
				{
					Type: api.Job_DEFRAG,
					Job: &api.Job_Defrag{
						Defrag: &api.DefragJob{
							MaxDurationMinutes: 60,
							MaxNodesInParallel: 1,
						},
					},
				},
			},
		},
	}

	s.MockCluster().
		EXPECT().
		InspectSchedule(gomock.Any(), gomock.Any()).
		Return(resp, nil)

	// Setup client
	c := api.NewOpenStorageScheduleClient(s.Conn())

	// Get info
	r, err := c.Inspect(
		context.Background(), 
		&api.SdkInspectScheduleRequest{
			Type: "DEFRAG", 
			Id: "12345",
		},
	)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, r.Schedule)
	assert.Equal(t, resp.Schedule.Id, r.Schedule.Id)
	assert.Equal(t, 1, len(r.Schedule.Tasks))
	assert.NotNil(t, r.Schedule.Tasks[0].GetDefrag())
	assert.Equal(t, resp.Schedule.Tasks[0].GetDefrag().MaxDurationMinutes, r.Schedule.Tasks[0].GetDefrag().MaxDurationMinutes)
}

func TestEnumerateSchedules(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create response
	schedule := &api.Schedule{
		Id:		"12345",
		StartTime: "daily=19:15",
		MaxDurationMinutes: 60,
		Type: api.Job_DEFRAG,
		Tasks: []*api.Job{
			{
				Type: api.Job_DEFRAG,
				Job: &api.Job_Defrag{
					Defrag: &api.DefragJob{
						MaxDurationMinutes: 60,
						MaxNodesInParallel: 1,
					},
				},
			},
		},
	}
	resp := &api.SdkEnumerateSchedulesResponse{
		Schedules: []*api.Schedule{schedule},
	}

	s.MockCluster().
		EXPECT().
		EnumerateSchedules(gomock.Any(), gomock.Any()).
		Return(resp, nil)

	// Setup client
	c := api.NewOpenStorageScheduleClient(s.Conn())

	// Get info
	r, err := c.Enumerate(
		context.Background(), 
		&api.SdkEnumerateSchedulesRequest{
			Type: "DEFRAG",
		},
	)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Schedules))
	assert.Equal(t, schedule.Id, r.Schedules[0].Id)
	assert.Equal(t, 1, len(r.Schedules[0].Tasks))
	assert.NotNil(t, r.Schedules[0].Tasks[0].GetDefrag())
	assert.Equal(t, resp.Schedules[0].Tasks[0].GetDefrag().MaxDurationMinutes,
		r.Schedules[0].Tasks[0].GetDefrag().MaxDurationMinutes)
}

func TestDeleteSchedule(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	s.MockCluster().
		EXPECT().
		DeleteSchedule(gomock.Any(), gomock.Any()).
		Return(&api.SdkDeleteScheduleResponse{}, nil)

	// Setup client
	c := api.NewOpenStorageScheduleClient(s.Conn())

	// Delete object store
	resp, err := c.Delete(
		context.Background(),
		&api.SdkDeleteScheduleRequest{
			Id: "12345",
			Type: "DEFRAG",
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
