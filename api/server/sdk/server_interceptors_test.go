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
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizationServerInterceptorCreateVolume(t *testing.T) {
	tt := []struct {
		TestName                     string
		PublicVolumeCreationDisabled bool
		RequestAuthenticated         bool

		ExpectSuccess      bool
		ExpectedError      string
		ExpectedDriverCall gomock.Call
	}{
		{
			TestName:                     "1-1: Authenticated volume creation should succeed with public vol creation disabled",
			PublicVolumeCreationDisabled: true,
			RequestAuthenticated:         true,

			ExpectSuccess: true,
			ExpectedError: "",
		},
		{
			TestName:                     "1-2: Unauthenticated volume creation should fail with public volume creation disabled",
			PublicVolumeCreationDisabled: true,
			RequestAuthenticated:         false,

			ExpectSuccess: false,
			ExpectedError: "rpc error: code = PermissionDenied desc = Access denied without authentication token",
		},
		{
			TestName:                     "2-1: Authenticated volume creation should succeed with public vol creation enabled",
			PublicVolumeCreationDisabled: false,
			RequestAuthenticated:         true,

			ExpectSuccess: true,
			ExpectedError: "",
		},
		{
			TestName:                     "2-2: Unauthenticated volume creation should succeed with public volume creation enabled",
			PublicVolumeCreationDisabled: false,
			RequestAuthenticated:         false,

			ExpectSuccess: true,
			ExpectedError: "",
		},
	}

	s := newTestServerAuth(t)
	defer s.Stop()
	for _, tc := range tt {
		volumeAPIs := "*"
		if tc.PublicVolumeCreationDisabled {
			volumeAPIs = "!create"
		}
		rc := api.NewOpenStorageRoleClient(s.Conn())
		updateCtx, err := contextWithToken(context.Background(), "jim.stevens", "system.admin", "mysecret")
		assert.NoError(t, err, "Expected context with token to succeed")
		_, err = rc.Update(updateCtx, &api.SdkRoleUpdateRequest{
			Role: &api.SdkRole{
				Name: "system.guest",
				Rules: []*api.SdkRule{
					&api.SdkRule{
						Services: []string{"volume"},
						Apis:     []string{volumeAPIs},
					},
					&api.SdkRule{
						Services: []string{"mountattach", "cloudbackup", "migrate"},
						Apis:     []string{"*"},
					},
					&api.SdkRule{
						Services: []string{"identity"},
						Apis:     []string{"version"},
					},
				},
			},
		})
		assert.NoError(t, err, "Expected role update to succeed")

		name := "myvol"
		size := uint64(1234)
		req := &api.SdkVolumeCreateRequest{
			Name: name,
			Spec: &api.VolumeSpec{
				Size: size,
			},
		}

		id := "myid"
		gomock.InOrder(
			s.MockDriver().
				EXPECT().
				Inspect([]string{name}).
				Return(nil, fmt.Errorf("not found")).
				AnyTimes(),
			s.MockDriver().
				EXPECT().
				Enumerate(&api.VolumeLocator{Name: name}, nil).
				Return(nil, fmt.Errorf("not found")).
				AnyTimes(),
			s.MockDriver().
				EXPECT().
				Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(id, nil).
				AnyTimes(),
		)

		// Setup client
		c := api.NewOpenStorageVolumeClient(s.Conn())

		// Call create with or without auth
		ctx := context.Background()
		if tc.RequestAuthenticated {
			ctx, err = contextWithToken(ctx, "jim.stevens", "system.admin", "mysecret")
			assert.NoError(t, err)
		}
		r, err := c.Create(ctx, req)

		if tc.ExpectSuccess {
			assert.NoError(t, err)
			assert.Equal(t, r.GetVolumeId(), "myid")
		} else {
			assert.Error(t, err)
			assert.Equal(t, tc.ExpectedError, err.Error())
		}
	}
}

func TestAuthorizationServerInterceptorStreamingCreateAlert(t *testing.T) {
	tt := []struct {
		TestName             string
		RequestAuthenticated bool

		ExpectSuccess      bool
		ExpectedError      string
		ExpectedDriverCall gomock.Call
	}{
		{
			TestName:             "1-1: Authenticated alerts API call should succeed",
			RequestAuthenticated: true,

			ExpectSuccess: true,
			ExpectedError: "",
		},
		{
			TestName:             "1-2: Unauthenticated alerts API call should not succeed",
			RequestAuthenticated: false,

			ExpectSuccess: false,
			ExpectedError: "rpc error: code = PermissionDenied desc = Access denied without authentication token",
		},
	}

	s := newTestServerAuth(t)
	defer s.Stop()
	for _, tc := range tt {
		req := &api.SdkAlertsEnumerateWithFiltersRequest{
			Queries: []*api.SdkAlertsQuery{
				{
					Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
					Opts: []*api.SdkAlertsOption{
						{
							Opt: testNewCountSpanOption(0, 10),
						},
					},
				},
			},
		}

		// Setup client
		c := api.NewOpenStorageAlertsClient(s.Conn())
		var filters []interface{}
		for _, filter := range getFilters([]*api.SdkAlertsQuery{
			{
				Query: testNewResourceTypeQuery(api.ResourceType_RESOURCE_TYPE_DRIVE),
				Opts: []*api.SdkAlertsOption{
					{
						Opt: testNewCountSpanOption(0, 10),
					},
				},
			},
		}) {
			filters = append(filters, filter)
		}
		expectedAlerts := 4
		myAlerts := make([]*api.Alert, expectedAlerts)
		for i := range myAlerts {
			myAlerts[i] = &api.Alert{}
		}
		s.MockFilterDeleter().EXPECT().Enumerate(filters...).Return(myAlerts, nil).AnyTimes()

		// Call create with or without auth
		ctx := context.Background()
		var err error
		if tc.RequestAuthenticated {
			ctx, err = contextWithToken(ctx, "jim.stevens", "system.admin", "mysecret")
			assert.NoError(t, err)
		}
		enumerateClient, err := c.EnumerateWithFilters(ctx, req)
		assert.NoError(t, err)

		resp := &api.SdkAlertsEnumerateWithFiltersResponse{}
		for {
			r, err := enumerateClient.Recv()
			if err == io.EOF {
				break
			}
			if tc.ExpectSuccess {
				assert.NoError(t, err)
				resp.Alerts = append(resp.Alerts, r.Alerts...)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.ExpectedError, err.Error())
				break
			}
		}

		if tc.ExpectSuccess {
			assert.Len(t, resp.Alerts, expectedAlerts)
		}
	}
}
