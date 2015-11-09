package main

import (
	"fmt"
	"net/url"
	"os"
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/reexec"
	"github.com/fsouza/go-dockerclient"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/etcd"
	"github.com/portworx/kvdb/mem"

	"github.com/libopenstorage/openstorage/api"
	apiserver "github.com/libopenstorage/openstorage/api/server"
	osdcli "github.com/libopenstorage/openstorage/cli"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/drivers"
	_ "github.com/libopenstorage/openstorage/graph/layer0"
	_ "github.com/libopenstorage/openstorage/graph/proxy"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	version = "0.3"
)

func start(c *cli.Context) {
	var cm *cluster.ClusterManager

	if !osdcli.DaemonMode(c) {
		cli.ShowAppHelp(c)
		return
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
	kvdbURL := c.String("kvdb")
	u, err := url.Parse(kvdbURL)
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

	// Start the cluster state machine, if enabled.
	if cfg.Osd.ClusterConfig.NodeId != "" && cfg.Osd.ClusterConfig.ClusterId != "" {
		dockerClient, err := docker.NewClientFromEnv()
		if err != nil {
			fmt.Println("Failed to initialize docker client: ", err)
			return
		}
		cm = cluster.New(cfg.Osd.ClusterConfig, kv, dockerClient)
	}

	// Start the volume drivers.
	for d, v := range cfg.Osd.Drivers {
		fmt.Println("Starting volume driver: ", d)
		_, err := volume.New(d, v)
		if err != nil {
			fmt.Println("Unable to start volume driver: ", d, err)
			return
		}
		err = apiserver.StartServerAPI(d, 0, config.DriverAPIBase)
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

	// Start the graph drivers.
	for d, _ := range cfg.Osd.GraphDrivers {
		fmt.Println("Starting graph driver: ", d)
		err = apiserver.StartGraphAPI(d, 0, config.PluginAPIBase)
		if err != nil {
			fmt.Println("Unable to start graph plugin: ", err)
			return
		}
	}

	if cm != nil {
		err = cm.Start()
		if err != nil {
			fmt.Println("Unable to start cluster manager: ", err)
			return
		}
	}

	// Daemon does not exit.
	select {}
}

func showVersion(c *cli.Context) {
	fmt.Println("OSD Version:", version)
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("OS:", runtime.GOOS)
	fmt.Println("Arch:", runtime.GOARCH)
}

func main() {
	if reexec.Init() {
		return
	}
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
		{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "Display version",
			Action:  showVersion,
		},
	}

	// Start all drivers.
	for _, v := range drivers.AllDrivers {
		if v.DriverType&api.Block == api.Block {
			bCmds := osdcli.BlockVolumeCommands(v.Name)
			clstrCmds := osdcli.ClusterCommands(v.Name)
			cmds := append(bCmds, clstrCmds...)
			c := cli.Command{
				Name:        v.Name,
				Usage:       fmt.Sprintf("Manage %s storage", v.Name),
				Subcommands: cmds,
			}
			app.Commands = append(app.Commands, c)
		} else if v.DriverType&api.File == api.File {
			fCmds := osdcli.FileVolumeCommands(v.Name)
			clstrCmds := osdcli.ClusterCommands(v.Name)
			cmds := append(fCmds, clstrCmds...)
			c := cli.Command{
				Name:        v.Name,
				Usage:       fmt.Sprintf("Manage %s volumes", v.Name),
				Subcommands: cmds,
			}
			app.Commands = append(app.Commands, c)
		}

		if v.DriverType&api.Graph == api.Graph {
			// TODO - register this as a graph driver with Docker.
		}
	}

	app.Run(os.Args)
}
