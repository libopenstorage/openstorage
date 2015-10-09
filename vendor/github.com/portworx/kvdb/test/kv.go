package test

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/portworx/kvdb"
	"github.com/stretchr/testify/assert"
)

type watchData struct {
	t            *testing.T
	key          string
	stop         string
	localIndex   uint64
	updateIndex  uint64
	kvp          *kvdb.KVPair
	watchStopped bool
	iterations   int
	action       kvdb.KVAction
	writer       int32
	reader       int32
}

func Run(t *testing.T) {
	get(t)
	getInterface(t)
	create(t)
	update(t)
	deleteKey(t)
	deleteTree(t)
	enumerate(t)
	lock(t)
	watchKey(t)
	watchTree(t)
	cas(t)
}

func get(t *testing.T) {
	fmt.Println("get")

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	kvPair, err := kv.Get("DEADCAFE")
	assert.Error(t, err, "Expecting error value for non-existent value")

	key := "foo/docker"
	val := "great"
	defer func() {
		kv.Delete(key)
	}()

	kvPair, err = kv.Put(key, []byte(val), 10)
	assert.NoError(t, err, "Unxpected error in Put")

	kvPair, err = kv.Get(key)
	assert.NoError(t, err, "Failed in Get")

	assert.Equal(t, kvPair.Key, key, "Key mismatch in Get")
	assert.Equal(t, string(kvPair.Value), val, "value mismatch in Get")
}

func getInterface(t *testing.T) {

	fmt.Println("getInterface")
	expected := struct {
		N int
		S string
	}{
		N: 10,
		S: "Ten",
	}

	actual := expected
	actual.N = 0
	actual.S = "zero"

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	key := "DEADBEEF"
	_, err := kv.Delete(key)
	_, err = kv.Put(key, &expected, 0)
	assert.NoError(t, err, "Failed in Put")

	_, err = kv.GetVal(key, &actual)
	assert.NoError(t, err, "Failed in Get")

	assert.Equal(t, expected, actual, "Expected %#v but got %#v",
		expected, actual)
}

func create(t *testing.T) {
	fmt.Println("create")

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	key := "create/foo"
	kv.Delete(key)

	kvp, err := kv.Create(key, []byte("bar"), 0)
	assert.NoError(t, err, "Error on create")

	defer func() {
		kv.Delete(key)
	}()
	assert.Equal(t, kvp.Action, kvdb.KVCreate,
		"Expected action KVCreate, actual %v", kvp.Action)

	_, err = kv.Create(key, []byte("bar"), 0)
	assert.Error(t, err, "Create on existing key should have errored.")
}

func update(t *testing.T) {
	fmt.Println("update")

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	key := "update/foo"
	kv.Delete(key)

	kvp, err := kv.Update(key, []byte("bar"), 0)
	assert.Error(t, err, "Update should error on non-existent key")

	defer func() {
		kv.Delete(key)
	}()

	kvp, err = kv.Create(key, []byte("bar"), 0)
	assert.NoError(t, err, "Unexpected error on create")

	kvp, err = kv.Update(key, []byte("bar"), 0)
	assert.NoError(t, err, "Unexpected error on update")

	assert.Equal(t, kvp.Action, kvdb.KVSet,
		"Expected action KVSet, actual %v", kvp.Action)
}

func deleteKey(t *testing.T) {
	fmt.Println("deleteKey")

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	key := "delete_key"
	_, err := kv.Delete(key)

	_, err = kv.Put(key, []byte("delete_me"), 10)
	assert.NoError(t, err, "Unexpected error on Put")

	_, err = kv.Get(key)
	assert.NoError(t, err, "Unexpected error on Get")

	_, err = kv.Delete(key)
	assert.NoError(t, err, "Unexpected error on Delete")

	_, err = kv.Get(key)
	assert.Error(t, err, "Get should fail on deleted key")

	_, err = kv.Delete(key)
	assert.Error(t, err, "Delete should fail on non existent key")
}

func deleteTree(t *testing.T) {
	fmt.Println("deleteTree")

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	prefix := "tree"
	keys := map[string]string{
		prefix + "/1cbc9a98-072a-4793-8608-01ab43db96c8": "bar",
		prefix + "/foo":                                  "baz",
	}

	for key, val := range keys {
		_, err := kv.Put(key, []byte(val), 10)
		assert.NoError(t, err, "Unexpected error on Put")
	}

	for key, _ := range keys {
		_, err := kv.Get(key)
		assert.NoError(t, err, "Unexpected error on Get")
	}
	err := kv.DeleteTree(prefix)
	assert.NoError(t, err, "Unexpected error on DeleteTree")

	for key, _ := range keys {
		_, err := kv.Get(key)
		assert.Error(t, err, "Get should fail on all keys after DeleteTree")
	}
}

func enumerate(t *testing.T) {

	fmt.Println("enumerate")

	prefix := "enumerate"
	keys := map[string]string{
		prefix + "/1cbc9a98-072a-4793-8608-01ab43db96c8": "bar",
		prefix + "/foo":                                  "baz",
	}

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	kv.DeleteTree(prefix)
	defer func() {
		kv.DeleteTree(prefix)
	}()

	for key, val := range keys {
		_, err := kv.Put(key, []byte(val), 10)
		assert.NoError(t, err, "Unexpected error on Put")
	}
	kvPairs, err := kv.Enumerate(prefix)
	assert.NoError(t, err, "Unexpected error on Enumerate")

	assert.Equal(t, len(kvPairs), len(keys),
		"Expecting %d keys under %s got: %d",
		len(keys), prefix, len(kvPairs))

	for i := range kvPairs {
		v, ok := keys[kvPairs[i].Key]
		assert.True(t, ok, "unexpected kvpair (%s)->(%s)",
			kvPairs[i].Key, kvPairs[i].Value)
		assert.Equal(t, v, string(kvPairs[i].Value),
			"Invalid kvpair (%s)->(%s) expect value %s",
			kvPairs[i].Key, kvPairs[i].Value, v)
	}
}

func lock(t *testing.T) {

	fmt.Println("lock")

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	key := "locktest"
	kvPair, err := kv.Lock(key, 100)
	assert.NoError(t, err, "Unexpected error in lock")

	stash := *kvPair
	stash.Value = []byte("hoohah")
	err = kv.Unlock(&stash)
	assert.Error(t, err, "Unlock should fail for bad KVPair")

	err = kv.Unlock(kvPair)
	assert.NoError(t, err, "Unexpected error from Unlock")

	kvPair, err = kv.Lock(key, 20)
	assert.NoError(t, err, "Failed to lock after unlock")

	err = kv.Unlock(kvPair)
	assert.NoError(t, err, "Unexpected error from Unlock")
}

func watchFn(prefix string,
	opaque interface{},
	kvp *kvdb.KVPair,
	err error) error {
	data := opaque.(*watchData)

	atomic.AddInt32(&data.reader, 1)
	if err != nil {
		assert.Equal(data.t, err, kvdb.ErrWatchStopped)
		data.watchStopped = true
		return err

	}
	fmt.Printf("+")

	// Doesn't work for ETCD because HTTP header contains Etcd-Index
	/*
		assert.True(data.t, kvp.KVDBIndex >= data.updateIndex,
			"KVDBIndex %v must be >= than updateIndex %v",
			kvp.KVDBIndex, data.updateIndex)
	*/

	assert.True(data.t, kvp.KVDBIndex > data.localIndex,
		"KVDBIndex %v must be > than localIndex %v",
		kvp.KVDBIndex, data.updateIndex)

	assert.True(data.t, kvp.ModifiedIndex > data.localIndex,
		"ModifiedIndex %v must be > than localIndex %v",
		kvp.ModifiedIndex, data.localIndex)

	data.localIndex = kvp.KVDBIndex

	assert.Equal(data.t, kvp.Key, data.key,
		"Bad kvpair key %s expecting %s",
		kvp.Key, data.key)

	assert.Equal(data.t, kvp.Action, data.action,
		"Expected action %v actual %v",
		data.action, kvp.Action)

	if string(kvp.Value) == data.stop {
		return errors.New(data.stop)
	}

	return nil
}

func watchUpdate(data *watchData) error {
	var err error
	var kvp *kvdb.KVPair

	kv := kvdb.Instance()
	assert.NotNil(data.t, kv, "Default KVDB is not set")

	data.reader, data.writer = 0, 0
	atomic.AddInt32(&data.writer, 1)
	data.action = kvdb.KVCreate
	kvp, err = kv.Create(data.key, []byte("bar"), 10)
	for i := 0; i < data.iterations && err == nil; i++ {
		fmt.Printf("-")

		for data.writer != data.reader {
			time.Sleep(time.Millisecond * 100)
		}
		atomic.AddInt32(&data.writer, 1)
		data.action = kvdb.KVSet
		kvp, err = kv.Put(data.key, []byte("bar"), 10)

		data.updateIndex = kvp.KVDBIndex
		assert.NoError(data.t, err, "Unexpected error in Put")
	}

	for data.writer != data.reader {
		time.Sleep(time.Millisecond * 100)
	}
	atomic.AddInt32(&data.writer, 1)
	data.action = kvdb.KVDelete
	kv.Delete(data.key)

	for data.writer != data.reader {
		time.Sleep(time.Millisecond * 100)
	}
	atomic.AddInt32(&data.writer, 1)
	data.action = kvdb.KVCreate
	_, err = kv.Create(data.key, []byte(data.stop), 0)
	return err
}

func watchKey(t *testing.T) {
	fmt.Println("watchKey")

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	watchData := watchData{
		t:          t,
		key:        "tree/key",
		stop:       "stop",
		iterations: 2,
	}
	kv.Delete(watchData.key)
	kv.WatchKey(watchData.key, 0, &watchData, watchFn)
	go watchUpdate(&watchData)

	for watchData.watchStopped == false {
		time.Sleep(time.Millisecond * 100)
	}
}

func watchTree(t *testing.T) {
	fmt.Println("watchTree")

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")

	tree := "tree"

	watchData := watchData{
		t:          t,
		key:        tree + "/key",
		stop:       "stop",
		iterations: 2,
	}
	kv.Delete(watchData.key)
	time.Sleep(time.Second)
	kv.WatchTree(tree, 0, &watchData, watchFn)
	go watchUpdate(&watchData)

	for watchData.watchStopped == false {
		time.Sleep(time.Millisecond * 100)
	}
}

func cas(t *testing.T) {
	fmt.Println("cas")

	kv := kvdb.Instance()
	assert.NotNil(t, kv, "Default KVDB is not set")
	key := "foo/docker"
	val := "great"
	defer func() {
		kv.Delete(key)
	}()

	kvPair, err := kv.Put(key, []byte(val), 10)
	assert.NoError(t, err, "Unxpected error in Put")

	kvPair, err = kv.Get(key)
	assert.NoError(t, err, "Failed in Get")

	_, err = kv.CompareAndSet(kvPair, kvdb.KVFlags(0), []byte("badval"))
	assert.Error(t, err, "CompareAndSet should fail on an incorrect previous value")

	kvPair.ModifiedIndex++
	_, err = kv.CompareAndSet(kvPair, kvdb.KVModifiedIndex, nil)
	assert.Error(t, err, "CompareAndSet should fail on an incorrect modified index")

	kvPair.ModifiedIndex--
	kvPair, err = kv.CompareAndSet(kvPair, kvdb.KVModifiedIndex, nil)
	assert.NoError(t, err, "CompareAndSet should succeed on an correct modified index")

	kvPair, err = kv.CompareAndSet(kvPair, kvdb.KVFlags(0), []byte(val))
	assert.NoError(t, err, "CompareAndSet should succeed on an correct value")

	kvPair, err = kv.CompareAndSet(kvPair, kvdb.KVModifiedIndex, []byte(val))
	assert.NoError(t, err, "CompareAndSet should succeed on an correct value and modified index")
}
