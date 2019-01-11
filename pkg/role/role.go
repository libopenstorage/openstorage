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
	"github.com/libopenstorage/openstorage/api"
)

// RoleManager provides an implementation of the SDK Role handler
// and the necessary verification methods
type RoleManager interface {
	api.OpenStorageRoleServer

	// Verify returns no error if the role exists and is allowed
	// to run the requested method
	Verify(ctx context.Context, roles []string, method string) error
}
