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
	jwt "github.com/dgrijalva/jwt-go"
	"time"
)

// Options provide any options to apply to the token
type Options struct {
	// Expiration time in Unix format as per JWT standard
	Expiration int64
}

// Token returns a signed JWT containing the claims provided
func Token(
	claims *Claims,
	signature *Signature,
	options *Options,
) (string, error) {

	mapclaims := jwt.MapClaims{
		"sub":   claims.Subject,
		"iss":   claims.Issuer,
		"email": claims.Email,
		"name":  claims.Name,
		"roles": claims.Roles,
		"iat":   time.Now().Unix(),
		"exp":   options.Expiration,
	}
	if claims.Groups != nil {
		mapclaims["groups"] = claims.Groups
	}
	token := jwt.NewWithClaims(signature.Type, mapclaims)
	signedtoken, err := token.SignedString(signature.Key)
	if err != nil {
		return "", err
	}

	return signedtoken, nil
}
