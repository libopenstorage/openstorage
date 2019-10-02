package systemtoken

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestSystemTokenGenerator(t *testing.T) {
	config := &Config{
		ClusterId:    "uid",
		NodeId:       "id",
		SharedSecret: "mysecret",
	}
	stm, err := NewManager(config)
	assert.NoError(t, err)
	assert.Equal(t, stm.Issuer(), config.ClusterId)

	token, err := stm.GetToken(&auth.Options{
		Expiration: time.Now().Add(5 * auth.Year).Unix(),
	})
	assert.NotEmpty(t, token)
	assert.Equal(t, strings.Count(token, "."), 2)
	issuer, err := auth.TokenIssuer(token)
	assert.NoError(t, err)
	assert.Equal(t, issuer, stm.Issuer())
}

func TestNewManager(t *testing.T) {
	_, err := NewManager(&Config{
		ClusterId:    "",
		NodeId:       "id",
		SharedSecret: "mysecret",
	})
	assert.Error(t, err)
}

func TestGetToken(t *testing.T) {
	secret := "hivailable"

	stm, err := NewManager(&Config{
		ClusterId:    "abcd",
		NodeId:       "id",
		SharedSecret: secret,
	})
	assert.NoError(t, err)
	authctr, err := stm.GetAuthenticator()
	assert.NoError(t, err)

	token, err := stm.GetToken(&auth.Options{
		Expiration: time.Now().Add(5 * auth.Year).Unix(),
	})
	assert.NoError(t, err)

	isToken := auth.IsJwtToken(token)
	assert.True(t, isToken)

	_, err = authctr.AuthenticateToken(context.Background(), token)
	assert.NoError(t, err)
}
