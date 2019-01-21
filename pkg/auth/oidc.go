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
	"fmt"

	oidc "github.com/coreos/go-oidc"
)

// OIDCAuthConfig configures an OIDC connection
type OIDCAuthConfig struct {
	// Issuer of the OIDC tokens
	// e.g. https://accounts.google.com
	Issuer string
	// ClientID is the client id provided by the OIDC
	ClientID string
	// SkipClientIDCheck skips a verification on tokens which are returned
	// from the OIDC without the client ID set
	SkipClientIDCheck bool
	// UsernameClaim has the location of the unique id for the user.
	// If empty, "sub" will be used for the user name unique id.
	UsernameClaim UsernameClaimType
}

// OIDCAuthenticator is used to validate tokens with an OIDC
type OIDCAuthenticator struct {
	url           string
	provider      *oidc.Provider
	verifier      *oidc.IDTokenVerifier
	usernameClaim UsernameClaimType
}

// NewOIDC returns a new OIDC authenticator
func NewOIDC(config *OIDCAuthConfig) (*OIDCAuthenticator, error) {
	p, err := oidc.NewProvider(context.Background(), config.Issuer)
	if err != nil {
		return nil, fmt.Errorf("Unable to communicate with ODIC provider %s: %v",
			config.Issuer,
			err)
	}

	v := p.Verifier(&oidc.Config{
		ClientID:          config.ClientID,
		SkipClientIDCheck: config.SkipClientIDCheck,
	})
	return &OIDCAuthenticator{
		url:           config.Issuer,
		usernameClaim: config.UsernameClaim,
		provider:      p,
		verifier:      v,
	}, nil
}

func (o *OIDCAuthenticator) AuthenticateToken(ctx context.Context, rawtoken string) (*Claims, error) {
	idToken, err := o.verifier.Verify(ctx, rawtoken)
	if err != nil {
		return nil, fmt.Errorf("Token failed validation: %v", err)
	}

	// Check for required claims
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("Unable to get claim map from token: %v", err)
	}
	for _, requiredClaim := range requiredClaims {
		if _, ok := claims[requiredClaim]; !ok {
			// Claim missing
			return nil, fmt.Errorf("Required claim %v missing from token", requiredClaim)
		}
	}

	// Return claims
	var sdkClaims Claims
	if err := idToken.Claims(&sdkClaims); err != nil {
		return nil, fmt.Errorf("Unable to get claims from token: %v", err)
	}

	return &sdkClaims, nil
}

func (o *OIDCAuthenticator) Username(claims *Claims) string {
	return getUsername(o.usernameClaim, claims)
}
