package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"

	"github.com/libopenstorage/openstorage/apiserver"
	osdcli "github.com/libopenstorage/openstorage/cli"
	"github.com/libopenstorage/openstorage/drivers/aws"
	"github.com/libopenstorage/openstorage/drivers/nfs"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	version       = "0.3"
	driverApiBase = "/var/lib/osd/driver/"
	pluginApiBase = "/var/lib/osd/plugin/"
)

var (
	drivers = []string{aws.Name, nfs.Name}
)

type osd struct {
	// Drivers map[string][]volume.DriverParams
	Drivers map[string]volume.DriverParams
}

type Config struct {
	Osd osd
}

func start(c *cli.Context) {
	cfg := Config{}

	err := os.MkdirAll(driverApiBase, 0744)
	if err != nil {
		fmt.Println("Unable to create UNIX socket path: ", err)
		os.Exit(-1)
	}

	err = os.MkdirAll(pluginApiBase, 0744)
	if err != nil {
		fmt.Println("Unable to create UNIX socket path: ", err)
		os.Exit(-1)
	}

	if !osdcli.DaemonMode(c) {
		cli.ShowAppHelp(c)
	}

	// We are in daemon mode.  Bring up the volume drivers.

	file := c.String("file")
	if file == "" {
		fmt.Println("OSD configuration file not specified.  Visit openstorage.org for an example.")
		os.Exit(-1)
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Unable to read the OSD configuration file.")
		os.Exit(-1)
	}

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		fmt.Println("Unable to parse OSD configuration: ", err)
		os.Exit(-1)
	}

	// Start the volume drivers.
	for d, v := range cfg.Osd.Drivers {
		// 1. Create a new volume driver of the requested type.
		// 2. Start the driver API server for this volume.
		// 3. Start the plugin API server for this volume.

		fmt.Println("Starting volume driver: ", d)
		_, err := volume.New(d, v)
		if err != nil {
			fmt.Println("Unable to start volume driver: ", err)
			os.Exit(-1)
		}

		// Create a unique path for a UNIX socket that the driver will listen on.
		out, err := exec.Command("uuidgen").Output()
		if err != nil {
			fmt.Println("Unable to create UUID: ", err)
			os.Exit(-1)
		}
		uuid := string(out)
		uuid = strings.TrimSuffix(uuid, "\n")

		sock := driverApiBase + uuid
		err = apiserver.StartDriverApi(d, 0, sock)
		if err != nil {
			fmt.Println("Unable to start volume driver: ", err)
			os.Exit(-1)
		}

		sock = pluginApiBase + uuid
		err = apiserver.StartPluginApi(d, sock)
		if err != nil {
			fmt.Println("Unable to start volume plugin: ", err)
			os.Exit(-1)
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "osd"
	app.Usage = "Open Storage CLI"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "json,j",
			Usage: "output in json",
		},
		cli.BoolFlag{
			Name:  osdcli.DaemonAlias,
			Usage: "Start OSD in daemon mode",
		},
		cli.StringSliceFlag{
			Name:  "provider, p",
			Usage: "provider name and options: name=btrfs,root_vol=/var/openstorage/btrfs",
			Value: new(cli.StringSlice),
		},
		cli.StringFlag{
			Name:  "file,f",
			Usage: "file to read the OSD configuration from.",
			Value: "",
		},
	}
	app.Action = start
	app.Commands = []cli.Command{
		{
			Name:        "volume",
			Aliases:     []string{"v"},
			Usage:       "Manage volumes",
			Subcommands: osdcli.VolumeCommands(),
		},
		{
			Name:        "driver",
			Aliases:     []string{"d"},
			Usage:       "Manage drivers",
			Subcommands: osdcli.DriverCommands(),
		},
	}
	app.Run(os.Args)
}

func init() {
}
