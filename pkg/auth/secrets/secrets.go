package secrets

import (
	"errors"
	"fmt"

	osecrets "github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/k8s"
)

const (
	// SecretNameKey is a label on the openstorage.Volume object
	// which corresponds to the name of the secret which holds the
	// token information. Used for all secret providers
	SecretNameKey = "openstorage.io/auth-secret-name"

	// SecretNamespaceKey is a label on the openstorage.Volume object
	// which corresponds to the namespeace of the secret which holds the
	// token information. Used for all secret providers
	SecretNamespaceKey = "openstorage.io/auth-secret-namespace"

	// SecretTokenKey corresponds to the key at which the auth token is stored
	// in the secret. Used when secrets endpoint is kubernetes secrets
	SecretTokenKey = "auth-token"
)

var (
	// ErrSecretsNotInitialized is returned when the auth secret provider is not initialized
	ErrSecretsNotInitialized = errors.New("auth token secret instance not initialized")
	// ErrAuthTokenNotFound is returned when the auth token was not found in the secret
	ErrAuthTokenNotFound = errors.New("auth token was not found in the configured secret")
)

// Auth interface provides helper routines to fetch authorization tokens
// from a secrets store
type Auth interface {
	// GetToken returns the auth token obtained from the secret with
	// secretName with the provided secretContext from the configured secretContext
	GetToken(secretName string, secretContext string) (string, error)
	// Type returns the type of AuthTokenProvider
	Type() AuthTokenProviders
}

// AuthTokenProviders is an enum indicating the type of secret store that is storing
// the auth token
type AuthTokenProviders int

const (
	TypeNone AuthTokenProviders = iota
	TypeK8s
	TypeDCOS
)

// NewAuth returns a new instance of Auth implementation
func NewAuth(
	authProviderType AuthTokenProviders,
	s osecrets.Secrets,
) (Auth, error) {
	if s == nil {
		return nil, ErrSecretsNotInitialized
	}
	switch authProviderType {
	case TypeK8s:
		return &k8sAuth{s}, nil
	case TypeDCOS:
		return &dcosAuth{s}, nil
	}
	return nil, fmt.Errorf("secrets type %v not supported", authProviderType)
}

// Kubernetes as the auth token secrets provider

type k8sAuth struct {
	s osecrets.Secrets
}

func (k *k8sAuth) GetToken(secretName string, secretContext string) (string, error) {
	keyContext := make(map[string]string)
	keyContext[k8s.SecretNamespace] = secretContext

	secretValue, err := k.s.GetSecret(secretName, keyContext)
	if err != nil {
		return "", err
	}
	authToken, exists := secretValue[SecretTokenKey]
	if !exists {
		return "", ErrAuthTokenNotFound
	}
	return authToken.(string), nil
}

func (k *k8sAuth) Type() AuthTokenProviders {
	return TypeK8s
}

// DCOS as the auth token secrets provider

type dcosAuth struct {
	s osecrets.Secrets
}

func (d *dcosAuth) GetToken(secretName string, secretContext string) (string, error) {
	key := secretName
	if secretContext != "" {
		key = secretContext + "/" + secretName
	}
	secretValue, err := d.s.GetSecret(key, nil)
	if err != nil {
		return "", err
	}
	authToken, exists := secretValue[key]
	if !exists {
		return "", ErrAuthTokenNotFound
	}
	return authToken.(string), nil
}

func (d *dcosAuth) Type() AuthTokenProviders {
	return TypeDCOS
}
