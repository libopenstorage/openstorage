package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoAuth(t *testing.T) {
	assert.False(t, Enabled())
	assert.Empty(t, inst.Issuer())

	a, err := inst.GetAuthenticator()
	assert.Nil(t, a)
	assert.Error(t, err)

	token, err := inst.GetSystemToken()
	assert.Empty(t, token)
	assert.NoError(t, err)
}

func TestSystemTokenGenerator(t *testing.T) {
	oldinst := inst
	defer func() {
		inst = oldinst
	}()
	assert.False(t, Enabled())
	claims := &Claims{
		Issuer:  "uid",
		Subject: "456",
		Name:    "Internal cluster communication",
		Email:   "abcd@abcd.com",
		Roles:   []string{"system.admin"},
		Groups:  []string{"*"},
	}
	config := &Config{
		ClusterUuid: "uid",
		NodeId:      "id",
		Claims:      claims,
		Secret:      "mysecret",
	}
	stg, err := Init(config)
	assert.NoError(t, err)
	assert.Equal(t, stg.Issuer(), config.ClusterUuid)
	assert.True(t, Enabled())

	token, err := stg.GetSystemToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, strings.Count(token, "."), 2)
	issuer, err := TokenIssuer(token)
	assert.NoError(t, err)
	assert.Equal(t, issuer, stg.Issuer())
}
