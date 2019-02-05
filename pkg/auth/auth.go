/*
Package auth can be used for authentication and authorization
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
package auth

import (
	"context"
)

// UsernameClaimType holds the claims type to be use as the unique id for the user
type UsernameClaimType string

const (
	// default type is sub
	UsernameClaimTypeDefault UsernameClaimType = ""
	// UsernameClaimTypeSubject requests to use "sub" as the claims for the
	// ID of the user
	UsernameClaimTypeSubject UsernameClaimType = "sub"
	// UsernameClaimTypeEmail requests to use "name" as the claims for the
	// ID of the user
	UsernameClaimTypeEmail UsernameClaimType = "email"
	// UsernameClaimTypeName requests to use "name" as the claims for the
	// ID of the user
	UsernameClaimTypeName UsernameClaimType = "name"
)

var (
	// Required claim keys
	requiredClaims = []string{"iss", "sub", "exp", "iat", "name", "email"}
)

// Authenticator interface validates and extracts the claims from a raw token
type Authenticator interface {
	// AuthenticateToken validates the token and returns the claims
	AuthenticateToken(context.Context, string) (*Claims, error)

	// Username returns the unique id according to the configuration. Default
	// it will return the value for "sub" in the token claims, but it can be
	// configured to return the email or name as the unique id.
	Username(*Claims) string
}

// utility function to get username
func getUsername(usernameClaim UsernameClaimType, claims *Claims) string {
	switch usernameClaim {
	case UsernameClaimTypeEmail:
		return claims.Email
	case UsernameClaimTypeName:
		return claims.Name
	}
	return claims.Subject
}

// Enabled returns if authentication is enabled in the system.
// If node-node authentication is enabled, then system authentication is enabled.
func Enabled() bool {
	return len(inst.Issuer()) != 0
}
