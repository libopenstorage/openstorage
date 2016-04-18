package coprhd

import (
	"errors"
	"fmt"
	"github.com/ModelRocket/coprhd"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/common"
	"github.com/portworx/kvdb"
	"go.pedge.io/dlog"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	// Name is the name of the coprhd driver
	Name = "coprhd"

	// Type is the coprhd driver type
	Type = api.DriverType_DRIVER_TYPE_BLOCK

	// minVolumeSize is the minimum volume size for coprhd
	minVolumeSize = 1024 * 1024 * 1000
)

var (
	// ErrApiUrlRequired is returned when the config is missing the url parameter
	ErrApiUrlRequired = errors.New("coprhd api url parameter must be set")

	// ErrApiAuthTokenRequired is returned when the auth token is missing from the config file
	ErrApiAuthTokenRequired = errors.New("coprhd auth token parameter is required")

	// ErrArrayRequired is returned when the varray name is not provided in the config
	ErrArrayRequired = errors.New("varray name is required")

	// ErrPoolRequired is returned when the vpool name is not provided in the config
	ErrPoolRequired = errors.New("vpool name is required")

	// ErrPortRequired is returned when the initiator name is not provided in the config
	ErrPortRequired = errors.New("port name is required")

	// ErrProjectRequired is returned when the project name is not provided in the config
	ErrProjectRequired = errors.New("project name is required")

	// ErrInvalidPool is returned when the virtual pool can not be located for the resource
	ErrInvalidPool = errors.New("pool does not support block type")

	// ErrInvalidPort is returned when the initiator does not exist or is of an unsupported type
	ErrInvalidPort = errors.New("port does not support pool initiator type")
)

type (
	driver struct {
		*volume.IoNotSupported
		*volume.DefaultEnumerator

		client *coprhd.Client

		// driver defaults
		project   *coprhd.Project
		varray    *coprhd.VArray
		vpool     *coprhd.VPool
		initiator *coprhd.Initiator
	}
)

func init() {
	volume.Register(Name, Init)
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	host, ok := params["url"]
	if !ok {
		return nil, ErrApiUrlRequired
	}

	token, ok := params["token"]
	if !ok {
		return nil, ErrApiAuthTokenRequired
	}

	// create a coprhd api client instance
	client := coprhd.NewClient(host, token)

	d := &driver{
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
		client:            client,
	}

	if projectName, ok := params["project"]; ok {
		if project, err := client.Project().Name(projectName).Query(); err != nil {
			return nil, err
		} else {
			d.project = project
		}
	} else {
		dlog.Warnln("Default coprhd 'project' not set")
	}

	if varrayName, ok := params["varray"]; ok {
		if varray, err := client.VArray().Name(varrayName).Query(); err != nil {
			return nil, err
		} else {
			d.varray = varray
		}
	} else {
		dlog.Warnf("Default coprhd 'varray' not set")
	}

	if vpoolName, ok := params["vpool"]; ok {
		if vpool, err := client.VPool().Name(vpoolName).Query(); err != nil {
			return nil, err
		} else {
			d.vpool = vpool
		}
	} else {
		dlog.Warnf("Default coprhd 'vpool' not set")
	}

	if port, ok := params["port"]; ok {
		if initiator, err := client.Initiator().Port(port).Query(); err != nil {
			return nil, err
		} else {
			d.initiator = initiator
		}
	} else {
		return nil, ErrPortRequired
	}

	return d, nil
}

func (d *driver) String() string {
	return Name
}

func (d *driver) Type() api.DriverType {
	return Type
}

func (d *driver) Create(
	locator *api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec) (string, error) {

	var err error

	project := d.project
	varray := d.varray
	vpool := d.vpool
	initiator := d.initiator

	if name, ok := locator.VolumeLabels["project"]; ok {
		project, err = d.client.Project().
			Name(name).
			Query()
		if err != nil {
			return "", err
		}
	}
	if project == nil {
		return "", ErrProjectRequired
	}

	if name, ok := locator.VolumeLabels["varray"]; ok {
		varray, err = d.client.VArray().
			Name(name).
			Query()
		if err != nil {
			return "", err
		}
	}
	if varray == nil {
		return "", ErrArrayRequired
	}

	if name, ok := locator.VolumeLabels["vpool"]; ok {
		vpool, err = d.client.VPool().
			Name(name).
			Query()
		if err != nil {
			return "", err
		}
	}
	if vpool == nil {
		return "", ErrPoolRequired
	}

	if !vpool.IsBlock() {
		return "", ErrInvalidPool
	}

	// make sure this pool supports the initiator protocol
	if !vpool.HasProtocol(initiator.Protocol) {
		return "", ErrInvalidPort
	}

	sz := spec.Size
	if sz < minVolumeSize {
		sz = minVolumeSize
	}

	vol, err := d.client.Volume().
		Name(locator.Name).
		Project(project.Id).
		Array(varray.Id).
		Pool(vpool.Id).
		Create(sz)
	if err != nil {
		return "", err
	}

	volumeID := strings.ToLower(vol.WWN)

	dlog.Infof("coprhd volume %s created", volumeID)

	volume := common.NewVolume(
		volumeID,
		api.FSType_FS_TYPE_EXT4,
		locator,
		source,
		spec)

	if err := d.UpdateVol(volume); err != nil {
		return "", err
	}

	if _, err := d.Attach(volumeID); err != nil {
		return "", err
	}

	dlog.Infof("coprhd preparing volume %s...", volumeID)

	if err := d.Format(volumeID); err != nil {
		return "", err
	}

	if err := d.Detach(volumeID); err != nil {
		return "", err
	}

	return volumeID, nil
}

func (d *driver) Delete(volumeID string) error {
	if _, err := d.GetVol(volumeID); err != nil {
		return fmt.Errorf("Volume %q could not be located", volumeID)
	}

	if err := d.client.Volume().
		WWN(volumeID).
		Delete(true); err != nil {
		return err
	}

	if err := d.DeleteVol(volumeID); err != nil {
		return err
	}

	return nil
}

func (d *driver) Stats(volumeID string) (*api.Stats, error) {
	return nil, volume.ErrNotSupported
}

func (d *driver) Alerts(volumeID string) (*api.Alerts, error) {
	return nil, volume.ErrNotSupported
}

func (d *driver) Attach(volumeID string) (path string, err error) {
	volume, err := d.GetVol(volumeID)
	if err != nil {
		return "", fmt.Errorf("Volume %q could not be located", volumeID)
	}

	coprVolume, err := d.client.Volume().
		WWN(volumeID).
		Query()
	if err != nil {
		return "", err
	}

	group, err := d.client.Export().
		Name(coprVolume.Name).
		Initiators(d.initiator.Id).
		Volumes(coprVolume.Id).
		Project(coprVolume.Project.Id).
		Array(coprVolume.VArray.Id).
		Create()
	if err != nil {
		return "", err
	}

	protocol := d.initiator.Protocol
	device := ""

	switch protocol {
	case coprhd.InitiatorTypeScaleIO:
		device, err = d.waitAttachDevice(volumeID, time.Second*180)
		if err != nil {
			return "", err
		}
	default:
		return "", ErrInvalidPort
	}

	volume.AttachedOn = group.Id
	volume.DevicePath = device
	volume.Status = api.VolumeStatus_VOLUME_STATUS_UP

	if err := d.UpdateVol(volume); err != nil {
		return "", err
	}
	return device, nil
}

func (d *driver) Detach(volumeID string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Volume %q is not attached", volumeID)
	}

	export := v.AttachedOn
	if export != "" {
		err = d.client.Export().
			Id(export).
			Delete()
		if err != nil {
			return err
		}
	}

	v.AttachedOn = ""
	v.DevicePath = ""
	v.Status = api.VolumeStatus_VOLUME_STATUS_DOWN

	if err := d.UpdateVol(v); err != nil {
		return err
	}

	return nil
}

func (d *driver) Mount(volumeID string, mountpath string) error {
	volume, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate attached volume %q", volumeID)
	}
	devicePath := volume.DevicePath
	if devicePath == "" {
		return fmt.Errorf("Invalid device path")
	}

	fstype := volume.Spec.Format.SimpleString()

	if err := syscall.Mount(devicePath, mountpath, fstype, 0, ""); err != nil {
		return err
	}

	volume.AttachPath = mountpath

	if err := d.UpdateVol(volume); err != nil {
		return err
	}
	return nil
}

func (d *driver) Unmount(volumeID string, mountpath string) error {
	volume, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate attached volume %q", volumeID)
	}
	if err := syscall.Unmount(mountpath, 0); err != nil {
		return err
	}

	volume.AttachPath = ""

	if err := d.UpdateVol(volume); err != nil {
		return err
	}

	return nil
}

func (d *driver) Set(volumeID string, locator *api.VolumeLocator, spec *api.VolumeSpec) error {
	return volume.ErrNotSupported
}

func (d *driver) Shutdown() {
	dlog.Infof("%s Shutting down", Name)
}

func (d *driver) Snapshot(volumeID string, readonly bool, locator *api.VolumeLocator) (string, error) {
	return "", volume.ErrNotSupported
}

func (v *driver) Status() [][2]string {
	return [][2]string{}
}

func (d *driver) Inspect(volumeIDs []string) ([]*api.Volume, error) {
	volumes, err := d.DefaultEnumerator.Inspect(volumeIDs)
	if err != nil {
		return nil, err
	}

	for _, volume := range volumes {
		protocol := d.initiator.Protocol
		device := ""

		switch protocol {
		case coprhd.InitiatorTypeScaleIO:
			device, err = d.waitAttachDevice(volume.Id, time.Second*1)
			if err == nil {
				volume.DevicePath = device
			}
		default:
		}
	}
	return volumes, nil
}

func (d *driver) Format(volumeID string) error {
	volume, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Volume %q is not attached", volumeID)
	}

	fstype := volume.Spec.Format.SimpleString()
	devicePath := volume.DevicePath

	cmd := "/sbin/mkfs." + fstype
	if _, err := exec.Command(cmd, devicePath).Output(); err != nil {
		return err
	}
	volume.Format = volume.Spec.Format
	if err := d.UpdateVol(volume); err != nil {
		return err
	}
	return nil
}

func (d *driver) waitAttachDevice(volumeID string, to time.Duration) (string, error) {
	if _, err := d.GetVol(volumeID); err != nil {
		return "", fmt.Errorf("Volume %s is not attached", volumeID)
	}

	timeout := time.After(to)
	timer := time.Tick(time.Millisecond * 100)

	protocol := d.initiator.Protocol
	device := ""

	for {
		switch protocol {
		case coprhd.InitiatorTypeScaleIO:
			device, _ = d.GetScaleIoDevice(volumeID)
			if device != "" {
				return device, nil
			}
		default:
			return "", ErrInvalidPort
		}

		select {
		case <-timer:
		case <-timeout:
			return "", fmt.Errorf("Timeout waiting for device to attach")
		}
	}
}
