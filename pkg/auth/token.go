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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Options provide any options to apply to the token
type Options struct {
	// Expiration time in Unix format as per JWT standard
	Expiration int64
}

// TokenClaims returns the claims for the raw JWT token.
func TokenClaims(rawtoken string) (*Claims, error) {
	parts := strings.Split(rawtoken, ".")

	// There are supposed to be three parts for the token
	if len(parts) < 3 {
		return nil, fmt.Errorf("Token is invalid: %v", rawtoken)
	}

	// Access claims in the token
	claimBytes, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return nil, fmt.Errorf("Failed to decode claims: %v", err)
	}
	var claims *Claims

	// Unmarshal claims
	err = json.Unmarshal(claimBytes, &claims)
	if err != nil {
		return nil, fmt.Errorf("Unable to get information from the claims in the token: %v", err)
	}

	return claims, nil
}

// TokenIssuer returns the issuer for the raw JWT token.
func TokenIssuer(rawtoken string) (string, error) {
	claims, err := TokenClaims(rawtoken)
	if err != nil {
		return "", err
	}

	// Return issuer
	if len(claims.Issuer) != 0 {
		return claims.Issuer, nil
	} else {
		return "", fmt.Errorf("Issuer was not specified in the token")
	}
}

// IsJwtToken returns true if the provided string is a valid jwt token
func IsJwtToken(authstring string) bool {
	_, _, err := new(jwt.Parser).ParseUnverified(authstring, jwt.MapClaims{})
	return err == nil
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
