package osdconfig

import (
	"context"
	"sync"

	"time"

	"github.com/portworx/kvdb"
)

// configManager implements ConfigManager
type configManager struct {
	// wrap a handle to kvdb
	cc kvdb.Kvdb

	// hashmap for callback bookkeeping
	cb map[string]*callbackData

	// value to be passed to callback
	opt interface{}

	// execution status
	status map[string]*Status

	//
	dataToCallback chan *DataWrite

	// placeholder for parent context
	parentContext context.Context

	// global context (derived from parent context)
	ctx    context.Context
	cancel context.CancelFunc

	// local context (derived from ctx)
	runCtx    context.Context
	runCancel context.CancelFunc

	// lock for executing kvdb watch
	watchLock sync.Mutex

	// mutex for locking during key operations
	sync.Mutex
}

// osdconfigError is for declaring error strings
type osdconfigError string

// Error returns string representation to satisfy Error interface
func (err osdconfigError) Error() string {
	return string(err)
}

// Watcher is a classifier for registering function
type Watcher string

// Status stores status of execution
type Status struct {
	Err      error
	Duration time.Duration
}

// DataToKvdb is data to be sent to kvdb as a state to run on
type DataToKvdb struct {
	ctx  context.Context
	Type Watcher
	wd   chan *DataWrite
}

// DataWrite is data to be sent to callbacks
// The contents here are populated based on what is received from kvdb
// Callback sends an instance of this on a channel that others can only write on
type DataWrite struct {
	// kvdb key received in kvdb.KvPair
	Key string

	// kvdb byte buffer received in kvdb.KvPair
	Value []byte

	// Type
	Type Watcher

	// kvdb error received in callback executed by kvdb
	Err error
}

// DataRead is data to be received from callback at callback completion
// Callback sends an instance of this on a channel that can others can only read from
type DataRead struct {
	// name of the callback
	Name string

	// error during callback execution
	Err error
}

// callbackData is callback metadata required for callback management
type callbackData struct {
	// functional literal that is registered
	f func(ctx context.Context, opt interface{}) (chan<- *DataWrite, <-chan *DataRead)

	// value to be passed to the function during execution
	opt interface{}

	// type of watcher
	Type Watcher
}
