package main

import (
	"fmt"
	"net/url"
	"os"
	"runtime"

	"go.pedge.io/dlog"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/reexec"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/server"
	osdcli "github.com/libopenstorage/openstorage/cli"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/graph/drivers"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/consul"
	"github.com/portworx/kvdb/etcd"
	"github.com/portworx/kvdb/mem"
)

func start(c *cli.Context) {

	if !osdcli.DaemonMode(c) {
		cli.ShowAppHelp(c)
		return
	}

	datastores := []string{mem.Name, etcd.Name, consul.Name}

	// We are in daemon mode.
	file := c.String("file")
	if file == "" {
		dlog.Warnln("OSD configuration file not specified.  Visit openstorage.org for an example.")
		return
	}

	cfg, err := config.Parse(file)
	if err != nil {
		dlog.Errorln(err)
		return
	}
	kvdbURL := c.String("kvdb")
	u, err := url.Parse(kvdbURL)
	scheme := u.Scheme
	u.Scheme = "http"

	kv, err := kvdb.New(scheme, "openstorage", []string{u.String()}, nil)
	if err != nil {
		dlog.Warnf("Failed to initialize KVDB: %v (%v)", scheme, err)
		dlog.Warnf("Supported datastores: %v", datastores)
		return
	}
	err = kvdb.SetInstance(kv)
	if err != nil {
		dlog.Warnf("Failed to initialize KVDB: %v", err)
		return
	}

	// Start the cluster state machine, if enabled.
	clusterInit := false
	if cfg.Osd.ClusterConfig.NodeId != "" && cfg.Osd.ClusterConfig.ClusterId != "" {
		dlog.Infof("OSD enabling cluster mode.")

		if err := cluster.Init(cfg.Osd.ClusterConfig); err != nil {
			dlog.Errorln("Unable to init cluster server: %v", err)
			return
		}
		clusterInit = true

		if err := server.StartClusterAPI(config.ClusterAPIBase); err != nil {
			dlog.Warnf("Unable to start cluster API server: %v", err)
			return
		}
	}

	isDefaultSet := false
	// Start the volume drivers.
	for d, v := range cfg.Osd.Drivers {
		dlog.Infof("Starting volume driver: %v", d)
		if _, err := volume.New(d, v); err != nil {
			dlog.Warnf("Unable to start volume driver: %v, %v", d, err)
			return
		}

		if err := server.StartPluginAPI(d, config.DriverAPIBase, config.PluginAPIBase); err != nil {
			dlog.Warnf("Unable to start volume plugin: %v", err)
			return
		}
		if d != "" && cfg.Osd.ClusterConfig.DefaultDriver == d {
			isDefaultSet = true
		}
	}

	if cfg.Osd.ClusterConfig.DefaultDriver != "" && !isDefaultSet {
		dlog.Warnf("Invalid OSD config file: Default Driver specified but driver not initialized")
		return
	}

	if err := server.StartFlexVolumeAPI(config.FlexVolumePort, cfg.Osd.ClusterConfig.DefaultDriver); err != nil {
		dlog.Warnf("Unable to start flexvolume API: %v", err)
		return
	}

	// Start the graph drivers.
	for d, _ := range cfg.Osd.GraphDrivers {
		dlog.Infof("Starting graph driver: %v", d)
		if err := server.StartGraphAPI(d, config.PluginAPIBase); err != nil {
			dlog.Warnf("Unable to start graph plugin: %v", err)
			return
		}
	}

	if clusterInit {
		cm, err := cluster.Inst()
		if err != nil {
			dlog.Warnf("Unable to find cluster instance: %v", err)
			return
		}
		if err := cm.Start(); err != nil {
			dlog.Warnf("Unable to start cluster manager: %v", err)
			return
		}
	}

	// Daemon does not exit.
	select {}
}

func showVersion(c *cli.Context) {
	fmt.Println("OSD Version:", config.Version)
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
	app.Version = config.Version
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
			Name:        "cluster",
			Aliases:     []string{"c"},
			Usage:       "Manage cluster",
			Subcommands: osdcli.ClusterCommands(),
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
		// TODO(pedge): was an and, but we have no drivers that have two types
		if v.DriverType == api.DriverType_DRIVER_TYPE_BLOCK {
			bCmds := osdcli.BlockVolumeCommands(v.Name)
			cmds := append(bCmds)
			c := cli.Command{
				Name:        v.Name,
				Usage:       fmt.Sprintf("Manage %s storage", v.Name),
				Subcommands: cmds,
			}
			app.Commands = append(app.Commands, c)
			// TODO(pedge): was an and, but we have no drivers that have two types
		} else if v.DriverType == api.DriverType_DRIVER_TYPE_FILE {
			fCmds := osdcli.FileVolumeCommands(v.Name)
			cmds := append(fCmds)
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
		// TODO(pedge): was an and, but we have no drivers that have two types
		if v.DriverType == api.DriverType_DRIVER_TYPE_GRAPH {
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
