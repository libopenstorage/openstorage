package secrets

import (
	"errors"
	"sync"
)

var (
	// ErrNotImplemented default secrets in OSD
	ErrNotImplemented = errors.New("Not Implemented")
)

type Secrets interface {
	// Login create session with secret store
	Login(secretType string, secretConfig map[string]string) error
	// SetDefaultSecretKey  sets the cluster wide secret key
	SetDefaultSecretKey(secretKey string, override bool) error
	// GetDefaultSecretKey returns cluster wide secret key's value
	GetDefaultSecretKey() (interface{}, error)
	// CheckLogin validates session with secret store
	CheckLogin() error
	// Set the given value/data against the key
	Set(key string, value interface{}) error
	// Get retrieves the value/data for given key
	Get(key string) (interface{}, error)
}

type Manager struct {
	Secrets
	lock sync.Mutex
}

func NewSecretManager(sec Secrets) *Manager {
	return &Manager{
		Secrets: sec,
	}
}

type nullSecrets struct {
}

// New returns a new Default secrets implementation
func New() Secrets {
	return &nullSecrets{}
}

func (f *nullSecrets) Login(secretType string, secretConfig map[string]string) error {
	return ErrNotImplemented
}

func (f *nullSecrets) SetDefaultSecretKey(secretKey string, override bool) error {
	return ErrNotImplemented
}

func (f *nullSecrets) GetDefaultSecretKey() (interface{}, error) {
	return nil, ErrNotImplemented
}

func (f *nullSecrets) CheckLogin() error {
	return ErrNotImplemented
}

func (f *nullSecrets) Set(secretKey string, secretValue interface{}) error {
	return ErrNotImplemented
}

func (f *nullSecrets) Get(secretKey string) (interface{}, error) {
	return nil, ErrNotImplemented
}
