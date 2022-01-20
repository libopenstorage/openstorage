package generictoken

import (
	"fmt"

	"github.com/libopenstorage/openstorage/pkg/auth"
)

// Config represents identifiers and information
// need to manage system tokens
type Config struct {
	Issuer       string
	Subject      string
	Name         string
	Email        string
	Roles        []string
	Groups       []string
	SharedSecret string
}

// manager provides access to tokens needed for node to
// node communication
type manager struct {
	config *Config
	claims *auth.Claims
}

var _ auth.TokenGenerator = &manager{}

// NewManager initializes the system token generator
func NewManager(cfg *Config) (auth.TokenGenerator, error) {
	if cfg == nil ||
		len(cfg.Issuer) == 0 ||
		len(cfg.Subject) == 0 ||
		len(cfg.Name) == 0 ||
		len(cfg.Email) == 0 ||
		len(cfg.Roles) == 0 ||
		len(cfg.Groups) == 0 ||
		len(cfg.SharedSecret) == 0 {
		return nil, fmt.Errorf("apps token generator must supply issuer, subject, name, email, roles, groups, and shared secret")
	}

	claims := &auth.Claims{
		Issuer:  cfg.Issuer,
		Subject: cfg.Subject,
		Name:    cfg.Name,
		Email:   cfg.Email,
		Roles:   cfg.Roles,
		Groups:  cfg.Groups,
	}

	return &manager{
		config: cfg,
		claims: claims,
	}, nil
}

// Issuer returns the token issuer for this generator necessary
// for registering the authenticator in the SDK.
func (m *manager) Issuer() string {
	return m.config.Issuer
}

// GetAuthenticator returns an authenticator for this issuer used by the SDK
func (m *manager) GetAuthenticator() (auth.Authenticator, error) {
	return auth.NewJwtAuth(&auth.JwtAuthConfig{
		SharedSecret: []byte(m.config.SharedSecret),
	})
}

// GetToken returns a token which can be used for
// authentication and communication from node to node.
func (m *manager) GetToken(opts *auth.Options) (string, error) {
	signature, err := auth.NewSignatureSharedSecret(m.config.SharedSecret)
	if err != nil {
		return "", err
	}

	return auth.Token(m.claims, signature, opts)
}
