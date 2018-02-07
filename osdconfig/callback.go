package osdconfig

import (
	"errors"

	"context"
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/portworx/kvdb"
)

// cb is a callback to be registered with kvdb.
// this callback simply receives data from kvdb and reflects it on a channel it receives in opaque
func cb(prefix string, opaque interface{}, kvp *kvdb.KVPair, err error) error {
	c, ok := opaque.(*DataToKvdb)
	if !ok {
		return errors.New("opaque value type is incorrect")
	}

	wd := new(DataWrite)
	if kvp != nil {
		wd.Key = kvp.Key
		wd.Value = kvp.Value
	}
	wd.Type = c.Type
	wd.Err = err
	select {
	case c.wd <- wd:
		return nil
	case <-c.ctx.Done():
		return errors.New("context done")
	}
}

// GetCallback build a function literal that can be registered.
// This is a helper util function for users of this library who wish to register callbacks of following forms:
// func(config *ClusterConfig) error and func(config *NodeConfig) error
func GetCallback(name string, funcLiteral interface{}) (func(ctx context.Context, opt interface{}) (chan<- *DataWrite, <-chan *DataRead), error) {
	var watcherType Watcher

	// determine the watcher type based on funcLiteral type
	// and error out if funcLiteral type is not acceptable
	switch funcLiteral.(type) {
	case func(conifg *ClusterConfig) error:
		watcherType = ClusterWatcher
	case func(config *NodeConfig) error:
		watcherType = NodeWatcher
	default:
		return nil, errors.New("invalid funcLiteral signature")
	}

	// define the behavior of native callback that can be registered with osdconfig
	f := func(ctx context.Context, opt interface{}) (chan<- *DataWrite, <-chan *DataRead) {
		writeChan := make(chan *DataWrite)
		readChan := make(chan *DataRead)

		var D interface{}
		switch watcherType {
		case ClusterWatcher:
			D = new(ClusterConfig)
		case NodeWatcher:
			D = new(NodeConfig)
		}

		var b []byte
		go func() {
			select {
			case msg := <-writeChan:
				if msg.Err != nil {
					logrus.Error(msg.Err)
				} else {
					b = msg.Value
				}
			case <-ctx.Done():
				logrus.Info("context cancellation received in: ", name)
				return
			}

			// process data and execute user function in a goroutine
			// receiving error in a channel
			err_ch := make(chan error)
			go func(c chan error) {
				if err := json.Unmarshal(b, D); err != nil {
					c <- err
					return
				}

				if D == nil {
					c <- errors.New("invalid data received")
					return
				}

				switch watcherType {
				case ClusterWatcher:
					c <- funcLiteral.(func(conifg *ClusterConfig) error)(D.(*ClusterConfig))
				case NodeWatcher:
					c <- funcLiteral.(func(conifg *NodeConfig) error)(D.(*NodeConfig))
				}

			}(err_ch)

			// prepare data to be send back
			d := new(DataRead)
			d.Name = name

			// wait for user function to complete or context to be done
			select {
			case err := <-err_ch:
				d.Err = err
				if err != nil {
					logrus.Error(err)
				}
			case <-ctx.Done():
				logrus.Info("context cancellation received in: ", name)
				return
			}

			// send data or exit on context cancellation
			select {
			case readChan <- d:
				return
			case <-ctx.Done():
				logrus.Info("context cancellation received in: ", name)
				return
			}
		}()
		return writeChan, readChan
	}

	return f, nil
}
