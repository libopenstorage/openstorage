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
	Name = "coprhd"
	Type = api.DriverType_DRIVER_TYPE_BLOCK
)

var (
	ErrArrayRequired   = errors.New("varray name is required")
	ErrPoolRequired    = errors.New("vpool name is required")
	ErrPortRequired    = errors.New("port name is required")
	ErrProjectRequired = errors.New("project name is required")
	ErrInvalidPool     = errors.New("pool does not support block type")
	ErrInvalidPort     = errors.New("port does not support pool initiator type")
)

type (
	driver struct {
		*volume.IoNotSupported
		*volume.DefaultEnumerator

		coprhd *coprhd.Client

		// driver defaults
		project *coprhd.Project
		varray  *coprhd.VArray
		vpool   *coprhd.VPool
		itr     *coprhd.Initiator
	}
)

func init() {
	volume.Register(Name, Init)
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	host, ok := params["url"]
	if !ok {
		return nil, fmt.Errorf("coprhd rest api 'url' configuration parameter must be set")
	}

	token, ok := params["token"]
	if !ok {
		return nil, fmt.Errorf("coprhd rest auth 'token' must be set")
	}

	// create a coprhd api client instance
	client := coprhd.NewClient(host, token)

	d := &driver{
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
		coprhd:            client,
	}

	if p, ok := params["project"]; ok {
		project, err := client.Project().
			Name(p).
			Query()
		if err != nil {
			return nil, err
		}
		d.project = project
	} else {
		dlog.Warnf("Default coprhd 'project' not set")
	}

	if va, ok := params["varray"]; ok {
		varray, err := client.VArray().
			Name(va).
			Query()
		if err != nil {
			return nil, err
		}
		d.varray = varray
	} else {
		dlog.Warnf("Default coprhd 'varray' not set")
	}

	if vp, ok := params["vpool"]; ok {
		vpool, err := client.VPool().
			Name(vp).
			Query()
		if err != nil {
			return nil, err
		}
		d.vpool = vpool
	} else {
		dlog.Warnf("Default coprhd 'vpool' not set")
	}

	if port, ok := params["port"]; ok {
		itr, err := client.Initiator().
			Port(port).
			Query()
		if err != nil {
			return nil, err
		}
		d.itr = itr
	} else {
		return nil, fmt.Errorf("coprhd initiator 'port' must be set")
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
	itr := d.itr

	if name, ok := locator.VolumeLabels["project"]; ok {
		project, err = d.coprhd.Project().
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
		varray, err = d.coprhd.VArray().
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
		vpool, err = d.coprhd.VPool().
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
	found := false
	for _, t := range vpool.Protocols {
		if t == itr.Protocol {
			found = true
		}
	}
	if !found {
		return "", ErrInvalidPort
	}

	vol, err := d.coprhd.Volume().
		Name(locator.Name).
		Project(project.Id).
		Array(varray.Id).
		Pool(vpool.Id).
		Create(spec.Size)
	if err != nil {
		return "", err
	}

	dlog.Infof("coprhd volume %s created", vol.WWN)

	volumeID := strings.ToLower(vol.WWN)

	volume := common.NewVolume(
		volumeID,
		api.FSType_FS_TYPE_NONE,
		locator,
		source,
		spec)

	err = d.UpdateVol(volume)
	if err != nil {
		return "", err
	}

	_, err = d.Attach(volumeID)
	if err != nil {
		return "", err
	}

	dlog.Infof("coprhd preparing volume %s...", vol.WWN)

	err = d.Format(volumeID)
	if err != nil {
		return "", err
	}

	err = d.Detach(volumeID)
	if err != nil {
		return "", err
	}

	return volumeID, nil
}

func (d *driver) Delete(volumeID string) error {
	_, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Volume could not be located")
	}

	err = d.coprhd.Volume().
		WWN(volumeID).
		Delete(true)
	if err != nil {
		return err
	}

	err = d.DeleteVol(volumeID)
	if err != nil {
		dlog.Println(err)
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
	v, err := d.GetVol(volumeID)
	if err != nil {
		return "", fmt.Errorf("Volume %s could not be located", volumeID)
	}

	vol, err := d.coprhd.Volume().
		WWN(volumeID).
		Query()
	if err != nil {
		return "", err
	}

	group, err := d.coprhd.Export().
		Name(vol.Name).
		Initiators(d.itr.Id).
		Volumes(vol.Id).
		Project(vol.Project.Id).
		Array(vol.VArray.Id).
		Create()
	if err != nil {
		return "", err
	}

	proto := d.itr.Protocol
	dev := ""

	switch proto {
	case coprhd.InitiatorTypeScaleIO:
		dev, err = d.waitAttachDevice(volumeID, time.Second*180)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("Unsupported initiator protocol")
	}

	v.AttachedOn = group.Id
	v.DevicePath = dev
	v.Status = api.VolumeStatus_VOLUME_STATUS_UP

	err = d.UpdateVol(v)
	if err != nil {
		return "", err
	}
	return dev, nil
}

func (d *driver) Detach(volumeID string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Volume is not attached")
	}

	export := v.AttachedOn
	if export != "" {
		err = d.coprhd.Export().
			Id(export).
			Delete()
		if err != nil {
			return err
		}
	}

	v.AttachedOn = ""
	v.DevicePath = ""
	v.Status = api.VolumeStatus_VOLUME_STATUS_DOWN

	err = d.UpdateVol(v)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) Mount(volumeID string, mountpath string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate attached volume %q", volumeID)
	}
	devicePath := v.DevicePath
	if devicePath == "" {
		return fmt.Errorf("Invalid device path")
	}

	fstype := v.Spec.Format.SimpleString()

	err = syscall.Mount(devicePath, mountpath, fstype, 0, "")
	if err != nil {
		return err
	}

	v.AttachPath = mountpath

	err = d.UpdateVol(v)
	if err != nil {
		return err
	}
	return nil
}

func (d *driver) Unmount(volumeID string, mountpath string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Failed to locate attached volume %q", volumeID)
	}
	err = syscall.Unmount(mountpath, 0)
	if err != nil {
		v.AttachPath = ""
	}

	err = d.UpdateVol(v)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) Set(
	volumeID string,
	locator *api.VolumeLocator,
	spec *api.VolumeSpec) error {
	return volume.ErrNotSupported
}

func (d *driver) Shutdown() {
	dlog.Infof("%s Shutting down", Name)
}

func (d *driver) Snapshot(
	volumeID string,
	readonly bool,
	locator *api.VolumeLocator) (string, error) {
	return "", volume.ErrNotSupported
}

func (v *driver) Status() [][2]string {
	return [][2]string{}
}

func (d *driver) Inspect(volumeIDs []string) ([]*api.Volume, error) {
	vols, err := d.DefaultEnumerator.Inspect(volumeIDs)
	if err != nil {
		return nil, err
	}

	for i, v := range vols {
		proto := d.itr.Protocol
		dev := ""

		switch proto {
		case coprhd.InitiatorTypeScaleIO:
			dev, err = d.waitAttachDevice(v.Id, time.Second*1)
			if err == nil {
				vols[i].DevicePath = dev
			}
		default:
		}
	}
	return vols, nil
}

func (d *driver) Format(volumeID string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return fmt.Errorf("Volume is not attached")
	}

	fstype := v.Spec.Format.SimpleString()
	devicePath := v.DevicePath

	cmd := "/sbin/mkfs." + fstype
	o, err := exec.Command(cmd, devicePath).Output()
	if err != nil {
		dlog.Warnf("Failed to run command %v %v: %v", cmd, devicePath, o)
		return err
	}
	v.Format = v.Spec.Format
	err = d.UpdateVol(v)
	return err
}

func (d *driver) waitAttachDevice(volumeID string, to time.Duration) (string, error) {
	_, err := d.GetVol(volumeID)
	if err != nil {
		return "", fmt.Errorf("Volume is not attached")
	}

	timeout := time.After(to)
	timer := time.Tick(time.Millisecond * 100)

	proto := d.itr.Protocol
	dev := ""

	for {
		switch proto {
		case coprhd.InitiatorTypeScaleIO:
			dev, _ = d.getScaleIoDevice(volumeID)
			if dev != "" {
				return dev, nil
			}
		default:
			return "", fmt.Errorf("Unsupported initiator protocol")
		}

		select {
		case <-timer:
		case <-timeout:
			return "", fmt.Errorf("Timeout waiting for device to attach")
		}
	}
}
