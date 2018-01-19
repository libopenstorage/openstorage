package main

import (
	"encoding/json"
	"os"
	"testing"

	"net"
	"time"

	"github.com/sdeoras/openstorage/osdconfig"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	GRPC_ADDR = "127.0.0.1:7555"
)

type MyGrpcObj struct {
	conn *grpc.ClientConn
}

func (m *MyGrpcObj) Handler() *grpc.ClientConn {
	return m.conn
}

// implement a grpc server first
type server struct{}

func (s *server) Get(ctx context.Context, in *proto.Empty) (*proto.Config, error) {
	file, err := os.Open(ConfigFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	client := osdconfig.NewIOConnection(&MyIOObj{file})
	return client.Get(context.Background(), &proto.Empty{})
}
func (s *server) Set(ctx context.Context, in *proto.Config) (*proto.Ack, error) {
	file, err := os.OpenFile(ConfigFile, os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	client := osdconfig.NewIOConnection(&MyIOObj{file})
	return client.Set(context.Background(), in)
}

func TestGrpc(t *testing.T) {
	config := new(proto.Config)
	config.Description = "this is description text"
	config.Global = new(proto.GlobalConfig)
	config.Global.AlertingUrl = "this is alerting url"

	//start grpc server on localhost
	lis, err := net.Listen("tcp", GRPC_ADDR)
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	proto.RegisterClusterSpecServer(s, &server{})
	reflection.Register(s)
	cerr := make(chan error)
	go func(c chan error) {
		c <- s.Serve(lis)
	}(cerr)

	select {
	case err := <-cerr:
		t.Fatal(err)
	case <-time.After(time.Second * 2): // wait 2 seconds for grpc server to kick in
		t.Log("grpc server probably up and running")
	}

	//dial to grpc server
	conn, err := grpc.Dial(GRPC_ADDR, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func(c chan struct{}) {
		client := osdconfig.NewGrpcConnection(&MyGrpcObj{conn})

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

		client := osdconfig.NewGrpcConnection(&MyGrpcObj{conn})
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
