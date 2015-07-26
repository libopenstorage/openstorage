package main

import (
	"github.com/codegangsta/cli"

	osd "github.com/libopenstorage/openstorage/cli"
	"github.com/libopenstorage/openstorage/drivers/ebs"
)

const (
	version = "0.3"
)

var (
	providers = []string{ebs.Name}
)

func start(c *cli.Context) {
	if !osd.DaemonMode(c) {
		cli.ShowAppHelp(c)
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
			Name:  osd.DaemonAlias,
			Usage: "Start OSD in daemon mode",
		},
		cli.StringSliceFlag{
			Name:  "provider, p",
			Usage: "provider name and options: name=btrfs,root_vol=/var/openstorage/btrfs",
		},
	}
	app.Action = start
	app.Commands = []cli.Command{
		{
			Name:        "volume",
			Aliases:     []string{"v"},
			Usage:       "Manage volumes",
			Subcommands: osd.VolumeCommands(),
		},
		{
			Name:        "provider",
			Aliases:     []string{"p"},
			Usage:       "Manage providers",
			Subcommands: osd.ProviderCommands(),
		},
	}
}

func init() {
}
