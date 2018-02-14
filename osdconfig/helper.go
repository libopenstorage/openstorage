package osdconfig

import (
	"context"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/libopenstorage/openstorage/pkg/dbg"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

// helper function to get a new callback function that can be registered
func newCallback(name string, minSleep, maxSleep int) func(ctx context.Context,
	opt interface{}) (chan<- *dataWrite, <-chan *dataRead) {
	f := func(ctx context.Context, opt interface{}) (chan<- *dataWrite, <-chan *dataRead) {
		send := make(chan *dataWrite)
		recv := make(chan *dataRead)
		go func() {
			select {
			case msg := <-send:
				if msg.Err != nil {
					logrus.Error(msg.Err)
				}
			case <-ctx.Done():
				logrus.WithField("source", sourceCallback).
					WithField("callback", name).
					Warn(msgCtxCancelled)
				return
			}

			dur := time.Millisecond * time.Duration(rand.Intn(maxSleep)+minSleep)
			logrus.WithField("source", sourceCallback).
				WithField("callback", name).
				WithField("duration", dur).
				Debug("sleeping")
			select {
			case <-time.After(dur):
			case <-ctx.Done():
				logrus.WithField("source", sourceCallback).
					WithField("callback", name).
					Warn(msgCtxCancelled)
				return
			}

			d := new(dataRead)
			d.Name = name

			select {
			case recv <- d:
				return
			case <-ctx.Done():
				logrus.WithField("source", sourceCallback).
					WithField("callback", name).
					Warn(msgCtxCancelled)
				return
			}
		}()
		return send, recv
	}
	return f
}

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
	return filepath.Join(rootKey, nodeKey, nodeID)
}
