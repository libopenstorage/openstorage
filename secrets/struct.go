package secrets

const (
	DefaultSecretKey = "defaultsecretkey"
	OverrideSecrets  = "override"
	SecretKey        = "id"
	SecretValue      = "secretvalue"
	SecretType       = "type"
	APIVersion       = "v1"
)

const (
	TypeNone   = "none"
	TypeKvdb   = "kvdb"
	TypeVault  = "vault"
	TypeAWS    = "aws-kms"
	TypeDocker = "docker"
	TypeK8s    = "k8s"
	TypeDCOS   = "dcos"
)

// SecretLoginRequest specify secret store and config to initiate
// secret store session
// swagger:model
type SecretLoginRequest struct {
	SecretType   string
	SecretConfig map[string]string
}

// DefaultSecretKeyRequest specify request to set cluster secret key
// swagger:model
type DefaultSecretKeyRequest struct {
	DefaultSecretKey string
	Override         bool
}

// SetSecretRequest stores the given value/data against the key
// swagger:model
type SetSecretRequest struct {
	SecretValue interface{}
}

// GetSecretResponse gets secret value for given key
// swagger:model
type GetSecretResponse struct {
	SecretValue interface{}
}
