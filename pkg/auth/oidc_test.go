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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOidcParseClaims(t *testing.T) {
	o := &OIDCAuthenticator{}

	// create a claims
	claims := map[string]interface{}{
		"name":  "Test",
		"email": "test@test.com",
		"sub":   "Subject",
		"roles": []interface{}{
			"role.1",
			"role.2",
		},
		"groups": []interface{}{
			"group.1",
			"group.2",
		},
	}
	sdkClaims, err := o.parseClaims(claims)
	assert.NoError(t, err)
	assert.Equal(t, "Test", sdkClaims.Name)
	assert.Equal(t, "test@test.com", sdkClaims.Email)
	assert.Equal(t, "Subject", sdkClaims.Subject)
	assert.Equal(t, []string{"role.1", "role.2"}, sdkClaims.Roles)
	assert.Equal(t, []string{"group.1", "group.2"}, sdkClaims.Groups)
}

func TestOidcParseClaimsWithNamespace(t *testing.T) {
	namespace := "https://namespace/"
	o := &OIDCAuthenticator{
		namespace: namespace,
	}

	// create a claims
	claims := map[string]interface{}{
		"name":  "Test",
		"email": "test@test.com",
		"sub":   "Subject",
		namespace + "roles": []interface{}{
			"role.1",
			"role.2",
		},
		namespace + "groups": []interface{}{
			"group.1",
			"group.2",
		},
		"roles":  "bad",
		"groups": 12345,
	}
	sdkClaims, err := o.parseClaims(claims)
	assert.NoError(t, err)
	assert.Equal(t, "Test", sdkClaims.Name)
	assert.Equal(t, "test@test.com", sdkClaims.Email)
	assert.Equal(t, "Subject", sdkClaims.Subject)
	assert.Equal(t, []string{"role.1", "role.2"}, sdkClaims.Roles)
	assert.Equal(t, []string{"group.1", "group.2"}, sdkClaims.Groups)
}
