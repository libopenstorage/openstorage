package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/client"
	"github.com/libopenstorage/openstorage/volume"
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

var (
	clnt      *client.Client
	volDriver volume.VolumeDriver
)

func initVolumeClient(volumeUrl *url.URL) {
	volDriver = clnt.VolumeDriver()
}

func processLabels(s string) (api.Labels, error) {
	m := make(map[string]string)
	labels := strings.Split(s, ",")
	for _, v := range labels {
		label := strings.Split(v, "=")
		if len(label) != 2 {
			return nil, fmt.Errorf("Malformed label: %s", v)
		}
		if _, ok := m[label[0]]; ok {
			return nil, fmt.Errorf("Duplicate label: %s", v)
		}
		m[label[0]] = label[1]
	}
	return m, nil
}

func volumeOptions(c *cli.Context) {
	var volumeUrl *url.URL
	var err error

	if err != nil {
		b, _ := json.MarshalIndent(err, "", " ")
		fmt.Printf("%+v\n", string(b))
		os.Exit(1)
	}
	initVolumeClient(volumeUrl)
}

func volumeCreate(c *cli.Context) {
	var err error
	var labels api.Labels
	var locator api.VolumeLocator
	var id api.VolumeID
	fn := "create"

	if len(c.Args()) != 1 {
		missingParameter(c, fn, "name", "Invalid number of arguments")
		return
	}

	volumeOptions(c)
	if l := c.String("label"); l != "" {
		if labels, err = processLabels(l); err != nil {
			cmdError(c, fn, err)
			return
		}
	}
	locator = api.VolumeLocator{
		Name:         c.Args()[0],
		VolumeLabels: labels,
	}
	spec := &api.VolumeSpec{
		Size:             uint64(VolumeSzUnits(c.Int("s")) * MB),
		Format:           api.Filesystem(c.String("fs")),
		BlockSize:        c.Int("b") * 1024,
		HALevel:          c.Int("r"),
		Cos:              api.VolumeCos(c.Int("cos")),
		SnapShotInterval: c.Int("si"),
	}
	if id, err = volDriver.Create(locator, nil, spec); err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{Uuid: []string{string(id)}})
}

func volumeFormat(c *cli.Context) {
	volumeOptions(c)
	fn := "format"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := c.Args()[0]

	err := volDriver.Format(api.VolumeID(volumeID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{Uuid: []string{c.Args()[0]}})
}

func volumeAttach(c *cli.Context) {
	fn := "attach"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeOptions(c)
	volumeID := c.Args()[0]

	devicePath, err := volDriver.Attach(api.VolumeID(volumeID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{Result: devicePath})
}

func volumeDetach(c *cli.Context) {
	fn := "detach"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := c.Args()[0]
	volumeOptions(c)
	err := volDriver.Detach(api.VolumeID(volumeID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{Uuid: []string{c.Args()[0]}})
}

func volumeInspect(c *cli.Context) {

	volumeOptions(c)
	fn := "inspect"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}

	d := make([]api.VolumeID, len(c.Args()))
	for i, v := range c.Args() {
		d[i] = api.VolumeID(v)
	}

	volumes, err := volDriver.Inspect(d)
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	cmdOutput(c, volumes)
}

func volumeEnumerate(c *cli.Context) {
	var locator api.VolumeLocator
	var err error

	fn := "enumerate"
	locator.Name = c.String("name")
	if l := c.String("label"); l != "" {
		locator.VolumeLabels, err = processLabels(l)
		if err != nil {
			cmdError(c, fn, err)
			return
		}
	}

	volumeOptions(c)
	volumes, err := volDriver.Enumerate(locator, nil)
	if err != nil {
		cmdError(c, fn, err)
		return
	}
	if volumes == nil {
		cmdError(c, fn, err)
		return
	}
	cmdOutput(c, volumes)
}

func volumeDelete(c *cli.Context) {
	fn := "delete"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := c.Args()[0]
	volumeOptions(c)
	err := volDriver.Delete(api.VolumeID(volumeID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{Uuid: []string{c.Args()[0]}})
}

func snapCreate(c *cli.Context) {
}

func snapInspect(c *cli.Context) {

	volumeOptions(c)
	fn := "inspect"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "snapID", "Invalid number of arguments")
		return
	}
	d := make([]api.SnapID, len(c.Args()))
	for i, v := range c.Args() {
		d[i] = api.SnapID(v)
	}

	snaps, err := volDriver.SnapInspect(d)
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	cmdOutput(c, snaps)
}

func snapEnumerate(c *cli.Context) {
	var locator api.VolumeLocator
	var err error

	fn := "enumerate"
	locator.Name = c.String("name")
	if l := c.String("label"); l != "" {
		locator.VolumeLabels, err = processLabels(l)
		if err != nil {
			cmdError(c, fn, err)
			return
		}
	}

	volumeOptions(c)
	snaps, err := volDriver.Enumerate(locator, nil)
	if err != nil {
		cmdError(c, fn, err)
		return
	}
	if snaps == nil {
		cmdError(c, fn, err)
		return
	}
	cmdOutput(c, snaps)
}

func snapDelete(c *cli.Context) {
	fn := "delete"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "snapID", "Invalid number of arguments")
		return
	}
	volumeOptions(c)
	snapID := c.Args()[0]
	err := volDriver.SnapDelete(api.SnapID(snapID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{Uuid: []string{c.Args()[0]}})
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
		{
			Name:    "snap",
			Aliases: []string{"sc"},
			Usage:   "create snap",
			Action:  snapCreate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "label,l",
					Usage: "Comma separated name=value pairs, e.g name=sqlvolume,type=production",
				},
			},
		},
		{
			Name:    "snapInspect",
			Aliases: []string{"si"},
			Usage:   "Inspect snap",
			Action:  snapInspect,
		},
		{
			Name:    "snapEnumerate",
			Aliases: []string{"se"},
			Usage:   "Enumerate snap",
			Action:  snapEnumerate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "snap name used during creation",
				},
				cli.StringFlag{
					Name:  "label,l",
					Usage: "Comma separated name=value pairs, e.g name=sqlvolume,type=production",
				},
			},
		},
		{
			Name:    "snapDelete",
			Aliases: []string{"si"},
			Usage:   "Delete snap",
			Action:  snapDelete,
		},
	}
	return commands
}
