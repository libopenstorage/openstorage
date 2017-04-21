package volume

import (
	"github.com/libopenstorage/openstorage/api"
)

var (
	// BlockNotSupported is a default (null) block driver implementation.  This can be
	// used by drivers that do not want to (or care about) implementing the attach,
	// format and detach interfaces.
	BlockNotSupported = &blockNotSupported{}
	// SnapshotNotSupported is a null snapshot driver implementation. This can be used
	// by drivers that do not want to implement the snapshot interface
	SnapshotNotSupported = &snapshotNotSupported{}
	// IONotSupported is a null IODriver interface
	IONotSupported = &ioNotSupported{}
	// StatsNotSupported is a null stats driver implementation. This can be used
	// by drivers that do not want to implement the stats interface.
	StatsNotSupported = &statsNotSupported{}
	// AlertsNotSupported is a null alerts driver implementation. This can be used
	// by drivers that do not want to implement the alerts interface.
	AlertsNotSupported = &alertsNotSupported{}
)

type blockNotSupported struct{}

func (b *blockNotSupported) Attach(volumeID string, attachOptions map[string]string) (string, error) {
	return "", ErrNotSupported
}

func (b *blockNotSupported) Detach(volumeID string) error {
	return ErrNotSupported
}

type snapshotNotSupported struct{}

func (s *snapshotNotSupported) Snapshot(volumeID string, readonly bool, locator *api.VolumeLocator) (string, error) {
	return "", ErrNotSupported
}

type ioNotSupported struct{}

func (i *ioNotSupported) Read(volumeID string, buffer []byte, size uint64, offset int64) (int64, error) {
	return 0, ErrNotSupported
}

func (i *ioNotSupported) Write(volumeID string, buffer []byte, size uint64, offset int64) (int64, error) {
	return 0, ErrNotSupported
}

func (i *ioNotSupported) Flush(volumeID string) error {
	return ErrNotSupported
}

type statsNotSupported struct{}

// Stats returns stats
func (s *statsNotSupported) Stats(
	volumeID string,
	cumulative bool,
) (*api.Stats, error) {
	return nil, ErrNotSupported
}

// UsedSize returns allocated size
func (s *statsNotSupported) UsedSize(volumeID string) (uint64, error) {
	return 0, ErrNotSupported
}

// GetActiveRequests gets active requests
func (s *statsNotSupported) GetActiveRequests() (*api.ActiveRequests, error) {
	return nil, nil
}

type alertsNotSupported struct {}

// Alerts returns active alerts
func (s *alertsNotSupported) Alerts() (*api.Alerts, error) {
	return nil, ErrNotSupported
}

// ClearAlert clears an active alert
func (a *alertsNotSupported) ClearAlert(alertId int64) error {
	return ErrNotSupported
}

// EraseAlert erases an alert
func (a *alertsNotSupported) EraseAlert(alertId int64) error {
	return ErrNotSupported
}
