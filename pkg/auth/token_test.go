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
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestTokenSharedSecretSimple(t *testing.T) {

	key := []byte("mysecret")
	claims := Claims{
		Email: "my@email.com",
		Name:  "myname",
		Roles: []string{"tester"},
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
