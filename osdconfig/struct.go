package osdconfig

import "time"

// NodesConfig contains all of node level data
// swagger:model
type NodesConfig []*NodeConfig

// NodeConfig is a node level config data
// swagger:model
type NodeConfig struct {
	NodeId      string         `json:"node_id,omitempty" enable:"true" hidden:"false" usage:"ID for the node"`
	CSIEndpoint string         `json:"csi_endpoint,omitempty" enable:"true" hidden:"false" usage:"CSI endpoint"`
	Network     *NetworkConfig `json:"network,omitempty" enable:"true" hidden:"false" usage:"Network configuration" description:"Configure network values for a node"`
	Storage     *StorageConfig `json:"storage,omitempty" enable:"true" hidden:"false" usage:"Storage configuration" description:"Configure storage values for a node"`
	Geo         *GeoConfig     `json:"geo,omitempty" enable:"true" hidden:"false" usage:"Geographic configuration" description:"Stores geo info for node"`
	Private     interface{}    `json:"private,omitempty" enable:"false" hidden:"false" usage:"Private node data"`
}

func (conf *NodeConfig) Init() *NodeConfig {
	conf.Network = new(NetworkConfig).Init()
	conf.Storage = new(StorageConfig).Init()
	conf.Geo = new(GeoConfig).Init()
	return conf
}

// KvdbConfig stores parameters defining kvdb configuration
// swagger:model
type KvdbConfig struct {
	Name               string   `json:"name,omitempty" enable:"true" hidden:"false" usage:"Name for kvdb"`
	Username           string   `json:"username,omitempty" enable:"true" hidden:"false" usage:"Username for kvdb"`
	Password           string   `json:"password,omitempty" enable:"true" hidden:"false" usage:"Passwd for kvdb"`
	CAFile             string   `json:"ca_file,omitempty" enable:"true" hidden:"false" usage:"CA file for kvdb"`
	CertFile           string   `json:"cert_file,omitempty" enable:"true" hidden:"false" usage:"Cert file for kvdb"`
	CertKeyFile        string   `json:"cert_key_file,omitempty" enable:"true" hidden:"false" usage:"Cert key file for kvdb"`
	TrustedCAFile      string   `json:"trusted_ca_file,omitempty" enable:"true" hidden:"false" usage:"Trusted CA file for kvdb"`
	ClientCertAuth     string   `json:"client_cert_auth,omitempty" enable:"true" hidden:"false" usage:"Client cert auth"`
	AclToken           string   `json:"acl_token,omitempty" enable:"true" hidden:"false" usage:"ACL token"`
	CAAuthAddress      string   `json:"ca_auth_address,omitempty" enable:"true" hidden:"false" usage:"Address of CA auth server (only for consul)"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify,omitempty" enable:"true" hidden:"false" usage:"Insecure skip verify bool (only for consul)"`
	TransportScheme    string   `json:"transport_scheme,omitempty" enable:"true" hidden:"false" usage:"Transport method http or https (only for consul)"`
	Discovery          []string `json:"discovery,omitempty" enable:"true" hidden:"false" usage:"List of etcd endpoints"`
}

func (conf *KvdbConfig) Init() *KvdbConfig {
	conf.Discovery = make([]string, 0, 0)
	return conf
}

// ClusterConfig is a cluster level config parameter struct
// swagger:model
type ClusterConfig struct {
	Description string         `json:"description,omitempty" enable:"true" hidden:"false" usage:"Cluster description"`
	Mode        string         `json:"mode,omitempty" enable:"true" hidden:"false" usage:"Mode for cluster"`
	Version     string         `json:"version,omitempty" enable:"true" hidden:"false" usage:"Version info for cluster"`
	Created     time.Time      `json:"created,omitempty" enable:"true" hidden:"false" usage:"Creation info for cluster"`
	ClusterId   string         `json:"cluster_id,omitempty" enable:"true" hidden:"false" usage:"Cluster ID info"`
	Domain      string         `json:"domain,omitempty" enable:"true" hidden:"false" usage:"usage to be added"`
	Secrets     *SecretsConfig `json:"secrets,omitempty" enable:"true" hidden:"false" usage:"usage to be added" description:"description to be added"`
	Kvdb        *KvdbConfig    `json:"kvdb,omitempty" enable:"false" hidden:"false" usage:"usage to be added" description:"description to be added"`
	Private     interface{}    `json:"private,omitempty" enable:"true" hidden:"false" usage:"usage to be added"`
}

func (conf *ClusterConfig) Init() *ClusterConfig {
	conf.Secrets = new(SecretsConfig).Init()
	conf.Kvdb = new(KvdbConfig).Init()
	return conf
}

// NetworkConfig is a network configuration parameters struct
// swagger:model
type NetworkConfig struct {
	MgtIface  string `json:"mgt_interface,omitempty" enable:"true" hidden:"false" usage:"Management interface"`
	DataIface string `json:"data_interface,omitempty" enable:"true" hidden:"false" usage:"Data interface"`
}

func (conf *NetworkConfig) Init() *NetworkConfig {
	return conf
}

// GeoConfig holds geographic information
type GeoConfig struct {
	Rack   string `json:"rack,omitempty" enable:"true" hidden:"false" usage:"Rack info"`
	Zone   string `json:"zone,omitempty" enable:"true" hidden:"false" usage:"Zone info"`
	Region string `json:"region,omitempty" enable:"true" hidden:"false" usage:"Region info"`
}

func (conf *GeoConfig) Init() *GeoConfig {
	return conf
}

// SecretsConfig is a secrets configuration parameters struct
// swagger:model
type SecretsConfig struct {
	SecretType       string       `json:"secret_type,omitempty" enable:"true" hidden:"false" usage:"Secret type"`
	ClusterSecretKey string       `json:"cluster_secret_key,omitempty" enable:"true" hidden:"false" usage:"Secret key"`
	Vault            *VaultConfig `json:"vault,omitempty" enable:"true" hidden:"false" usage:"Vault configuration"`
	Aws              *AWSConfig   `json:"aws,omitempty" enable:"true" hidden:"false" usage:"AWS configuration"`
}

func (conf *SecretsConfig) Init() *SecretsConfig {
	conf.Vault = new(VaultConfig).Init()
	conf.Aws = new(AWSConfig).Init()
	return conf
}

// VaultConfig is a vault configuration parameters struct
// swagger:model
type VaultConfig struct {
	Token         string `json:"token,omitempty" enable:"true" hidden:"false" usage:"Vault token"`
	Address       string `json:"address,omitempty" enable:"true" hidden:"false" usage:"Vault address"`
	CACert        string `json:"ca_cert,omitempty" enable:"true" hidden:"false" usage:"Vault CA certificate"`
	CAPath        string `json:"ca_path,omitempty" enable:"true" hidden:"false" usage:"Vault CA path"`
	ClientCert    string `json:"client_cert,omitempty" enable:"true" hidden:"false" usage:"Vault client certificate"`
	ClientKey     string `json:"client_key,omitempty" enable:"true" hidden:"false" usage:"Vault client key"`
	TLSSkipVerify string `json:"skip_verify,omitempty" enable:"true" hidden:"false" usage:"Vault skip verification"`
	TLSServerName string `json:"tls_server_name,omitempty" enable:"true" hidden:"false" usage:"Vault TLS server name"`
	BasePath      string `json:"base_path,omitempty" enable:"true" hidden:"false" usage:"Vault base path"`
	BackendPath   string `json:"backend_path,omitempty" enable:"true" hidden:"false" usage:"Vault secrets backend mount path"`
}

func (conf *VaultConfig) Init() *VaultConfig {
	return conf
}

// AWS configuration parameters struct
// swagger:model
type AWSConfig struct {
	AccessKeyId     string `json:"aws_access_key_id,omitempty" enable:"true" hidden:"false" usage:"AWS access key ID"`
	SecretAccessKey string `json:"aws_secret_access_key,omitempty" enable:"true" hidden:"false" usage:"AWS secret access key"`
	SecretTokenKey  string `json:"aws_secret_token_key,omitempty" enable:"true" hidden:"false" usage:"AWS secret token key"`
	Cmk             string `json:"aws_cmk,omitempty" enable:"true" hidden:"false" usage:"AWS CMK"`
	Region          string `json:"aws_region,omitempty" enable:"true" hidden:"false" usage:"AWS region"`
}

func (conf *AWSConfig) Init() *AWSConfig {
	return conf
}

// StorageConfig is a storage configuration parameters struct
// swagger:model
type StorageConfig struct {
	DevicesMd        []string `json:"devices_md,omitempty" enable:"true" hidden:"false" usage:"Devices MD"`
	Devices          []string `json:"devices,omitempty" enable:"true" hidden:"false" usage:"Devices list"`
	MaxCount         uint32   `json:"max_count,omitempty" enable:"true" hidden:"false" usage:"Maximum count"`
	MaxDriveSetCount uint32   `json:"max_drive_set_count,omitempty" enable:"true" hidden:"false" usage:"Max drive set count"`
	RaidLevel        string   `json:"raid_level,omitempty" enable:"true" hidden:"false" usage:"RAID level info"`
	RaidLevelMd      string   `json:"raid_level_md,omitempty" enable:"true" hidden:"false" usage:"RAID level MD"`
}

func (conf *StorageConfig) Init() *StorageConfig {
	conf.Devices = make([]string, 0, 0)
	conf.Devices = make([]string, 0, 0)
	return conf
}
