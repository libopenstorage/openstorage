/*
Package default contains the contant definitions for default
roles provided by Openstorage.

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

package defaults

import "github.com/libopenstorage/openstorage/api"

const (
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
	// Roles are the default roles to load on system startup
	// Should be prefixed by `system.` to avoid collisions
	Roles = map[string]*DefaultRole{
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
