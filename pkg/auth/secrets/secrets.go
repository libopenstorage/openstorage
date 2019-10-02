package secrets

import (
	"errors"

	"github.com/libopenstorage/openstorage/api"
	lsecrets "github.com/libopenstorage/secrets"
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

// GetToken returns the token for a given secret name and context
// based on the configured auth secrets provider.
func GetToken(tokenSecretContext *api.TokenSecretContext) (string, error) {
	var inputSecretKey string
	var outputSecretKey string
	secretName := tokenSecretContext.SecretName
	secretsInst := lsecrets.Instance()

	// Handle edge cases for different providers.
	switch secretsInst.String() {
	case lsecrets.TypeDCOS:
		inputSecretKey = tokenSecretContext.SecretName
		namespace := tokenSecretContext.SecretNamespace
		if namespace != "" {
			inputSecretKey = namespace + "/" + secretName
		}
		outputSecretKey = inputSecretKey

	case lsecrets.TypeK8s:
		inputSecretKey = tokenSecretContext.SecretName
		outputSecretKey = SecretTokenKey

	default:
		inputSecretKey = tokenSecretContext.SecretName
		outputSecretKey = SecretNameKey
	}

	// Get secret value with standardized interface
	secretValue, err := secretsInst.GetSecret(inputSecretKey, requestToContext(tokenSecretContext))
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

func requestToContext(request *api.TokenSecretContext) map[string]string {
	context := make(map[string]string)

	// Add namespace for providers that support it.
	switch lsecrets.Instance().String() {
	case lsecrets.TypeK8s:
		context[k8s.SecretNamespace] = request.SecretNamespace
	}

	return context
}
