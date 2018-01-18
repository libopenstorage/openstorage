package osdconfig

import (
	"os"

	"github.com/sdeoras/openstorage/osdconfig/client"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
)

const (
	ConfigFile = "/tmp/config.pb"
)

func Get() (*proto.Config, error) {
	file, err := os.Open(ConfigFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if pcc, err := client.New(file); err != nil {
		return nil, err
	} else {
		return pcc.Get(context.Background(), &proto.Empty{})
	}
}

func Set(config *proto.Config) (*proto.Ack, error) {
	file, err := os.OpenFile(ConfigFile, os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if pcc, err := client.New(file); err != nil {
		return nil, err
	} else {
		return pcc.Set(context.Background(), config)
	}
}
