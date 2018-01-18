// Code generated by protoc-gen-go. DO NOT EDIT.
// source: config.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	config.proto

It has these top-level messages:
	Empty
	Ack
	Config
	NodeConfig
	KVDBConfig
	GlobalConfig
	NetworkConfig
	SecretsConfig
	VaultConfig
	AWSConfig
	StorageConfig
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto1.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Ack struct {
	N int64 `protobuf:"varint,1,opt,name=n" json:"n,omitempty"`
}

func (m *Ack) Reset()                    { *m = Ack{} }
func (m *Ack) String() string            { return proto1.CompactTextString(m) }
func (*Ack) ProtoMessage()               {}
func (*Ack) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Ack) GetN() int64 {
	if m != nil {
		return m.N
	}
	return 0
}

type Config struct {
	Description string        `protobuf:"bytes,1,opt,name=description" json:"description,omitempty"`
	Global      *GlobalConfig `protobuf:"bytes,2,opt,name=global" json:"global,omitempty"`
}

func (m *Config) Reset()                    { *m = Config{} }
func (m *Config) String() string            { return proto1.CompactTextString(m) }
func (*Config) ProtoMessage()               {}
func (*Config) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Config) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Config) GetGlobal() *GlobalConfig {
	if m != nil {
		return m.Global
	}
	return nil
}

type NodeConfig struct {
	Network *NetworkConfig `protobuf:"bytes,1,opt,name=network" json:"network,omitempty"`
	Secrets *SecretsConfig `protobuf:"bytes,2,opt,name=secrets" json:"secrets,omitempty"`
	Storage *StorageConfig `protobuf:"bytes,3,opt,name=storage" json:"storage,omitempty"`
}

func (m *NodeConfig) Reset()                    { *m = NodeConfig{} }
func (m *NodeConfig) String() string            { return proto1.CompactTextString(m) }
func (*NodeConfig) ProtoMessage()               {}
func (*NodeConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *NodeConfig) GetNetwork() *NetworkConfig {
	if m != nil {
		return m.Network
	}
	return nil
}

func (m *NodeConfig) GetSecrets() *SecretsConfig {
	if m != nil {
		return m.Secrets
	}
	return nil
}

func (m *NodeConfig) GetStorage() *StorageConfig {
	if m != nil {
		return m.Storage
	}
	return nil
}

type KVDBConfig struct {
	Username       string   `protobuf:"bytes,1,opt,name=username" json:"username,omitempty"`
	Password       string   `protobuf:"bytes,2,opt,name=password" json:"password,omitempty"`
	CaFile         string   `protobuf:"bytes,3,opt,name=ca_file,json=caFile" json:"ca_file,omitempty"`
	CertFile       string   `protobuf:"bytes,4,opt,name=cert_file,json=certFile" json:"cert_file,omitempty"`
	TrustedCaFile  string   `protobuf:"bytes,5,opt,name=trusted_ca_file,json=trustedCaFile" json:"trusted_ca_file,omitempty"`
	ClientCertAuth string   `protobuf:"bytes,6,opt,name=client_cert_auth,json=clientCertAuth" json:"client_cert_auth,omitempty"`
	AclToken       string   `protobuf:"bytes,7,opt,name=acl_token,json=aclToken" json:"acl_token,omitempty"`
	KvdbAddr       []string `protobuf:"bytes,8,rep,name=kvdb_addr,json=kvdbAddr" json:"kvdb_addr,omitempty"`
}

func (m *KVDBConfig) Reset()                    { *m = KVDBConfig{} }
func (m *KVDBConfig) String() string            { return proto1.CompactTextString(m) }
func (*KVDBConfig) ProtoMessage()               {}
func (*KVDBConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *KVDBConfig) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *KVDBConfig) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *KVDBConfig) GetCaFile() string {
	if m != nil {
		return m.CaFile
	}
	return ""
}

func (m *KVDBConfig) GetCertFile() string {
	if m != nil {
		return m.CertFile
	}
	return ""
}

func (m *KVDBConfig) GetTrustedCaFile() string {
	if m != nil {
		return m.TrustedCaFile
	}
	return ""
}

func (m *KVDBConfig) GetClientCertAuth() string {
	if m != nil {
		return m.ClientCertAuth
	}
	return ""
}

func (m *KVDBConfig) GetAclToken() string {
	if m != nil {
		return m.AclToken
	}
	return ""
}

func (m *KVDBConfig) GetKvdbAddr() []string {
	if m != nil {
		return m.KvdbAddr
	}
	return nil
}

type GlobalConfig struct {
	Mode           string        `protobuf:"bytes,1,opt,name=mode" json:"mode,omitempty"`
	Version        string        `protobuf:"bytes,2,opt,name=version" json:"version,omitempty"`
	Created        string        `protobuf:"bytes,3,opt,name=created" json:"created,omitempty"`
	ClusterId      string        `protobuf:"bytes,4,opt,name=cluster_id,json=clusterId" json:"cluster_id,omitempty"`
	LoggingUrl     string        `protobuf:"bytes,5,opt,name=logging_url,json=loggingUrl" json:"logging_url,omitempty"`
	AlertingUrl    string        `protobuf:"bytes,6,opt,name=alerting_url,json=alertingUrl" json:"alerting_url,omitempty"`
	Scheduler      string        `protobuf:"bytes,7,opt,name=scheduler" json:"scheduler,omitempty"`
	Multicontainer bool          `protobuf:"varint,8,opt,name=multicontainer" json:"multicontainer,omitempty"`
	Nolh           bool          `protobuf:"varint,9,opt,name=nolh" json:"nolh,omitempty"`
	Callhome       bool          `protobuf:"varint,10,opt,name=callhome" json:"callhome,omitempty"`
	Bootstrap      bool          `protobuf:"varint,11,opt,name=bootstrap" json:"bootstrap,omitempty"`
	TunnelEndPoint string        `protobuf:"bytes,12,opt,name=tunnel_end_point,json=tunnelEndPoint" json:"tunnel_end_point,omitempty"`
	TunnelCerts    []string      `protobuf:"bytes,13,rep,name=tunnel_certs,json=tunnelCerts" json:"tunnel_certs,omitempty"`
	Driver         string        `protobuf:"bytes,14,opt,name=driver" json:"driver,omitempty"`
	DebugLevel     string        `protobuf:"bytes,15,opt,name=debug_level,json=debugLevel" json:"debug_level,omitempty"`
	Domain         string        `protobuf:"bytes,16,opt,name=domain" json:"domain,omitempty"`
	Mgmtip         string        `protobuf:"bytes,17,opt,name=mgmtip" json:"mgmtip,omitempty"`
	Dataip         string        `protobuf:"bytes,18,opt,name=dataip" json:"dataip,omitempty"`
	Nodes          []*NodeConfig `protobuf:"bytes,19,rep,name=nodes" json:"nodes,omitempty"`
	Kvdb           *KVDBConfig   `protobuf:"bytes,20,opt,name=kvdb" json:"kvdb,omitempty"`
}

func (m *GlobalConfig) Reset()                    { *m = GlobalConfig{} }
func (m *GlobalConfig) String() string            { return proto1.CompactTextString(m) }
func (*GlobalConfig) ProtoMessage()               {}
func (*GlobalConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *GlobalConfig) GetMode() string {
	if m != nil {
		return m.Mode
	}
	return ""
}

func (m *GlobalConfig) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *GlobalConfig) GetCreated() string {
	if m != nil {
		return m.Created
	}
	return ""
}

func (m *GlobalConfig) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *GlobalConfig) GetLoggingUrl() string {
	if m != nil {
		return m.LoggingUrl
	}
	return ""
}

func (m *GlobalConfig) GetAlertingUrl() string {
	if m != nil {
		return m.AlertingUrl
	}
	return ""
}

func (m *GlobalConfig) GetScheduler() string {
	if m != nil {
		return m.Scheduler
	}
	return ""
}

func (m *GlobalConfig) GetMulticontainer() bool {
	if m != nil {
		return m.Multicontainer
	}
	return false
}

func (m *GlobalConfig) GetNolh() bool {
	if m != nil {
		return m.Nolh
	}
	return false
}

func (m *GlobalConfig) GetCallhome() bool {
	if m != nil {
		return m.Callhome
	}
	return false
}

func (m *GlobalConfig) GetBootstrap() bool {
	if m != nil {
		return m.Bootstrap
	}
	return false
}

func (m *GlobalConfig) GetTunnelEndPoint() string {
	if m != nil {
		return m.TunnelEndPoint
	}
	return ""
}

func (m *GlobalConfig) GetTunnelCerts() []string {
	if m != nil {
		return m.TunnelCerts
	}
	return nil
}

func (m *GlobalConfig) GetDriver() string {
	if m != nil {
		return m.Driver
	}
	return ""
}

func (m *GlobalConfig) GetDebugLevel() string {
	if m != nil {
		return m.DebugLevel
	}
	return ""
}

func (m *GlobalConfig) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

func (m *GlobalConfig) GetMgmtip() string {
	if m != nil {
		return m.Mgmtip
	}
	return ""
}

func (m *GlobalConfig) GetDataip() string {
	if m != nil {
		return m.Dataip
	}
	return ""
}

func (m *GlobalConfig) GetNodes() []*NodeConfig {
	if m != nil {
		return m.Nodes
	}
	return nil
}

func (m *GlobalConfig) GetKvdb() *KVDBConfig {
	if m != nil {
		return m.Kvdb
	}
	return nil
}

type NetworkConfig struct {
	MgtIface  string `protobuf:"bytes,1,opt,name=mgt_iface,json=mgtIface" json:"mgt_iface,omitempty"`
	DataIface string `protobuf:"bytes,2,opt,name=data_iface,json=dataIface" json:"data_iface,omitempty"`
}

func (m *NetworkConfig) Reset()                    { *m = NetworkConfig{} }
func (m *NetworkConfig) String() string            { return proto1.CompactTextString(m) }
func (*NetworkConfig) ProtoMessage()               {}
func (*NetworkConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *NetworkConfig) GetMgtIface() string {
	if m != nil {
		return m.MgtIface
	}
	return ""
}

func (m *NetworkConfig) GetDataIface() string {
	if m != nil {
		return m.DataIface
	}
	return ""
}

type SecretsConfig struct {
	SecretType       string       `protobuf:"bytes,1,opt,name=secret_type,json=secretType" json:"secret_type,omitempty"`
	ClusterSecretKey string       `protobuf:"bytes,2,opt,name=cluster_secret_key,json=clusterSecretKey" json:"cluster_secret_key,omitempty"`
	Vault            *VaultConfig `protobuf:"bytes,3,opt,name=vault" json:"vault,omitempty"`
	Aws              *AWSConfig   `protobuf:"bytes,4,opt,name=aws" json:"aws,omitempty"`
}

func (m *SecretsConfig) Reset()                    { *m = SecretsConfig{} }
func (m *SecretsConfig) String() string            { return proto1.CompactTextString(m) }
func (*SecretsConfig) ProtoMessage()               {}
func (*SecretsConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *SecretsConfig) GetSecretType() string {
	if m != nil {
		return m.SecretType
	}
	return ""
}

func (m *SecretsConfig) GetClusterSecretKey() string {
	if m != nil {
		return m.ClusterSecretKey
	}
	return ""
}

func (m *SecretsConfig) GetVault() *VaultConfig {
	if m != nil {
		return m.Vault
	}
	return nil
}

func (m *SecretsConfig) GetAws() *AWSConfig {
	if m != nil {
		return m.Aws
	}
	return nil
}

type VaultConfig struct {
	VaultToken         string `protobuf:"bytes,1,opt,name=vault_token,json=vaultToken" json:"vault_token,omitempty"`
	VaultAddr          string `protobuf:"bytes,2,opt,name=vault_addr,json=vaultAddr" json:"vault_addr,omitempty"`
	VaultCacert        string `protobuf:"bytes,3,opt,name=vault_cacert,json=vaultCacert" json:"vault_cacert,omitempty"`
	VaultCapath        string `protobuf:"bytes,4,opt,name=vault_capath,json=vaultCapath" json:"vault_capath,omitempty"`
	VaultClientCert    string `protobuf:"bytes,5,opt,name=vault_client_cert,json=vaultClientCert" json:"vault_client_cert,omitempty"`
	VaultClientKey     string `protobuf:"bytes,6,opt,name=vault_client_key,json=vaultClientKey" json:"vault_client_key,omitempty"`
	VaultSkipVerify    string `protobuf:"bytes,7,opt,name=vault_skip_verify,json=vaultSkipVerify" json:"vault_skip_verify,omitempty"`
	VaultTlsServerName string `protobuf:"bytes,8,opt,name=vault_tls_server_name,json=vaultTlsServerName" json:"vault_tls_server_name,omitempty"`
	VaultBasePath      string `protobuf:"bytes,9,opt,name=vault_base_path,json=vaultBasePath" json:"vault_base_path,omitempty"`
}

func (m *VaultConfig) Reset()                    { *m = VaultConfig{} }
func (m *VaultConfig) String() string            { return proto1.CompactTextString(m) }
func (*VaultConfig) ProtoMessage()               {}
func (*VaultConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *VaultConfig) GetVaultToken() string {
	if m != nil {
		return m.VaultToken
	}
	return ""
}

func (m *VaultConfig) GetVaultAddr() string {
	if m != nil {
		return m.VaultAddr
	}
	return ""
}

func (m *VaultConfig) GetVaultCacert() string {
	if m != nil {
		return m.VaultCacert
	}
	return ""
}

func (m *VaultConfig) GetVaultCapath() string {
	if m != nil {
		return m.VaultCapath
	}
	return ""
}

func (m *VaultConfig) GetVaultClientCert() string {
	if m != nil {
		return m.VaultClientCert
	}
	return ""
}

func (m *VaultConfig) GetVaultClientKey() string {
	if m != nil {
		return m.VaultClientKey
	}
	return ""
}

func (m *VaultConfig) GetVaultSkipVerify() string {
	if m != nil {
		return m.VaultSkipVerify
	}
	return ""
}

func (m *VaultConfig) GetVaultTlsServerName() string {
	if m != nil {
		return m.VaultTlsServerName
	}
	return ""
}

func (m *VaultConfig) GetVaultBasePath() string {
	if m != nil {
		return m.VaultBasePath
	}
	return ""
}

type AWSConfig struct {
	AwsAccessKeyId     string `protobuf:"bytes,1,opt,name=aws_access_key_id,json=awsAccessKeyId" json:"aws_access_key_id,omitempty"`
	AwsSecretAccessKey string `protobuf:"bytes,2,opt,name=aws_secret_access_key,json=awsSecretAccessKey" json:"aws_secret_access_key,omitempty"`
	AwsSecretTokenKey  string `protobuf:"bytes,3,opt,name=aws_secret_token_key,json=awsSecretTokenKey" json:"aws_secret_token_key,omitempty"`
	AwsCmk             string `protobuf:"bytes,4,opt,name=aws_cmk,json=awsCmk" json:"aws_cmk,omitempty"`
	AwsRegion          string `protobuf:"bytes,5,opt,name=aws_region,json=awsRegion" json:"aws_region,omitempty"`
}

func (m *AWSConfig) Reset()                    { *m = AWSConfig{} }
func (m *AWSConfig) String() string            { return proto1.CompactTextString(m) }
func (*AWSConfig) ProtoMessage()               {}
func (*AWSConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *AWSConfig) GetAwsAccessKeyId() string {
	if m != nil {
		return m.AwsAccessKeyId
	}
	return ""
}

func (m *AWSConfig) GetAwsSecretAccessKey() string {
	if m != nil {
		return m.AwsSecretAccessKey
	}
	return ""
}

func (m *AWSConfig) GetAwsSecretTokenKey() string {
	if m != nil {
		return m.AwsSecretTokenKey
	}
	return ""
}

func (m *AWSConfig) GetAwsCmk() string {
	if m != nil {
		return m.AwsCmk
	}
	return ""
}

func (m *AWSConfig) GetAwsRegion() string {
	if m != nil {
		return m.AwsRegion
	}
	return ""
}

type StorageConfig struct {
	DevicesMd        []string `protobuf:"bytes,1,rep,name=devices_md,json=devicesMd" json:"devices_md,omitempty"`
	MaxCount         int32    `protobuf:"varint,2,opt,name=max_count,json=maxCount" json:"max_count,omitempty"`
	MaxDriveSetCount int32    `protobuf:"varint,3,opt,name=max_drive_set_count,json=maxDriveSetCount" json:"max_drive_set_count,omitempty"`
	Devices          []string `protobuf:"bytes,4,rep,name=devices" json:"devices,omitempty"`
	RaidLevel        string   `protobuf:"bytes,5,opt,name=raid_level,json=raidLevel" json:"raid_level,omitempty"`
	RaidLevelMd      string   `protobuf:"bytes,6,opt,name=raid_level_md,json=raidLevelMd" json:"raid_level_md,omitempty"`
	AsyncIo          bool     `protobuf:"varint,7,opt,name=async_io,json=asyncIo" json:"async_io,omitempty"`
	NumThreads       int32    `protobuf:"varint,8,opt,name=num_threads,json=numThreads" json:"num_threads,omitempty"`
}

func (m *StorageConfig) Reset()                    { *m = StorageConfig{} }
func (m *StorageConfig) String() string            { return proto1.CompactTextString(m) }
func (*StorageConfig) ProtoMessage()               {}
func (*StorageConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *StorageConfig) GetDevicesMd() []string {
	if m != nil {
		return m.DevicesMd
	}
	return nil
}

func (m *StorageConfig) GetMaxCount() int32 {
	if m != nil {
		return m.MaxCount
	}
	return 0
}

func (m *StorageConfig) GetMaxDriveSetCount() int32 {
	if m != nil {
		return m.MaxDriveSetCount
	}
	return 0
}

func (m *StorageConfig) GetDevices() []string {
	if m != nil {
		return m.Devices
	}
	return nil
}

func (m *StorageConfig) GetRaidLevel() string {
	if m != nil {
		return m.RaidLevel
	}
	return ""
}

func (m *StorageConfig) GetRaidLevelMd() string {
	if m != nil {
		return m.RaidLevelMd
	}
	return ""
}

func (m *StorageConfig) GetAsyncIo() bool {
	if m != nil {
		return m.AsyncIo
	}
	return false
}

func (m *StorageConfig) GetNumThreads() int32 {
	if m != nil {
		return m.NumThreads
	}
	return 0
}

func init() {
	proto1.RegisterType((*Empty)(nil), "proto.Empty")
	proto1.RegisterType((*Ack)(nil), "proto.Ack")
	proto1.RegisterType((*Config)(nil), "proto.Config")
	proto1.RegisterType((*NodeConfig)(nil), "proto.NodeConfig")
	proto1.RegisterType((*KVDBConfig)(nil), "proto.KVDBConfig")
	proto1.RegisterType((*GlobalConfig)(nil), "proto.GlobalConfig")
	proto1.RegisterType((*NetworkConfig)(nil), "proto.NetworkConfig")
	proto1.RegisterType((*SecretsConfig)(nil), "proto.SecretsConfig")
	proto1.RegisterType((*VaultConfig)(nil), "proto.VaultConfig")
	proto1.RegisterType((*AWSConfig)(nil), "proto.AWSConfig")
	proto1.RegisterType((*StorageConfig)(nil), "proto.StorageConfig")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ClusterSpec service

type ClusterSpecClient interface {
	Get(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Config, error)
	Set(ctx context.Context, in *Config, opts ...grpc.CallOption) (*Ack, error)
}

type clusterSpecClient struct {
	cc *grpc.ClientConn
}

func NewClusterSpecClient(cc *grpc.ClientConn) ClusterSpecClient {
	return &clusterSpecClient{cc}
}

func (c *clusterSpecClient) Get(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Config, error) {
	out := new(Config)
	err := grpc.Invoke(ctx, "/proto.ClusterSpec/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterSpecClient) Set(ctx context.Context, in *Config, opts ...grpc.CallOption) (*Ack, error) {
	out := new(Ack)
	err := grpc.Invoke(ctx, "/proto.ClusterSpec/Set", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ClusterSpec service

type ClusterSpecServer interface {
	Get(context.Context, *Empty) (*Config, error)
	Set(context.Context, *Config) (*Ack, error)
}

func RegisterClusterSpecServer(s *grpc.Server, srv ClusterSpecServer) {
	s.RegisterService(&_ClusterSpec_serviceDesc, srv)
}

func _ClusterSpec_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterSpecServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ClusterSpec/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterSpecServer).Get(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterSpec_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Config)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterSpecServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ClusterSpec/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterSpecServer).Set(ctx, req.(*Config))
	}
	return interceptor(ctx, in, info, handler)
}

var _ClusterSpec_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ClusterSpec",
	HandlerType: (*ClusterSpecServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _ClusterSpec_Get_Handler,
		},
		{
			MethodName: "Set",
			Handler:    _ClusterSpec_Set_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "config.proto",
}

// Client API for NodeSpec service

type NodeSpecClient interface {
	Get(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Config, error)
	Set(ctx context.Context, in *Config, opts ...grpc.CallOption) (*Ack, error)
}

type nodeSpecClient struct {
	cc *grpc.ClientConn
}

func NewNodeSpecClient(cc *grpc.ClientConn) NodeSpecClient {
	return &nodeSpecClient{cc}
}

func (c *nodeSpecClient) Get(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Config, error) {
	out := new(Config)
	err := grpc.Invoke(ctx, "/proto.NodeSpec/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeSpecClient) Set(ctx context.Context, in *Config, opts ...grpc.CallOption) (*Ack, error) {
	out := new(Ack)
	err := grpc.Invoke(ctx, "/proto.NodeSpec/Set", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for NodeSpec service

type NodeSpecServer interface {
	Get(context.Context, *Empty) (*Config, error)
	Set(context.Context, *Config) (*Ack, error)
}

func RegisterNodeSpecServer(s *grpc.Server, srv NodeSpecServer) {
	s.RegisterService(&_NodeSpec_serviceDesc, srv)
}

func _NodeSpec_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeSpecServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.NodeSpec/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeSpecServer).Get(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeSpec_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Config)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeSpecServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.NodeSpec/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeSpecServer).Set(ctx, req.(*Config))
	}
	return interceptor(ctx, in, info, handler)
}

var _NodeSpec_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.NodeSpec",
	HandlerType: (*NodeSpecServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _NodeSpec_Get_Handler,
		},
		{
			MethodName: "Set",
			Handler:    _NodeSpec_Set_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "config.proto",
}

func init() { proto1.RegisterFile("config.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1178 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x56, 0xdb, 0x72, 0x1b, 0x45,
	0x13, 0x8e, 0x2c, 0xeb, 0xd4, 0xb2, 0x1c, 0x7b, 0x9c, 0xff, 0xcf, 0x12, 0x48, 0x61, 0xb6, 0x20,
	0x98, 0x53, 0x28, 0xcc, 0x13, 0x38, 0x4a, 0x48, 0xb9, 0x4c, 0x52, 0x29, 0xd9, 0x24, 0x97, 0x5b,
	0xe3, 0x99, 0xb6, 0xb4, 0xa5, 0x3d, 0xd5, 0xcc, 0xac, 0x64, 0xbd, 0x00, 0x2f, 0xc0, 0x2d, 0x8f,
	0xc0, 0xd3, 0xf0, 0x1a, 0xdc, 0xf0, 0x08, 0x54, 0xf7, 0xcc, 0xae, 0x2d, 0xae, 0xb9, 0x8a, 0xfa,
	0xeb, 0xaf, 0x7b, 0xa7, 0x0f, 0x5f, 0xc7, 0xb0, 0xa7, 0xca, 0xe2, 0x26, 0x9d, 0x3f, 0xaf, 0x4c,
	0xe9, 0x4a, 0xd1, 0xe3, 0x7f, 0xe2, 0x01, 0xf4, 0x5e, 0xe5, 0x95, 0xdb, 0xc4, 0x47, 0xd0, 0x3d,
	0x53, 0x4b, 0xb1, 0x07, 0x9d, 0x22, 0xea, 0x1c, 0x77, 0x4e, 0xba, 0xb3, 0x4e, 0x11, 0x7f, 0x80,
	0xfe, 0x94, 0x83, 0xc4, 0x31, 0x8c, 0x35, 0x5a, 0x65, 0xd2, 0xca, 0xa5, 0xa5, 0x67, 0x8c, 0x66,
	0xf7, 0x21, 0xf1, 0x0d, 0xf4, 0xe7, 0x59, 0x79, 0x2d, 0xb3, 0x68, 0xe7, 0xb8, 0x73, 0x32, 0x3e,
	0x3d, 0xf2, 0x1f, 0x7a, 0xfe, 0x9a, 0x41, 0x9f, 0x66, 0x16, 0x28, 0xf1, 0xef, 0x1d, 0x80, 0xb7,
	0xa5, 0xc6, 0x90, 0xfd, 0x39, 0x0c, 0x0a, 0x74, 0xeb, 0xd2, 0x2c, 0x39, 0xf3, 0xf8, 0xf4, 0x51,
	0x08, 0x7e, 0xeb, 0xd1, 0x10, 0xdd, 0x90, 0x88, 0x6f, 0x51, 0x19, 0x74, 0x36, 0x7c, 0xac, 0xe1,
	0x5f, 0x7a, 0xb4, 0xe1, 0x07, 0x12, 0xf3, 0x5d, 0x69, 0xe4, 0x1c, 0xa3, 0xee, 0x36, 0xdf, 0xa3,
	0x2d, 0xdf, 0x9b, 0xf1, 0xaf, 0x3b, 0x00, 0x17, 0xef, 0x5f, 0xbe, 0x08, 0xcf, 0x7b, 0x02, 0xc3,
	0xda, 0xa2, 0x29, 0x64, 0x8e, 0xa1, 0xf2, 0xd6, 0x26, 0x5f, 0x25, 0xad, 0x5d, 0x97, 0x46, 0xf3,
	0x5b, 0x46, 0xb3, 0xd6, 0x16, 0x8f, 0x61, 0xa0, 0x64, 0x72, 0x93, 0x66, 0xfe, 0xb3, 0xa3, 0x59,
	0x5f, 0xc9, 0x9f, 0xd2, 0x0c, 0xc5, 0xc7, 0x30, 0x52, 0x68, 0x9c, 0x77, 0xed, 0xfa, 0x28, 0x02,
	0xd8, 0xf9, 0x0c, 0x1e, 0x3a, 0x53, 0x5b, 0x87, 0x3a, 0x69, 0xa2, 0x7b, 0x4c, 0x99, 0x04, 0x78,
	0xea, 0x93, 0x9c, 0xc0, 0x81, 0xca, 0x52, 0x2c, 0x5c, 0xc2, 0xb9, 0x64, 0xed, 0x16, 0x51, 0x9f,
	0x89, 0xfb, 0x1e, 0x9f, 0xa2, 0x71, 0x67, 0xb5, 0x5b, 0xd0, 0xe7, 0xa4, 0xca, 0x12, 0x57, 0x2e,
	0xb1, 0x88, 0x06, 0xfe, 0x73, 0x52, 0x65, 0x57, 0x64, 0x93, 0x73, 0xb9, 0xd2, 0xd7, 0x89, 0xd4,
	0xda, 0x44, 0xc3, 0xe3, 0x2e, 0x39, 0x09, 0x38, 0xd3, 0xda, 0xc4, 0x7f, 0xed, 0xc2, 0xde, 0xfd,
	0x01, 0x0a, 0x01, 0xbb, 0x79, 0xa9, 0x9b, 0x36, 0xf0, 0x6f, 0x11, 0xc1, 0x60, 0x85, 0xc6, 0xd2,
	0x5e, 0xf8, 0x0e, 0x34, 0x26, 0x79, 0x94, 0x41, 0xe9, 0x50, 0x87, 0x06, 0x34, 0xa6, 0x78, 0x0a,
	0xa0, 0x32, 0xaa, 0xc6, 0x24, 0xa9, 0x0e, 0x2d, 0x18, 0x05, 0xe4, 0x5c, 0x8b, 0x4f, 0x61, 0x9c,
	0x95, 0xf3, 0x79, 0x5a, 0xcc, 0x93, 0xda, 0x64, 0xa1, 0x7e, 0x08, 0xd0, 0x2f, 0x26, 0x13, 0x9f,
	0xc1, 0x9e, 0xcc, 0xd0, 0xb8, 0x86, 0xe1, 0x0b, 0x1f, 0x37, 0x18, 0x51, 0x3e, 0x81, 0x91, 0x55,
	0x0b, 0xd4, 0x75, 0x86, 0x26, 0x54, 0x7d, 0x07, 0x88, 0x67, 0xb0, 0x9f, 0xd7, 0x99, 0x4b, 0x55,
	0x59, 0x38, 0x99, 0x16, 0x48, 0xb5, 0x77, 0x4e, 0x86, 0xb3, 0x7f, 0xa1, 0x54, 0x70, 0x51, 0x66,
	0x8b, 0x68, 0xc4, 0x5e, 0xfe, 0x4d, 0x33, 0x57, 0x32, 0xcb, 0x16, 0x65, 0x8e, 0x11, 0x30, 0xde,
	0xda, 0xf4, 0xd5, 0xeb, 0xb2, 0x74, 0xd6, 0x19, 0x59, 0x45, 0x63, 0x76, 0xde, 0x01, 0x34, 0x33,
	0x57, 0x17, 0x05, 0x66, 0x09, 0x16, 0x3a, 0xa9, 0xca, 0xb4, 0x70, 0xd1, 0x9e, 0x9f, 0x99, 0xc7,
	0x5f, 0x15, 0xfa, 0x1d, 0xa1, 0x54, 0x60, 0x60, 0xd2, 0x74, 0x6d, 0x34, 0xe1, 0xc9, 0x8c, 0x3d,
	0x46, 0x93, 0xb5, 0xe2, 0xff, 0xd0, 0xd7, 0x26, 0x5d, 0xa1, 0x89, 0xf6, 0xfd, 0x76, 0x79, 0x8b,
	0x9a, 0xa7, 0xf1, 0xba, 0x9e, 0x27, 0x19, 0xae, 0x30, 0x8b, 0x1e, 0xfa, 0xe6, 0x31, 0xf4, 0x33,
	0x21, 0x1c, 0x58, 0xe6, 0x32, 0x2d, 0xa2, 0x83, 0x10, 0xc8, 0x16, 0xe1, 0xf9, 0x3c, 0x77, 0x69,
	0x15, 0x1d, 0x7a, 0xdc, 0x5b, 0xcc, 0x97, 0x4e, 0xa6, 0x55, 0x24, 0x02, 0x9f, 0x2d, 0xf1, 0x25,
	0xf4, 0x8a, 0x52, 0xa3, 0x8d, 0x8e, 0x8e, 0xbb, 0x27, 0xe3, 0xd3, 0xc3, 0x46, 0xb4, 0xad, 0xb0,
	0x67, 0xde, 0x2f, 0xbe, 0x80, 0x5d, 0x5a, 0xa9, 0xe8, 0x11, 0x8b, 0xaf, 0xe1, 0xdd, 0x29, 0x6c,
	0xc6, 0xee, 0xf8, 0x02, 0x26, 0x5b, 0x82, 0xa7, 0xdd, 0xcc, 0xe7, 0x2e, 0x49, 0x6f, 0xa4, 0x6a,
	0x95, 0x97, 0xcf, 0xdd, 0x39, 0xd9, 0xb4, 0x42, 0xf4, 0x8e, 0xe0, 0xf5, 0x9b, 0x37, 0x22, 0x84,
	0xdd, 0xf1, 0x1f, 0x1d, 0x98, 0x6c, 0x9d, 0x03, 0xea, 0x8b, 0x3f, 0x08, 0x89, 0xdb, 0x54, 0x4d,
	0x3e, 0xf0, 0xd0, 0xd5, 0xa6, 0x42, 0xf1, 0x2d, 0x88, 0x66, 0x29, 0x03, 0x71, 0x89, 0x9b, 0x90,
	0xf9, 0x20, 0x78, 0x7c, 0xca, 0x0b, 0xdc, 0x88, 0x13, 0xe8, 0xad, 0x64, 0x9d, 0xb9, 0x70, 0x52,
	0x44, 0xa8, 0xea, 0x3d, 0x61, 0x4d, 0xf9, 0x4c, 0x10, 0x31, 0x74, 0xe5, 0xda, 0xf2, 0x96, 0x8f,
	0x4f, 0x0f, 0x02, 0xef, 0xec, 0xc3, 0x65, 0x60, 0x91, 0x33, 0xfe, 0x7b, 0x07, 0xc6, 0xf7, 0x42,
	0xe9, 0xb1, 0x1c, 0x1c, 0x54, 0x1b, 0x1e, 0xcb, 0x90, 0xd7, 0xed, 0x53, 0xf0, 0x96, 0x17, 0x6e,
	0x28, 0x9f, 0x11, 0x52, 0x2e, 0xed, 0x8f, 0x77, 0x2b, 0x49, 0x0b, 0x14, 0xf4, 0xe7, 0x73, 0x4e,
	0x19, 0xba, 0x4f, 0xa9, 0xa4, 0x5b, 0x04, 0x15, 0x36, 0x14, 0x82, 0xc4, 0xd7, 0x70, 0x18, 0x28,
	0x77, 0x97, 0x26, 0xa8, 0xf1, 0xa1, 0xe7, 0xb5, 0x97, 0x86, 0x76, 0x7b, 0x8b, 0x4b, 0xbd, 0x0b,
	0xf7, 0xe8, 0x1e, 0x95, 0x3a, 0xd7, 0x66, 0xb5, 0xcb, 0xb4, 0x4a, 0x56, 0x68, 0xd2, 0x9b, 0x4d,
	0x50, 0xa8, 0xcf, 0x7a, 0xb9, 0x4c, 0xab, 0xf7, 0x0c, 0x8b, 0x1f, 0xe0, 0x7f, 0xa1, 0x0f, 0x99,
	0x4d, 0x2c, 0x9a, 0x15, 0x9a, 0x84, 0x0f, 0xf1, 0x90, 0xf9, 0xc2, 0x77, 0x24, 0xb3, 0x97, 0xec,
	0x7a, 0x4b, 0x27, 0xf9, 0x19, 0xf8, 0x2c, 0xc9, 0xb5, 0xb4, 0x98, 0x70, 0x69, 0x23, 0x7f, 0x40,
	0x19, 0x7e, 0x21, 0x2d, 0xbe, 0x93, 0x6e, 0x11, 0xff, 0xd9, 0x81, 0x51, 0x3b, 0x05, 0xf1, 0x15,
	0x1c, 0xca, 0xb5, 0x4d, 0xa4, 0x52, 0x68, 0x2d, 0x3d, 0x9e, 0x0e, 0x93, 0x6f, 0xfb, 0xbe, 0x5c,
	0xdb, 0x33, 0xc6, 0x2f, 0x70, 0x73, 0xae, 0xe9, 0x4d, 0x44, 0x0d, 0x3b, 0x72, 0x17, 0x11, 0xa6,
	0x20, 0xe4, 0xda, 0xfa, 0x35, 0x69, 0x83, 0xc4, 0xf7, 0xf0, 0xe8, 0x5e, 0x08, 0xcf, 0x94, 0x23,
	0xfc, 0x58, 0x0e, 0xdb, 0x08, 0x9e, 0x2d, 0x05, 0x3c, 0x86, 0x01, 0x05, 0xa8, 0x7c, 0x19, 0xe6,
	0xd2, 0x97, 0x6b, 0x3b, 0xcd, 0x97, 0x34, 0x77, 0x72, 0x18, 0x9c, 0xd3, 0xc1, 0xf5, 0xb3, 0x18,
	0xc9, 0xb5, 0x9d, 0x31, 0x10, 0xff, 0xb6, 0x03, 0x93, 0xad, 0xff, 0xd5, 0x58, 0x27, 0xb8, 0x4a,
	0x15, 0xda, 0x24, 0xa7, 0x8a, 0xba, 0xac, 0x13, 0x8f, 0xbc, 0xd1, 0xac, 0x31, 0x79, 0x9b, 0xa8,
	0xb2, 0x2e, 0x1c, 0x17, 0xd0, 0x9b, 0x0d, 0x73, 0x79, 0x3b, 0x25, 0x5b, 0x7c, 0x07, 0x47, 0xe4,
	0xe4, 0xc3, 0x92, 0x58, 0x74, 0x81, 0xd6, 0x65, 0xda, 0x41, 0x2e, 0x6f, 0x5f, 0x92, 0xe7, 0x12,
	0x9d, 0xa7, 0x47, 0x30, 0x08, 0x89, 0xa3, 0x5d, 0xfe, 0x4e, 0x63, 0xd2, 0x23, 0x8c, 0x4c, 0x75,
	0x38, 0x49, 0xe1, 0xd5, 0x84, 0xf8, 0x8b, 0x14, 0xc3, 0xe4, 0xce, 0x4d, 0xcf, 0x0c, 0xf7, 0xbc,
	0x65, 0xbc, 0xd1, 0xe2, 0x23, 0x18, 0x4a, 0xbb, 0x29, 0x54, 0x92, 0x96, 0xbc, 0x2c, 0xc3, 0xd9,
	0x80, 0xed, 0xf3, 0x92, 0xc4, 0x52, 0xd4, 0x79, 0xe2, 0x16, 0x06, 0xa5, 0xb6, 0xbc, 0x1a, 0xbd,
	0x19, 0x14, 0x75, 0x7e, 0xe5, 0x91, 0xd3, 0x0f, 0x30, 0x9e, 0x06, 0xfd, 0x56, 0xa8, 0xc4, 0xe7,
	0xd0, 0x7d, 0x8d, 0x4e, 0xec, 0x05, 0x29, 0xf2, 0x5f, 0x40, 0x4f, 0x26, 0xc1, 0xf2, 0x6d, 0x8b,
	0x1f, 0x90, 0x6c, 0x2f, 0xd1, 0x89, 0x6d, 0xfc, 0x09, 0x34, 0xfa, 0x55, 0xcb, 0xf8, 0xc1, 0xe9,
	0x15, 0x0c, 0xe9, 0xdc, 0xfd, 0xb7, 0x59, 0xaf, 0xfb, 0x6c, 0xfc, 0xf8, 0x4f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x49, 0xe3, 0xb4, 0xfd, 0xb3, 0x09, 0x00, 0x00,
}
