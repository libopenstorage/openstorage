package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"bufio"
	"bytes"
	"io/ioutil"

	"github.com/sdeoras/openstorage/osdconfig"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
)

const (
	ConfigFile = "/tmp/config.pb"
)

func TestFileIO(t *testing.T) {
	globalConf := proto.NewGlobalConfig()
	globalConf.ClusterConf.Description = "this is description text"
	globalConf.ClusterConf.AlertingUrl = "this is alerting url"
	nodeConf := proto.NewNodeConfig()
	nodeConf.NodeId = "Node1"
	if err := globalConf.SetNode(nodeConf); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	done := make(chan error)
	go func(c chan error) {
		if err := ioutil.WriteFile(ConfigFile, []byte{}, os.ModeAppend); err != nil {
			c <- err
		}

		// read from file and create a new reader
		bf, err := ioutil.ReadFile(ConfigFile)
		if err != nil {
			c <- err
		}
		br := bufio.NewReader(bytes.NewReader(bf))

		// create a new writer to bytes
		var bb bytes.Buffer
		bw := bufio.NewWriter(&bb)

		// create a new read writer
		brw := bufio.NewReadWriter(br, bw)

		// get a new client connection to osdconfig library using read writer
		client := osdconfig.NewIOConnection(brw)

		ack, err := client.SetGlobalSpec(ctx, globalConf)
		if err != nil {
			c <- err
		}

		ack, err = client.SetClusterSpec(ctx, globalConf.ClusterConf)
		if err != nil {
			c <- err
		}

		if ack != nil {
			t.Log("Bytes written:", ack.N)
		}

		if err := brw.Flush(); err != nil {
			c <- err
		}
		if err := ioutil.WriteFile(ConfigFile, bb.Bytes(), os.ModeAppend); err != nil {
			c <- err
		}

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
		// read from file and create a new reader
		bf, err := ioutil.ReadFile(ConfigFile)
		if err != nil {
			c <- err
		}
		br := bufio.NewReader(bytes.NewReader(bf))

		// create a new writer to bytes
		var bb bytes.Buffer
		bw := bufio.NewWriter(&bb)

		// create a new read writer
		brw := bufio.NewReadWriter(br, bw)

		// get a new client connection to osdconfig library using read writer
		client := osdconfig.NewIOConnection(brw)

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
