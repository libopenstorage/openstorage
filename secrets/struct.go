package secrets

const (
	ClusterSecretKey = "clustersecretkey"
	OverrideSecrets  = "override"
	SecretKey        = "secretid"
	SecretValue      = "secretvalue"
	SecretType       = "secret"
)

// SecretLoginRequest specify secret store and config to initiate
// secret store session
// swagger: parameters secret
type SecretLoginRequest struct {
	SecretConfig map[string]string
}

// ClusterSecretKeyRequest  specify request to set cluster secret key
// swagger: parameters clusterKey
type ClusterSecretKeyRequest struct {
	Clustersecretkey string
	Override         bool
}

// SetsecretsLogin setsecrets
// swagger: parameters secret
type SetSecretsRequest struct {
	Secretid    string
	SecretValue string
}
