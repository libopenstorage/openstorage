package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/volume"
)

// VolumeSzUnits number representing size units.
type VolumeSzUnits uint64

const (
	_ = iota
	// KiB 1024 bytes
	KiB VolumeSzUnits = 1 << (10 * iota)
	// MiB 1024 KiB
	MiB
	// GiB 1024 MiB
	GiB
	// TiB 1024 GiB
	TiB
	// PiB 1024 PiB
	PiB
)

type volDriver struct {
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

func (v *volDriver) volumeOptions(context *cli.Context) {
	clnt, err := client.NewDriverClient(v.name)
	if err != nil {
		fmt.Printf("Failed to initialize client library: %v\n", err)
		os.Exit(1)
	}
	v.volDriver = clnt.VolumeDriver()
}

func (v *volDriver) volumeCreate(context *cli.Context) {
	var err error
	var labels api.Labels
	var locator api.VolumeLocator
	var id api.VolumeID
	fn := "create"

	if len(context.Args()) != 1 {
		missingParameter(context, fn, "name", "Invalid number of arguments")
		return
	}

	v.volumeOptions(context)
	if l := context.String("label"); l != "" {
		if labels, err = processLabels(l); err != nil {
			cmdError(context, fn, err)
			return
		}
	}
	locator = api.VolumeLocator{
		Name:         context.Args()[0],
		VolumeLabels: labels,
	}
	spec := &api.VolumeSpec{
		Size:             uint64(VolumeSzUnits(context.Int("s")) * MiB),
		Format:           api.Filesystem(context.String("fs")),
		BlockSize:        context.Int("b") * 1024,
		HALevel:          context.Int("r"),
		Cos:              api.VolumeCos(context.Int("cos")),
		SnapshotInterval: context.Int("si"),
	}
	source := &api.Source{
		Seed: context.String("seed"),
	}
	if id, err = v.volDriver.Create(locator, source, spec); err != nil {
		cmdError(context, fn, err)
		return
	}

	fmtOutput(context, &Format{UUID: []string{string(id)}})
}

func (v *volDriver) volumeMount(context *cli.Context) {
	v.volumeOptions(context)
	fn := "mount"

	if len(context.Args()) < 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := context.Args()[0]

	path := context.String("path")
	if path == "" {
		missingParameter(context, fn, "path", "Target mount path")
		return
	}

	err := v.volDriver.Mount(api.VolumeID(volumeID), path)
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	fmtOutput(context, &Format{UUID: []string{volumeID}})
}

func (v *volDriver) volumeUnmount(context *cli.Context) {
	v.volumeOptions(context)
	fn := "unmount"

	if len(context.Args()) < 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := context.Args()[0]

	path := context.String("path")

	err := v.volDriver.Unmount(api.VolumeID(volumeID), path)
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	fmtOutput(context, &Format{UUID: []string{volumeID}})
}

func (v *volDriver) volumeFormat(context *cli.Context) {
	v.volumeOptions(context)
	fn := "format"
	if len(context.Args()) < 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := context.Args()[0]

	err := v.volDriver.Format(api.VolumeID(volumeID))
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	fmtOutput(context, &Format{UUID: []string{volumeID}})
}

func (v *volDriver) volumeAttach(context *cli.Context) {
	fn := "attach"
	if len(context.Args()) < 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}
	v.volumeOptions(context)
	volumeID := context.Args()[0]

	devicePath, err := v.volDriver.Attach(api.VolumeID(volumeID))
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	fmtOutput(context, &Format{Result: devicePath})
}

func (v *volDriver) volumeDetach(context *cli.Context) {
	fn := "detach"
	if len(context.Args()) < 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := context.Args()[0]
	v.volumeOptions(context)
	err := v.volDriver.Detach(api.VolumeID(volumeID))
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	fmtOutput(context, &Format{UUID: []string{context.Args()[0]}})
}

func (v *volDriver) volumeInspect(context *cli.Context) {
	v.volumeOptions(context)
	fn := "inspect"
	if len(context.Args()) < 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}

	d := make([]api.VolumeID, len(context.Args()))
	for i, v := range context.Args() {
		d[i] = api.VolumeID(v)
	}

	volumes, err := v.volDriver.Inspect(d)
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	cmdOutput(context, volumes)
}

func (v *volDriver) volumeStats(context *cli.Context) {
	v.volumeOptions(context)
	fn := "stats"
	if len(context.Args()) != 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}

	stats, err := v.volDriver.Stats(api.VolumeID(context.Args()[0]))
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	cmdOutput(context, stats)
}

func (v *volDriver) volumeAlerts(context *cli.Context) {
	v.volumeOptions(context)
	fn := "alerts"
	if len(context.Args()) != 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}

	alerts, err := v.volDriver.Alerts(api.VolumeID(context.Args()[0]))
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	cmdOutput(context, alerts)
}

func (v *volDriver) volumeEnumerate(context *cli.Context) {
	var locator api.VolumeLocator
	var err error

	fn := "enumerate"
	locator.Name = context.String("name")
	if l := context.String("label"); l != "" {
		locator.VolumeLabels, err = processLabels(l)
		if err != nil {
			cmdError(context, fn, err)
			return
		}
	}

	v.volumeOptions(context)
	volumes, err := v.volDriver.Enumerate(locator, nil)
	if err != nil {
		cmdError(context, fn, err)
		return
	}
	cmdOutput(context, volumes)
}

func (v *volDriver) volumeDelete(context *cli.Context) {
	fn := "delete"
	if len(context.Args()) < 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := context.Args()[0]
	v.volumeOptions(context)
	err := v.volDriver.Delete(api.VolumeID(volumeID))
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	fmtOutput(context, &Format{UUID: []string{context.Args()[0]}})
}

func (v *volDriver) snapCreate(context *cli.Context) {
	var err error
	var labels api.Labels
	fn := "snapCreate"

	if len(context.Args()) != 1 {
		missingParameter(context, fn, "volumeID", "Invalid number of arguments")
		return
	}
	volumeID := api.VolumeID(context.Args()[0])

	v.volumeOptions(context)
	if l := context.String("label"); l != "" {
		if labels, err = processLabels(l); err != nil {
			cmdError(context, fn, err)
			return
		}
	}
	locator := api.VolumeLocator{
		Name:         context.String("name"),
		VolumeLabels: labels,
	}
	readonly := context.Bool("readonly")
	id, err := v.volDriver.Snapshot(volumeID, readonly, locator)
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	fmtOutput(context, &Format{UUID: []string{string(id)}})
}

func (v *volDriver) snapEnumerate(context *cli.Context) {
	var locator api.VolumeLocator
	var err error

	fn := "snap enumerate"
	locator.Name = context.String("name")
	if l := context.String("label"); l != "" {
		locator.VolumeLabels, err = processLabels(l)
		if err != nil {
			cmdError(context, fn, err)
			return
		}
	}

	v.volumeOptions(context)
	snaps, err := v.volDriver.Enumerate(locator, nil)
	if err != nil {
		cmdError(context, fn, err)
		return
	}
	if snaps == nil {
		cmdError(context, fn, err)
		return
	}
	cmdOutput(context, snaps)
}

// baseVolumeCommand exports commands common to block and file volume drivers.
func baseVolumeCommand(v *volDriver) []cli.Command {

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
				cli.StringFlag{
					Name:  "seed",
					Usage: "optional data that the volume should be seeded with",
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
			Name:   "alerts",
			Usage:  "volume alerts",
			Action: v.volumeAlerts,
		},
		{
			Name:   "stats",
			Usage:  "volume stats",
			Action: v.volumeStats,
		},
		{
			Name:    "snap",
			Aliases: []string{"sc"},
			Usage:   "create snap",
			Action:  v.snapCreate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "user friendly name",
				},
				cli.StringFlag{
					Name:  "label,l",
					Usage: "Comma separated name=value pairs, e.g name=sqlvolume,type=production",
				},
				cli.BoolFlag{
					Name:  "readonly",
					Usage: "true if snapshot is readonly",
				},
			},
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
	}
	return commands
}

// BlockVolumeCommands exports CLI comamnds for a Block VolumeDriver.
func BlockVolumeCommands(name string) []cli.Command {
	v := &volDriver{name: name}

	blockCommands := []cli.Command{
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
			Name:    "detach",
			Aliases: []string{"d"},
			Usage:   "Detach specified volume",
			Action:  v.volumeDetach,
		},
	}

	baseCommands := baseVolumeCommand(v)

	return append(baseCommands, blockCommands...)
}

// FileVolumeCommands exports CLI comamnds for File VolumeDriver
func FileVolumeCommands(name string) []cli.Command {
	v := &volDriver{name: name}

	return baseVolumeCommand(v)
}
