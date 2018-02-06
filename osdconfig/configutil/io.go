// package configutil is a higher level abstraction on osdconfig.
// It allows easier interface to setup kvdb watches
package configutil

import (
	"context"
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/libopenstorage/openstorage/osdconfig"
	"github.com/pkg/errors"
	"github.com/portworx/kvdb"
)

// NewManager instantiates a new osdconfig manager
func NewManager(kv kvdb.Kvdb) (osdconfig.ConfigManager, error) {
	return osdconfig.NewManager(context.Background(), kv)
}

// WatchCluster registers a function literal for watch on cluster level changes
func WatchCluster(manager osdconfig.ConfigManager, cb func(clusterConfig *osdconfig.ClusterConfig) error) error {
	name := "ioutil/custom/cluster"
	f := func(ctx context.Context, opt interface{}) (chan<- *osdconfig.DataWrite, <-chan *osdconfig.DataRead) {
		send := make(chan *osdconfig.DataWrite)
		recv := make(chan *osdconfig.DataRead)

		D := new(osdconfig.ClusterConfig)
		var b []byte
		go func() {
			select {
			case msg := <-send:
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

				c <- cb(D)

			}(err_ch)

			// prepare data to be send back
			d := new(osdconfig.DataRead)
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
			case recv <- d:
				return
			case <-ctx.Done():
				logrus.Info("context cancellation received in: ", name)
				return
			}
		}()
		return send, recv
	}

	if err := manager.Register(name, osdconfig.ClusterWatcher, nil, f); err != nil {
		return err
	}

	return nil
}

// WatchNodes registers a function literal for watch on nodes level changes
func WatchNode(manager osdconfig.ConfigManager, cb func(nodeConfig *osdconfig.NodeConfig) error) error {
	name := "ioutil/custom/nodes"
	f := func(ctx context.Context, opt interface{}) (chan<- *osdconfig.DataWrite, <-chan *osdconfig.DataRead) {
		send := make(chan *osdconfig.DataWrite)
		recv := make(chan *osdconfig.DataRead)

		D := new(osdconfig.NodeConfig)
		var b []byte
		go func() {
			select {
			case msg := <-send:
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

				c <- cb(D)

			}(err_ch)

			// prepare data to be send back
			d := new(osdconfig.DataRead)
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
			case recv <- d:
				return
			case <-ctx.Done():
				logrus.Info("context cancellation received in: ", name)
				return
			}
		}()
		return send, recv
	}

	if err := manager.Register(name, osdconfig.NodesWatcher, nil, f); err != nil {
		return err
	}

	return nil
}
