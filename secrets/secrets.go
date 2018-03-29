package secrets

import "sync"

type SecretManager interface {
	// Login create session with secret store
	SecretLogin(secretType int, secretConfig map[string]string) error
	// SetClusterSecretKey sets the cluster wide secret key
	SetClusterSecretKey(secretKey string, override bool) error
	//CheckSecretLogin validates session with secret store
	CheckSecretLogin() error
	//SetSecret sets secret key against data
	SetSecret(key string, value string) error
	// GetSecret retrieves the data for the given feature and key
	GetSecret(key string) (string, error)
}

type Manager struct {
	Secret SecretManager
	lock   sync.Mutex
}

func NewSecretManager(sec SecretManager) *Manager {
	return &Manager{
		Secret: sec,
	}
}
