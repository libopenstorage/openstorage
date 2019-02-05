package auth

import (
	"fmt"
	"time"
)

// SystemTokenGenerator provides access to tokens needed for node to
// node communication
type SystemTokenGenerator interface {

	// Issuer returns the token issuer for this generator necessary
	// for registering the authenticator in the SDK.
	Issuer() string

	// GetAuthenticator returns an authenticator for this issuer used by the SDK
	GetAuthenticator() (*JwtAuthenticator, error)

	// GetSystemToken returns a token which can be used for
	// authentication and communication from node to node.
	GetSystemToken() (string, error)
}

type Config struct {
	ClusterUuid string
	NodeId      string
	Claims      *Claims
	Secret      string
}

type systemTokenGenerator struct {
	clusteruuid string
	nodeId      string
	claims      *Claims
	secret      string
}

var _ SystemTokenGenerator = &systemTokenGenerator{}
var _ SystemTokenGenerator = &noauth{}

var (
	inst SystemTokenGenerator = &noauth{}

	// Inst returns an instance of an already instantiated object.
	// This function can be overridden for testing purposes
	Inst = func() (SystemTokenGenerator, error) {
		return systemTokenGeneratorInst()
	}
)

func systemTokenGeneratorInst() (SystemTokenGenerator, error) {
	return inst, nil
}

// Init initializes the system token generator
func Init(c *Config) (SystemTokenGenerator, error) {

	if c == nil ||
		len(c.ClusterUuid) == 0 ||
		len(c.NodeId) == 0 ||
		c.Claims == nil ||
		len(c.Secret) == 0 {
		return nil, fmt.Errorf("Must supply claims, clusterUuid, nodeId, and system secret")
	}

	inst = &systemTokenGenerator{
		clusteruuid: c.ClusterUuid,
		nodeId:      c.NodeId,
		claims:      c.Claims,
		secret:      c.Secret,
	}
	return inst, nil
}

func (s *systemTokenGenerator) Issuer() string {
	return s.clusteruuid
}

func (s *systemTokenGenerator) GetAuthenticator() (*JwtAuthenticator, error) {
	return NewJwtAuth(&JwtAuthConfig{
		SharedSecret: []byte(s.secret),
	})
}

func (s *systemTokenGenerator) GetSystemToken() (string, error) {
	options := &Options{
		Expiration: time.Now().Add(5 * Year).Unix(),
	}
	signature, err := NewSignatureSharedSecret(s.secret)
	if err != nil {
		return "", err
	}

	return Token(s.claims, signature, options)
}
