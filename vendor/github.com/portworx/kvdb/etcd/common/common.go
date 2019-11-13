package common

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/coreos/etcd/pkg/transport"
	"github.com/coreos/etcd/version"
	"github.com/portworx/kvdb"
)

const (
	// DefaultRetryCount for etcd operations
	DefaultRetryCount = 60
	// DefaultIntervalBetweenRetries for etcd failed operations
	DefaultIntervalBetweenRetries = time.Millisecond * 500
	// Bootstrap key
	Bootstrap = "kvdb/bootstrap"
	// DefaultDialTimeout in etcd http requests
	// the maximum amount of time a dial will wait for a connection to setup.
	// 30s is long enough for most of the network conditions.
	DefaultDialTimeout = 30 * time.Second
	// DefaultLockTTL is the ttl for an etcd lock
	DefaultLockTTL = 16
	// DefaultLockRefreshDuration is the time interval for refreshing an etcd lock
	DefaultLockRefreshDuration = 2 * time.Second
)

// EtcdCommon defined the common functions between v2 and v3 etcd implementations.
type EtcdCommon interface {
	// GetAuthInfoFromOptions
	GetAuthInfoFromOptions() (transport.TLSInfo, string, string, error)

	// GetRetryCount
	GetRetryCount() int
}

// EtcdLock combines Mutex and channel
type EtcdLock struct {
	Done            chan struct{}
	Unlocked        bool
	Err             error
	Tag             string
	AcquisitionTime time.Time
	sync.Mutex
}

// LockerIDInfo id of locker
type LockerIDInfo struct {
	LockerID string
}

type etcdCommon struct {
	options map[string]string
}

var (
	cmd *exec.Cmd
)

// NewEtcdCommon returns the EtcdCommon interface
func NewEtcdCommon(options map[string]string) EtcdCommon {
	return &etcdCommon{
		options: options,
	}
}

func (ec *etcdCommon) GetRetryCount() int {
	retryCount, ok := ec.options[kvdb.RetryCountKey]
	if !ok {
		return DefaultRetryCount
	}
	retry, err := strconv.ParseInt(retryCount, 10, 0)
	if err != nil {
		// use default value
		return DefaultRetryCount
	}
	return int(retry)
}

func (ec *etcdCommon) GetAuthInfoFromOptions() (transport.TLSInfo, string, string, error) {
	var (
		username       string
		password       string
		caFile         string
		certFile       string
		keyFile        string
		trustedCAFile  string
		clientCertAuth bool
		err            error
	)
	// options provided. Probably auth options
	if ec.options != nil || len(ec.options) > 0 {
		// Check if username provided
		username, _ = ec.options[kvdb.UsernameKey]
		// Check if password provided
		password, _ = ec.options[kvdb.PasswordKey]
		// Check if CA file provided
		caFile, _ = ec.options[kvdb.CAFileKey]
		// Check if certificate file provided
		certFile, _ = ec.options[kvdb.CertFileKey]
		// Check if certificate key is provided
		keyFile, _ = ec.options[kvdb.CertKeyFileKey]
		// Check if trusted ca file is provided
		trustedCAFile, _ = ec.options[kvdb.TrustedCAFileKey]
		// Check if client cert auth is provided
		clientCertAuthStr, ok := ec.options[kvdb.ClientCertAuthKey]
		if !ok {
			clientCertAuth = false
		} else {
			clientCertAuth, err = strconv.ParseBool(clientCertAuthStr)
			if err != nil {
				clientCertAuth = false
			}
		}
	}
	t := transport.TLSInfo{
		CertFile:       certFile,
		KeyFile:        keyFile,
		CAFile:         caFile,
		TrustedCAFile:  trustedCAFile,
		ClientCertAuth: clientCertAuth,
	}

	_ = t.CAFile // required since CAFile is apparently not used and it fails errcheck

	return t, username, password, nil
}

// Version returns the version of the provided etcd server
func Version(uri string, options map[string]string) (string, error) {
	useTLS := false
	tlsConfig := &tls.Config{}
	// Check if CA file provided
	caFile, ok := options[kvdb.CAFileKey]
	if ok && caFile != "" {
		useTLS = true
		// Load CA cert
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return "", err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	// Check if certificate file provided
	certFile, certOk := options[kvdb.CertFileKey]
	// Check if certificate key is provided
	keyFile, keyOk := options[kvdb.CertKeyFileKey]
	if certOk && keyOk && certFile != "" && keyFile != "" {
		useTLS = true
		// Load client cert
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return "", err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	var client *http.Client
	if useTLS {
		tlsConfig.BuildNameToCertificate()
		client = &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	} else {
		tempURL, _ := url.Parse(uri)
		if tempURL.Scheme == "https" {
			transport := &http.Transport{TLSClientConfig: &tls.Config{}}
			client = &http.Client{Transport: transport}
		} else {
			client = &http.Client{}
		}
	}

	// Do GET something
	resp, err := client.Get(uri + "/version")
	if err != nil {
		return "", fmt.Errorf("error in obtaining etcd version: %v", err)
	}
	defer resp.Body.Close()

	// Dump response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error in obtaining etcd version: %v", err)
	}

	var ver version.Versions
	err = json.Unmarshal(data, &ver)
	if err != nil {
		// Probably a version less than 2.3. Default to using v2 apis
		return kvdb.EtcdBaseVersion, nil
	}
	if ver.Server == "" {
		// This should never happen in an ideal scenario unless
		// etcd messes up. To avoid a crash further in this code
		// we return an error
		return "", fmt.Errorf("unable to determine etcd version (empty response from etcd)")
	}
	if ver.Server[0] == '2' || ver.Server[0] == '1' {
		return kvdb.EtcdBaseVersion, nil
	} else if ver.Server[0] == '3' {
		return kvdb.EtcdVersion3, nil
	} else {
		return "", fmt.Errorf("unsupported etcd version: %v", ver.Server)
	}
}

// TestStart starts test
func TestStart(removeData bool) error {
	dataDir := "/tmp/etcd"
	if removeData {
		os.RemoveAll(dataDir)
	}
	cmd = exec.Command("etcd", "--advertise-client-urls", "http://127.0.0.1:2379", "--data-dir", dataDir)
	err := cmd.Start()
	time.Sleep(5 * time.Second)
	return err
}

// TestStop stops test
func TestStop() error {
	return cmd.Process.Kill()
}
