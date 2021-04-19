/*
Package role manages roles in Kvdb and provides validation
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
package role

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/stretchr/testify/assert"

	"github.com/libopenstorage/openstorage/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPrefixWithName(t *testing.T) {
	assert.Equal(t, prefixWithName("hello"), rolePrefix+"/"+"hello")
}

func TestMatchDenyRule(t *testing.T) {

	tests := []struct {
		matchFound bool
		role       string
		s          string
	}{
		{
			matchFound: false,
			role:       "",
			s:          "",
		},
		{
			matchFound: true,
			role:       "!*",
			s:          "test",
		},
		{
			matchFound: true,
			role:       "!test",
			s:          "test",
		},
		{
			matchFound: true,
			role:       "!!!!!!!!!!!!!*******************test****",
			s:          "test",
		},
		{
			matchFound: false,
			role:       "test",
			s:          "test",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.matchFound, denyRule(test.role, test.s))
	}
}

func TestMatchRule(t *testing.T) {

	tests := []struct {
		matchFound bool
		role       string
		s          string
	}{
		{
			matchFound: false,
			role:       "",
			s:          "",
		},
		{
			matchFound: true,
			role:       "*",
			s:          "test",
		},
		{
			matchFound: true,
			role:       "***********",
			s:          "test",
		},
		{
			matchFound: false,
			role:       "nomatch",
			s:          "test",
		},
		{
			matchFound: false,
			role:       "*nomatch",
			s:          "test",
		},
		{
			matchFound: false,
			role:       "nomatch*",
			s:          "test",
		},
		{
			matchFound: false,
			role:       "*nomatch*",
			s:          "test",
		},
		{
			matchFound: true,
			role:       "*test",
			s:          "thisisatest",
		},
		{
			matchFound: true,
			role:       "this*",
			s:          "thisisatest",
		},
		{
			matchFound: true,
			role:       "*isa*",
			s:          "thisisatest",
		},
		{
			matchFound: false,
			role:       "isa",
			s:          "thisisatest",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.matchFound, matchRule(test.role, test.s))
	}
}

func TestSdkRuleCreateBadArguments(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	req := &api.SdkRoleCreateRequest{}
	_, err = s.Create(context.Background(), req)
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "role")

	req.Role = &api.SdkRole{}
	_, err = s.Create(context.Background(), req)
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "name")

	req.Role = &api.SdkRole{
		Name: "hello",
	}
	_, err = s.Create(context.Background(), req)
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "rules")

	// No service
	req.Role = &api.SdkRole{
		Name: "helloworld",
		Rules: []*api.SdkRule{
			{
				Services: []string{},
			},
		},
	}
	_, err = s.Create(context.Background(), req)
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "service in rules")

	// No APIs
	req.Role = &api.SdkRole{
		Name: "helloworld",
		Rules: []*api.SdkRule{
			{
				Services: []string{"hello"},
				Apis:     []string{},
			},
		},
	}
	_, err = s.Create(context.Background(), req)
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "apis in rules")

	// Name with invalid characters
	for _, ic := range invalidChars {
		req.Role = &api.SdkRole{
			Name: fmt.Sprintf("hello%cworld", ic),
			Rules: []*api.SdkRule{
				{
					Services: []string{"service"},
					Apis:     []string{"api"},
				},
			},
		}
		_, err = s.Create(context.Background(), req)
		assert.Error(t, err)
		serverError, ok = status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, serverError.Code(), codes.InvalidArgument)
		assert.Contains(t, serverError.Message(), "invalid")
	}
}

func TestSdkRuleCreateCollisionSystemRole(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	for roleName, defaultRole := range DefaultRoles {
		req := &api.SdkRoleCreateRequest{
			Role: &api.SdkRole{
				Name:  roleName,
				Rules: defaultRole.Rules,
			},
		}
		_, err := s.Create(context.Background(), req)
		assert.Error(t, err)
		serverError, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, serverError.Code(), codes.InvalidArgument)
		assert.Contains(t, serverError.Message(), "system role")
	}

}

func TestSdkRuleCreate(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	name := "test.volume"
	req := &api.SdkRoleCreateRequest{
		Role: &api.SdkRole{
			Name: name,
			Rules: []*api.SdkRule{
				{
					Services: []string{"volume", "credentials", "cloudbackup"},
					Apis:     []string{"*"},
				},
				{
					Services: []string{"identity"},
					Apis:     []string{"*"},
				},
			},
		},
	}
	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	// Assert idempotency
	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	// Assert the information is in kvdb
	var elem *api.SdkRole
	_, err = kv.GetVal(prefixWithName(name), &elem)
	assert.NoError(t, err)
	assert.Equal(t, elem.GetName(), name)
	assert.Len(t, elem.Rules, len(req.GetRole().GetRules()))
	assert.True(t, reflect.DeepEqual(elem.GetRules(), req.GetRole().GetRules()))

	// Assert that creating the same name with different roles fails
	req.Role.Rules = append(req.Role.Rules, &api.SdkRule{
		Services: []string{"hello"},
		Apis:     []string{"world"},
	})
	_, err = s.Create(context.Background(), req)
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.AlreadyExists)
	assert.Contains(t, serverError.Message(), "role differs")
}

func TestSdkRuleEnumerate(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	req := &api.SdkRoleCreateRequest{
		Role: &api.SdkRole{
			Name: "one",
			Rules: []*api.SdkRule{
				{
					Services: []string{"service"},
					Apis:     []string{"api"},
				},
			},
		},
	}
	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	req.Role.Name = "two"
	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	r, err := s.Enumerate(context.Background(), &api.SdkRoleEnumerateRequest{})
	assert.NoError(t, err)
	assert.Len(t, r.GetNames(), 2+len(DefaultRoles))
	assert.Contains(t, r.GetNames(), "one")
	assert.Contains(t, r.GetNames(), "two")
}

func TestSdkRuleInspectBadArgument(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	// Not name passed in
	_, err = s.Inspect(context.Background(), &api.SdkRoleInspectRequest{})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "name")

	// Name does not exist
	_, err = s.Inspect(context.Background(), &api.SdkRoleInspectRequest{
		Name: "doesnotexist",
	})
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")
}

func TestSdkRuleInspectDelete(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	req := &api.SdkRoleCreateRequest{
		Role: &api.SdkRole{
			Name: "one",
			Rules: []*api.SdkRule{
				{
					Services: []string{"service"},
					Apis:     []string{"api"},
				},
			},
		},
	}
	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	r, err := s.Inspect(context.Background(), &api.SdkRoleInspectRequest{
		Name: "one",
	})
	assert.NoError(t, err)
	assert.NotNil(t, r.GetRole())
	assert.Equal(t, r.GetRole().GetName(), "one")
	assert.Contains(t, r.GetRole().GetName(), "one")
	assert.Len(t, r.GetRole().GetRules(), len(req.GetRole().GetRules()))
	assert.True(t, reflect.DeepEqual(r.GetRole().GetRules(), req.GetRole().GetRules()))

	_, err = s.Delete(context.Background(), &api.SdkRoleDeleteRequest{
		Name: "one",
	})
	assert.NoError(t, err)

	_, err = s.Inspect(context.Background(), &api.SdkRoleInspectRequest{
		Name: "one",
	})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")
}

func TestSdkRuleDeleteCollisionSystemRole(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	for systemRole, _ := range DefaultRoles {
		req := &api.SdkRoleDeleteRequest{
			Name: systemRole,
		}
		_, err := s.Delete(context.Background(), req)
		assert.Error(t, err)
		serverError, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, serverError.Code(), codes.InvalidArgument)
		assert.Contains(t, serverError.Message(), "system role")
	}

}

func TestSdkRuleUpdateBadArguments(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	req := &api.SdkRoleUpdateRequest{}
	_, err = s.Update(context.Background(), req)
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "role")

	req.Role = &api.SdkRole{}
	_, err = s.Update(context.Background(), req)
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "name")

	req.Role = &api.SdkRole{
		Name: "hello",
	}
	_, err = s.Update(context.Background(), req)
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "rules")

	// key does not exist
	req = &api.SdkRoleUpdateRequest{
		Role: &api.SdkRole{
			Name: "one",
			Rules: []*api.SdkRule{
				{
					Services: []string{"service"},
					Apis:     []string{"api"},
				},
			},
		},
	}

	_, err = s.Update(context.Background(), req)
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")
}

func TestSdkRuleUpdateCollisionSystemRole(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	for roleName, defaultRole := range DefaultRoles {
		req := &api.SdkRoleUpdateRequest{
			Role: &api.SdkRole{
				Name:  roleName,
				Rules: defaultRole.Rules,
			},
		}
		_, err := s.Update(context.Background(), req)
		if defaultRole.Mutable {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
			serverError, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, serverError.Code(), codes.InvalidArgument)
			assert.Contains(t, serverError.Message(), "System role")
		}
	}

}

func TestSdkRuleUpdate(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	role := &api.SdkRole{
		Name: "test",
		Rules: []*api.SdkRule{
			{
				Services: []string{"service"},
				Apis:     []string{"api"},
			},
		},
	}

	_, err = s.Create(context.Background(), &api.SdkRoleCreateRequest{
		Role: role,
	})
	assert.NoError(t, err)

	// Assert the information is in kvdb
	role.Rules = append(role.Rules, &api.SdkRule{
		Services: []string{"test"},
		Apis:     []string{"test2"},
	})
	_, err = s.Update(context.Background(), &api.SdkRoleUpdateRequest{
		Role: role,
	})
	assert.NoError(t, err)

	// Check db
	var elem *api.SdkRole
	_, err = kv.GetVal(prefixWithName("test"), &elem)
	assert.NoError(t, err)
	assert.Equal(t, elem.GetName(), "test")
	assert.Len(t, elem.Rules, len(role.GetRules()))
	assert.True(t, reflect.DeepEqual(elem.GetRules(), role.GetRules()))
}

func TestSdkRoleVerifyRules(t *testing.T) {

	tests := []struct {
		denied     bool
		fullmethod string
		rules      []*api.SdkRule
		roles      []string
	}{
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Enumerate",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"*"},
					Apis:     []string{"!enumerate"},
				},
			},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Enumerate",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"!volumes"},
					Apis:     []string{"*"},
				},
			},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Create",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"volumes"},
					Apis:     []string{"*", "!create"},
				},
			},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Create",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"volumes"},
					Apis:     []string{"!create", "*"},
				},
			},
		},
		{
			// Denials have more priority
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Create",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"volumes"},
					Apis:     []string{"!*", "create"},
				},
			},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Create",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"*", "!*"},
					Apis:     []string{"*"},
				},
			},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Create",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"*"},
					Apis:     []string{"*", "!*"},
				},
			},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Create",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"volumes"},
					Apis:     []string{"*"},
				},
			},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageVolumes/Enumerate",
			roles:      []string{"system.admin"},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			roles:      []string{"system.admin"},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules:      []*api.SdkRule{},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"futureservice"},
					Apis:     []string{"*"},
				},
			},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"futureservice"},
					Apis:     []string{"anothercall"},
				},
			},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"*"},
					Apis:     []string{"anothercall"},
				},
			},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"cluster", "volume", "futureservice"},
					Apis:     []string{"somecallinthefuture"},
				},
			},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFutureEnumerate",
			roles:      []string{"system.view"},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageFutureService/SomeCallInTheFuture",
			roles:      []string{"system.view"},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageCluster/InspectCurrent",
			roles:      []string{"system.guest"},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageNode/Enumerate",
			roles:      []string{"system.guest"},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageNode/Inspect",
			roles:      []string{"system.guest"},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageCluster/InspectCurrent",
			roles:      []string{"system.user"},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageNode/Enumerate",
			roles:      []string{"system.user"},
		},
		{
			denied:     false,
			fullmethod: "/openstorage.api.OpenStorageNode/Inspect",
			roles:      []string{"system.user"},
		},
		{
			denied:     true,
			fullmethod: "/openstorage.api.OpenStorageNode/FutureCall",
			roles:      []string{"system.user"},
		},
	}

	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)

	s, err := NewSdkRoleManager(kv)
	assert.NoError(t, err)

	for _, test := range tests {
		var err error
		if len(test.roles) != 0 {
			err = s.Verify(context.Background(), test.roles, test.fullmethod)
		} else {
			err = VerifyRules(test.rules, api.SdkRootPath, test.fullmethod)
		}

		if test.denied {
			assert.NotNil(t, err, test.fullmethod, fmt.Sprintf("%v", test))
		} else {
			assert.Nil(t, err, test.fullmethod)
		}
	}
}
