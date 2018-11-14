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
	"testing"

	sdk_auth "github.com/libopenstorage/openstorage-sdk-auth/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizeClaims(t *testing.T) {

	tests := []struct {
		denied     bool
		fullmethod string
		rules      []sdk_auth.Rule
		role       string
	}{
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Enumerate",
			role:       "admin",
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			role:       "admin",
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules:      []sdk_auth.Rule{},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules: []sdk_auth.Rule{
				{
					Services: []string{"futureservice"},
					Apis:     []string{"*"},
				},
			},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules: []sdk_auth.Rule{
				{
					Services: []string{"futureservice"},
					Apis:     []string{"anothercall"},
				},
			},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules: []sdk_auth.Rule{
				{
					Services: []string{"*"},
					Apis:     []string{"anothercall"},
				},
			},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules: []sdk_auth.Rule{
				{
					Services: []string{"cluster", "volume", "futureservice"},
					Apis:     []string{"somecallinthefuture"},
				},
			},
		},
	}

	for _, test := range tests {
		var rules []sdk_auth.Rule
		if len(test.role) != 0 {
			rules = defaultRoles[test.role]
		} else {
			rules = test.rules
		}

		err := authorizeClaims(rules, test.fullmethod)
		if test.denied {
			assert.NotNil(t, err, test.fullmethod, rules)
		} else {
			assert.Nil(t, err, test.fullmethod, rules)
		}
	}
}
