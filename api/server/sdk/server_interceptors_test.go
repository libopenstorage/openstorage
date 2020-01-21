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
	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizationServerInterceptorCreate(t *testing.T) {
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
			ExpectedError: "rpc error: code = PermissionDenied desc = Public volume creation is disabled",
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
		fmt.Printf("Running %s\n", tc.TestName)
		s.server.config.Security.PublicVolumeCreationDisabled = tc.PublicVolumeCreationDisabled

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
				Create(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(id, nil).
				AnyTimes(),
		)

		// Setup client
		c := api.NewOpenStorageVolumeClient(s.Conn())

		// Call create with or without auth
		ctx := context.Background()
		var err error
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
