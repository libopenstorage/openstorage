package osdconfig

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

// Registers a few callbacks and executes them concurrently
func TestExecCB(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// create in memory kvdb
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	// get new config manager using handle to kvdb
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	manager, err := newManager(ctx, kv)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// register a few callback functions
	names := []string{"f0", "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9"}
	for i, name := range names {
		name := name
		if err := manager.register(name, TuneCluster, i, newCallback(name, 0, 1000)); err != nil {
			t.Fatal(err)
		}
	}

	// execute callbacks manually
	wd := new(dataWrite)
	wd.Type = TuneCluster // only trigger cluster watcher callbacks
	manager.run(wd)

	time.Sleep(time.Second * 2)
	manager.printStatus()
}
