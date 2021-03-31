package role

import (
	"errors"
	"fmt"
	"strings"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
)

// Datastore provides APIs for storing and fetching SDK Roles
// from a persistent datastore
type Datastore interface {

	// Exists returns true if the role exists in the datastore
	Exists(roleName string) (bool, error)

	// Update updates the role in the datastore. It will return an error
	// if the role does not exist in the datastore
	Update(roleName string, sdkRole *api.SdkRole) error

	// Put puts the role in the datastore. It will overwrite the
	// role object if it already exists
	Put(roleName string, sdkRole *api.SdkRole) error

	// Get gets the role object from the datastore for the given name
	Get(roleName string) (*api.SdkRole, error)

	// Create creates the role in the datastore. It will return an error
	// if the role with the name already exists
	Create(roleName string, sdkRole *api.SdkRole) error

	// EnumerateRoleNames enumerates all the roles from the datastore and returns their names
	EnumerateRoleNames() ([]string, error)

	// Delete deletes the role from the datastore
	Delete(roleName string) error
}

var (
	// ErrRoleNotFound is returned when role is not found in datastore
	ErrRoleNotFound = errors.New("role not found in datastore")
	// ErrRoleExists is returned when a role with the same name exists
	ErrRoleExists = errors.New("role exists with the same name")
)

const (
	rolePrefix = "cluster/roles"
)

// Simple function which creates key for Kvdb
func prefixWithName(name string) string {
	return rolePrefix + "/" + name
}

type kvRoleDatastore struct {
}

// NewKvdbRoleDatastore returns a kvdb implementation of Datastore interface
func NewKvdbRoleDatastore() (Datastore, error) {
	return newKvdbRoleDatastore()
}

func newKvdbRoleDatastore() (*kvRoleDatastore, error) {
	if kvdb.Instance() == nil {
		return nil, fmt.Errorf("kvdb is not initialized")
	}

	return &kvRoleDatastore{}, nil
}

func (k *kvRoleDatastore) Exists(roleName string) (bool, error) {
	if kvdb.Instance() == nil {
		return false, fmt.Errorf("kvdb is not initialized")
	}

	_, err := kvdb.Instance().Get(prefixWithName(roleName))
	if err == nil {
		return true, nil
	} else if err == kvdb.ErrNotFound {
		return false, nil
	}
	return false, err

}

func (k *kvRoleDatastore) Update(roleName string, sdkRole *api.SdkRole) error {
	if kvdb.Instance() == nil {
		return fmt.Errorf("kvdb is not initialized")
	}

	_, err := kvdb.Instance().Update(prefixWithName(roleName), sdkRole, 0)
	if err == kvdb.ErrNotFound {
		return ErrRoleNotFound
	}
	return err
}

func (k *kvRoleDatastore) Put(roleName string, sdkRole *api.SdkRole) error {
	if kvdb.Instance() == nil {
		return fmt.Errorf("kvdb is not initialized")
	}

	_, err := kvdb.Instance().Put(prefixWithName(roleName), sdkRole, 0)
	return err
}

func (k *kvRoleDatastore) Get(roleName string) (*api.SdkRole, error) {
	if kvdb.Instance() == nil {
		return nil, fmt.Errorf("kvdb is not initialized")
	}

	elem := &api.SdkRole{}
	_, err := kvdb.Instance().GetVal(prefixWithName(roleName), elem)
	if err == kvdb.ErrNotFound {
		return nil, ErrRoleNotFound
	} else if err != nil {
		return nil, err
	}
	return elem, nil

}

func (k *kvRoleDatastore) Create(roleName string, sdkRole *api.SdkRole) error {
	if kvdb.Instance() == nil {
		return fmt.Errorf("kvdb is not initialized")
	}

	_, err := kvdb.Instance().Create(prefixWithName(roleName), sdkRole, 0)
	if err == kvdb.ErrExist {
		return ErrRoleExists
	}
	return err
}

func (k *kvRoleDatastore) EnumerateRoleNames() ([]string, error) {
	if kvdb.Instance() == nil {
		return nil, fmt.Errorf("kvdb is not initialized")
	}

	kvPairs, err := kvdb.Instance().Enumerate(rolePrefix + "/")
	if err != nil {
		return nil, fmt.Errorf("failed to access roles from datastore: %v", err)
	}

	var names []string
	for _, kvPair := range kvPairs {
		names = append(names, strings.TrimPrefix(kvPair.Key, rolePrefix+"/"))
	}
	return names, nil

}

func (k *kvRoleDatastore) Delete(roleName string) error {
	if kvdb.Instance() == nil {
		return fmt.Errorf("kvdb is not initialized")
	}

	_, err := kvdb.Instance().Delete(prefixWithName(roleName))
	if err != kvdb.ErrNotFound && err != nil {
		return fmt.Errorf("failed to delete role: %v", err)
	}
	return nil

}
