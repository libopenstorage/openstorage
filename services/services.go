//go:generate mockgen -package=mock -destination=mock/services.mock.go github.com/libopenstorage/openstorage/services Service
package services

import "errors"

var (
	ErrNotImplemented = errors.New("Not Implemented")
)

type Service interface {
	// ServiceAddDrive adds the specified drive
	ServiceAddDrive(op, drive string, journal bool) (string, error)

	// ServiceReplaceDrive source with target.
	ServiceReplaceDrive(op, source, target string) (string, error)

	// ServiceRebalance the storage pool
	ServiceRebalancePool(op string, poolID int) (string, error)

	// ServiceExitMaintenanceMode exits maintenance mode
	ServiceExitMaintenanceMode() error

	// ServiceEnterMaintenanceMode enters maintenance mode and exits the proces
	ServiceEnterMaintenanceMode(exitOut bool) error
}

type nullServiceMgr struct {
}

func NewDefaultService() *nullServiceMgr {
	return &nullServiceMgr{}
}

func (s *nullServiceMgr) ServiceAddDrive(op, drive string, journal bool) (string, error) {
	return "", ErrNotImplemented
}

func (s *nullServiceMgr) ServiceReplaceDrive(op, source, target string) (string, error) {
	return "", ErrNotImplemented
}

func (s *nullServiceMgr) ServiceRebalancePool(op string, PoolID int) (string, error) {
	return "", ErrNotImplemented
}

func (s *nullServiceMgr) ServiceExitMaintenanceMode() error {
	return ErrNotImplemented
}

func (s *nullServiceMgr) ServiceEnterMaintenanceMode(exitOut bool) error {
	return ErrNotImplemented
}
