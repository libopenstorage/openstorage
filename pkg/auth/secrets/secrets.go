package secrets

import (
	"errors"
	"fmt"

	"github.com/libopenstorage/openstorage/api"
	osecrets "github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/k8s"
)

const (
	// SecretNameKey is a label on the openstorage.Volume object
	// which corresponds to the name of the secret which holds the
	// token information. Used for all secret providers
	SecretNameKey = "openstorage.io/auth-secret-name"

	// SecretNamespaceKey is a label on the openstorage.Volume object
	// which corresponds to the namespace of the secret which holds the
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
type Auth struct {
	ProviderClient osecrets.Secrets
	ProviderType   AuthTokenProviders
}

// AuthTokenProviders is an enum indicating the type of secret store that is storing
// the auth token
type AuthTokenProviders int

const (
	TypeNone AuthTokenProviders = iota
	TypeK8s
	TypeDCOS
	TypeVault
	TypeKVDB
	TypeDocker
)

// NewAuth returns a new instance of Auth implementation
func NewAuth(
	p AuthTokenProviders,
	s osecrets.Secrets,
) (*Auth, error) {
	if s == nil {
		return nil, ErrSecretsNotInitialized
	}

	switch p {
	case TypeK8s:
		return &Auth{s, p}, nil
	case TypeDCOS:
		return &Auth{s, p}, nil
	case TypeVault:
		return &Auth{s, p}, nil
	case TypeKVDB:
		return &Auth{s, p}, nil
	case TypeDocker:
		return &Auth{s, p}, nil
	}

	return nil, fmt.Errorf("secrets type %v not supported", p)
}

// GetToken returns the token for a given secret name and context
// based on the configured auth secrets provider.
func (a *Auth) GetToken(tokenSecretContext *api.TokenSecretContext) (string, error) {
	var inputSecretKey string
	var outputSecretKey string
	secretName := tokenSecretContext.SecretName

	// Handle edge cases for different providers.
	switch a.ProviderType {
	case TypeDCOS:
		inputSecretKey = tokenSecretContext.SecretName
		namespace := tokenSecretContext.SecretNamespace
		if namespace != "" {
			inputSecretKey = namespace + "/" + secretName
		}
		outputSecretKey = inputSecretKey

	case TypeK8s:
		inputSecretKey = tokenSecretContext.SecretName
		outputSecretKey = SecretTokenKey

	default:
		inputSecretKey = tokenSecretContext.SecretName
		outputSecretKey = SecretNameKey
	}

	// Get secret value with standardized interface
	secretValue, err := a.ProviderClient.GetSecret(inputSecretKey, a.requestToContext(tokenSecretContext))
	if err != nil {
		return "", err
	}

	// Retrieve auth token
	authToken, exists := secretValue[outputSecretKey]
	if !exists {
		return "", ErrAuthTokenNotFound
	}
	return authToken.(string), nil

}

func (a *Auth) requestToContext(request *api.TokenSecretContext) map[string]string {
	context := make(map[string]string)

	// Add namespace for providers that support it.
	switch a.ProviderType {
	case TypeK8s:
		context[k8s.SecretNamespace] = request.SecretNamespace
	}

	return context
}

func (a *Auth) Type() AuthTokenProviders {
	return a.ProviderType
}
