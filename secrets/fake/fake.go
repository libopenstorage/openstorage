package fake

import (
	"errors"
	"github.com/libopenstorage/openstorage/secrets"
)

var (
	// ErrNotImplemented default secrets in OSD
	ErrNotImplemented = errors.New("Not Implemented")
)

type Fake struct {
	secrets.SecretManager
}

// New returns a new Fake secret implementation
func New() *Fake {
	return &Fake{}
}

func (f *Fake) SecretLogin(secretType int, secretConfig map[string]string) error {
	return ErrNotImplemented
}

func (f *Fake) SetClusterSecretKey(secretKey string, override bool) error {
	return ErrNotImplemented
}

func (f *Fake) CheckSecretLogin() error {
	return ErrNotImplemented
}

func (f *Fake) SetSecret(secretKey string, secretValue string) error {
	return ErrNotImplemented
}

func (f *Fake) GetSecret(secretKey string) (string, error) {
	return "", ErrNotImplemented
}
