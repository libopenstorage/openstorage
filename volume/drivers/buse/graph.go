package buse

import (
	"github.com/docker/docker/pkg/archive"
)

type graph struct {
}

//
// These functions below implement the graph graph interface.
//

func (g *graph) String() string {
	return Name
}

// Create creates a new, empty, filesystem layer with the
// specified id and parent. Parent may be "".
func (g *graph) Create(id, parent string) error {
	return nil
}

// Remove attempts to remove the filesystem layer with this id.
func (g *graph) Remove(id string) error {
	return nil
}

// Get returns the mountpoint for the layered filesystem referred
// to by this id. You can optionally specify a mountLabel or "".
// Returns the absolute path to the mounted layered filesystem.
func (g *graph) Get(id, mountLabel string) (dir string, err error) {
	return "", nil
}

// Put releases the system resources for the specified id,
// e.g, unmounting layered filesystem.
func (g *graph) Put(id string) error {
	return nil
}

// Exists returns whether a filesystem layer with the specified
// ID exists on this graph.
func (g *graph) Exists(id string) bool {
	return false
}

// Status returns a set of key-value pairs which give low
// level diagnostic status about this graph.
func (g *graph) Status() [][2]string {
	return [][2]string{
		{Name, "unknown"},
	}
}

// Returns a set of key-value pairs which give low level information
// about the image/container graph is managing.
func (g *graph) GetMetadata(id string) (map[string]string, error) {
	return nil, nil
}

// Cleanup performs necessary tasks to release resources
// held by the graph, e.g., unmounting all layered filesystems
// known to this graph.
func (g *graph) Cleanup() error {
	return nil
}

// Diff produces an archive of the changes between the specified
// layer and its parent layer which may be "".
func (g *graph) Diff(id, parent string) (archive.Archive, error) {
	return nil, nil
}

// Changes produces a list of changes between the specified layer
// and its parent layer. If parent is "", then all changes will be ADD changes.
func (g *graph) Changes(id, parent string) ([]archive.Change, error) {
	return nil, nil
}

// ApplyDiff extracts the changeset from the given diff into the
// layer with the specified id and parent, returning the size of the
// new layer in bytes.
// The archive.Reader must be an uncompressed stream.
func (g *graph) ApplyDiff(id, parent string, diff archive.Reader) (size int64, err error) {
	return 0, nil
}

// DiffSize calculates the changes between the specified id
// and its parent and returns the size in bytes of the changes
// relative to its base filesystem directory.
func (g *graph) DiffSize(id, parent string) (size int64, err error) {
	return 0, nil
}
