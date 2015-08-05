package cli

import (
	"fmt"
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

type VolDriver struct {
	volDriver volume.VolumeDriver
	name      string
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

func (v *VolDriver) volumeOptions(c *cli.Context) {
	clnt, err := client.NewDriverClient(v.name)
	if err != nil {
		fmt.Printf("Failed to initialize client library: %v\n", err)
		os.Exit(1)
	}
	v.volDriver = clnt.VolumeDriver()
}

func (v *VolDriver) volumeCreate(c *cli.Context) {
	var err error
	var labels api.Labels
	var locator api.VolumeLocator
	var id api.VolumeID
	fn := "create"

	if len(c.Args()) != 1 {
		missingParameter(c, fn, "name", "Invalid number of arguments")
		return
	}

	v.volumeOptions(c)
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
		SnapshotInterval: c.Int("si"),
	}
	if id, err = v.volDriver.Create(locator, nil, spec); err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{UUID: []string{string(id)}})
}

func (v *VolDriver) volumeMount(c *cli.Context) {
	v.volumeOptions(c)
	fn := "mount"

	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := c.Args()[0]

	path := c.String("path")
	if path == "" {
		missingParameter(c, fn, "path", "Target mount path")
		return

	}

	err := v.volDriver.Mount(api.VolumeID(volumeID), path)
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{UUID: []string{volumeID}})
}

func (v *VolDriver) volumeUnmount(c *cli.Context) {
	v.volumeOptions(c)
	fn := "unmount"

	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := c.Args()[0]

	path := c.String("path")

	err := v.volDriver.Unmount(api.VolumeID(volumeID), path)
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{UUID: []string{volumeID}})
}

func (v *VolDriver) volumeFormat(c *cli.Context) {
	v.volumeOptions(c)
	fn := "format"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := c.Args()[0]

	err := v.volDriver.Format(api.VolumeID(volumeID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{UUID: []string{volumeID}})
}

func (v *VolDriver) volumeAttach(c *cli.Context) {
	fn := "attach"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	v.volumeOptions(c)
	volumeID := c.Args()[0]

	devicePath, err := v.volDriver.Attach(api.VolumeID(volumeID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{Result: devicePath})
}

func (v *VolDriver) volumeDetach(c *cli.Context) {
	fn := "detach"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := c.Args()[0]
	v.volumeOptions(c)
	err := v.volDriver.Detach(api.VolumeID(volumeID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{UUID: []string{c.Args()[0]}})
}

func (v *VolDriver) volumeInspect(c *cli.Context) {

	v.volumeOptions(c)
	fn := "inspect"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}

	d := make([]api.VolumeID, len(c.Args()))
	for i, v := range c.Args() {
		d[i] = api.VolumeID(v)
	}

	volumes, err := v.volDriver.Inspect(d)
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	cmdOutput(c, volumes)
}

func (v *VolDriver) volumeEnumerate(c *cli.Context) {
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

	v.volumeOptions(c)
	volumes, err := v.volDriver.Enumerate(locator, nil)
	if err != nil {
		cmdError(c, fn, err)
		return
	}
	cmdOutput(c, volumes)
}

func (v *VolDriver) volumeDelete(c *cli.Context) {
	fn := "delete"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := c.Args()[0]
	v.volumeOptions(c)
	err := v.volDriver.Delete(api.VolumeID(volumeID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{UUID: []string{c.Args()[0]}})
}

func (v *VolDriver) snapCreate(c *cli.Context) {
}

func (v *VolDriver) snapInspect(c *cli.Context) {

	v.volumeOptions(c)
	fn := "inspect"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "snapID", "Invalid number of arguments")
		return
	}
	d := make([]api.SnapID, len(c.Args()))
	for i, v := range c.Args() {
		d[i] = api.SnapID(v)
	}

	snaps, err := v.volDriver.SnapInspect(d)
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	cmdOutput(c, snaps)
}

func (v *VolDriver) snapEnumerate(c *cli.Context) {
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

	v.volumeOptions(c)
	snaps, err := v.volDriver.Enumerate(locator, nil)
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

func (v *VolDriver) snapDelete(c *cli.Context) {
	fn := "delete"
	if len(c.Args()) < 1 {
		missingParameter(c, fn, "snapID", "Invalid number of arguments")
		return
	}
	v.volumeOptions(c)
	snapID := c.Args()[0]
	err := v.volDriver.SnapDelete(api.SnapID(snapID))
	if err != nil {
		cmdError(c, fn, err)
		return
	}

	fmtOutput(c, &Format{UUID: []string{c.Args()[0]}})
}

func BlockVolumeCommands(name string) []cli.Command {
	v := &VolDriver{name: name}

	commands := []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create a new volume",
			Action:  v.volumeCreate,
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
			Action:  v.volumeFormat,
		},
		{
			Name:    "attach",
			Aliases: []string{"a"},
			Usage:   "Attach volume to specified path",
			Action:  v.volumeAttach,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path,p",
					Usage: "Path on local filesystem",
				},
			},
		},
		{
			Name:    "mount",
			Aliases: []string{"m"},
			Usage:   "Mount specified volume",
			Action:  v.volumeMount,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path",
					Usage: "destination path at which this volume must be mounted on",
				},
			},
		},
		{
			Name:    "unmount",
			Aliases: []string{"u"},
			Usage:   "Unmount specified volume",
			Action:  v.volumeUnmount,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path",
					Usage: "destination path at which this volume must be mounted on",
				},
			},
		},
		{
			Name:    "detach",
			Aliases: []string{"d"},
			Usage:   "Detach specified volume",
			Action:  v.volumeDetach,
		},
		{
			Name:    "delete",
			Aliases: []string{"rm"},
			Usage:   "Detach specified volume",
			Action:  v.volumeDelete,
		},
		{
			Name:    "enumerate",
			Aliases: []string{"e"},
			Usage:   "Enumerate volumes",
			Action:  v.volumeEnumerate,
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
			Action:  v.volumeInspect,
		},
		{
			Name:    "snap",
			Aliases: []string{"sc"},
			Usage:   "create snap",
			Action:  v.snapCreate,
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
			Action:  v.snapInspect,
		},
		{
			Name:    "snapEnumerate",
			Aliases: []string{"se"},
			Usage:   "Enumerate snap",
			Action:  v.snapEnumerate,
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
			Action:  v.snapDelete,
		},
	}
	return commands
}

func FileVolumeCommands(name string) []cli.Command {
	v := &VolDriver{name: name}

	commands := []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create a new volume",
			Action:  v.volumeCreate,
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
			Name:    "mount",
			Aliases: []string{"m"},
			Usage:   "Mount specified volume",
			Action:  v.volumeMount,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path",
					Usage: "destination path at which this volume must be mounted on",
				},
			},
		},
		{
			Name:    "unmount",
			Aliases: []string{"u"},
			Usage:   "Unmount specified volume",
			Action:  v.volumeUnmount,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path",
					Usage: "destination path at which this volume must be mounted on",
				},
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"rm"},
			Usage:   "Detach specified volume",
			Action:  v.volumeDelete,
		},
		{
			Name:    "enumerate",
			Aliases: []string{"e"},
			Usage:   "Enumerate volumes",
			Action:  v.volumeEnumerate,
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
			Action:  v.volumeInspect,
		},
		{
			Name:    "snap",
			Aliases: []string{"sc"},
			Usage:   "create snap",
			Action:  v.snapCreate,
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
			Action:  v.snapInspect,
		},
		{
			Name:    "snapEnumerate",
			Aliases: []string{"se"},
			Usage:   "Enumerate snap",
			Action:  v.snapEnumerate,
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
			Action:  v.snapDelete,
		},
	}
	return commands
}
