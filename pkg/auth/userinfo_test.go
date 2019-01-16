/*
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
package auth

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfoContext(t *testing.T) {
	original := &UserInfo{
		Username: "username",
		Claims: Claims{
			Subject: "subject",
			Name:    "name",
			Email:   "email",
			Roles:   []string{"one", "two"},
			Groups:  []string{"three", "four"},
		},
	}
	ctx := ContextSaveUserInfo(context.Background(), original)
	u, ok := NewUserInfoFromContext(ctx)
	assert.True(t, ok)
	assert.True(t, reflect.DeepEqual(u, original))
}
