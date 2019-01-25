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
	"context"

	"github.com/libopenstorage/openstorage/pkg/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// AdminGroup is the value that can be set in the token claims Group which
	// gives the user access to any resource
	AdminGroup = "*"
)

// OwnershipSetUsernameFromContext is used to create a new ownership object for
// a volume. It takes an ownership value if passed in by the user, then
// sets the `owner` value to the user name referred to in the user context
func OwnershipSetUsernameFromContext(ctx context.Context, srcOwnership *Ownership) *Ownership {
	// Check if the context has information about the user. If not,
	// then security is not enabled.
	if userinfo, ok := auth.NewUserInfoFromContext(ctx); ok {

		// Merge the previous acls which may have been set by the user
		var acls *Ownership_AccessControl
		if srcOwnership != nil {
			acls = srcOwnership.GetAcls()
		}

		return &Ownership{
			Owner: userinfo.Username,
			Acls:  acls,
		}
	}

	return srcOwnership
}

// NewLocatorOwnershipFromContext is used by the Enumerate functions as a
// simple way to filter for volumes only accessible by the user
func NewLocatorOwnershipFromContext(ctx context.Context) *Ownership {
	// Check if the context has information about the user. If not,
	// then security is not enabled.
	if userinfo, ok := auth.NewUserInfoFromContext(ctx); ok {
		return &Ownership{
			Owner: userinfo.Username,
			Acls: &Ownership_AccessControl{
				Groups:        userinfo.Claims.Groups,
				Collaborators: []string{userinfo.Username},
			},
		}
	}

	return nil
}

// IsPermitted returns true if the user has access to the resource
// according to the ownership. If there is no owner, then it is public
func (o *Ownership) IsPermitted(user *auth.UserInfo) bool {
	// There is no owner, so it is a public resource
	if !o.HasAnOwner() {
		return true
	}

	// If we are missing user information then do not allow.
	// It is ok for the the user claims to have an empty Groups setting
	if user == nil ||
		len(user.Username) == 0 {
		return false
	}

	if o.IsOwner(user) ||
		o.IsUserAllowedByGroup(user) ||
		o.IsUserAllowedByCollaborators(user) {
		return true
	}

	return false
}

// GetGroups returns the groups in the ownership
func (o *Ownership) GetGroups() []string {
	if o.GetAcls() == nil {
		return nil
	}
	return o.GetAcls().GetGroups()
}

// GetCollaborators returns the collaborators in the ownership
func (o *Ownership) GetCollaborators() []string {
	if o.GetAcls() == nil {
		return nil
	}
	return o.GetAcls().GetCollaborators()
}

// IsUserAllowedByGroup returns true if the user is allowed access
// by belonging to the appropriate group
func (o *Ownership) IsUserAllowedByGroup(user *auth.UserInfo) bool {

	// If it is the admin user for any group
	if o.IsAdminByUser(user) {
		return true
	}

	ownergroups := o.GetGroups()
	if len(ownergroups) == 0 {
		return false
	}

	// Check if any group is allowed
	if listContains(ownergroups, "*") {
		return true
	}

	// Check each of the groups from the user
	for _, group := range user.Claims.Groups {
		// Check if the user is in the admin group and can access
		// any resource
		if group == AdminGroup {
			return true
		}

		// Check if the user group has permission
		if listContains(ownergroups, group) {
			return true
		}
	}
	return false
}

// IsUserAllowedByCollaborators returns true if the user is allowed access
// because they are part of the collaborators list
func (o *Ownership) IsUserAllowedByCollaborators(user *auth.UserInfo) bool {
	collaborators := o.GetCollaborators()
	if len(collaborators) == 0 {
		return false
	}

	// Check any user is allowed
	if listContains(collaborators, "*") {
		return true
	}

	// Check each of the groups from the user
	return listContains(collaborators, user.Username)
}

// HasAnOwner returns true if the resource has an owner
func (o *Ownership) HasAnOwner() bool {
	return len(o.Owner) != 0
}

// IsPublic returns true if there is no ownership in this resource
func (o *Ownership) IsPublic() bool {
	return !o.HasAnOwner()
}

// IsOwner returns if the user is the owner of the resource
func (o *Ownership) IsOwner(user *auth.UserInfo) bool {
	return o.Owner == user.Username
}

// IsAdminByUser returns true if the user is an ownership admin, meaning,
// that they belong to any group
func (o *Ownership) IsAdminByUser(user *auth.UserInfo) bool {

	// If there is a user, then auth is enabled
	if user != nil {
		return listContains(user.Claims.Groups, AdminGroup)
	}

	// No auth enabled, so everyone is an admin
	return true
}

// Update can be used to update an ownership with new ownership information. It
// takes into account who is trying to change the ownership values
func (o *Ownership) Update(newownerInfo *Ownership, user *auth.UserInfo) error {
	if user == nil {
		// There is no auth, just copy the whole thing
		o = newownerInfo
	} else {
		// Auth is enabled

		// Only the owner or admin can change the group
		if user.Username != o.Owner && !o.IsAdminByUser(user) {
			return status.Error(codes.PermissionDenied,
				"Only owner can update volume acls")
		}
		o.Acls = newownerInfo.GetAcls()

		// Only the admin can change the owner
		if o.IsAdminByUser(user) && len(newownerInfo.Owner) != 0 {
			o.Owner = newownerInfo.Owner
		}
	}
	return nil
}

// IsMatch returns true if the ownership has at least one similar
// owner, group, or collaborator
func (o *Ownership) IsMatch(check *Ownership) bool {
	if check == nil {
		return false
	}

	// Check user
	if o.Owner == check.GetOwner() {
		return true
	}
	if check.GetAcls() == nil || o.GetAcls() == nil {
		return false
	}

	// Check groups
	for _, group := range check.GetAcls().GetGroups() {
		if listContains(o.GetAcls().GetGroups(), group) {
			return true
		}
	}

	// Check collaborators
	for _, collaborator := range check.GetAcls().GetCollaborators() {
		if listContains(o.GetAcls().GetCollaborators(), collaborator) {
			return true
		}
	}

	return false
}

func listContains(list []string, s string) bool {
	for _, value := range list {
		if value == s {
			return true
		}
	}
	return false
}
