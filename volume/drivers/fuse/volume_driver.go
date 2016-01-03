package fuse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/common"
	"github.com/pborman/uuid"
	"github.com/portworx/kvdb"
)

type volumeDriver struct {
	name        string
	baseDirPath string
	provider    Provider
	*volume.IoNotSupported
	*volume.DefaultBlockDriver
	*volume.DefaultEnumerator
	*volume.SnapshotNotSupported
}

func newVolumeDriver(
	name string,
	baseDirPath string,
	provider Provider,
) *volumeDriver {
	return &volumeDriver{
		name,
		baseDirPath,
		provider,
		&volume.IoNotSupported{},
		&volume.DefaultBlockDriver{},
		volume.NewDefaultEnumerator(
			name,
			kvdb.Instance(),
		),
		&volume.SnapshotNotSupported{},
	}
}

func (v *volumeDriver) String() string {
	return v.name
}

func (v *volumeDriver) Type() api.DriverType {
	return api.DriverType_DRIVER_TYPE_FILE
}

func (v *volumeDriver) Create(
	volumeLocator *api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec,
) (string, error) {
	volumeID := strings.TrimSpace(string(uuid.New()))
	dirPath := filepath.Join(v.baseDirPath, volumeID)
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return "", err
	}
	volume := common.NewVolume(
		volumeID,
		api.FSType_FS_TYPE_FUSE,
		volumeLocator,
		source,
		spec,
	)
	volume.DevicePath = dirPath
	if err := v.CreateVol(volume); err != nil {
		return "", err
	}
	if err := v.UpdateVol(volume); err != nil {
		return "", err
	}
	return volume.Id, nil
}

func (v *volumeDriver) Delete(volumeID string) error {
	if _, err := v.GetVol(volumeID); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(v.baseDirPath, string(volumeID))); err != nil {
		return err
	}
	return v.DeleteVol(volumeID)
}

func (v *volumeDriver) Mount(volumeID string, mountpath string) error {
	volume, err := v.GetVol(volumeID)
	if err != nil {
		return err
	}
	mountOptions, err := v.provider.GetMountOptions(volume.Spec)
	if err != nil {
		return err
	}
	conn, err := fuse.Mount(mountpath, mountOptions...)
	if err != nil {
		return err
	}
	filesystem, err := v.provider.GetFS(volume.Spec)
	if err != nil {
		return err
	}
	go func() {
		// TODO: track error once we understand driver model better
		_ = fs.Serve(conn, filesystem)
		_ = conn.Close()
	}()
	<-conn.Ready
	return conn.MountError
}

func (v *volumeDriver) Unmount(volumeID string, mountpath string) error {
	volume, err := v.GetVol(volumeID)
	if err != nil {
		return err
	}
	if volume.AttachPath == "" {
		return fmt.Errorf("openstorage: device not mounted: %v", volumeID)
	}
	if err := fuse.Unmount(volume.AttachPath); err != nil {
		return err
	}
	volume.AttachPath = ""
	return v.UpdateVol(volume)
}

func (v *volumeDriver) Set(volumeID string, locator *api.VolumeLocator, spec *api.VolumeSpec) error {
	return volume.ErrNotSupported

}

func (v *volumeDriver) Stats(volumeID string) (*api.Stats, error) {
	return nil, volume.ErrNotSupported
}

func (v *volumeDriver) Alerts(volumeID string) (*api.Alerts, error) {
	return nil, volume.ErrNotSupported
}

func (v *volumeDriver) Status() [][2]string {
	return [][2]string{}
}

func (v *volumeDriver) Shutdown() {}
