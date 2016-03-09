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
)

const (
	Name = "coprhd"
	Type = api.DriverType_DRIVER_TYPE_BLOCK
)

var (
	ErrArrayRequired       = errors.New("varray is required")
	ErrPoolRequired        = errors.New("vpool is required")
	ErrProjectRequired     = errors.New("project label is required")
	ErrInvalidPool         = errors.New("pool does not support block type")
	ErrInvalidPoolProtocol = errors.New("pool does not iscsi")
)

type (
	driver struct {
		*volume.IoNotSupported
		*volume.DefaultEnumerator

		client *coprhd.Client
		itr    *coprhd.Initiator
	}
)

func init() {
	volume.Register(Name, Init)
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	host, ok := params["url"]
	if !ok {
		return nil, fmt.Errorf("rest api 'url' configuration parameter must be set")
	}

	token, ok := params["token"]
	if !ok {
		return nil, fmt.Errorf("rest auth 'token' must be set")
	}

	iqn, ok := params["iqn"]
	if !ok {
		return nil, fmt.Errorf("iscsi 'iqn' must be set")
	}

	// create a coprhd api client instance
	client := coprhd.NewClient(host, token)

	// search for the specified initiator
	itr, err := client.Initiator().
		Search("initiator_port=" + iqn)
	if err != nil {
		return nil, fmt.Errorf("iSCSI initiator %s could not be located: %s", iqn, err.Error())
	}

	dlog.Infof("iSCSI initiator found %s", itr.Id)

	d := &driver{
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
		client:            client,
		itr:               itr,
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

	name, ok := locator.VolumeLabels["project"]
	if !ok {
		return "", ErrProjectRequired
	}

	project, err := d.client.Project().
		Search("name=" + name)
	if err != nil {
		return "", err
	}

	name, ok = locator.VolumeLabels["varray"]
	if !ok {
		return "", ErrArrayRequired
	}

	varray, err := d.client.VArray().
		Search("name=" + name)
	if err != nil {
		return "", err
	}

	name, ok = locator.VolumeLabels["vpool"]
	if !ok {
		return "", ErrPoolRequired
	}

	vpool, err := d.client.VPool().
		Search("name=" + name)
	if err != nil {
		return "", err
	}

	if !vpool.IsBlock() {
		return "", ErrInvalidPool
	}

	if !vpool.HasProtocol(coprhd.InitiatorTypeISCSI) {
		return "", ErrInvalidPoolProtocol
	}

	vol, err := d.client.Volume().
		Project(project.Id).
		Array(varray.Id).
		Pool(vpool.Id).
		Create(locator.Name, spec.Size)
	if err != nil {
		return "", err
	}

	volume := common.NewVolume(
		vol.Id,
		api.FSType_FS_TYPE_NONE,
		locator,
		source,
		spec)

	err = d.UpdateVol(volume)
	if err != nil {
		return "", nil
	}

	return vol.Id, nil
}

func (d *driver) Delete(volumeID string) error {
	return d.client.Volume().
		Id(volumeID).
		Delete(true)
}

func (d *driver) Stats(volumeID string) (*api.Stats, error) {
	return nil, volume.ErrNotSupported
}

func (d *driver) Alerts(volumeID string) (*api.Alerts, error) {
	return nil, volume.ErrNotSupported
}

func (d *driver) Attach(volumeID string) (path string, err error) {
	return "", nil
}

func (d *driver) Detach(volumeID string) error {
	return nil
}

func (d *driver) Mount(volumeID string, mountpath string) error {
	return nil
}

func (d *driver) Unmount(volumeID string, mountpath string) error {
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
	return "", nil
}

func (v *driver) Status() [][2]string {
	return [][2]string{}
}
