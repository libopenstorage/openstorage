package k8s

import (
	"fmt"

	"github.com/libopenstorage/secrets"
	"github.com/portworx/sched-ops/k8s/core"
)

const (
	Name            = secrets.TypeK8s
	SecretNamespace = "namespace"
)

type k8sSecrets struct{}

func New(
	secretConfig map[string]interface{},
) (secrets.Secrets, error) {
	return &k8sSecrets{}, nil
}

func (s *k8sSecrets) String() string {
	return Name
}

func (s *k8sSecrets) GetSecret(
	secretName string,
	keyContext map[string]string,
) (map[string]interface{}, secrets.Version, error) {
	namespace, exists := keyContext[SecretNamespace]
	if !exists {
		return nil, secrets.NoVersion, fmt.Errorf("Namespace cannot be empty.")
	}

	secret, err := core.Instance().GetSecret(secretName, namespace)
	if err != nil {
		return nil, secrets.NoVersion, fmt.Errorf("Failed to get secret [%s]. Err: %v",
			secretName, err)
	}
	if secret == nil {
		return nil, secrets.NoVersion, secrets.ErrInvalidSecretId
	}

	data := make(map[string]interface{})
	for key, val := range secret.Data {
		data[key] = fmt.Sprintf("%s", val)
	}
	return data, secrets.Version(secret.ResourceVersion), nil
}

func (s *k8sSecrets) PutSecret(
	secretName string,
	secretData map[string]interface{},
	keyContext map[string]string,
) (secrets.Version, error) {
	namespace, exists := keyContext[SecretNamespace]
	if !exists {
		return secrets.NoVersion, fmt.Errorf("Namespace cannot be empty.")
	}
	if len(secretData) == 0 {
		return secrets.NoVersion, nil
	}

	data := make(map[string][]byte)
	for key, val := range secretData {
		if v, ok := val.(string); ok {
			data[key] = []byte(v)
		} else if v, ok := val.([]byte); ok {
			data[key] = v
		} else {
			return secrets.NoVersion, fmt.Errorf("Unsupported data type for data: %s", key)
		}
	}

	updatedSecret, err := core.Instance().UpdateSecretData(secretName, namespace, data)
	if err != nil {
		return secrets.NoVersion, err
	}
	return secrets.Version(updatedSecret.ResourceVersion), nil
}

func (s *k8sSecrets) DeleteSecret(
	secretName string,
	keyContext map[string]string,
) error {
	namespace, exists := keyContext[SecretNamespace]
	if !exists {
		return fmt.Errorf("Namespace cannot be empty.")
	}

	return core.Instance().DeleteSecret(secretName, namespace)
}

func (s *k8sSecrets) ListSecrets() ([]string, error) {
	return nil, secrets.ErrNotSupported
}

func (s *k8sSecrets) Encrypt(
	secretId string,
	plaintTextData string,
	keyContext map[string]string,
) (string, error) {
	return "", secrets.ErrNotSupported
}

func (s *k8sSecrets) Decrypt(
	secretId string,
	encryptedData string,
	keyContext map[string]string,
) (string, error) {
	return "", secrets.ErrNotSupported
}

func (s *k8sSecrets) Rencrypt(
	originalSecretId string,
	newSecretId string,
	originalKeyContext map[string]string,
	newKeyContext map[string]string,
	encryptedData string,
) (string, error) {
	return "", secrets.ErrNotSupported
}

func init() {
	if err := secrets.Register(Name, New); err != nil {
		panic(err.Error())
	}
}
