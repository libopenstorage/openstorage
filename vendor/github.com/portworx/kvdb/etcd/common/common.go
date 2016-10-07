package common

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	
	"github.com/coreos/etcd/version"
	"github.com/portworx/kvdb"
)

// Version returns the version of the provided etcd server
func Version(url string) (string, error) {
	response, err := http.Get(url+"/version")
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
