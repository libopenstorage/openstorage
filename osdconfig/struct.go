package osdconfig

// NodesConfig contains all of node level data in the form of a map with node ID's as keys
type NodesConfig struct {
	NodeConf map[string]*NodeConfig `json:"node_conf,omitempty"`
}

// NodeConfig is a node level config data
type NodeConfig struct {
	NodeId  string         `json:"node_id,omitempty"`
	Network *NetworkConfig `json:"network,omitempty"`
	Storage *StorageConfig `json:"storage,omitempty"`
	Private interface{}    `json:"generic,omitempty"`
}

// KvdbConfig stores parameters defining kvdb configuration
type KvdbConfig struct {
	Username       string   `json:"username,omitempty"`
	Password       string   `json:"password,omitempty"`
	CaFile         string   `json:"ca_file,omitempty"`
	CertFile       string   `json:"cert_file,omitempty"`
	TrustedCaFile  string   `json:"trusted_ca_file,omitempty"`
	ClientCertAuth string   `json:"client_cert_auth,omitempty"`
	AclToken       string   `json:"acl_token,omitempty"`
	KvdbAddr       []string `json:"kvdb_addr,omitempty"`
}

// ClusterConfig is a cluster level config parameter struct
type ClusterConfig struct {
	Description    string         `json:"description,omitempty"`
	Mode           string         `json:"mode,omitempty"`
	Version        string         `json:"version,omitempty"`
	Created        string         `json:"created,omitempty"`
	ClusterId      string         `json:"cluster_id,omitempty"`
	LoggingUrl     string         `json:"logging_url,omitempty"`
	AlertingUrl    string         `json:"alerting_url,omitempty"`
	Scheduler      string         `json:"scheduler,omitempty"`
	Multicontainer bool           `json:"multicontainer,omitempty"`
	Nolh           bool           `json:"nolh,omitempty"`
	Callhome       bool           `json:"callhome,omitempty"`
	Bootstrap      bool           `json:"bootstrap,omitempty"`
	TunnelEndPoint string         `json:"tunnel_end_point,omitempty"`
	TunnelCerts    []string       `json:"tunnel_certs,omitempty"`
	Driver         string         `json:"driver,omitempty"`
	DebugLevel     string         `json:"debug_level,omitempty"`
	Domain         string         `json:"domain,omitempty"`
	Secrets        *SecretsConfig `json:"secrets,omitempty"`
	Kvdb           *KvdbConfig    `json:"kvdb,omitempty"`
	Private        interface{}    `json:"generic,omitempty"`
}

// NetworkConfig is a network configuration parameters struct
type NetworkConfig struct {
	MgtIface  string `json:"mgt_iface,omitempty"`
	DataIface string `json:"data_iface,omitempty"`
}

// SecretsConfig is a secrets configuration parameters struct
type SecretsConfig struct {
	SecretType       string       `json:"secret_type,omitempty"`
	ClusterSecretKey string       `json:"cluster_secret_key,omitempty"`
	Vault            *VaultConfig `json:"vault,omitempty"`
	Aws              *AWSConfig   `json:"aws,omitempty"`
}

// VaultConfig is a vault configuration parameters struct
type VaultConfig struct {
	VaultToken         string `json:"vault_token,omitempty"`
	VaultAddr          string `json:"vault_addr,omitempty"`
	VaultCacert        string `json:"vault_cacert,omitempty"`
	VaultCapath        string `json:"vault_capath,omitempty"`
	VaultClientCert    string `json:"vault_client_cert,omitempty"`
	VaultClientKey     string `json:"vault_client_key,omitempty"`
	VaultSkipVerify    string `json:"vault_skip_verify,omitempty"`
	VaultTlsServerName string `json:"vault_tls_server_name,omitempty"`
	VaultBasePath      string `json:"vault_base_path,omitempty"`
}

// AWS configuration parameters struct
type AWSConfig struct {
	AwsAccessKeyId     string `json:"aws_access_key_id,omitempty"`
	AwsSecretAccessKey string `json:"aws_secret_access_key,omitempty"`
	AwsSecretTokenKey  string `json:"aws_secret_token_key,omitempty"`
	AwsCmk             string `json:"aws_cmk,omitempty"`
	AwsRegion          string `json:"aws_region,omitempty"`
}

// StorageConfig is a storage configuration parameters struct
type StorageConfig struct {
	DevicesMd        []string `json:"devices_md,omitempty"`
	MaxCount         int32    `json:"max_count,omitempty"`
	MaxDriveSetCount int32    `json:"max_drive_set_count,omitempty"`
	Devices          []string `json:"devices,omitempty"`
	RaidLevel        string   `json:"raid_level,omitempty"`
	RaidLevelMd      string   `json:"raid_level_md,omitempty"`
	AsyncIo          bool     `json:"async_io,omitempty"`
	NumThreads       int32    `json:"num_threads,omitempty"`
}
