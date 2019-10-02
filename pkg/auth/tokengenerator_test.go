package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoAuth(t *testing.T) {
	// Check for interface implementation
	var na TokenGenerator = &noauth{}

	// Get no auth Instance
	na = NoAuth()
	assert.NotNil(t, na)

	assert.Equal(t, na.Issuer(), "")
	authctr, err := na.GetAuthenticator()
	assert.Error(t, err)
	assert.Nil(t, authctr)
	token, err := na.GetToken(&Options{})
	assert.NoError(t, err)
	assert.Equal(t, token, "")
}
