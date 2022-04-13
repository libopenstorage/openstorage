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
	"strings"

	"github.com/portworx/kvdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
)

const (
	rolePrefix   = "cluster/roles"
	invalidChars = "/ "
	negMatchChar = "!"

	SystemAdminRoleName = "system.admin"
	SystemViewRoleName  = "system.view"
	SystemUserRoleName  = "system.user"
	SystemGuestRoleName = "system.guest"
)

// DefaultRole is a role loaded into the system on startup
type DefaultRole struct {
	Rules   []*api.SdkRule
	Mutable bool
}

var (
	// DefaultRoles are the default roles to load on system startup
	// Should be prefixed by `system.` to avoid collisions
	DefaultRoles = map[string]*DefaultRole{
		// system:admin role can run any command
		SystemAdminRoleName: &DefaultRole{
			Rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"*"},
					Apis:     []string{"*"},
				},
			},
			Mutable: false,
		},

		// system:view role can only run read-only commands
		SystemViewRoleName: &DefaultRole{
			Rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"*"},
					Apis: []string{
						"*enumerate*",
						"inspect*",
						"stats",
						"status",
						"validate",
						"capacityusage",
					},
				},
				&api.SdkRule{
					Services: []string{"identity"},
					Apis:     []string{"*"},
				},
			},
			Mutable: false,
		},
		// system:user role can only access volume lifecycle commands
		SystemUserRoleName: &DefaultRole{
			Rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{
						"volume",
						"cloudbackup",
						"credentials",
						"objectstore",
						"schedulepolicy",
						"mountattach",
						"migrate",
					},
					Apis: []string{"*"},
				},
				&api.SdkRule{
					Services: []string{
						"cluster",
						"node",
					},
					Apis: []string{
						"inspect*",
						"enumerate*",
					},
				},
				&api.SdkRule{
					Services: []string{"identity"},
					Apis:     []string{"*"},
				},
				&api.SdkRule{
					Services: []string{"policy"},
					Apis: []string{
						"*enumerate*",
						// This will allow system.user to view default policy also
						"*inspect*",
					},
				},
			},
			Mutable: false,
		},

		// system:guest role is used for any unauthenticated user.
		// They can only use standard volume lifecycle commands.
		SystemGuestRoleName: &DefaultRole{
			Rules: []*api.SdkRule{
				&api.SdkRule{
					Services: []string{"mountattach", "volume", "cloudbackup", "migrate"},
					Apis:     []string{"*"},
				},
				&api.SdkRule{
					Services: []string{"identity"},
					Apis:     []string{"version"},
				},
				&api.SdkRule{
					Services: []string{
						"cluster",
						"node",
					},
					Apis: []string{
						"inspect*",
						"enumerate*",
					},
				},
			},
			Mutable: true,
		},
	}
)

// SdkRoleManager is an implementation of the RoleManager for the SDK
type SdkRoleManager struct {
	kv kvdb.Kvdb
}

// Check interface
var _ RoleManager = &SdkRoleManager{}

// Simple function which creates key for Kvdb
func prefixWithName(name string) string {
	return rolePrefix + "/" + name
}

// Determines if the rules deny string s
func denyRule(rule, s string) bool {
	if strings.HasPrefix(rule, negMatchChar) {
		return matchRule(strings.TrimSpace(strings.Join(strings.Split(rule, negMatchChar), "")), s)
	}

	return false
}

// Determines if the rules apply to string s
// rule can be:
// '*' - match all
// '*xxx' - ends with xxx
// 'xxx*' - starts with xxx
// '*xxx*' - contains xxx
func matchRule(rule, s string) bool {
	// no rule
	rl := len(rule)
	if rl == 0 {
		return false
	}

	// '*xxx' || 'xxx*'
	if rule[0:1] == "*" || rule[rl-1:rl] == "*" {
		// get the matching string from the rule
		match := strings.TrimSpace(strings.Join(strings.Split(rule, "*"), ""))

		// '*' or '*******'
		if len(match) == 0 {
			return true
		}

		// '*xxx*'
		if rule[0:1] == "*" && rule[rl-1:rl] == "*" {
			return strings.Contains(s, match)
		}

		// '*xxx'
		if rule[0:1] == "*" {
			return strings.HasSuffix(s, match)
		}

		// 'xxx*'
		return strings.HasPrefix(s, match)
	}

	// no wildcard stars given in rule
	return rule == s
}

// NewSdkRoleManager returns a new SDK role manager
func NewSdkRoleManager(kv kvdb.Kvdb) (*SdkRoleManager, error) {
	s := &SdkRoleManager{
		kv: kv,
	}

	// Load all default roles
	for roleName, defaultRole := range DefaultRoles {
		roleExists := false
		if _, err := kv.Get(prefixWithName(roleName)); err == nil {
			roleExists = true
		}

		// always re-initialize immutable default roles.
		// if the role is mutable and does exist, skip kvdb put.
		if !roleExists || !defaultRole.Mutable {
			role := &api.SdkRole{
				Name:  roleName,
				Rules: defaultRole.Rules,
			}
			if _, err := kv.Put(prefixWithName(roleName), role, 0); err != nil {
				return nil, err
			}
		}
	}

	return s, nil
}

// Create saves a role in Kvdb
func (r *SdkRoleManager) Create(
	ctx context.Context,
	req *api.SdkRoleCreateRequest,
) (*api.SdkRoleCreateResponse, error) {
	if req.GetRole() == nil {
		return nil, status.Error(codes.InvalidArgument, "Must supply a role")
	} else if len(req.GetRole().GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a name for the role")
	} else if err := r.validateRole(req.GetRole()); err != nil {
		return nil, err
	}

	// Determine if there is collision with default roles
	if _, ok := DefaultRoles[req.GetRole().GetName()]; ok {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Name %s already used by system role", req.GetRole().GetName())
	}

	// Save value in kvdb
	_, err := r.kv.Create(prefixWithName(req.GetRole().GetName()), req.GetRole(), 0)
	if err == kvdb.ErrExist {
		// Idempotency check.
		// Check that the new rules are the same.
		oldrole, err := r.Inspect(ctx, &api.SdkRoleInspectRequest{
			Name: req.GetRole().GetName(),
		})
		if err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(oldrole.GetRole(), req.GetRole()) {
			return nil, status.Error(
				codes.AlreadyExists,
				"Existing role differs from requested role")
		}
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save role: %v", err)
	}

	return &api.SdkRoleCreateResponse{
		Role: req.GetRole(),
	}, nil
}

// Enumerate returns a list of role names
func (r *SdkRoleManager) Enumerate(
	ctx context.Context,
	req *api.SdkRoleEnumerateRequest,
) (*api.SdkRoleEnumerateResponse, error) {
	kvPairs, err := r.kv.Enumerate(rolePrefix + "/")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to access roles from database: %v", err)
	}

	var names []string
	for _, kvPair := range kvPairs {
		names = append(names, strings.TrimPrefix(kvPair.Key, rolePrefix+"/"))
	}

	return &api.SdkRoleEnumerateResponse{
		Names: names,
	}, nil
}

// Inspect returns a role object
func (r *SdkRoleManager) Inspect(
	ctx context.Context,
	req *api.SdkRoleInspectRequest,
) (*api.SdkRoleInspectResponse, error) {
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a name for role")
	}

	elem := &api.SdkRole{}
	_, err := r.kv.GetVal(prefixWithName(req.GetName()), elem)
	if err == kvdb.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "Role %s not found", req.GetName())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get role %s information: %v", req.GetName(), err)
	}

	return &api.SdkRoleInspectResponse{
		Role: elem,
	}, nil
}

// Delete removes a role from Kvdb
func (r *SdkRoleManager) Delete(
	ctx context.Context,
	req *api.SdkRoleDeleteRequest,
) (*api.SdkRoleDeleteResponse, error) {
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a name for role")
	}

	// Determine if there is collision with default roles
	if _, ok := DefaultRoles[req.GetName()]; ok {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Cannot delete system role %s", req.GetName())
	}

	_, err := r.kv.Delete(prefixWithName(req.GetName()))
	if err != kvdb.ErrNotFound && err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete role %s: %v", req.GetName(), err)
	}

	return &api.SdkRoleDeleteResponse{}, nil
}

// Update replaces an existing role.
func (r *SdkRoleManager) Update(
	ctx context.Context,
	req *api.SdkRoleUpdateRequest,
) (*api.SdkRoleUpdateResponse, error) {
	if req.GetRole() == nil {
		return nil, status.Error(codes.InvalidArgument, "Must supply a role")
	}
	if err := r.validateRole(req.GetRole()); err != nil {
		return nil, err
	}

	// Determine if there is collision with default roles.
	// We can still update mutable default roles.
	if defaultRole, ok := DefaultRoles[req.GetRole().GetName()]; ok && !defaultRole.Mutable {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"System role %s cannot be updated", req.GetRole().GetName())
	}

	_, err := r.kv.Update(prefixWithName(req.GetRole().GetName()), req.GetRole(), 0)
	if err == kvdb.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "Role %s not found", req.GetRole())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get role %s information: %v", req.GetRole().GetName(), err)
	}

	return &api.SdkRoleUpdateResponse{
		Role: req.GetRole(),
	}, nil
}

// Verify determines if the role has access to `fullmethod`
func (r *SdkRoleManager) Verify(ctx context.Context, roles []string, fullmethod string) error {

	// Check all roles
	for _, role := range roles {
		// Get the role rules
		resp, err := r.Inspect(ctx, &api.SdkRoleInspectRequest{
			Name: role,
		})
		if err != nil || resp == nil || resp.GetRole() == nil {
			continue
		}

		if err := VerifyRules(resp.GetRole().GetRules(), api.SdkRootPath, fullmethod); err == nil {
			return nil
		}
	}

	return status.Errorf(codes.PermissionDenied, "Access denied to roles: %+s", roles)
}

// VerifyRules checks if the rules authorize use of the API called `fullmethod`
func VerifyRules(rules []*api.SdkRule, rootPath, fullmethod string) error {

	reqService, reqApi := grpcserver.GetMethodInformation(rootPath, fullmethod)

	// Look for denials first
	for _, rule := range rules {
		for _, service := range rule.Services {
			// if the service is denied, then return here
			if denyRule(service, reqService) {
				return fmt.Errorf("access denied to service by role")
			}

			// If there is a match to the service now check the apis
			if matchRule(service, reqService) {
				for _, api := range rule.Apis {
					if denyRule(api, reqApi) {
						return fmt.Errorf("access denied to api by role")
					}
				}
			}
		}
	}

	// Look for permissions
	for _, rule := range rules {
		for _, service := range rule.Services {
			if matchRule(service, reqService) {
				for _, api := range rule.Apis {
					if matchRule(api, reqApi) {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("no accessible rule to authorize access found")
}

func (r *SdkRoleManager) validateRole(role *api.SdkRole) error {
	if len(role.GetName()) == 0 {
		return status.Error(codes.InvalidArgument, "Must supply a name for role")
	} else if len(role.GetRules()) == 0 {
		return status.Error(codes.InvalidArgument, "Must supply rules for role")
	} else if strings.ContainsAny(role.GetName(), invalidChars) {
		return status.Errorf(codes.InvalidArgument, "Name cannot contain any of the following invalid characters: %s", invalidChars)
	}

	// Check if the rules have services and apis
	for _, rule := range role.GetRules() {
		if len(rule.GetServices()) == 0 {
			return status.Error(codes.InvalidArgument, "Must supply a service in rules for the role")
		} else if len(rule.GetApis()) == 0 {
			return status.Error(codes.InvalidArgument, "Must supply apis in rules for the role")
		}
	}

	return nil
}
