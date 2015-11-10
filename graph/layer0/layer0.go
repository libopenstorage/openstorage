package graph

import (
	"fmt"
	"path"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/daemon/graphdriver/overlay"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/parsers"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/graph"
	"github.com/libopenstorage/openstorage/volume"
)

type Layer0Vol struct {
	// id self referential ID
	id string
	// parent image string
	parent string
	// path where the external volume is mounted.
	path string
	// volumeID mapping to this external volume
	volumeID api.VolumeID
	// ref keeps track of mount and unmounts.
	ref int32
}

type Layer0 struct {
	sync.Mutex
	// Driver is an implementation of GraphDriver. Only select methods are overridden
	graphdriver.Driver
	// home base string
	home string
	// volumes maintains a map of currently mounted volumes.
	volumes map[string]*Layer0Vol
	// volDriver is the volume driver used for the writeable layer.
	volDriver volume.VolumeDriver
}

// Layer0Graphdriver options
const (
	Layer0VolumeDriver = "layer0.volume_driver"
)

func init() {
	graph.Register("layer0", Init)
}

func Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {

	var volumeDriver string
	var params volume.DriverParams
	for _, option := range options {
		key, val, err := parsers.ParseKeyValueOpt(option)
		if err != nil {
			return nil, err
		}
		switch key {
		case Layer0VolumeDriver:
			volumeDriver = val
		default:
			return nil, fmt.Errorf("Unknown option %s\n", key)
		}
	}
	// XXX populate params
	volDriver, err := volume.New(volumeDriver, params)
	if err != nil {
		return nil, err
	}
	ov, err := overlay.Init(home, options, uidMaps, gidMaps)
	if err != nil {
		volDriver.Shutdown()
		return nil, err
	}
	d := &Layer0{
		Driver:    ov,
		home:      home,
		volumes:   make(map[string]*Layer0Vol),
		volDriver: volDriver,
	}

	return d, nil
}
func (l *Layer0) isLayer0Parent(id string) (string, bool) {
	// This relies on an <instance_id>-init volume being created for
	// every new container.
	if strings.HasSuffix(id, "-init") {
		return strings.TrimSuffix(id, "-init"), true
	}
	return "", false
}

func (l *Layer0) isLayer0(id string) bool {
	if strings.HasSuffix(id, "-init") {
		baseID := strings.TrimSuffix(id, "-init")
		if _, ok := l.volumes[baseID]; !ok {
			l.volumes[baseID] = &Layer0Vol{id: baseID}
		}
		return false
	}
	_, ok := l.volumes[id]
	return ok
}

func (l *Layer0) loID(id string) string {
	return id + "-vol"
}

func (l *Layer0) upperBase(id string) string {
	return path.Join(l.home, l.loID(id))
}

func (l *Layer0) realID(id string) string {
	if l.isLayer0(id) {
		return path.Join(l.loID(id), id)
	}
	return id
}

func (l *Layer0) create(id, parent string) (string, error) {

	l.Lock()
	defer l.Unlock()

	// If this is the parent of the Layer0, add an entry for it.
	baseID, l0 := l.isLayer0Parent(id)
	if l0 {
		l.volumes[baseID] = &Layer0Vol{id: baseID, parent: parent}
		return id, nil
	}

	// Don't do anything if this is not layer 0
	if !l.isLayer0(id) {
		return id, nil
	}

	vol, ok := l.volumes[id]
	if !ok {
		log.Warnf("Failed to find layer0 volume for id %v", id)
		return id, nil
	}

	// Query volume for Layer 0
	vols, err := l.volDriver.Enumerate(api.VolumeLocator{Name: vol.parent}, nil)

	// If we don't find a volume configured for this image,
	// then don't track layer0
	if err != nil || vols == nil {
		log.Warnf("Failed to find configured volume for id %v", vol.parent)
		delete(l.volumes, id)
		return id, nil
	}

	// Find a volume that is available.
	index := -1
	for i, v := range vols {
		if len(v.AttachPath) == 0 {
			index = i
			break
		}
	}
	if index == -1 {
		log.Warnf("Failed to find free volume for id %v", vol.parent)
		delete(l.volumes, id)
		return id, nil
	}

	mountPath := path.Join(l.home, l.loID(id))
	err = l.volDriver.Mount(vols[index].ID, mountPath)
	if err != nil {
		log.Errorf("Failed to mount volume %v at path %v",
			vols[index].ID, mountPath)
		delete(l.volumes, id)
		return id, nil
	}
	vol.path = mountPath
	vol.volumeID = vols[index].ID
	vol.ref = 1

	return l.realID(id), nil
}

func (l *Layer0) Create(id string, parent string) error {
	id, err := l.create(id, parent)
	if err != nil {
		return err
	}
	return l.Driver.Create(id, parent)
}
func (l *Layer0) Remove(id string) error {
	id = l.realID(id)
	return l.Driver.Remove(id)
}

func (l *Layer0) Get(id string, mountLabel string) (string, error) {
	id = l.realID(id)
	return l.Driver.Get(id, mountLabel)
}

func (l *Layer0) Put(id string) error {
	id = l.realID(id)
	return l.Driver.Put(id)
}

func (l *Layer0) ApplyDiff(id string, parent string, diff archive.Reader) (size int64, err error) {
	id = l.realID(id)
	return l.Driver.ApplyDiff(id, parent, diff)
}

func (l *Layer0) Exists(id string) bool {
	id = l.realID(id)
	return l.Driver.Exists(id)
}
