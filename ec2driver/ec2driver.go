package ec2driver

import (
	"fmt"
	"math/rand"
	"os"

	api "github.com/libopenstorage/api"
	"github.com/libopenstorage/volume"
)

const (
	DriverName = "EC2_LIBOPENSTORAGE"
)

var (
	r        *rand.Rand
	devMinor int32
)

type EC2Driver struct {
}

func init() {
	fmt.Printf("EC2 init\n")
	volume.SetDefaultDriver(&EC2Driver{})
	r = rand.New(rand.NewSource(99))
	devMinor = 1
}

func (d *EC2Driver) String() string {
	return DriverName
}

func (d *EC2Driver) Create(diskInfo *api.VolumeInfo) error {
	return nil
}

func (d *EC2Driver) AttachInfo(diskInfo *api.VolumeInfo) (int32, string, error) {
	s := fmt.Sprintf("/tmp/gdd_%v", int(devMinor))
	return devMinor, s, nil
}

func (d *EC2Driver) Attach(diskInfo api.VolumeID, path string) (string, error) {
	devMinor++
	s := fmt.Sprintf("/tmp/gdd_%v", int(devMinor))
	os.Create(s)
	return s, nil
}

func (d *EC2Driver) Detach(diskInfo *api.VolumeInfo) error {
	return nil
}

func (d *EC2Driver) Delete(diskInfo *api.VolumeInfo) error {

	return nil
}

func (d *EC2Driver) Snapshot(
	diskID api.VolumeID,
	labels *api.Labels) (snap api.SnapID, err error) {

	return 0, nil
}

func (d *EC2Driver) SnapDelete(
	snapID api.SnapID) (err error) {

	return nil
}
func (d *EC2Driver) SnapInspect(
	snapID api.SnapID) (snap api.VolumeSnap, err error) {

	return api.VolumeSnap{}, nil
}
func (d *EC2Driver) SnapEnumerate(
	locator api.VolumeLocator, labels *api.Labels) *[]api.SnapID {

	return nil
}

func (d *EC2Driver) Stats(
	diskID api.VolumeID) (stats api.VolumeStats, err error) {

	return api.VolumeStats{}, nil
}

func (d *EC2Driver) Alerts(
	diskID api.VolumeID) (stats api.VolumeAlerts, err error) {

	return api.VolumeAlerts{}, nil
}

func (d *EC2Driver) Shutdown() {
	fmt.Printf("%s Shutting down", DriverName)
}
