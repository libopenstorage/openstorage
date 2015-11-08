package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/volume"
)

type osd struct {
	ClusterConfig cluster.Config `yaml:"cluster"`
	Drivers       map[string]volume.DriverParams
	GraphDrivers  map[string]volume.DriverParams
}

type Config struct {
	Osd osd
}

const (
	PluginAPIBase      = "/run/docker/plugins/"
	DriverAPIBase      = "/var/lib/osd/driver/"
	GraphDriverAPIBase = "/var/lib/osd/graphdriver/"
	UrlKey             = "url"
	VersionKey         = "version"
	MountBase          = "/var/lib/osd/mounts/"
	DataDir            = ".data"
	Version            = "v1"
)

var (
	cfg Config
)

func Parse(file string) (*Config, error) {

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to read the OSD configuration file (%s): %s", file, err.Error())
	}

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		fmt.Println("Unable to parse OSD configuration: ", err)
		return nil, fmt.Errorf("Unable to parse OSD configuration: %s", err.Error())
	}
	return &cfg, nil
}
func init() {
	os.MkdirAll(MountBase, 0755)
	os.MkdirAll(GraphDriverAPIBase, 0755)
}
