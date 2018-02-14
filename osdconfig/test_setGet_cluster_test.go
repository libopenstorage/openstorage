package osdconfig

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestSetGetCluster(t *testing.T) {
	// create in memory kvdb
	kv, err := newInMemKvdb()
	if err != nil {
		t.Fatal(err)
	}

	// get new config manager using handle to kvdb
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	manager, err := NewManager(ctx, kv)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

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
}
