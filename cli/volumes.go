package cli

import (
	"net/url"

	"github.com/codegangsta/cli"
)

type VolumeSzUnits uint64

const (
	_                = iota
	KB VolumeSzUnits = 1 << (10 * iota)
	MB
	GB
	TB
	PB
)

func initVolumeClient(volumeUrl *url.URL) {
}

func volumeOptions(c *cli.Context) {
}

func volumeCreate(c *cli.Context) {
}

func volumeEnumerate(c *cli.Context) {
}

func volumeInspect(c *cli.Context) {
}

func volumeFormat(c *cli.Context) {
}

func volumeAttach(c *cli.Context) {
}

func volumeDetach(c *cli.Context) {
}

func volumeDelete(c *cli.Context) {
}

func VolumeCommands() []cli.Command {

	commands := []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create a new volume",
			Action:  volumeCreate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "label,l",
					Usage: "Comma separated name=value pairs, e.g name=sqlvolume,type=production",
					Value: "",
				},
				cli.IntFlag{
					Name:  "size,s",
					Usage: "specify size in MB",
					Value: 1000,
				},
				cli.StringFlag{
					Name:  "fs",
					Usage: "filesystem to be laid out: none|xfs|ext4",
					Value: "ext4",
				},
				cli.IntFlag{
					Name:  "block_size,b",
					Usage: "block size in Kbytes",
					Value: 32,
				},
				cli.IntFlag{
					Name:  "repl,r",
					Usage: "replication factor [1..2]",
					Value: 1,
				},
				cli.IntFlag{
					Name:  "cos",
					Usage: "Class of Service [1..9]",
					Value: 1,
				},
				cli.IntFlag{
					Name:  "snap_interval,si",
					Usage: "snapshot interval in minutes, 0 disables snaps",
					Value: 0,
				},
			},
		},
		{
			Name:    "format",
			Aliases: []string{"f"},
			Usage:   "Format volume to spec in create",
			Action:  volumeFormat,
		},
		{
			Name:    "attach",
			Aliases: []string{"a"},
			Usage:   "Attach volume to specified path",
			Action:  volumeAttach,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path,p",
					Usage: "Path on local filesystem",
				},
			},
		},
		{
			Name:    "detach",
			Aliases: []string{"d"},
			Usage:   "Detach specified volume",
			Action:  volumeDetach,
		},
		{
			Name:    "delete",
			Aliases: []string{"rm"},
			Usage:   "Detach specified volume",
			Action:  volumeDelete,
		},
		{
			Name:    "enumerate",
			Aliases: []string{"e"},
			Usage:   "Enumerate volumes",
			Action:  volumeEnumerate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "volume name used during creation if any",
				},
				cli.StringFlag{
					Name:  "label,l",
					Usage: "Comma separated name=value pairs, e.g name=sqlvolume,type=production",
				},
			},
		},
		{
			Name:    "inspect",
			Aliases: []string{"i"},
			Usage:   "Inspect volume",
			Action:  volumeInspect,
		},
	}
	return commands
}
