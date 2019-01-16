/*
Package ownership manages access to resources
Copyright 2019 Portworx

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
package api

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/libopenstorage/openstorage/pkg/auth"
)

func TestOwnershipIsPermitted(t *testing.T) {
	tests := []struct {
		owner     *Ownership
		user      *auth.UserInfo
		permitted bool
	}{
		{
			owner: &Ownership{
				Owner: "me",
			},
			permitted: false,
		},
		{
			owner: &Ownership{
				Owner: "me",
			},
			user: &auth.UserInfo{
				Username: "notme",
			},
			permitted: false,
		},
		{
			owner: &Ownership{
				Owner: "me",
			},
			user: &auth.UserInfo{
				Username: "me",
			},
			permitted: true,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups: []string{},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
			},
			permitted: false,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups: []string{},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{},
				},
			},
			permitted: false,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups: []string{"*"},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{},
				},
			},
			permitted: true,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups: []string{"*"},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{"group1", "group2"},
				},
			},
			permitted: true,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group2"},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{"group1", "group2"},
				},
			},
			permitted: true,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group3"},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{"group1", "group2"},
				},
			},
			permitted: false,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group3"},
					Collaborators: []string{},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{"group1", "group2"},
				},
			},
			permitted: false,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group3"},
					Collaborators: []string{"*"},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{"group1", "group2"},
				},
			},
			permitted: true,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group3"},
					Collaborators: []string{"user1", "user2", "user3"},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{"group1", "group2"},
				},
			},
			permitted: false,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group3"},
					Collaborators: []string{"user1", "user2", "user3", "notme"},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{"group1", "group2"},
				},
			},
			permitted: true,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group3"},
					Collaborators: []string{"user1", "user2", "user3"},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{"group1", "group2", AdminGroup},
				},
			},
			permitted: true,
		},
		{
			owner: &Ownership{
				Owner: "me",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group3"},
					Collaborators: []string{"user1", "user2", "user3"},
				},
			},
			user: &auth.UserInfo{
				Username: "notme",
				Claims: auth.Claims{
					Groups: []string{"*"},
				},
			},
			permitted: true,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.owner.IsPermitted(test.user), test.permitted)
	}
}

func TestOwnershipUpdate(t *testing.T) {

	tests := []struct {
		owner  *Ownership
		update *Ownership
		result *Ownership
		user   *auth.UserInfo
	}{
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group1"},
				}},
			update: &Ownership{
				Owner: "user2",
				Acls: &Ownership_AccessControl{
					Collaborators: []string{"user1", "user2", "user3"},
				},
			},
			result: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Collaborators: []string{"user1", "user2", "user3"},
				},
			},
			user: &auth.UserInfo{
				Username: "user1",
				Claims: auth.Claims{
					Groups: []string{"group"},
				},
			},
		},
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group1"},
				}},
			update: &Ownership{
				Owner: "user2",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group1", "group2"},
					Collaborators: []string{"user1", "user2", "user3"},
				},
			},
			result: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group1", "group2"},
					Collaborators: []string{"user1", "user2", "user3"},
				},
			},
			user: &auth.UserInfo{
				Username: "user1",
				Claims: auth.Claims{
					Groups: []string{"group"},
				},
			},
		},
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group1"},
				}},
			update: &Ownership{
				Owner: "user2",
				Acls:  &Ownership_AccessControl{},
			},
			result: &Ownership{
				Owner: "user2",
				Acls:  &Ownership_AccessControl{},
			},
			user: &auth.UserInfo{
				Claims: auth.Claims{
					Groups: []string{"*"},
				},
			},
		},
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group1"},
				}},
			update: &Ownership{
				Acls: &Ownership_AccessControl{},
			},
			result: &Ownership{
				Owner: "user1",
				Acls:  &Ownership_AccessControl{},
			},
			user: &auth.UserInfo{
				Claims: auth.Claims{
					Groups: []string{"*"},
				},
			},
		},
	}

	for _, test := range tests {
		r := test.owner.Update(test.update, test.user)
		assert.True(t, reflect.DeepEqual(r, test.result), fmt.Sprintf("%v | %v", r, test.result))
	}
}

func TestOwnershipIsMatch(t *testing.T) {
	tests := []struct {
		owner   *Ownership
		match   *Ownership
		isMatch bool
	}{
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group1"},
				},
			},
			match: &Ownership{
				Owner: "user2",
				Acls:  &Ownership_AccessControl{},
			},
			isMatch: false,
		},
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group1"},
				},
			},
			isMatch: false,
		},
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group1"},
				},
			},
			match: &Ownership{
				Owner: "user2",
			},
			isMatch: false,
		},
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group1"},
				},
			},
			match: &Ownership{
				Owner: "user2",
				Acls: &Ownership_AccessControl{
					Groups: []string{"group1"},
				},
			},
			isMatch: true,
		},
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group1"},
					Collaborators: []string{"one", "two", "three"},
				},
			},
			match: &Ownership{
				Owner: "user2",
				Acls: &Ownership_AccessControl{
					Groups: []string{"one"},
				},
			},
			isMatch: false,
		},
		{
			owner: &Ownership{
				Owner: "user1",
				Acls: &Ownership_AccessControl{
					Groups:        []string{"group1"},
					Collaborators: []string{"one", "two", "three"},
				},
			},
			match: &Ownership{
				Owner: "user2",
				Acls: &Ownership_AccessControl{
					Collaborators: []string{"three"},
				},
			},
			isMatch: true,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.owner.IsMatch(test.match), test.isMatch)
	}
}
