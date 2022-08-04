package pwx

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/libopenstorage/openstorage/pkg/grpcserver"
)

// ConnectionParamsBuilder contains dependencies needed for building Dial options and endpoints
// to SDK/iSDK and legacy REST API connections
type ConnectionParamsBuilder struct {
	Config  *ConnectionParamsBuilderConfig
	kubeOps core.Ops
}

// ConnectionParamsBuilderConfig contains default values for all PX legacy rest API and SDk API connections
// also can be used for redefining names of Env variables which are used for external configuration
type ConnectionParamsBuilderConfig struct {
	// DefaultServiceName is the default name of the Portworx service
	DefaultServiceName string
	// DefaultServiceNamespaceName is the kubernetes namespace in which Portworx daemon set runs
	DefaultServiceNamespaceName string
	// DefaultRestPortName is name of Porx legacy REST API port in service
	DefaultRestPortName string
	// DefaultRestPortNameSecured is the name of the TLS Secured (if enabled) Porx legacy REST API port in service
	DefaultRestPortNameSecured string
	// DefaultSDKPortName is name of Porx SDK/iSDK API port in service
	DefaultSDKPortName string
	// DefaultTokenIssuer is the default value for token issuer
	DefaultTokenIssuer string

	// Environment variables names to get  values config properties
	// EnableTLSEnv is used to set environment variable name which should be read to enable TLS on the connections
	EnableTLSEnv string
	// NamespaceNameEnv is used to set environment variable name which should be read to fetch Porx namespace
	NamespaceNameEnv string
	// ServiceNameEnv is used to set environment variable name which should be read to fetch Porx service
	ServiceNameEnv string
	// CaCertSecretEnv is used to set environment variable name which should be read to fetch secret
	// containing certificate used for protecting Porx APIs
	CaCertSecretEnv    string
	CaCertSecretKeyEnv string
	// TokenIssuerEnv is used to set environment variable name which should be read to fetch the token issuer
	TokenIssuerEnv string

	// StaticEndpointEnv can be used to overwrite Porx endpoint
	StaticEndpointEnv string
	// StaticSDKPortEnv can be used to overwrite Porx SDK port
	StaticSDKPortEnv string
	// StaticRestPortEnv can be used to overwrite Porx Legacy Rest API port
	StaticRestPortEnv string
	// if AuthEnabled is set to true params builder generates additional dial option which injects authorization token
	AuthEnabled bool
	// AuthTokenGen function need to be used to generate Authorization token
	AuthTokenGenerator func() (string, error)
}

// NewConnectionParamsBuilderDefaultConfig returns ConnectionParamsBuilderConfig with default values set
func NewConnectionParamsBuilderDefaultConfig() *ConnectionParamsBuilderConfig {
	return &ConnectionParamsBuilderConfig{
		DefaultServiceName:          "portworx-service",
		DefaultServiceNamespaceName: "kube-system",
		DefaultRestPortName:         "px-api",
		DefaultRestPortNameSecured:  "px-api-tls",
		DefaultTokenIssuer:          "apps.portworx.io",
		DefaultSDKPortName:          "px-sdk",
		EnableTLSEnv:                "PX_ENABLE_TLS",
		NamespaceNameEnv:            "PX_NAMESPACE",
		ServiceNameEnv:              "PX_SERVICE_NAME",
		CaCertSecretEnv:             "PX_CA_CERT_SECRET",
		CaCertSecretKeyEnv:          "PX_CA_CERT_SECRET_KEY",
		TokenIssuerEnv:              "PX_JWT_ISSUER",
		StaticEndpointEnv:           "PX_ENDPOINT",
		StaticSDKPortEnv:            "PX_SDK_PORT",
		StaticRestPortEnv:           "PX_API_PORT",
		AuthEnabled:                 false,
		AuthTokenGenerator:          func() (string, error) { return "", fmt.Errorf("auth token generator func is not set") },
	}
}

// NewConnectionParamsBuilder constructor function to create ConnectionParamsBuilder with needed dependencies
func NewConnectionParamsBuilder(ops core.Ops, params *ConnectionParamsBuilderConfig) (*ConnectionParamsBuilder, error) {
	if params == nil {
		return nil, fmt.Errorf("ConnectionParamsBuilderConfig cannot be nil")
	}

	return &ConnectionParamsBuilder{
		Config:  params,
		kubeOps: ops,
	}, nil
}

// BuildClientsEndpoints returns two endpoints for PX MGMT API and gRPC SDK/iSDL API
func (cpb *ConnectionParamsBuilder) BuildClientsEndpoints() (string, string, error) {
	var endpoint string

	pxMgmtEndpoint, sdkEndpoint, err := cpb.checkStaticEndpoints()
	if err != nil && !os.IsNotExist(err) {
		return "", "", err
	}
	if err == nil {
		logrus.Infof("Using static %s endpoint for portworx REST API", pxMgmtEndpoint)
		logrus.Infof("Using static %s endpoint for portworx gRPC API", sdkEndpoint)

		return pxMgmtEndpoint, sdkEndpoint, nil
	}

	// Check if service name and namespace is provided
	// as environment variables
	serviceName := getPxService(cpb.Config.ServiceNameEnv, cpb.Config.DefaultServiceName)
	ns := getPxNamespace(cpb.Config.NamespaceNameEnv, cpb.Config.DefaultServiceNamespaceName)

	svc, err := cpb.kubeOps.GetService(serviceName, ns)
	if err != nil {
		return "", "", fmt.Errorf("failed to get k8s service specification: %v", err)
	}

	endpoint = fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)

	var restPort int
	var restPortSecured int
	var finalRestPort int // the port passed to the caller (restPortSecured if available, else restPort)
	var sdkPort int

	// Get the ports from service
	for _, svcPort := range svc.Spec.Ports {
		if svcPort.Name == cpb.Config.DefaultSDKPortName &&
			svcPort.Port != 0 {
			sdkPort = int(svcPort.Port)
		} else if svcPort.Name == cpb.Config.DefaultRestPortName &&
			svcPort.Port != 0 {
			restPort = int(svcPort.Port)
		} else if svcPort.Name == cpb.Config.DefaultRestPortNameSecured &&
			svcPort.Port != 0 {
			restPortSecured = int(svcPort.Port)
		}
	}
	// if secured REST port (9023) is available, use it instead of legacy 9001
	if restPortSecured != 0 {
		finalRestPort = restPortSecured
	} else {
		finalRestPort = restPort
	}

	// check if the ports were parsed
	if sdkPort == 0 || finalRestPort == 0 {
		err := fmt.Errorf("%s in %s namespace does not contain %s and either of %s or %s ports set", serviceName, ns, cpb.Config.DefaultSDKPortName, cpb.Config.DefaultRestPortName, cpb.Config.DefaultRestPortNameSecured)
		logrus.Errorf(err.Error())
		return "", "", err
	}

	scheme := "http"
	var isTLSEnabled = isTLSEnabled(cpb.Config.EnableTLSEnv)
	if isTLSEnabled && finalRestPort == restPortSecured {
		scheme = "https" // legacy 9001 port is never TLS secured, irrespective of the value of PX_ENABLE_TLS
	}
	pxMgmtEndpoint = fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(endpoint, strconv.Itoa(finalRestPort)))
	sdkEndpoint = net.JoinHostPort(endpoint, strconv.Itoa(sdkPort))

	logrus.Infof("Using %s as endpoint for portworx REST API", pxMgmtEndpoint)
	logrus.Infof("Using %s as endpoint for portworx gRPC API", sdkEndpoint)

	return pxMgmtEndpoint, sdkEndpoint, nil
}

// BuildTlsConfig returns the TLS configuration (if needed) to connect to the Porx API
func (cpb *ConnectionParamsBuilder) BuildTlsConfig() (*tls.Config, error) {
	if !isTLSEnabled(cpb.Config.EnableTLSEnv) {
		return nil, nil
	}
	rootCA, err := cpb.getCaCertBytes()
	if err != nil {
		return nil, err
	}
	tlsCfg := &tls.Config{}
	setRootCA(tlsCfg, rootCA)

	return tlsCfg, nil
}

func setRootCA(tlsCfg *tls.Config, rootCA []byte) error {
	clientCertPool, err := x509.SystemCertPool()
	if err != nil || clientCertPool == nil {
		logrus.Warnf("Warning: Failed to load system certs, root CA param data only: %v\n", err)
	}

	if clientCertPool == nil && len(rootCA) > 0 {
		// Only create if system cert is nil && rootCA exists
		clientCertPool = x509.NewCertPool()
	}

	if len(rootCA) > 0 { // rootCA exists, append it
		clientCertPool.AppendCertsFromPEM(rootCA)
	}

	tlsCfg.RootCAs = clientCertPool
	return nil
}

// BuildDialOps build slice of grpc.DialOption to connect to SDK API
func (cpb *ConnectionParamsBuilder) BuildDialOps() ([]grpc.DialOption, error) {
	var dialOptions []grpc.DialOption
	var isTLSEnabled = isTLSEnabled(cpb.Config.EnableTLSEnv)

	if cpb.Config.AuthEnabled {
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(grpcserver.NewCredsInjector(cpb.Config.AuthTokenGenerator, isTLSEnabled)))
	}

	if !isTLSEnabled {
		dialOptions = append(dialOptions, grpc.WithInsecure())
		return dialOptions, nil
	}

	rootCA, err := cpb.getCaCertBytes()
	if err != nil {
		return nil, err
	}

	tlsDialOptions, err := grpcserver.GetTlsDialOptions(rootCA)
	if err != nil {
		return nil, fmt.Errorf("unable to build TLS gRPC connection options: %v", err)
	}

	dialOptions = append(dialOptions, tlsDialOptions...)

	return dialOptions, nil
}

func (cpb *ConnectionParamsBuilder) getCaCertBytes() ([]byte, error) {
	var rootCA []byte
	var caCertSecretName = strings.TrimSpace(os.Getenv(cpb.Config.CaCertSecretEnv))
	var caCertSecretKey = strings.TrimSpace(os.Getenv(cpb.Config.CaCertSecretKeyEnv))
	var pxNamespace = getPxNamespace(cpb.Config.NamespaceNameEnv, cpb.Config.DefaultServiceNamespaceName)

	if caCertSecretName == "" {
		logrus.Infof("CA cert secret name was not provided using env %s", cpb.Config.CaCertSecretEnv)
		return rootCA, nil
	} else if caCertSecretKey == "" {
		return nil, fmt.Errorf("failed to load CA cert from secret: %s, secret key should be defined using env %s",
			caCertSecretName, cpb.Config.CaCertSecretKeyEnv)
	}

	secret, err := cpb.kubeOps.GetSecret(caCertSecretName, pxNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA cert secret: %v", err)
	}

	exist := false
	rootCA, exist = secret.Data[caCertSecretKey]
	if !exist {
		return nil, fmt.Errorf("failed to load CA cert from secret: %s using key: %s", caCertSecretName, caCertSecretKey)
	}
	if len(rootCA) == 0 {
		return nil, fmt.Errorf("CA cert fetchecd from secret: %s using key: %s is empty", caCertSecretName, caCertSecretKey)
	}
	return rootCA, nil
}

func (cpb *ConnectionParamsBuilder) checkStaticEndpoints() (string, string, error) {
	if cpb.Config.StaticEndpointEnv == "" || cpb.Config.StaticRestPortEnv == "" || cpb.Config.StaticSDKPortEnv == "" {
		return "", "", os.ErrNotExist
	}

	endpoint, staticRESTPort, staticSDKPort := os.Getenv(cpb.Config.StaticEndpointEnv), os.Getenv(cpb.Config.StaticRestPortEnv), os.Getenv(cpb.Config.StaticSDKPortEnv)
	if endpoint == "" || staticRESTPort == "" || staticSDKPort == "" {
		return "", "", os.ErrNotExist
	}

	restPort, err := strconv.Atoi(staticRESTPort)
	if err != nil {
		return "", "", fmt.Errorf("cannot parse static REST port value, err: %s", err.Error())
	}
	sdkPort, err := strconv.Atoi(staticSDKPort)
	if err != nil {
		return "", "", fmt.Errorf("cannot parse static SDK port value, err: %s", err.Error())
	}

	if sdkPort < 1 {
		return "", "", fmt.Errorf("static SDK port value should be greater than 0")
	}

	if restPort < 1 {
		return "", "", fmt.Errorf("static REST port value should be greater than 0")
	}

	scheme := "http"
	if isTLSEnabled(cpb.Config.EnableTLSEnv) {
		scheme = "https"
	}
	pxMgmtEndpoint := fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(endpoint, strconv.Itoa(restPort)))
	sdkEndpoint := net.JoinHostPort(endpoint, strconv.Itoa(sdkPort))

	return pxMgmtEndpoint, sdkEndpoint, nil
}

func isTLSEnabled(pxEnableTLSEnv string) bool {
	if v, err := strconv.ParseBool(os.Getenv(pxEnableTLSEnv)); err == nil {
		return v
	}
	return false
}

func getPxNamespace(pxNamespaceEnv, defaultNamespace string) string {
	namespace := os.Getenv(pxNamespaceEnv)
	if len(namespace) == 0 {
		namespace = defaultNamespace
	}
	return namespace
}

func getPxService(pxServiceEnv, defaultServiceName string) string {
	service := os.Getenv(pxServiceEnv)
	if len(service) == 0 {
		service = defaultServiceName
	}
	return service
}
