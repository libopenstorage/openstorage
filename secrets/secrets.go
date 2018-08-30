package secrets

import (
	"errors"
)

var (
	// ErrNotImplemented default secrets in OSD
	ErrNotImplemented = errors.New("Not Implemented")
	// ErrKeyEmpty returned when the secrety key provided is empty
	ErrKeyEmpty = errors.New("Secret key cannot be empty")
	// ErrNotAuthenticated returned when not authenticated with secrets endpoint
	ErrNotAuthenticated = errors.New("Not authenticated with the secrets endpoint")
	// ErrInvalidSecretId returned when no secret data is found associated with the id
	ErrInvalidSecretId = errors.New("No Secret Data found for Secret Id")
	// ErrEmptySecretData returned when no secret data is provided to store the secret
	ErrEmptySecretData = errors.New("Secret data cannot be empty")
	// ErrSecretExists returned when a secret for the given secret id already exists
	ErrSecretExists = errors.New("Secret ID already exists")
)

type Secrets interface {
	// SecretLogin create session with secret store
	SecretLogin(secretType string, secretConfig map[string]string) error
	// SecretSetDefaultSecretKey  sets the cluster wide secret key
	SecretSetDefaultSecretKey(secretKey string, override bool) error
	// SecretGetDefaultSecretKey returns cluster wide secret key
	SecretGetDefaultSecretKey() (interface{}, error)
	// SecretCheckLogin validates session with secret store
	SecretCheckLogin() error
	// SecretSet the given value/data against the key
	SecretSet(key string, value interface{}) error
	// SecretGet retrieves the value/data for given key
	SecretGet(key string) (interface{}, error)
}

type NullSecrets struct {
}

// New returns null secrets implementation
func NewDefaultSecrets() Secrets {
	return &NullSecrets{}
}

func (f *NullSecrets) SecretLogin(secretType string, secretConfig map[string]string) error {
	return ErrNotImplemented
}

func (f *NullSecrets) SecretSetDefaultSecretKey(secretKey string, override bool) error {
	return ErrNotImplemented
}

func (f *NullSecrets) SecretGetDefaultSecretKey() (interface{}, error) {
	return nil, ErrNotImplemented
}

func (f *NullSecrets) SecretCheckLogin() error {
	return ErrNotImplemented
}

func (f *NullSecrets) SecretSet(secretKey string, secretValue interface{}) error {
	return ErrNotImplemented
}

func (f *NullSecrets) SecretGet(secretKey string) (interface{}, error) {
	return nil, ErrNotImplemented
}
