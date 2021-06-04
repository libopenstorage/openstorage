package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Version                   = "v1"
	GraphDriverAPIBase        = "/var/lib/osd/graphdriver/"
	UrlKey                    = "url"
	MgmtPortKey               = "mgmtPort"
	PluginPortKey             = "pluginPort"
	VersionKey                = "version"
	DataDir                   = ".data"
	FlexVolumePort     uint16 = 2345
)

func init() {
	os.MkdirAll(volume.MountBase, 0755)
	os.MkdirAll(GraphDriverAPIBase, 0755)
}

// swagger:model
type ClusterConfig struct {
	ClusterId            string
	ClusterUuid          string
	NodeId               string
	SchedulerNodeName    string
	MgtIface             string
	DataIface            string
	DefaultDriver        string
	MgmtIp               string
	DataIp               string
	ManagementURL        string
	FluentDHost          string
	SystemSharedSecret   string
	AllowSecurityRemoval bool
	HWType               api.HardwareType
	// QuorumTimeoutInSeconds configures time after which an
	// out of quorum node will restart
	QuorumTimeoutInSeconds int
	// SnapLockTryDurationInMinutes is the time for which
	// the cluster manager will try acquiring a lock for cluster snapshot
	SnapLockTryDurationInMinutes int
}

// swagger:model
type Config struct {
	Osd struct {
		ClusterConfig ClusterConfig `yaml:"cluster"`
		// map[string]string is volume.VolumeParams equivalent
		Drivers map[string]map[string]string
		// map[string]string is volume.VolumeParams equivalent
		GraphDrivers map[string]map[string]string
	}
}

func Parse(filePath string) (*Config, error) {
	config := &Config{}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read the OSD configuration file (%s): %s", filePath, err.Error())
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("Unable to parse OSD configuration: %s", err.Error())
	}
	return config, nil
}
