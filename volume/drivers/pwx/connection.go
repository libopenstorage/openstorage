package pwx

import (
	"fmt"
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
	// DefaultSDKPortName is name of Porx SDK/iSDK API port in service
	DefaultSDKPortName string
	// DefaultTokenIssuer is the default value for token issuer
	DefaultTokenIssuer string

	// Environment variables names to get  values config properties
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
	var sdkPort int

	// Get the ports from service
	for _, svcPort := range svc.Spec.Ports {
		if svcPort.Name == cpb.Config.DefaultSDKPortName &&
			svcPort.Port != 0 {
			sdkPort = int(svcPort.Port)
		} else if svcPort.Name == cpb.Config.DefaultRestPortName &&
			svcPort.Port != 0 {
			restPort = int(svcPort.Port)
		}
	}

	// check if the ports were parsed
	if sdkPort == 0 || restPort == 0 {
		err := fmt.Errorf("%s in %s namespace does not contain %s and %s ports set", serviceName, ns, cpb.Config.DefaultSDKPortName, cpb.Config.DefaultRestPortName)
		logrus.Errorf(err.Error())
		return "", "", err
	}

	// PX mgmt API was decided not to be protected with TLS
	pxMgmtEndpoint = fmt.Sprintf("http://%s:%d", endpoint, restPort)
	sdkEndpoint = fmt.Sprintf("%s:%d", endpoint, sdkPort)

	logrus.Infof("Using %s as endpoint for portworx REST API", pxMgmtEndpoint)
	logrus.Infof("Using %s as endpoint for portworx gRPC API", sdkEndpoint)

	return pxMgmtEndpoint, sdkEndpoint, nil
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

	var rootCA []byte
	var caCertSecretName = strings.TrimSpace(os.Getenv(cpb.Config.CaCertSecretEnv))
	var caCertSecretKey = strings.TrimSpace(os.Getenv(cpb.Config.CaCertSecretKeyEnv))
	var pxNamespace = getPxNamespace(cpb.Config.NamespaceNameEnv, cpb.Config.DefaultServiceNamespaceName)

	if caCertSecretName != "" && caCertSecretKey == "" {
		return nil, fmt.Errorf("failed to load CA cert from secret: %s, secret key should be defined using env PX_CA_CERT_SECRET_KEY", caCertSecretName)
	}

	if caCertSecretName != "" && caCertSecretKey != "" {
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
	}

	tlsDialOptions, err := grpcserver.GetTlsDialOptions(rootCA)
	if err != nil {
		return nil, fmt.Errorf("unable to build TLS gRPC connection options: %v", err)
	}

	dialOptions = append(dialOptions, tlsDialOptions...)

	return dialOptions, nil
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
		return "", "", fmt.Errorf("static SDK port value sould be greater than 0")
	}

	if restPort < 1 {
		return "", "", fmt.Errorf("static REST port value sould be greater than 0")
	}

	pxMgmtEndpoint, sdkEndpoint := fmt.Sprintf("http://%s:%d", endpoint, restPort), fmt.Sprintf("%s:%d", endpoint, sdkPort)

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
