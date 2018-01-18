package client

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/sdeoras/openstorage/pxconfig/proto"
	"golang.org/x/net/context"
)

const (
	ConfigFile = "/tmp/config.pb"
)

func TestNewClient(t *testing.T) {
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

		client, err := New(file)
		if err != nil {
			t.Fatal(err)
		}

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

		client := proto.NewClusterSpecClientIO(file)
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
