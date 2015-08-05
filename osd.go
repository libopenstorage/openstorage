package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/codegangsta/cli"

	"github.com/libopenstorage/kvdb"
	"github.com/libopenstorage/kvdb/etcd"
	"github.com/libopenstorage/kvdb/mem"
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

	datastores := []string{mem.Name, etcd.Name}

	// We are in daemon mode.
	file := c.String("file")
	if file == "" {
		fmt.Println("OSD configuration file not specified.  Visit openstorage.org for an example.")
		return
	}
	cfg, err := config.Parse(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	kvdb_url := c.String("kvdb")
	u, err := url.Parse(kvdb_url)
	scheme := u.Scheme
	u.Scheme = "http"

	kv, err := kvdb.New(scheme, "openstorage", []string{u.String()}, nil)
	if err != nil {
		fmt.Println("Failed to initialize KVDB: ", u.Scheme, err)
		fmt.Println("Supported datastores: ", datastores)
		return
	}
	err = kvdb.SetInstance(kv)
	if err != nil {
		fmt.Println("Failed to initialize KVDB: ", err)
		return
	}

	// Start the volume drivers.
	for d, v := range cfg.Osd.Drivers {

		fmt.Println("Starting volume driver: ", d)
		_, err := volume.New(d, v)
		if err != nil {
			fmt.Println("Unable to start volume driver: ", d, err)
			return
		}

		err = apiserver.StartDriverAPI(d, 0, config.DriverAPIBase)
		if err != nil {
			fmt.Println("Unable to start volume driver: ", err)
			return
		}

		err = apiserver.StartPluginAPI(d, config.PluginAPIBase)
		if err != nil {
			fmt.Println("Unable to start volume plugin: ", err)
			return
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
			Name:  "kvdb,k",
			Usage: "uri to kvdb e.g. kv-mem://localhost, etcd://localhost:4001",
			Value: "kv-mem://localhost",
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
		if v.driverType == volume.Block {
			c := cli.Command{
				Name:        v.name,
				Aliases:     []string{"v"},
				Usage:       fmt.Sprintf("Manage %s volumes", v.name),
				Subcommands: osdcli.BlockVolumeCommands(v.name),
			}
			app.Commands = append(app.Commands, c)
		} else if v.driverType == volume.File {
			c := cli.Command{
				Name:        v.name,
				Aliases:     []string{"v"},
				Usage:       fmt.Sprintf("Manage %s volumes", v.name),
				Subcommands: osdcli.FileVolumeCommands(v.name),
			}
			app.Commands = append(app.Commands, c)
		} else {
			fmt.Println("Unable to start volume plugin: ", errors.New("Unknown driver type."))
			return
		}
	}
	app.Run(os.Args)
}
