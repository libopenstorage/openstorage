package osdconfig

import (
	"context"
	"sync"
	"time"

	"net/http"

	"github.com/portworx/kvdb"
)

// configManager implements ConfigManager
type configManager struct {
	// wrap a handle to kvdb
	cc kvdb.Kvdb

	// hashmap for callback bookkeeping
	cb map[string]*callbackData

	// execution status
	status map[string]*Status

	// timestamp of last update
	lu *lastUpdate

	// trigger between kvdb callback and listener
	trigger chan struct{}

	// heap for storing job queue
	jobs *callbackHeap

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

// last update stores timestamp of last update based on a key
// and has a mutex to lock
type lastUpdate struct {
	Time map[string]int64
	sync.Mutex
}

// osdconfigError is for declaring error strings
type osdconfigError string

// Error returns string representation to satisfy Error interface
func (err osdconfigError) Error() string {
	return string(err)
}

// Band is a classifier for registering function
type Band string

// Status stores status of execution
type Status struct {
	Err      error
	Duration time.Duration
}

// dataToKvdb is data to be sent to kvdb as a state to run on
type dataToKvdb struct {
	ctx     context.Context
	Type    Band
	trigger chan struct{}
	cbh     *callbackHeap
	sync.Mutex
}

// dataWrite is data to be sent to callbacks
// The contents here are populated based on what is received from kvdb
// Callback sends an instance of this on a channel that others can only write on
type dataWrite struct {
	// time.UninxNano() timestamp
	Time int64

	// kvdb key received in kvdb.KvPair
	Key string

	// kvdb byte buffer received in kvdb.KvPair
	Value []byte

	// Type
	Type Band

	// kvdb error received in callback executed by kvdb
	Err error
}

// dataRead is data to be received from callback at callback completion
// Callback sends an instance of this on a channel that can others can only read from
type dataRead struct {
	// name of the callback
	Name string

	// error during callback execution
	Err error
}

// callbackData is callback metadata required for callback management
type callbackData struct {
	// functional literal that is registered
	f func(ctx context.Context, opt interface{}) (chan<- *dataWrite, <-chan *dataRead)

	// value to be passed to the function during execution
	opt interface{}

	// type of watcher
	Type Band
}

// callbackHeap is a heap for storing data from kvdb callback executions
type callbackHeap struct {
	jobHeap
	sync.Mutex
}

// RESTEndPoint is a struct containing REST endpoint details
type Routes struct {
	Method string
	Path   string
	Fn     func(w http.ResponseWriter, r *http.Request)
}
