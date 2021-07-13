package pwx_test

import (
	"fmt"
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"
	"github.com/portworx/sched-ops/k8s/core"
	v1 "k8s.io/api/core/v1"

	"github.com/libopenstorage/openstorage/api/server/mock"
	"github.com/libopenstorage/openstorage/volume/drivers/pwx"
)

const (
	pxNamespaceNameEnv   = "PX_NAMESPACE"
	pxServiceNameEnv     = "PX_SERVICE_NAME"
	pxCaCertSecretEnv    = "PX_CA_CERT_SECRET"
	pxCaCertSecretKeyEnv = "PX_CA_CERT_SECRET_KEY"
	pxEnableTLSEnv       = "PX_ENABLE_TLS"
	pxRestPort           = "px-api-tls"
	pxSdkPort            = "px-sdk"
	pxEndpointEnv        = "PX_ENDPOINT"
	StaticSDKPortEnv     = "PX_SDK_PORT"
	StaticRestPortEnv    = "PX_API_PORT"
)

func getSvc() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "portworx-service",
			Namespace: "kube-system",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: pxSdkPort,
					Port: 9999,
				},
				{
					Name: pxRestPort,
					Port: 9901,
				},
			},
		},
	}
}

func getNonDefaultNSSvc() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "portworx-service",
			Namespace: "non-default-ns",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: pxSdkPort,
					Port: 9999,
				},
				{
					Name: pxRestPort,
					Port: 9901,
				},
			},
		},
	}
}

func getSecret() *v1.Secret {
	return &v1.Secret{
		Data: map[string][]byte{
			"ca-cert": []byte("-----BEGIN CERTIFICATE-----\nMIIDdTCCAl2gAwIBAgIJAI6emGCW7kplMA0GCSqGSIb3DQEBCwUAMFExCzAJBgNV\nBAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQg\nQ29tcGFueSBMdGQxDTALBgNVBAMMBHB4dXAwHhcNMTkwODI0MDEyMzE1WhcNNDcw\nMTA4MDEyMzE1WjBRMQswCQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5\nMRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRkMQ0wCwYDVQQDDARweHVwMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAlmzy8sygifVaoFFbxpNIk39n\nybtiYBOIGa/AacACDw3opBnRLBd+uSY8P/nPFg3s3ZSs4b/kykHo+kMfK8m1BT7a\n3d4e4COcansb3qtp62ipEnp58e0dcJis0kTcvovStLN0gTin+IHfQtfQrVzw51KL\n134dNUeon0A6oSaXvnx0p3gMg9cS8L6l9Ih09/8hNVxm1KTYam1XXHf8vdi3RI5B\n9ClbDdiFvNjYvDJP//Bao+4yJrbyasRgmJBACLWbTxxp6Ph04WnxHdE2HYcxDmcP\nEsn3R/x5bQSQJr3gILS75sh3Xr0djbmdpKwi8zG7jzOuLeRdRu4uYWOzP6xJ+QID\nAQABo1AwTjAdBgNVHQ4EFgQUf7e2yL163c1SafdrzORVTEGDU4AwHwYDVR0jBBgw\nFoAUf7e2yL163c1SafdrzORVTEGDU4AwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0B\nAQsFAAOCAQEAX9gOeYdP/hbuTDJVRR4ipQrjhAW496vW0FIcSGG11H/P8PNidAQx\nveOXWEwKRZbSO+TxMbXw+fE+URstDq8eEQK7ns0yZve4GbeB+PVMyjRuc1hI/uR3\nwzTuYsEtGhXlojLkLaS8O6ARMwg11maBca3N4V7IlD/eNNtTqMiSNvYimpIEOKUj\nlvkx32Wi2bcSyAdqHkRGN3AArrzEz+ybtoSFC9M6/Xsduc/68bIZztZveuacmt9T\njCUt3FvwQhhoN2Ax6F60864QiMiPE5y9iFcoCQWz9yPb47qSr79SlqhOzKNAgE6R\nx/sTw+ZJP3ucSLJ5Fkxp5H8EzXv0HsS1fA==\n-----END CERTIFICATE-----"),
		},
	}
}

var ops core.Ops

func TestMain(m *testing.M) {
	mc := gomock.NewController(&utils.SafeGoroutineTester{})
	mockOps := mock.NewMockOps(mc)
	ops = mockOps

	mockOps.EXPECT().GetService("portworx-service-not-found", "kube-system").Return(nil, fmt.Errorf("not found"))
	mockOps.EXPECT().GetService("portworx-service", "kube-system").Return(getSvc(), nil)
	mockOps.EXPECT().GetService("portworx-service", "non-default-ns").Return(getNonDefaultNSSvc(), nil)
	mockOps.EXPECT().GetService("portworx-service", "non-default-ns").Return(getNonDefaultNSSvc(), nil)
	mockOps.EXPECT().GetService("portworx-service", "kube-system").Return(getSvc(), nil)
	mockOps.EXPECT().GetService("portworx-service-no-ip", "kube-system").Return(&v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: pxSdkPort,
					Port: 9999,
				},
				{
					Name: pxRestPort,
					Port: 9901,
				},
			},
		},
	}, nil)
	mockOps.EXPECT().GetService("portworx-service-no-ports", "kube-system").Return(&v1.Service{
		Spec: v1.ServiceSpec{
			ClusterIP: "8.8.4.4",
		},
	}, nil)
	mockOps.EXPECT().GetService("portworx-service", "kube-system").Return(getSvc(), nil)
	mockOps.EXPECT().GetService("portworx-service", "kube-system").Return(getSvc(), nil)
	mockOps.EXPECT().GetService("portworx-service-ports-zeroed", "kube-system").Return(&v1.Service{
		Spec: v1.ServiceSpec{
			ClusterIP: "8.8.4.4",
			Ports: []v1.ServicePort{
				{
					Name: pxSdkPort,
					Port: 0,
				},
				{
					Name: pxRestPort,
					Port: 0,
				},
			},
		},
	}, nil)

	mockOps.EXPECT().GetSecret("px-ca-cert-secret", "kube-system").Return(getSecret(), nil)
	mockOps.EXPECT().GetSecret("px-ca-cert-secret", "invalid").Return(nil, fmt.Errorf("not found"))
	mockOps.EXPECT().GetSecret("px-ca-cert-secret-is-empty", "kube-system").Return(&v1.Secret{
		Data: map[string][]byte{
			"ca-cert": nil,
		},
	}, nil)
	mockOps.EXPECT().GetSecret("px-ca-cert-secret-is-broken", "kube-system").Return(&v1.Secret{
		Data: map[string][]byte{
			"ca-cert": []byte("this is broken CA cert"),
		},
	}, nil)
	mockOps.EXPECT().GetSecret("px-ca-cert-secret", "kube-system").Return(getSecret(), nil)

	code := m.Run()

	os.Exit(code)
}

func TestPortworx_buildClientsEndpoints_Error_WhenServiceDoesNotExist(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxServiceNameEnv, "portworx-service-not-found", pxNamespaceNameEnv, "kube-system")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	_, _, err = paramsBuilder.BuildClientsEndpoints()
	if err == nil {
		t.Fatal("should return error when service does not exist")
	}
}

func TestPortworx_buildClientsEndpoints_Error_WhenServiceDoesNotHavePxRestPort(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxServiceNameEnv, "portworx-service-no-ports", pxNamespaceNameEnv, "kube-system")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	_, _, err = paramsBuilder.BuildClientsEndpoints()
	if err == nil {
		t.Fatal("should return error when service does not have ports")
	}
}

func TestPortworx_buildClientsEndpoints_Error_WhenServiceHasPxRestPortZeroed(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxServiceNameEnv, "portworx-service-ports-zeroed", pxNamespaceNameEnv, "kube-system")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	_, _, err = paramsBuilder.BuildClientsEndpoints()
	if err == nil {
		t.Fatal("should return error when service has ports equal to 0")
	}
}

func TestPortworx_buildClientsEndpoints_OK_WithDefaultNsAndService_TLS(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	pxMgmtEndpoint, sdkEndpoint, err := paramsBuilder.BuildClientsEndpoints()
	if err != nil {
		t.Fatalf("should build endpoints when service and ns is not defined in env varaibles: %+v", err)
	}

	if pxMgmtEndpoint != "https://portworx-service.kube-system:9901" {
		t.Fatalf("should build pxMgmtEndpoint actual: %q, required: %q", pxMgmtEndpoint, "https://portworx-service.kube-system:9901")
	}

	if sdkEndpoint != "portworx-service.kube-system:9999" {
		t.Fatalf("should build sdkEndpoint actual: %q, required: %q", sdkEndpoint, "portworx-service.kube-system:9999")
	}
}

func TestPortworx_buildClientsEndpoints_OK_WithDefaultNsAndService_NO_TLS(t *testing.T) {
	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	pxMgmtEndpoint, sdkEndpoint, err := paramsBuilder.BuildClientsEndpoints()
	if err != nil {
		t.Fatalf("should build endpoints when service and ns is not defined in env varaibles: %+v", err)
	}

	if pxMgmtEndpoint != "http://portworx-service.kube-system:9901" {
		t.Fatalf("should build pxMgmtEndpoint actual: %q, required: %q", pxMgmtEndpoint, "http://portworx-service.kube-system:9901")
	}

	if sdkEndpoint != "portworx-service.kube-system:9999" {
		t.Fatalf("should build sdkEndpoint actual: %q, required: %q", sdkEndpoint, "portworx-service.kube-system:9999")
	}
}

func TestPortworx_buildClientsEndpoints_OK_WithNonDefaultNsAndService_NO_TLS(t *testing.T) {
	cleaner := setEnvs(t, pxNamespaceNameEnv, "non-default-ns")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	pxMgmtEndpoint, sdkEndpoint, err := paramsBuilder.BuildClientsEndpoints()
	if err != nil {
		t.Fatalf("should build endpoints when service and ns is not defined in env varaibles: %+v", err)
	}

	if pxMgmtEndpoint != "http://portworx-service.non-default-ns:9901" {
		t.Fatalf("should build pxMgmtEndpoint actual: %q, required: %q", pxMgmtEndpoint, "http://portworx-service.non-default-ns:9901")
	}

	if sdkEndpoint != "portworx-service.non-default-ns:9999" {
		t.Fatalf("should build sdkEndpoint actual: %q, required: %q", sdkEndpoint, "portworx-service.non-default-ns:9999")
	}
}

func TestPortworx_buildClientsEndpoints_OK_WithNonDefaultNsAndService_TLS(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxNamespaceNameEnv, "non-default-ns")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	pxMgmtEndpoint, sdkEndpoint, err := paramsBuilder.BuildClientsEndpoints()
	if err != nil {
		t.Fatalf("should build endpoints when service and ns is not defined in env varaibles: %+v", err)
	}

	if pxMgmtEndpoint != "https://portworx-service.non-default-ns:9901" {
		t.Fatalf("should build pxMgmtEndpoint actual: %q, required: %q", pxMgmtEndpoint, "https://portworx-service.non-default-ns:9901")
	}

	if sdkEndpoint != "portworx-service.non-default-ns:9999" {
		t.Fatalf("should build sdkEndpoint actual: %q, required: %q", sdkEndpoint, "portworx-service.non-default-ns:9999")
	}
}

func TestPortworx_buildClientsEndpoints_OK_WithStaticEndpointAndPorts_NO_TLS(t *testing.T) {
	cleaner := setEnvs(t, pxNamespaceNameEnv, "non-default-ns", pxEndpointEnv, "k8s-node-0", StaticSDKPortEnv, "9020", StaticRestPortEnv, "9039")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	pxMgmtEndpoint, sdkEndpoint, err := paramsBuilder.BuildClientsEndpoints()
	if err != nil {
		t.Fatalf("should build endpoints when service and ns is not defined in env varaibles: %+v", err)
	}

	if pxMgmtEndpoint != "http://k8s-node-0:9039" {
		t.Fatalf("should build pxMgmtEndpoint actual: %q, required: %q", pxMgmtEndpoint, "http://k8s-node-0:9039")
	}

	if sdkEndpoint != "k8s-node-0:9020" {
		t.Fatalf("should build sdkEndpoint actual: %q, required: %q", sdkEndpoint, "k8s-node-0:9020")
	}
}

func TestPortworx_buildClientsEndpoints_OK_WithStaticEndpointAndPorts_TLS(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxNamespaceNameEnv, "non-default-ns", pxEndpointEnv, "k8s-node-0", StaticSDKPortEnv, "9020", StaticRestPortEnv, "9039")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	pxMgmtEndpoint, sdkEndpoint, err := paramsBuilder.BuildClientsEndpoints()
	if err != nil {
		t.Fatalf("should build endpoints when service and ns is not defined in env varaibles: %+v", err)
	}

	if pxMgmtEndpoint != "https://k8s-node-0:9039" {
		t.Fatalf("should build pxMgmtEndpoint actual: %q, required: %q", pxMgmtEndpoint, "https://k8s-node-0:9039")
	}

	if sdkEndpoint != "k8s-node-0:9020" {
		t.Fatalf("should build sdkEndpoint actual: %q, required: %q", sdkEndpoint, "k8s-node-0:9020")
	}
}

func TestPortworx_buildClientsEndpoints_Err_WithStaticSDKPortIncorrect(t *testing.T) {
	cleaner := setEnvs(t, pxNamespaceNameEnv, "non-default-ns", pxEndpointEnv, "k8s-node-0", StaticSDKPortEnv, "0", StaticRestPortEnv, "9999")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	_, _, err = paramsBuilder.BuildClientsEndpoints()
	if err == nil {
		t.Fatal("should return error when static port is negative")
	}
}

func TestPortworx_buildClientsEndpoints_Err_WithStaticRestPortIncorrect(t *testing.T) {
	cleaner := setEnvs(t, pxNamespaceNameEnv, "non-default-ns", pxEndpointEnv, "k8s-node-0", StaticSDKPortEnv, "9020", StaticRestPortEnv, "-12")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	_, _, err = paramsBuilder.BuildClientsEndpoints()
	if err == nil {
		t.Fatal("should return error when static port is negative")
	}
}

func TestPortworx_buildClientsEndpoints_Err_WithStaticRestPortRandomString(t *testing.T) {
	cleaner := setEnvs(t, pxNamespaceNameEnv, "non-default-ns", pxEndpointEnv, "k8s-node-0", StaticSDKPortEnv, "9020", StaticRestPortEnv, "ololo")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	_, _, err = paramsBuilder.BuildClientsEndpoints()
	if err == nil {
		t.Fatal("should return error when static port is negative")
	}
}

func TestPortworx_buildClientsEndpoints_Err_WithStaticSDKPortRandomString(t *testing.T) {
	cleaner := setEnvs(t, pxNamespaceNameEnv, "non-default-ns", pxEndpointEnv, "k8s-node-0", StaticSDKPortEnv, "hi", StaticRestPortEnv, "9001")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	_, _, err = paramsBuilder.BuildClientsEndpoints()
	if err == nil {
		t.Fatal("should return error when static port is negative")
	}
}

func TestPortworx_buildClientsEndpoints_OK_WithDefaultNsAndServiceWithEmptyStaticRestPort_NO_TLS(t *testing.T) {
	cleaner := setEnvs(t, pxEndpointEnv, "k8s-node-0", StaticSDKPortEnv, "9020")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	// without TLS enabled
	pxMgmtEndpoint, sdkEndpoint, err := paramsBuilder.BuildClientsEndpoints()
	if err != nil {
		t.Fatalf("should build endpoints when service and ns is not defined in env varaibles: %+v", err)
	}

	if pxMgmtEndpoint != "http://portworx-service.kube-system:9901" {
		t.Fatalf("should build pxMgmtEndpoint actual: %q, required: %q", pxMgmtEndpoint, "http://portworx-service.kube-system:9901")
	}

	if sdkEndpoint != "portworx-service.kube-system:9999" {
		t.Fatalf("should build sdkEndpoint actual: %q, required: %q", sdkEndpoint, "portworx-service.kube-system:9999")
	}
}

func TestPortworx_buildClientsEndpoints_OK_WithDefaultNsAndServiceWithEmptyStaticRestPort_TLS(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxEndpointEnv, "k8s-node-0", StaticSDKPortEnv, "9020")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}

	// with TLS enabled
	pxMgmtEndpoint, sdkEndpoint, err := paramsBuilder.BuildClientsEndpoints()
	if err != nil {
		t.Fatalf("should build endpoints when service and ns is not defined in env varaibles: %+v", err)
	}

	if pxMgmtEndpoint != "https://portworx-service.kube-system:9901" {
		t.Fatalf("should build pxMgmtEndpoint actual: %q, required: %q", pxMgmtEndpoint, "https://portworx-service.kube-system:9901")
	}

	if sdkEndpoint != "portworx-service.kube-system:9999" {
		t.Fatalf("should build sdkEndpoint actual: %q, required: %q", sdkEndpoint, "portworx-service.kube-system:9999")
	}
}

func TestPortworx_dialOptions_Error_WhenSecretDoesNotExist(t *testing.T) {
	cleaner := setEnvs(
		t,
		pxEnableTLSEnv, "true",
		pxCaCertSecretEnv, "px-ca-cert-secret",
		pxCaCertSecretKeyEnv, "ca-cert",
		pxNamespaceNameEnv, "invalid")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}
	_, err = paramsBuilder.BuildDialOps()
	if err == nil {
		t.Fatal("should return error when secret cannot be found")
	}
}

func TestPortworx_dialOptions_OK_WhenTLSDisabledAndInvalidNs(t *testing.T) {
	cleaner := setEnvs(
		t,
		pxEnableTLSEnv, "false",
		pxCaCertSecretEnv, "px-ca-cert-secret",
		pxCaCertSecretKeyEnv, "ca-cert",
		pxNamespaceNameEnv, "invalid")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}
	_, err = paramsBuilder.BuildDialOps()
	if err != nil {
		t.Fatalf("should not get error: %+v", err)
	}
}

func TestPortworx_dialOptions_OK_WhenTLSEnabled(t *testing.T) {
	cleaner := setEnvs(t,
		pxEnableTLSEnv, "true",
		pxCaCertSecretEnv, "px-ca-cert-secret",
		pxCaCertSecretKeyEnv, "ca-cert",
		pxNamespaceNameEnv, "kube-system")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}
	_, err = paramsBuilder.BuildDialOps()
	if err != nil {
		t.Fatalf("should not get error: %+v", err)
	}
}

func TestPortworx_dialOptions_OK_WhenTLSEnabledAndWrongSecretKey(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxCaCertSecretEnv, "px-ca-cert-secret", pxCaCertSecretKeyEnv, "ca-cert-wrong")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}
	_, err = paramsBuilder.BuildDialOps()
	if err == nil {
		t.Fatalf("should not get error when secret key is wrong")
	}
}

func TestPortworx_dialOptions_OK_WhenTLSEnabledWithoutCa(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(nil, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}
	_, err = paramsBuilder.BuildDialOps()
	if err != nil {
		t.Fatalf("should not get error: %+v", err)
	}
}

func TestPortworx_dialOptions_Error_WhenTLSEnabledAndCertIsEmpty(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxCaCertSecretEnv, "px-ca-cert-secret-is-empty", pxCaCertSecretKeyEnv, "ca-cert")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}
	_, err = paramsBuilder.BuildDialOps()
	if err == nil {
		t.Fatalf("should get error when certificate is empty")
	}
}

func TestPortworx_dialOptions_Error_WhenTLSEnabledAndCertIsBroken(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxCaCertSecretEnv, "px-ca-cert-secret-is-broken", pxCaCertSecretKeyEnv, "ca-cert")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}
	_, err = paramsBuilder.BuildDialOps()
	if err == nil {
		t.Fatalf("should not get error")
	}
}

func TestPortworx_dialOptions_Error_WhenTLSEnabledAndCertSecretKeyIsEmpty(t *testing.T) {
	cleaner := setEnvs(t, pxEnableTLSEnv, "true", pxCaCertSecretEnv, "px-ca-cert-secret")
	defer cleaner()

	paramsBuilder, err := pwx.NewConnectionParamsBuilder(ops, pwx.NewConnectionParamsBuilderDefaultConfig())
	if err != nil {
		t.Fatal("ConnectionParamsBuilder creation error")
	}
	_, err = paramsBuilder.BuildDialOps()
	if err == nil {
		t.Fatalf("should not get error: %+v", err)
	}
}

func TestPortworx_dialOptions_Error_WhenConfigIsNil(t *testing.T) {
	_, err := pwx.NewConnectionParamsBuilder(ops, nil)
	if err == nil {
		t.Fatal("ConnectionParamsBuilder creation should return error when config is nil")
	}
}

func setEnvs(t *testing.T, vars ...string) func() {
	t.Helper()
	if len(vars)%2 != 0 {
		t.Fatalf("number of args should be even")
	}

	var cleanUp []func()

	for i := 0; i < len(vars); i += 2 {
		env := vars[i]
		value := vars[i+1]
		existingValue := os.Getenv(env)
		err := os.Setenv(env, value)
		if err != nil {
			t.Fatalf("cannot set env varaibles for test: %+v", err)
		}

		cleanUp = append(cleanUp, func() {
			err = os.Setenv(env, existingValue)
			if err != nil {
				t.Errorf("cannot set back env varaibles for test: %+v", err)
			}
		})
	}

	cleaner := func() {
		for _, f := range cleanUp {
			f()
		}
	}

	return cleaner
}
