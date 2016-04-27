package volume

import "sync"

// VolumeDriverProvider provides VolumeDrivers.
type VolumeDriverProvider interface {
	// Get gets the VolumeDriver for the given name.
	// If a VolumeDriver was not created for the given name, the error ErrDriverNotFound is returned.
	Get(name string) (VolumeDriver, error)
	// Shutdown shuts down all volume drivers.
	Shutdown() error
}

// VolumeDriverRegistry registers VolumeDrivers.
type VolumeDriverRegistry interface {
	// New creates the VolumeDriver for the given name.
	// If a VolumeDriver was already created for the given name, the error ErrExist is returned.
	Register(name string, params map[string]string) error
}

// VolumeDriverRegistry constructs a new VolumeDriverRegistry.
func NewVolumeDriverRegistry(nameToInitFunc map[string]func(map[string]string) (VolumeDriver, error)) VolumeDriverRegistry {
	return newVolumeDriverRegistry(nameToInitFunc)
}

type volumeDriverRegistry struct {
	nameToInitFunc     map[string]func(map[string]string) (VolumeDriver, error)
	nameToVolumeDriver map[string]VolumeDriver
	lock               *sync.RWMutex
	isShutdown         bool
}

func newVolumeDriverRegistry(nameToInitFunc map[string]func(map[string]string) (VolumeDriver, error)) *volumeDriverRegistry {
	return &volumeDriverRegistry{
		nameToInitFunc,
		make(map[string]VolumeDriver),
		&sync.RWMutex{},
		false,
	}
}

func (v *volumeDriverRegistry) Get(name string) (VolumeDriver, error) {
	v.lock.RLock()
	defer v.lock.RUnlock()
	if v.isShutdown {
		return nil, ErrAlreadyShutdown
	}
	volumeDriver, ok := v.nameToVolumeDriver[name]
	if !ok {
		return nil, ErrDriverNotFound
	}
	return volumeDriver, nil
}

func (v *volumeDriverRegistry) Register(name string, params map[string]string) error {
	initFunc, ok := v.nameToInitFunc[name]
	if !ok {
		return ErrNotSupported
	}
	v.lock.Lock()
	defer v.lock.Unlock()
	if v.isShutdown {
		return ErrAlreadyShutdown
	}
	if _, ok := v.nameToVolumeDriver[name]; ok {
		return ErrExist
	}
	volumeDriver, err := initFunc(params)
	if err != nil {
		return err
	}
	v.nameToVolumeDriver[name] = volumeDriver
	return nil
}

func (v *volumeDriverRegistry) Shutdown() error {
	v.lock.Lock()
	if v.isShutdown {
		return ErrAlreadyShutdown
	}
	for _, volumeDriver := range v.nameToVolumeDriver {
		volumeDriver.Shutdown()
	}
	v.isShutdown = true
	return nil
}
