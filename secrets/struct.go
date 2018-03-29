package secrets

import "net/http"

const (
	ClusterSecretKey = "clustersecretkey"
	OverrideSecrets  = "override"
	SecretKey        = "secretid"
	SecretValue      = "secretvalue"
	SecretType       = "secret"
	APIVersion       = "v1"
)

// SecretLoginRequest specify secret store and config to initiate
// secret store session
// swagger: parameters secret
type SecretLoginRequest struct {
	SecretConfig map[string]string
}

type Route struct {
	Verb string
	Path string
	Fn   func(http.ResponseWriter, *http.Request)
}

// ClusterSecretKeyRequest  specify request to set cluster secret key
// swagger: parameters clusterKey
type ClusterSecretKeyRequest struct {
	ClusterSecretKey string
	Override         bool
}

// SetsecretsLogin setsecrets
// swagger: parameters secret
type SetSecretsRequest struct {
	SecretID    string
	SecretValue string
}
