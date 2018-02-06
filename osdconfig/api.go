package osdconfig

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
)

// watch starts a watch on kvdb.
// it listens on a channel forever as long as context is alive
func (manager *configManager) watch(c <-chan *DataWrite) {
	// watch lock guarantees only one watcher is spawned
	manager.watchLock.Lock()
	defer manager.watchLock.Unlock()

	// loop forever until watch context is alive
	for {
		select {
		case <-manager.ctx.Done():
			logrus.Info("context cancellation received in: kvdb watch")
			logrus.Info("no longer watching for kvdb updates")
			return
		case wd := <-c:
			logrus.Info("callback execution trigger received from kvdb. exec type: ", wd.Type)

			if wd.Err != nil {
				logrus.Error("error during callback execution at kvdb")
			} else {
				// execute callbacks
				manager.Run(wd)

				// wait for callbacks to complete
				manager.Wait()

				// print execution status
				manager.printStatus()
			}
		}
	}
}

// printStatus will print existing status
func (manager *configManager) printStatus() {
	for key, val := range manager.status {
		key := key
		if val != nil {
			if val.Err != nil {
				logrus.Error("callback ", key, " executed with errors. Last status: ", val.Err)
			} else {
				logrus.Info("callback ", key, " executed successfully in ", val.Duration)
			}
		}
	}
}

// GetContext returns global context for any associated processes to use
func (manager *configManager) GetContext() context.Context {
	return manager.ctx
}

// GetStatus returns execution status
func (manager *configManager) GetStatus() map[string]*Status {
	manager.Lock()
	M := make(map[string]*Status)
	for key, val := range manager.status {
		key := key
		M[key] = new(Status)
		M[key].Err = val.Err
		M[key].Duration = val.Duration
	}
	manager.Unlock()
	return M
}

// Error returns error status of callback execution
func (manager *configManager) Error() error {
	manager.Lock()
	defer manager.Unlock()
	if manager.status == nil {
		return nil
	}
	for _, val := range manager.status {
		if val != nil {
			if val.Err != nil {
				return ExecErr
			}
		}
	}
	return nil
}

// Close closes the contexts and releases bookkeeping memory
func (manager *configManager) Close() {
	// cancel contexts and wait for any pending tasks to complete
	manager.runCancel()
	manager.cancel()
	manager.Wait()

	// clean up resources
	manager.cc = nil
	manager.cb = nil
	manager.opt = nil
}

// Abort sends context cancellation to running callbacks if any
func (manager *configManager) Abort() {
	logrus.Info("aborting via context cancellation")
	manager.runCancel()
	manager.Lock()
	logrus.Info("abort successful")
	manager.Unlock()
}

// Wait waits for running callbacks to finish
func (manager *configManager) Wait() {
	manager.Lock()
	manager.Unlock()
}

// GetClusterConf retrieves cluster level data from kvdb
func (manager *configManager) GetClusterConf() (*ClusterConfig, error) {
	// get json from kvdb and unmarshal into config
	kvPair, err := manager.cc.Get(filepath.Join(baseKey, clusterKey))
	if err != nil {
		return nil, err
	}

	if kvPair == nil {
		return nil, DataErr
	}

	config := new(ClusterConfig)
	if err := json.Unmarshal(kvPair.Value, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SetClusterConf sets cluster config in kvdb
func (manager *configManager) SetClusterConf(config *ClusterConfig) error {
	manager.Lock()
	defer manager.Unlock()

	if config == nil {
		return InputErr
	}

	// push into kvdb
	_, err := manager.cc.Put(filepath.Join(baseKey, clusterKey), config, 0)
	if err != nil {
		return err
	}
	return nil
}

// GetNodeConf retrieves node config data using nodeID
func (manager *configManager) GetNodeConf(nodeID string) (*NodeConfig, error) {
	if len(nodeID) == 0 {
		return nil, InputErr
	}

	// get json from kvdb and unmarshal into config
	kvPair, err := manager.cc.Get(getNodeKeyFromNodeID(nodeID))
	if err != nil {
		return nil, err
	}

	if kvPair == nil {
		return nil, DataErr
	}

	config := new(NodeConfig)
	if err = json.Unmarshal(kvPair.Value, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SetNodeConf sets node config data in kvdb
func (manager *configManager) SetNodeConf(config *NodeConfig) error {
	manager.Lock()
	defer manager.Unlock()

	if config == nil {
		return InputErr
	}

	if len(config.NodeId) == 0 {
		return InputErr
	}

	// push into kvdb
	if _, err := manager.cc.Put(getNodeKeyFromNodeID(config.NodeId), config, 0); err != nil {
		return err
	}
	return nil
}

// Register registers callback functions
// callback to be registered is expected to be a service delivering a channel to write on
func (manager *configManager) Register(name string, watcherType Watcher, opt interface{},
	cb func(ctx context.Context,
		opt interface{}) (chan<- *DataWrite, <-chan *DataRead)) error {

	// obtain lock since registering a callback changes the state of manager object
	manager.Lock()
	defer manager.Unlock()

	// callbacks are registered by the "name" string. only one per name
	if _, present := manager.cb[name]; !present {
		cbi := new(callbackData)
		cbi.f = cb
		cbi.opt = opt
		cbi.Type = watcherType
		manager.cb[name] = cbi
	} else {
		return RegErr
	}

	return nil
}

// Run loops over all registered callbacks and executes them.
func (manager *configManager) Run(wd *DataWrite) {
	go func(dataToCallback *DataWrite) {
		manager.Lock()
		defer manager.Unlock()

		// re-establish the context
		// derive local contexts from parent context
		manager.runCtx, manager.runCancel = context.WithCancel(manager.ctx)

		// prepare status map
		if manager.status != nil {
			manager.status = make(map[string]*Status)
		}
		for name, val := range manager.cb {
			if val.Type == wd.Type {
				name := name
				manager.status[name] = new(Status)
			}
		}

		// create a channel to communicate with callbacks
		writeFanOut := make(chan *DataWrite)
		readFanIn := make(chan *DataRead)

		// start clock
		t := time.Now()

		// spawn callbacks
		for name, cb := range manager.cb {
			if cb.Type == wd.Type {
				name, cb := name, cb

				writeChannel, readChannel := cb.f(manager.runCtx, cb.opt)

				// update status
				dur := time.Since(t)
				manager.status[name].Err = errors.New("spawned")
				manager.status[name].Duration = dur

				logrus.Info("callback: ", name, " spawned in ", dur)

				// wire up writeChannel and receive channels for communicating with spawned routines
				go func(s chan *DataWrite, data chan<- *DataWrite) {
					data <- <-s
				}(writeFanOut, writeChannel)
				go func(r chan *DataRead, err <-chan *DataRead) {
					r <- <-err
				}(readFanIn, readChannel)
			}
		}

		// feed callbacks
		for _, val := range manager.cb {
			if val.Type == wd.Type {
				go func(c chan<- *DataWrite) {
					select {
					case <-manager.runCtx.Done():
						logrus.Info("context cancellation received in: callback scheduler")
						logrus.Error("not all callbacks received data")
					case c <- copyData(dataToCallback):
						logrus.Info("a callback received data in ", time.Since(t))
					}
				}(writeFanOut)
			}
		}

		// receive completion status from callbacks
	Loop:
		for _, val := range manager.cb {
			if val.Type == wd.Type {
				select {
				case <-manager.runCtx.Done():
					logrus.Info("context cancellation received in: callback scheduler")
					logrus.Error("not all callbacks delivered completion status")
					break Loop
				case mesg := <-readFanIn:
					name, err, dur := mesg.Name, mesg.Err, time.Since(t)
					if err != nil {
						logrus.Error(mesg.Name, " callback completed w/  errors: ", dur)
					} else {
						logrus.Info(mesg.Name, " callback completed w/o errors:", dur)
					}
					// update status
					manager.status[name].Err = err
					manager.status[name].Duration = dur
				}
			}
		}

		// cancel context
		manager.runCancel()

		// print execution status
		manager.printStatus()
	}(wd)
}

// copyData is a helper function to copy data to be fed to each callback
func copyData(wd *DataWrite) *DataWrite {
	wd2 := new(DataWrite)
	wd2.Key = wd.Key
	if wd.Err != nil {
		wd2.Err = errors.New(wd.Err.Error())
	}
	wd2.Value = make([]byte, len(wd.Value))
	copy(wd2.Value, wd.Value)
	return wd2
}
