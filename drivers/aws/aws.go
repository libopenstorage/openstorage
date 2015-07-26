package ebs

import (
	"fmt"
	"math/rand"
	"os"

	api "github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

const (
	Name = "ec2"
)

var (
	r        *rand.Rand
	devMinor int32
)

type Ebs struct {
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	return &Ebs{}, nil
}

func (d *Ebs) String() string {
	return Name
}

func (d *Ebs) Create(l api.VolumeLocator, opt *api.CreateOptions, spec *api.VolumeSpec) (api.VolumeID, error) {
	return 0, nil
}

func (d *Ebs) AttachInfo(volInfo *api.VolumeInfo) (int32, string, error) {
	s := fmt.Sprintf("/tmp/gdd_%v", int(devMinor))
	return devMinor, s, nil
}

func (d *Ebs) Attach(volInfo api.VolumeID, path string) (string, error) {
	devMinor++
	s := fmt.Sprintf("/tmp/gdd_%v", int(devMinor))
	os.Create(s)
	return s, nil
}

func (d *Ebs) Detach(volID api.VolumeID) error {
	return nil
}

func (d *Ebs) Delete(volID api.VolumeID) error {
	return nil
}

func (d *Ebs) Format(id api.VolumeID) error {
	return nil
}

func (d *Ebs) Inspect(ids []api.VolumeID) ([]api.Volume, error) {
	return nil, nil
}

func (d *Ebs) Enumerate(locator api.VolumeLocator, labels api.Labels) []api.Volume {
	return nil
}

func (d *Ebs) Snapshot(volID api.VolumeID, labels api.Labels) (snap api.SnapID, err error) {

	return 0, nil
}

func (d *Ebs) SnapDelete(snapID api.SnapID) (err error) {
	return nil
}

func (d *Ebs) SnapInspect(snapID api.SnapID) (snap api.VolumeSnap, err error) {
	return api.VolumeSnap{}, nil
}

func (d *Ebs) SnapEnumerate(locator api.VolumeLocator, labels api.Labels) *[]api.SnapID {
	return nil
}

func (d *Ebs) Stats(volID api.VolumeID) (stats api.VolumeStats, err error) {
	return api.VolumeStats{}, nil
}

func (d *Ebs) Alerts(volID api.VolumeID) (stats api.VolumeAlerts, err error) {
	return api.VolumeAlerts{}, nil
}

func (d *Ebs) Shutdown() {
	fmt.Printf("%s Shutting down", Name)
}

func init() {
	volume.Register(Name, Init)
}
