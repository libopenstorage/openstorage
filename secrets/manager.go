package secrets

import "sync"

type SecretsManager interface {
	// Register registers a new secret implementation
	Register(name string, secrets Secrets)
	// Get returns a secret implementation
	Get(name string) (Secrets, error)
}

type secretsRegistry struct {
	secrets map[string]Secrets
	lock    *sync.RWMutex
}

func (s *secretsRegistry) Register(name string, secrets Secrets) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.secrets[name]; ok {
		return ErrExists
	}
	s.secrets[name] = secrets
	return nil
}

func (s *secretsRegistry) Get(name string) (Secrets, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	secretsImplementor, ok := s.secrets[name]
	if !ok {
		return nil, ErrSecretsNotFound
	}
	return secretsImplementor, nil
}
