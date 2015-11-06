package api

// GraphDriverChanges represent a list of changes between the filesystem layers
// specified by the ID and Parent.  // Parent may be an empty string, in which
// case there is no parent.
// Where the Path is the filesystem path within the layered filesystem
// that is changed and Kind is an integer specifying the type of change that occurred:
// 0 - Modified
// 1 - Added
// 2 - Deleted
type GraphDriverChanges struct {
	Path string // "/some/path"
	Kind int
}
