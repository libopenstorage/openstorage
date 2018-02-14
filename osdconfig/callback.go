package osdconfig

import (
	"container/heap"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/portworx/kvdb"
)

// cb is a callback to be registered with kvdb.
// this callback simply receives data from kvdb and reflects it on a channel it receives in opaque
func cb(prefix string, opaque interface{}, kvp *kvdb.KVPair, err error) error {
	t := time.Now().UnixNano()
	c, ok := opaque.(*dataToKvdb)
	if !ok {
		return errors.New("opaque value type is incorrect")
	}

	wd := new(dataWrite)
	if kvp != nil {
		wd.Key = kvp.Key
		wd.Value = kvp.Value
	}
	wd.Type = c.Type
	wd.Err = err
	wd.Time = t

	// push all updates into a heap and never block
	c.cbh.Lock()
	heap.Push(c.cbh, wd)
	n := c.cbh.Len()
	c.cbh.Unlock()

	// debug log
	logrus.WithField("source", sourceKV).
		WithField("len", n).
		Debug(msgHeapStatus)

	// select statement below is never supposed to block, however,
	// it is required that it prioritizes trigger, when both writer
	//, i.e., this function, and the reader on the other end are in
	// lock step ensuring that jobs are served right away if they can.
	// However, if jobs are currently being processed and none can be
	// served right away, this function simply moves on by putting job
	// in the heap
	select {
	case <-c.ctx.Done():
	case c.trigger <- struct{}{}:
	default:
	}

	return nil
}

// getCallback build a function literal that can be registered.
// This is a helper util function for users of this library who wish to register callbacks of following forms:
// func(config *ClusterConfig) error and func(config *NodeConfig) error
func getCallback(name string, funcLiteral interface{}) (func(ctx context.Context, opt interface{}) (chan<- *dataWrite, <-chan *dataRead), error) {
	var watcherType Band

	// determine the watcher type based on funcLiteral type
	// and error out if funcLiteral type is not acceptable
	switch funcLiteral.(type) {
	case func(conifg *ClusterConfig) error:
		watcherType = TuneCluster
	case func(config *NodeConfig) error:
		watcherType = TuneNode
	default:
		return nil, errors.New("invalid funcLiteral signature")
	}

	// define the behavior of native callback that can be registered with osdconfig
	f := func(ctx context.Context, opt interface{}) (chan<- *dataWrite, <-chan *dataRead) {
		writeChan := make(chan *dataWrite)
		readChan := make(chan *dataRead)

		var D interface{}
		switch watcherType {
		case TuneCluster:
			D = new(ClusterConfig)
		case TuneNode:
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
				logrus.WithField("source", sourceCallback).
					WithField("callback", name).
					Warn(msgCtxCancelled)
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
					c <- errors.New(msgDataError)
					return
				}

				switch watcherType {
				case TuneCluster:
					c <- funcLiteral.(func(conifg *ClusterConfig) error)(D.(*ClusterConfig))
				case TuneNode:
					c <- funcLiteral.(func(conifg *NodeConfig) error)(D.(*NodeConfig))
				}

			}(err_ch)

			// prepare data to be send back
			d := new(dataRead)
			d.Name = name

			// wait for user function to complete or context to be done
			select {
			case err := <-err_ch:
				d.Err = err
				if err != nil {
					logrus.Error(err)
				}
			case <-ctx.Done():
				logrus.WithField("source", sourceCallback).
					WithField("callback", name).
					Warn(msgCtxCancelled)
				return
			}

			// send data or exit on context cancellation
			select {
			case readChan <- d:
				return
			case <-ctx.Done():
				logrus.WithField("source", sourceCallback).
					WithField("callback", name).
					Warn(msgCtxCancelled)
				return
			}
		}()
		return writeChan, readChan
	}

	return f, nil
}
