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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/libopenstorage/openstorage/api"
)

func TestSdkCreateDefragSchedule(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Create request and response
	defragTask := &api.DefragJob{
		MaxDurationHours: 1,
		MaxNodesInParallel: 1,
	}
	resp := &api.SdkCreateDefragScheduleResponse{
		Schedule: &api.Schedule{
			Id:		"12345",
			StartTime: "daily=19:15",
			MaxDurationHours: 1,
			Type: api.Job_DEFRAG,
			Tasks: []*api.Job{
				{
					Type: api.Job_DEFRAG,
					Job: &api.Job_Defrag{
						Defrag: defragTask,
					},
				},
			},
		},
	}

	s.MockCluster().
		EXPECT().
		CreateDefragSchedule(gomock.Any(), gomock.Any()).
		Return(resp, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageFilesystemDefragClient(s.Conn())

	// make the call and verify response
	r, err := c.CreateSchedule(context.Background(), &api.SdkCreateDefragScheduleRequest{
		StartTime: "daily=19:15",
		DefragTask: defragTask,
	})
	assert.NoError(t, err)
	assert.NotNil(t, r.Schedule)
	assert.Equal(t, resp.Schedule.Id, r.Schedule.Id)
	assert.Equal(t, resp.Schedule.StartTime, r.Schedule.StartTime)
	assert.Equal(t, 1, len(r.Schedule.Tasks))
	assert.NotNil(t, r.Schedule.Tasks[0].GetDefrag())
	assert.Equal(t, resp.Schedule.Tasks[0].GetDefrag().MaxNodesInParallel, r.Schedule.Tasks[0].GetDefrag().MaxNodesInParallel)
}
