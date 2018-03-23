package secrets

import "errors"

var (
	ErrSecretsNotFound = errors.New("Secrets Implementor not registerd")
	ErrExists          = errors.New("Already Exists")
)

type Secrets interface {
	// Login create session with secret store
	SecretLogin(secretType int, secretConfig map[string]string) error
	// SetClusterSecretKey sets the cluster wide secret key
	SetClusterSecretKey(secretKey string, override bool) error
	//CheckSecretLogin validates session with secret store
	CheckSecretLogin() error
}
