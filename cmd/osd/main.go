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
	"github.com/portworx/kvdb/consul"
	"github.com/portworx/kvdb/etcd"
	"github.com/portworx/kvdb/mem"

	"github.com/Sirupsen/logrus"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/server"
	osdcli "github.com/libopenstorage/openstorage/cli"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/graph/drivers"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers"
)

const (
	version = "0.3"
)

func start(c *cli.Context) {
	var cm cluster.Cluster

	if !osdcli.DaemonMode(c) {
		cli.ShowAppHelp(c)
		return
	}

	datastores := []string{mem.Name, etcd.Name, consul.Name}

	// We are in daemon mode.
	file := c.String("file")
	if file == "" {
		logrus.Warn("OSD configuration file not specified.  Visit openstorage.org for an example.")
		return
	}
	cfg, err := config.Parse(file)
	if err != nil {
		logrus.Error(err)
		return
	}
	kvdbURL := c.String("kvdb")
	u, err := url.Parse(kvdbURL)
	scheme := u.Scheme
	u.Scheme = "http"

	kv, err := kvdb.New(scheme, "openstorage", []string{u.String()}, nil)
	if err != nil {
		logrus.Warnf("Failed to initialize KVDB: %v (%v)", scheme, err)
		logrus.Warnf("Supported datastores: %v", datastores)
		return
	}
	err = kvdb.SetInstance(kv)
	if err != nil {
		logrus.Warnf("Failed to initialize KVDB: %v", err)
		return
	}

	// Start the cluster state machine, if enabled.
	if cfg.Osd.ClusterConfig.NodeId != "" && cfg.Osd.ClusterConfig.ClusterId != "" {
		dockerClient, err := docker.NewClientFromEnv()
		if err != nil {
			logrus.Warnf("Failed to initialize docker client: %v", err)
			return
		}
		cm = cluster.New(cfg.Osd.ClusterConfig, kv, dockerClient)
	}

	// Start the volume drivers.
	for d, v := range cfg.Osd.Drivers {
		logrus.Infof("Starting volume driver: %v", d)
		_, err := volume.New(d, v)
		if err != nil {
			logrus.Warnf("Unable to start volume driver: %v, %v", d, err)
			return
		}
		err = server.StartServerAPI(d, 0, config.DriverAPIBase)
		if err != nil {
			logrus.Warnf("Unable to start volume driver: %v", err)
			return
		}
		err = server.StartPluginAPI(d, config.PluginAPIBase)
		if err != nil {
			logrus.Warnf("Unable to start volume plugin: %v", err)
			return
		}
	}

	// Start the graph drivers.
	for d, _ := range cfg.Osd.GraphDrivers {
		logrus.Infof("Starting graph driver: %v", d)
		err = server.StartGraphAPI(d, 0, config.PluginAPIBase)
		if err != nil {
			logrus.Warnf("Unable to start graph plugin: %v", err)
			return
		}
	}

	if cm != nil {
		err = cm.Start()
		if err != nil {
			logrus.Warnf("Unable to start cluster manager: %v", err)
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
			Usage: "driver name and options: name=btrfs,home=/var/openstorage/btrfs",
			Value: new(cli.StringSlice),
		},
		cli.StringFlag{
			Name:  "kvdb,k",
			Usage: "uri to kvdb e.g. kv-mem://localhost, etcd-kv://localhost:4001, consul-kv://localhost:8500",
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

	// Register all volume drivers with the CLI.
	for _, v := range volumedrivers.AllDrivers {
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
	}

	// Register all graph drivers with the CLI.
	for _, v := range graphdrivers.AllDrivers {
		if v.DriverType&api.Graph == api.Graph {
			cmds := osdcli.GraphDriverCommands(v.Name)
			c := cli.Command{
				Name:        v.Name,
				Usage:       fmt.Sprintf("Manage %s graph storage", v.Name),
				Subcommands: cmds,
			}
			app.Commands = append(app.Commands, c)
		}
	}

	app.Run(os.Args)
}
