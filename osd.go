package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"

	"github.com/libopenstorage/openstorage/apiserver"
	osdcli "github.com/libopenstorage/openstorage/cli"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	version = "0.3"
)

func start(c *cli.Context) {

	if !osdcli.DaemonMode(c) {
		cli.ShowAppHelp(c)
	}

	// We are in daemon mode.  Bring up the volume drivers.
	file := c.String("file")
	if file == "" {
		fmt.Println("OSD configuration file not specified.  Visit openstorage.org for an example.")
		os.Exit(-1)
	}
	cfg, err := config.Parse(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// Start the volume drivers.
	for d, v := range cfg.Osd.Drivers {

		fmt.Println("Starting volume driver: ", d)
		_, err := volume.New(d, v)
		if err != nil {
			fmt.Println("Unable to start volume driver: ", err)
			os.Exit(-1)
		}

		err = apiserver.StartDriverAPI(d, 0, config.DriverAPIBase)
		if err != nil {
			fmt.Println("Unable to start volume driver: ", err)
			os.Exit(-1)
		}

		err = apiserver.StartPluginAPI(d, config.PluginAPIBase)
		if err != nil {
			fmt.Println("Unable to start volume plugin: ", err)
			os.Exit(-1)
		}
	}

	// Daemon does not exit.
	select {}
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
			Name:  "driver",
			Usage: "driver name and options: name=btrfs,root_vol=/var/openstorage/btrfs",
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
			Name:        "driver",
			Aliases:     []string{"d"},
			Usage:       "Manage drivers",
			Subcommands: osdcli.DriverCommands(),
		},
	}

	for _, v := range drivers {
		c := cli.Command{
			Name:        v,
			Aliases:     []string{"v"},
			Usage:       fmt.Sprintf("Manage %s volumes", v),
			Subcommands: osdcli.VolumeCommands(v),
		}
		app.Commands = append(app.Commands, c)
	}
	app.Run(os.Args)
}
