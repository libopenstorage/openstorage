package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"

	osdcli "github.com/libopenstorage/openstorage/cli"
	"github.com/libopenstorage/openstorage/drivers/aws"
	"github.com/libopenstorage/openstorage/drivers/nfs"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	version = "0.3"
)

var (
	providers = []string{aws.Name, nfs.Name}
)

type osd struct {
	Providers map[string]volume.DriverParams
}

type Config struct {
	Osd osd
}

func start(c *cli.Context) {
	cfg := Config{}

	file := c.String("file")
	if file != "" {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(b, &cfg)
		if err != nil {
			panic(err)
		}

	}

	fmt.Printf("%+v\n", cfg)

	if !osdcli.DaemonMode(c) {
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
			Name:        "provider",
			Aliases:     []string{"p"},
			Usage:       "Manage providers",
			Subcommands: osdcli.ProviderCommands(),
		},
	}
	app.Run(os.Args)
}

func init() {
}
