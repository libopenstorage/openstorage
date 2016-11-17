package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/coreos/etcd/version"
	"github.com/portworx/kvdb"
)

const (
	DefaultRetryCount             = 60
	DefaultIntervalBetweenRetries = time.Millisecond * 500
	Bootstrap                     = "kvdb/bootstrap"
	// the maximum amount of time a dial will wait for a connection to setup.
	// 30s is long enough for most of the network conditions.
	DefaultDialTimeout         = 30 * time.Second
	DefaultLockTTL             = 8
	DefaultLockRefreshDuration = 2 * time.Second
)

// Version returns the version of the provided etcd server
func Version(url string) (string, error) {
	response, err := http.Get(url + "/version")
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	contents, _ := ioutil.ReadAll(response.Body)

	var version version.Versions
	err = json.Unmarshal(contents, &version)
	if err != nil {
		return "", err
	}
	if version.Server[0] == '2' || version.Server[0] == '1' {
		return kvdb.EtcdBaseVersion, nil
	} else if version.Server[0] == '3' {
		return kvdb.EtcdVersion3, nil
	} else {
		return "", fmt.Errorf("Unsupported etcd version: %v", version.Server)
	}
}
