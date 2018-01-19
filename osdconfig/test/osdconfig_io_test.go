package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/sdeoras/openstorage/osdconfig"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
)

const (
	ConfigFile = "/tmp/config.pb"
)

func TestFileIO(t *testing.T) {
	config := new(proto.ClusterConfig)
	config.Description = "this is description text"
	config.AlertingUrl = "this is alerting url"

	done := make(chan error)
	go func(c chan error) {
		if err := ioutil.WriteFile(ConfigFile, []byte{}, 0644); err != nil {
			c <- err
		}

		file, err := os.OpenFile(ConfigFile, os.O_WRONLY, 0644)
		if err != nil {
			c <- err
		}
		defer file.Close()

		client := osdconfig.NewIOConnection(file)
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
	case <-time.After(time.Second * 5):
		t.Fatal("test 5 second timeout")
	}

	go func(c chan error) {
		file, err := os.Open(ConfigFile)
		if err != nil {
			c <- err
		}
		defer file.Close()

		client := osdconfig.NewIOConnection(file)
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
	case <-time.After(time.Second * 5):
		t.Fatal("test 5 second timeout")
	}
}
