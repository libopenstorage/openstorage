package osdconfig

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
)

// printStatus will print existing status
func (manager *configManager) printStatus() {
	for key, val := range manager.status {
		key := key
		if val != nil {
			if val.Err != nil {
				logrus.WithField("source", sourceManagerPrintStatus).
					WithField("callback", key).
					WithField("lastStatus", val.Err).
					Warn(msgExecError)
			} else {
				logrus.WithField("source", sourceManagerPrintStatus).
					WithField("duration", val.Duration).
					WithField("callback", key).
					Debug(msgExecSuccess)
			}
		}
	}
}

// getContext returns global context for any associated processes to use
func (manager *configManager) getContext() context.Context {
	return manager.ctx
}

// getStatus returns execution status
func (manager *configManager) getStatus() map[string]*Status {
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

// error returns error status of callback execution
func (manager *configManager) error() error {
	manager.Lock()
	defer manager.Unlock()
	if manager.status == nil {
		return nil
	}
	for _, val := range manager.status {
		if val != nil {
			if val.Err != nil {
				return ErrorExec
			}
		}
	}
	return nil
}

// Close closes the contexts and releases bookkeeping memory
func (manager *configManager) Close() {
	// cancel contexts and wait for any pending tasks to complete
	logrus.WithField("source", sourceManagerClose).
		Info(msgCtxCancelled)
	manager.runCancel()
	manager.cancel()

	// clean up resources
	logrus.WithField("source", sourceManagerClose).
		Info(msgMemRelease)
	manager.cc = nil
	manager.cb = nil
	logrus.WithField("source", sourceManagerClose).
		Info(msgCleanup)
}

// abort sends context cancellation to running callbacks if any
func (manager *configManager) abort() {
	logrus.WithField("source", sourceManagerAbort).
		Warn(msgAborting)
	manager.runCancel()
}

// GetClusterConf retrieves cluster level data from kvdb
func (manager *configManager) GetClusterConf() (*ClusterConfig, error) {
	// get json from kvdb and unmarshal into config
	kvPair, err := manager.cc.Get(filepath.Join(rootKey, clusterKey))
	if err != nil {
		return nil, err
	}

	if kvPair == nil {
		return nil, ErrorData
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
		return ErrorInput
	}

	// push into kvdb
	_, err := manager.cc.Put(filepath.Join(rootKey, clusterKey), config, 0)
	return err
}

// GetNodeConf retrieves node config data using nodeID
func (manager *configManager) GetNodeConf(nodeID string) (*NodeConfig, error) {
	if len(nodeID) == 0 {
		return nil, ErrorInput
	}

	// get json from kvdb and unmarshal into config
	kvPair, err := manager.cc.Get(getNodeKeyFromNodeID(nodeID))
	if err != nil {
		return nil, err
	}

	if kvPair == nil {
		return nil, ErrorData
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
		return ErrorInput
	}

	if len(config.NodeId) == 0 {
		return ErrorInput
	}

	// push node data into kvdb
	_, err := manager.cc.Put(getNodeKeyFromNodeID(config.NodeId), config, 0)
	return err
}

// WatchCluster registers user defined function as callback and sets a watch for changes
// to cluster configuration
func (manager *configManager) WatchCluster(name string, cb func(config *ClusterConfig) error) error {
	f, _ := getCallback(name, cb)
	return manager.register(name, TuneCluster, nil, f)
}

// WatchNode registers user defined function as callback and sets a watch for changes
// to node configuration
func (manager *configManager) WatchNode(name string, cb func(config *NodeConfig) error) error {
	f, _ := getCallback(name, cb)
	return manager.register(name, TuneNode, nil, f)
}

// register registers callback functions
// callback to be registered is expected to be a service delivering a channel to write on
func (manager *configManager) register(name string, watcherType Band, opt interface{},
	cb func(ctx context.Context,
		opt interface{}) (chan<- *dataWrite, <-chan *dataRead)) error {

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
		return ErrorRegister
	}

	return nil
}

// run loops over all registered callbacks and executes them.
func (manager *configManager) run(wd *dataWrite) {
	go func(dataToCallback *dataWrite) {
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
		writeFanOut := make(chan *dataWrite)
		readFanIn := make(chan *dataRead)

		// start clock
		t := time.Now()

		// spawn callbacks
		for name, cb := range manager.cb {
			if cb.Type == wd.Type {
				name, cb := name, cb

				writeChannel, readChannel := cb.f(manager.runCtx, cb.opt)

				// update status
				dur := time.Since(t)
				manager.status[name].Err = errors.New(msgSpawned)
				manager.status[name].Duration = dur

				logrus.WithField("source", sourceManagerRun).
					WithField("callback", name).
					WithField("duration", dur).
					Debug(msgSpawned)

				// wire up writeChannel and receive channels for communicating with spawned routines
				go func(s chan *dataWrite, data chan<- *dataWrite) {
					data <- <-s
				}(writeFanOut, writeChannel)
				go func(r chan *dataRead, err <-chan *dataRead) {
					r <- <-err
				}(readFanIn, readChannel)
			}
		}

		// feed callbacks
		for _, val := range manager.cb {
			if val.Type == wd.Type {
				go func(c chan<- *dataWrite) {
					select {
					case <-manager.runCtx.Done():
						logrus.WithField("source", sourceManagerRun).
							Warn(msgCtxCancelled)
						logrus.WithField("source", sourceManagerRun).
							Warn(msgDataError)
					case c <- copyData(dataToCallback):
						logrus.WithField("source", sourceManagerRun).
							WithField("callback", "unknown").
							WithField("duration", time.Since(t)).
							Debug(msgDataSuccess)
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
					logrus.WithField("source", sourceManagerRun).
						Warn(msgCtxCancelled)
					logrus.WithField("source", sourceManagerRun).
						Warn(msgStatusError)
					break Loop
				case mesg := <-readFanIn:
					name, err, dur := mesg.Name, mesg.Err, time.Since(t)
					if err != nil {
						logrus.WithField("source", sourceManagerRun).
							WithField("callback", mesg.Name).
							WithField("duration", dur).
							Warn(msgExecError)
					} else {
						logrus.WithField("source", sourceManagerRun).
							WithField("callback", mesg.Name).
							WithField("duration", dur).
							Debug(msgExecSuccess)
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
func copyData(wd *dataWrite) *dataWrite {
	wd2 := new(dataWrite)
	wd2.Key = wd.Key
	if wd.Err != nil {
		wd2.Err = errors.New(wd.Err.Error())
	}
	wd2.Value = make([]byte, len(wd.Value))
	copy(wd2.Value, wd.Value)
	return wd2
}
