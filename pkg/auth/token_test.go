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
	"context"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestTokenSharedSecretSimple(t *testing.T) {
	iss := "uid"
	key := []byte("mysecret")
	claims := Claims{
		Issuer:  iss,
		Email:   "my@email.com",
		Name:    "myname",
		Subject: "mysub",
		Roles:   []string{"tester"},
	}
	sig := Signature{
		Type: jwt.SigningMethodHS256,
		Key:  key,
	}
	opts := Options{
		Expiration: time.Now().Add(time.Minute * 10).Unix(),
	}

	// Create
	rawtoken, err := Token(&claims, &sig, &opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, rawtoken)

	// Verify
	token, err := jwt.Parse(rawtoken, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	assert.True(t, token.Valid)
	tokenClaims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Contains(t, tokenClaims, "email")
	assert.Equal(t, claims.Email, tokenClaims["email"])
	assert.Contains(t, tokenClaims, "name")
	assert.Equal(t, claims.Name, tokenClaims["name"])
	assert.Contains(t, tokenClaims, "roles")
	assert.Equal(t, claims.Roles[0], tokenClaims["roles"].([]interface{})[0].(string))

	// Get issuer from token
	newIss, err := TokenIssuer(rawtoken)
	assert.NoError(t, err)
	assert.Equal(t, newIss, iss)

	// Check if token is jwt
	assert.True(t, IsJwtToken(rawtoken))

	// Test authenticators
	authctr, err := NewJwtAuth(&JwtAuthConfig{
		SharedSecret:  key,
		UsernameClaim: UsernameClaimTypeDefault,
	})
	assert.NoError(t, err)

	authenticateClaims, err := authctr.AuthenticateToken(context.Background(), rawtoken)
	assert.NoError(t, err)
	assert.Equal(t, authenticateClaims.Issuer, claims.Issuer)
	assert.Equal(t, authenticateClaims.Email, claims.Email)
	assert.Equal(t, authenticateClaims.Name, claims.Name)
	assert.Equal(t, authenticateClaims.Roles, claims.Roles)
}

func TestTokenExpired(t *testing.T) {

	key := []byte("mysecret")
	claims := Claims{}
	sig := Signature{
		Type: jwt.SigningMethodHS256,
		Key:  key,
	}
	opts := Options{
		Expiration: time.Now().Add(-(time.Minute * 10)).Unix(),
	}

	// Create
	rawtoken, err := Token(&claims, &sig, &opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, rawtoken)

	// Verify
	token, err := jwt.Parse(rawtoken, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	assert.False(t, token.Valid)
}

func TestTokenIssuer(t *testing.T) {
	key := []byte("mysecret")
	issuer := "testiss"

	claims := Claims{
		Issuer: issuer,
	}
	sig := Signature{
		Type: jwt.SigningMethodHS256,
		Key:  key,
	}
	opts := Options{
		Expiration: time.Now().Add(-(time.Minute * 10)).Unix(),
	}

	// Create
	rawtoken, err := Token(&claims, &sig, &opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, rawtoken)

	parsedIssuer, err := TokenIssuer(rawtoken)
	assert.NoError(t, err)
	assert.Equal(t, issuer, parsedIssuer)
}

func TestTokenNTPDrift(t *testing.T) {
	goodTimeNow := time.Now()

	// Simulate a NTP drift of 15 seconds
	ntpDrift15Time := goodTimeNow.Add(-15 * time.Second)

	// Simulate a NTP drift of 90 seconds
	ntpDrift90Time := goodTimeNow.Add(-90 * time.Second)

	// Generate token
	key := []byte("mysecret")
	claims := Claims{}
	sig := Signature{
		Type: jwt.SigningMethodHS256,
		Key:  key,
	}
	opts := Options{
		Expiration:  time.Now().Add((time.Hour * 10)).Unix(),
		IATSubtract: time.Minute * 1,
	}

	// Create
	rawtoken, err := Token(&claims, &sig, &opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, rawtoken)

	// Verify
	token, err := jwt.Parse(rawtoken, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	assert.True(t, token.Valid)

	// Even though we have a 15 second NTP drift, our issue at time is valid.
	assert.True(t, token.Claims.(jwt.MapClaims)["iat"].(float64) < float64(ntpDrift15Time.Unix()))

	// We have a 90 second NTP drift, our issue at time is not valid.
	assert.False(t, token.Claims.(jwt.MapClaims)["iat"].(float64) < float64(ntpDrift90Time.Unix()))

}
