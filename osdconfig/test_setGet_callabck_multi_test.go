package osdconfig

import (
	"context"
	"testing"
	"time"
)

func TestSetGetCallbackMulti(t *testing.T) {
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

	// register a few callback functions
	names := []string{"f0", "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9"}
	for i, name := range names {
		name := name
		if err := manager.register(name, ClusterWatcher, i, newCallback(name, 0, 5000)); err != nil {
			t.Fatal(err)
		}
	}

	// prepare expected cluster config
	expectedConf := new(ClusterConfig)
	expectedConf.ClusterId = "myClusterID"
	expectedConf.Driver = "myDriver"

	// set the expected cluster config value
	if err := manager.SetClusterConf(expectedConf); err != nil {
		t.Fatal(err)
	}

Loop1:
	for i := 0; i < 10; i++ {
		t := time.Now()
		select {
		case <-time.After(time.Millisecond * 100):
			manager.wait()
			if time.Since(t) > time.Second { // done waiting for callback execution
				break Loop1
			}
		}
	}

	// set the expected cluster config value
	if err := manager.SetClusterConf(expectedConf); err != nil {
		t.Fatal(err)
	}

Loop2:
	for i := 0; i < 10; i++ {
		t := time.Now()
		select {
		case <-time.After(time.Millisecond * 100):
			manager.wait()
			if time.Since(t) > time.Second { // done waiting for callback execution
				break Loop2
			}
		}
	}
}
