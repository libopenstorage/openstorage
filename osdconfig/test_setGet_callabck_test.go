package osdconfig

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestSetGetCallback(t *testing.T) {
	// create in memory kvdb
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	// get new config manager using handle to kvdb
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	manager, err := NewManager(ctx, kv)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// register a few callback functions
	names := []string{"f0", "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9"}
	for i, name := range names {
		name := name
		if err := manager.Register(name, i, newCallback(name, 0, 5000)); err != nil {
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

	// get the cluster config value
	receivedConf, err := manager.GetClusterConf()
	if err != nil {
		t.Fatal(err)
	}

	// compare expected and received
	if !reflect.DeepEqual(expectedConf, receivedConf) {
		t.Fatal("expected and received values are not deep equal")
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
