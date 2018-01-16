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

	timeout := time.Second * 2

	// get new config manager using handle to kvdb
	ctx := context.Background()
	manager, err := NewManager(ctx, kv)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// register a few callback functions
	names := []string{"f0", "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9"}
	for i, name := range names {
		name := name
		if err := manager.Register(name, i, newCallback(name, 1000, 5000)); err != nil {
			t.Fatal(err)
		}
	}

	// execute callbacks manually
	manager.Run(new(DataToCallback))

	t0 := time.Now()
	select {
	case <-manager.GetContext().Done():
		if time.Since(t0) > timeout {
			t.Fatal("timeout did not occur")
		}
	case <-time.After(time.Second * 3):
		manager.Abort() // send abort signal
	}

Loop1:
	for i := 0; i < 10; i++ {
		t := time.Now()
		select {
		case <-time.After(time.Millisecond * 100):
			manager.Wait()
			if time.Since(t) > time.Second { // done waiting for callback execution
				break Loop1
			}
		}
	}
}
