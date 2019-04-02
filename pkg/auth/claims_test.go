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

func TestClaims(t *testing.T) {
	email, name, subject := "a@b.com", "hello", "123"
	claims := &Claims{
		Email:   email,
		Name:    name,
		Subject: subject,
	}

	// Check getUsername for correctness
	un := getUsername(UsernameClaimTypeEmail, claims)
	assert.Equal(t, un, email)
	un = getUsername(UsernameClaimTypeName, claims)
	assert.Equal(t, un, name)
	un = getUsername(UsernameClaimTypeSubject, claims)
	assert.Equal(t, un, subject)
}

func TestValidateUsername(t *testing.T) {
	email, name, subject := "a@b.com", "hello", "123"
	goodClaims := &Claims{
		Email:   email,
		Name:    name,
		Subject: subject,
	}
	badClaims := &Claims{
		Email:   "",
		Name:    "",
		Subject: "",
	}

	typesToTest := []UsernameClaimType{
		UsernameClaimTypeEmail,
		UsernameClaimTypeName,
		UsernameClaimTypeSubject,
		UsernameClaimTypeDefault,
	}
	for _, unType := range typesToTest {
		err := validateUsername(unType, goodClaims)
		assert.NoError(t, err)

		err = validateUsername(unType, badClaims)
		assert.Error(t, err)
	}
}
