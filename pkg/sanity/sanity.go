/*
Copyright 2017 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sanity

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"
)

var (
	osdAddress        string
	volumeDriver      string
	cloudBackupConfig *CloudBackupConfig
	lock              sync.Mutex
)

// CloudBackupConfig struct for cloud backup configuration
type CloudBackupConfig struct {
	//CloudProvider string `yaml:"providers"`
	// map[string]string is volume.VolumeParams equivalent
	CloudProviders map[string]map[string]string
}

// Test will test the CSI driver at the specified address
func Test(t *testing.T, address, driver, cloudBackupConfigPath string) {
	lock.Lock()
	defer lock.Unlock()

	if len(cloudBackupConfigPath) != 0 {
		cfg, err := CloudProviderConfigParse(cloudBackupConfigPath)
		if err != nil {
			t.Logf("Error in Cloud Backup Config , skipping the tests related to cloud backup and restore")
		}
		cloudBackupConfig = cfg
	}

	osdAddress = address
	volumeDriver = driver

	RegisterFailHandler(Fail)
	RunSpecs(t, "OSD API Test Suite")
}

// CloudProviderConfigParse parses the config file of cloudBackup
func CloudProviderConfigParse(filePath string) (*CloudBackupConfig, error) {

	config := &CloudBackupConfig{}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read the Cloud backup configuration file (%s): %s", filePath, err.Error())
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("Unable to parse Cloud backup configuration: %s", err.Error())
	}
	return config, nil

}
