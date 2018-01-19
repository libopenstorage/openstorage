package main

import (
	"encoding/json"
	"os"
	"testing"

	"time"

	"github.com/portworx/kvdb"
	"github.com/sdeoras/openstorage/osdconfig"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
)

type MyKVObj struct {
	kv kvdb.Kvdb
}

func (m *MyKVObj) Handler() kvdb.Kvdb {
	return m.kv
}

func TestKV(t *testing.T) {
	config := new(proto.ClusterConfig)
	config.Description = "this is description text"
	config.AlertingUrl = "this is alerting url"

	options := make(map[string]string)
	options["KvUseInterface"] = ""
	kv, err := kvdb.New("pwx/test", "", nil, options, nil)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error)
	go func(c chan error) {
		client := osdconfig.NewKVDBConnection(&MyKVObj{kv})

		ack, err := client.SetClusterSpec(context.Background(), config)
		if err != nil {
			c <- err
		}

		t.Log("Bytes written:", ack.N)

		c <- nil
	}(done)

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second * 2): // wait 2 seconds for grpc server to kick in
		t.Log("grpc server probably up and running")
	}

	go func(c chan error) {
		file, err := os.Open(ConfigFile)
		if err != nil {
			c <- err
		}
		defer file.Close()

		client := osdconfig.NewKVDBConnection(&MyKVObj{kv})
		config, err := client.GetClusterSpec(context.Background(), &proto.Empty{})
		if err != nil {
			c <- err
		}

		jb, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			c <- err
		}

		t.Log(string(jb))
		c <- nil
	}(done)

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second * 2): // wait 2 seconds for grpc server to kick in
		t.Log("grpc server probably up and running")
	}
}
