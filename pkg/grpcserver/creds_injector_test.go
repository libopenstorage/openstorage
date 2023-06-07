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
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var mySigningKey = []byte("superSecretKey")

func token(t *testing.T, expiresAt int64) string {
	t.Helper()

	claims := &jwt.StandardClaims{
		ExpiresAt: expiresAt,
		Issuer:    "test",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		t.Fatalf("cannot generate token")
	}
	return ss
}

func TestGetRequestMetadata_GetTokenSuccessful(t *testing.T) {
	ci := NewCredsInjector(func() (string, error) {
		return token(t, time.Now().Add(time.Hour).Unix()), nil
	}, true)
	m, err := ci.GetRequestMetadata(context.TODO())
	if err != nil {
		t.Fatalf("should not return error, err: %s", err.Error())
	}
	v, exists := m["authorization"]
	if !exists || v == "" {
		t.Fatal("did not get authorization token")
	}
}

func TestGetRequestMetadata_TokenShouldBeRegeneratedSuccessfulWhenExpired(t *testing.T) {
	ci := NewCredsInjector(func() (string, error) {
		exp := time.Now().Add(time.Second).Unix()
		t.Logf("%v", exp)
		return token(t, exp), nil
	}, true)
	m, err := ci.GetRequestMetadata(context.TODO())
	if err != nil {
		t.Fatalf("should not return error, err: %s", err.Error())
	}
	v, exists := m["authorization"]
	if !exists || v == "" {
		t.Fatal("did not get authorization token")
	}

	time.Sleep(time.Second * 2)

	m2, err := ci.GetRequestMetadata(context.TODO())
	if err != nil {
		t.Fatalf("should not return error, err: %s", err.Error())
	}
	v2, exists := m2["authorization"]
	if !exists || v == "" {
		t.Fatal("did not get authorization token")
	}

	if v == v2 {
		t.Fatal("token was not regenerated")
	}
}

func TestGetRequestMetadata_TokenShouldNotBeRegeneratedWhenNotExpired(t *testing.T) {
	ci := NewCredsInjector(func() (string, error) {
		exp := time.Now().Add(time.Hour * 1000).Unix()
		t.Logf("%v", exp)
		return token(t, exp), nil
	}, true)
	m, err := ci.GetRequestMetadata(context.TODO())
	if err != nil {
		t.Fatalf("should not return error, err: %s", err.Error())
	}
	v, exists := m["authorization"]
	if !exists || v == "" {
		t.Fatal("did not get authorization token")
	}

	time.Sleep(time.Second * 1)

	m2, err := ci.GetRequestMetadata(context.TODO())
	if err != nil {
		t.Fatalf("should not return error, err: %s", err.Error())
	}
	v2, exists := m2["authorization"]
	if !exists || v == "" {
		t.Fatal("did not get authorization token")
	}

	if v != v2 {
		t.Fatal("token should not have been regenerated")
	}
}
