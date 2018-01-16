package osdconfig

import (
	"errors"

	"github.com/portworx/kvdb"
)

// cb is a callback to be registered with kvdb.
// this callback simply receives data from kvdb and reflects it on a channel it receives in opaque
func cb(prefix string, opaque interface{}, kvp *kvdb.KVPair, err error) error {
	c, ok := opaque.(*DataToKvdb)
	if !ok {
		return errors.New("opaque value type is incorrect")
	}

	wd := new(DataToCallback)
	if kvp != nil {
		wd.Key = kvp.Key
		wd.Value = kvp.Value
	}
	wd.Err = err
	select {
	case c.wd <- wd:
		return nil
	case <-c.ctx.Done():
		return errors.New("context done")
	}
}
