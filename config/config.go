package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/libopenstorage/openstorage/volume"
)

type osd struct {
	PluginAPIBase string
	Drivers       map[string]volume.DriverParams
}

type Config struct {
	Osd osd
}

const (
	DriverAPIBase = "/var/lib/osd/driver/"
	PluginAPIBase = "/usr/share/docker/plugins/"
	Version       = "v1"
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
		return nil, err
	}
	return &cfg, nil
}
