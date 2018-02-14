package osdconfig

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

// Check timeout functionality in executing callbacks
func TestExecCBAbort(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// create in memory kvdb
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	// get new config manager using handle to kvdb
	ctx := context.Background()
	manager, err := newManager(ctx, kv)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// register a few callback functions that sleep for less than 3 seconds
	names := []string{"f0", "f1", "f2", "f3", "f4"}
	for i, name := range names {
		name := name
		if err := manager.register(name, TuneCluster, i, newCallback(name, 0, 3000)); err != nil {
			t.Fatal(err)
		}
	}
	// register a few callback functions that sleep for more than 5 seconds
	names = []string{"f5", "f6", "f7", "f8", "f9"}
	for i, name := range names {
		name := name
		if err := manager.register(name, TuneCluster, i, newCallback(name, 5000, 8000)); err != nil {
			t.Fatal(err)
		}
	}

	t0 := time.Now()

	// execute callbacks manually
	wd := new(dataWrite)
	wd.Type = TuneCluster // only trigger cluster watcher callbacks
	manager.run(wd)
	// abort after 3 second
	go func() { time.Sleep(time.Second * 3); manager.abort() }()
	// wait for execution completion
	time.Sleep(time.Second)
	manager.wait()

	if time.Since(t0) > time.Second*4 {
		t.Fatal("3 second abort did not work")
	}
}
