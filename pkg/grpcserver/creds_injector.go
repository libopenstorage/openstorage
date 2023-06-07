/*
Package grpcserver is a generic gRPC server manager
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
package grpcserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var minsBeforeExpiration = time.Minute * 5

// CredsInjector implements credentials.PerRPCCredentials interface
type CredsInjector struct {
	tokenGenerator  func() (string, error)
	m               sync.Mutex
	currentToken    string
	currentTokenExp int64
	tlsEnabled      bool
}

// GetRequestMetadata checks JWT token expiration time and invokes token generator function to get new token if that is needed
func (i *CredsInjector) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	i.m.Lock()
	defer i.m.Unlock()
	inAFewMins := time.Now().Unix() + int64(minsBeforeExpiration.Seconds())
	if i.currentToken == "" || i.currentTokenExp == 0 || inAFewMins >= i.currentTokenExp {
		token, err := i.tokenGenerator()
		if err != nil {
			return nil, fmt.Errorf("cannot generate token, err: %s", err.Error())
		}
		i.currentToken = token

		t, _, err := new(jwt.Parser).ParseUnverified(i.currentToken, &jwt.StandardClaims{})
		if err != nil {
			return nil, fmt.Errorf("failed to parse authorization token: %s", err.Error())
		}

		claims, ok := t.Claims.(*jwt.StandardClaims)
		if !ok {
			return nil, fmt.Errorf("failed to get token claims")
		}

		i.currentTokenExp = claims.ExpiresAt
	}

	return map[string]string{
		"authorization": "Bearer " + i.currentToken,
	}, nil
}

func (i *CredsInjector) RequireTransportSecurity() bool {
	return i.tlsEnabled
}

// ResetToken makes CredsInjector to invoke token generation next time when GetRequestMetadata is invoked
func (i *CredsInjector) ResetToken() {
	i.m.Lock()
	i.currentToken = ""
	i.currentTokenExp = 0
	i.m.Unlock()
}

// NewCredsInjector creates gRPC interceptor to inject Authorization token in requests
func NewCredsInjector(generator func() (string, error), tlsEnabled bool) *CredsInjector {
	return &CredsInjector{
		tokenGenerator: generator,
		tlsEnabled:     tlsEnabled,
	}
}
