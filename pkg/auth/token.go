/*
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
	jwt "github.com/dgrijalva/jwt-go"
	"time"
)

// Rule provides a method to provide a custom authorization
// Rules can also be set to `*` to allow all services or apis.
type Rule struct {
	// Services is the gRPC service name in `OpenStorage<service name>` in lowercase
	Services []string `json:"services,omitempty" yaml:"services,omitempty"`
	// Apis is the API name in the service in lowercase
	Apis []string `json:"apis,omitempty" yaml:"apis,omitempty"`
}

// Claims provides information about the claims in the token
// See https://openid.net/specs/openid-connect-core-1_0.html#IDToken
// for more information.
type Claims struct {
	// Subject identifier. Unique ID of this account
	Subject string `json:"sub" yaml:"sub"`
	// Account name
	Name string `json:"name" yaml:"name"`
	// Account email
	Email string `json:"email" yaml:"email"`
	// (optional) Role of this account
	Role string `json:"role,omitempty" yaml:"role,omitempty"`
	// (optional) Groups in which this account is part of
	Groups []string `json:"groups,omitempty" yaml:"groups,omitempty"`
	// (optional) RBAC rules for the OpenStorage SDK
	// (DO NOT USE) This will be removed from the claims
	Rules []Rule `json:"rules,omitempty" yaml:"rules,omitempty"`
}

// Options provide any options to apply to the token
type Options struct {
	// Expiration time in Unix format as per JWT standard
	Expiration int64
	// Issuer of the claims
	Issuer string
}

// Token returns a signed JWT containing the claims provided
func Token(
	claims *Claims,
	signature *Signature,
	options *Options,
) (string, error) {

	mapclaims := jwt.MapClaims{
		"sub":   claims.Subject,
		"iss":   options.Issuer,
		"email": claims.Email,
		"name":  claims.Name,
		"role":  claims.Role,
		"iat":   time.Now().Unix(),
		"exp":   options.Expiration,
	}
	if claims.Groups != nil {
		mapclaims["groups"] = claims.Groups
	}
	if claims.Rules != nil {
		mapclaims["rules"] = claims.Rules
	}
	token := jwt.NewWithClaims(signature.Type, mapclaims)
	signedtoken, err := token.SignedString(signature.Key)
	if err != nil {
		return "", err
	}

	return signedtoken, nil
}
