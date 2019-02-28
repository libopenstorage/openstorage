package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClaims(t *testing.T) {
	email, name, subject := "a@b.com", "hello", "123"
	claims := &Claims{
		Email:   email,
		Name:    name,
		Subject: subject,
	}

	// Check getUsername for correctness
	un := getUsername(UsernameClaimTypeEmail, claims)
	assert.Equal(t, un, email)
	un = getUsername(UsernameClaimTypeName, claims)
	assert.Equal(t, un, name)
	un = getUsername(UsernameClaimTypeSubject, claims)
	assert.Equal(t, un, subject)
}
