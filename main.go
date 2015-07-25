package main

import (
	"github.com/codegangsta/cli"
	"github.com/openstorage/drivers/ec2driver"
	"github.com/openstorage/volume"
	"os"
)

const (
	version = "0.3"
)

func daemonStart(c *cli.Context) {
	cmd := "start"
	if len(c.Args()) != 1 {
		missingParameter(c, cmd, "path", "path to a UNIX domain socket for the REST server")
		return
	}

	err := ec2driver.Init()
	if err != nil {
		cmdError(c, cmd, err)

	}

	v := volume.NewVolumePlugin()
	v.Listen(os.Args[0])
}

func main() {
	app := cli.NewApp()
	app.Name = "px"
	app.Usage = "pwx cli"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "json,j",
			Usage: "output in json",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "stary",
			Action:  daemonStart,
		},
	}
}
