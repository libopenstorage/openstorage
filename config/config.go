package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"go.pedge.io/dlog/logrus"
)

const (
	Version                   = "v1"
	PluginAPIBase             = "/run/docker/plugins/"
	DriverAPIBase             = "/var/lib/osd/driver/"
	GraphDriverAPIBase        = "/var/lib/osd/graphdriver/"
	ClusterAPIBase            = "/var/lib/osd/cluster/"
	UrlKey                    = "url"
	MgmtPortKey               = "mgmtPort"
	PluginPortKey             = "pluginPort"
	VersionKey                = "version"
	MountBase                 = "/var/lib/osd/mounts/"
	VolumeBase                = "/var/lib/osd/"
	DataDir                   = ".data"
	FlexVolumePort     uint16 = 2345
)

func init() {
	os.MkdirAll(MountBase, 0755)
	os.MkdirAll(GraphDriverAPIBase, 0755)
	// TODO(pedge) eventually move to osd main.go when everyone is comfortable with dlog
	dlog_logrus.Register()
}

type ClusterConfig struct {
	ClusterId     string
	NodeId        string
	MgtIface      string
	DataIface     string
	DefaultDriver string
}

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
