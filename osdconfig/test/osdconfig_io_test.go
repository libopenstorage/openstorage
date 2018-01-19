package main

import (
	"encoding/json"
	"os"
	"testing"

	"io"

	"github.com/sdeoras/openstorage/osdconfig"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
)

const (
	ConfigFile = "/tmp/config.pb"
)

type MyIOObj struct {
	file *os.File
}

func (m *MyIOObj) Handler() io.ReadWriter {
	return m.file
}

func TestFileIO(t *testing.T) {
	config := new(proto.Config)
	config.Description = "this is description text"
	config.Global = new(proto.GlobalConfig)
	config.Global.AlertingUrl = "this is alerting url"

	done := make(chan struct{})
	go func(c chan struct{}) {
		file, err := os.OpenFile(ConfigFile, os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		client := osdconfig.NewIOConnection(&MyIOObj{file})
		ack, err := client.Set(context.Background(), config)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("Bytes written:", ack.N)

		c <- struct{}{}
	}(done)
	<-done

	go func(c chan struct{}) {
		file, err := os.Open(ConfigFile)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		client := osdconfig.NewIOConnection(&MyIOObj{file})
		config, err := client.Get(context.Background(), &proto.Empty{})
		if err != nil {
			t.Fatal(err)
		}

		jb, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			t.Fatal(err)
		}

		t.Log(string(jb))
		c <- struct{}{}
	}(done)
	<-done

}
