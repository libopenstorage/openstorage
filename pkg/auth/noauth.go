package auth

import (
	"fmt"
)

// Default no auth
type noauth struct{}

func (s *noauth) Issuer() string {
	return ""
}

func (s *noauth) GetAuthenticator() (*JwtAuthenticator, error) {
	return nil, fmt.Errorf("No authentication set")
}

func (s *noauth) GetSystemToken() (string, error) {
	return "", nil
}
