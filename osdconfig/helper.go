package osdconfig

import (
	"path/filepath"

	"github.com/libopenstorage/openstorage/pkg/dbg"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

// helper function go get a new kvdb instance
func newInMemKvdb() (kvdb.Kvdb, error) {
	// create in memory kvdb
	if kv, err := kvdb.New(mem.Name, "", []string{}, nil, nil); err != nil {
		return nil, err
	} else {
		return kv, nil
	}
}

// helper function to obtain kvdb key for node based on nodeID
// the check for empty nodeID needs to be done elsewhere
func getNodeKeyFromNodeID(nodeID string) string {
	dbg.Assert(len(nodeID) > 0, "%s", "nodeID string can not be empty")
	return filepath.Join(baseKey, nodeKey, nodeID)
}

// copyData is a helper function to copy data to be fed to each callback
func copyData(wd *data) *data {
	wd2 := new(data)
	wd2.Key = wd.Key
	wd2.Value = make([]byte, len(wd.Value))
	copy(wd2.Value, wd.Value)
	return wd2
}
