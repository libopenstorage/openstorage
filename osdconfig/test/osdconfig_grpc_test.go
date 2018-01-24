package main

import (
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

	"bufio"
	"bytes"
	"io/ioutil"

	"github.com/sdeoras/openstorage/osdconfig"
	"github.com/sdeoras/openstorage/osdconfig/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	GRPC_ADDR = "127.0.0.1:7555"
)

// implement a grpc server first
type server struct{}

func (s *server) GetGlobalSpec(ctx context.Context, in *proto.Empty) (*proto.GlobalConfig, error) {
	// read from file and create a new reader
	bf, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(bf))

	// create a new writer to bytes
	var bb bytes.Buffer
	bw := bufio.NewWriter(&bb)

	// create a new read writer
	brw := bufio.NewReadWriter(br, bw)

	// get a new client connection to osdconfig library using read writer
	client := osdconfig.NewIOConnection(brw)

	return client.GetGlobalSpec(context.Background(), &proto.Empty{})
}
func (s *server) SetGlobalSpec(ctx context.Context, in *proto.GlobalConfig) (*proto.Ack, error) {
	// read from file and create a new reader
	bf, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(bf))

	// create a new writer to bytes
	var bb bytes.Buffer
	bw := bufio.NewWriter(&bb)

	// create a new read writer
	brw := bufio.NewReadWriter(br, bw)

	// get a new client connection to osdconfig library using read writer
	client := osdconfig.NewIOConnection(brw)

	ack, err := client.SetGlobalSpec(context.Background(), in)
	if err != nil {
		return nil, err
	}

	if err := brw.Flush(); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(ConfigFile, bb.Bytes(), os.ModeAppend); err != nil {
		return nil, err
	}

	return ack, nil
}
func (s *server) GetClusterSpec(ctx context.Context, in *proto.Empty) (*proto.ClusterConfig, error) {
	// read from file and create a new reader
	bf, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(bf))

	// create a new writer to bytes
	var bb bytes.Buffer
	bw := bufio.NewWriter(&bb)

	// create a new read writer
	brw := bufio.NewReadWriter(br, bw)

	// get a new client connection to osdconfig library using read writer
	client := osdconfig.NewIOConnection(brw)

	return client.GetClusterSpec(context.Background(), &proto.Empty{})
}
func (s *server) SetClusterSpec(ctx context.Context, in *proto.ClusterConfig) (*proto.Ack, error) {
	// read from file and create a new reader
	bf, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(bf))

	// create a new writer to bytes
	var bb bytes.Buffer
	bw := bufio.NewWriter(&bb)

	// create a new read writer
	brw := bufio.NewReadWriter(br, bw)

	// get a new client connection to osdconfig library using read writer
	client := osdconfig.NewIOConnection(brw)

	ack, err := client.SetClusterSpec(context.Background(), in)
	if err != nil {
		return nil, err
	}

	if err := brw.Flush(); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(ConfigFile, bb.Bytes(), os.ModeAppend); err != nil {
		return nil, err
	}

	return ack, nil
}
func (s *server) GetNodeSpec(ctx context.Context, in *proto.NodeID) (*proto.NodeConfig, error) {
	// read from file and create a new reader
	bf, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(bf))

	// create a new writer to bytes
	var bb bytes.Buffer
	bw := bufio.NewWriter(&bb)

	// create a new read writer
	brw := bufio.NewReadWriter(br, bw)

	// get a new client connection to osdconfig library using read writer
	client := osdconfig.NewIOConnection(brw)

	return client.GetNodeSpec(context.Background(), in)
}
func (s *server) SetNodeSpec(ctx context.Context, in *proto.NodeConfig) (*proto.Ack, error) {
	// read from file and create a new reader
	bf, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(bf))

	// create a new writer to bytes
	var bb bytes.Buffer
	bw := bufio.NewWriter(&bb)

	// create a new read writer
	brw := bufio.NewReadWriter(br, bw)

	// get a new client connection to osdconfig library using read writer
	client := osdconfig.NewIOConnection(brw)

	ack, err := client.SetNodeSpec(context.Background(), in)
	if err != nil {
		return nil, err
	}

	if err := brw.Flush(); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(ConfigFile, bb.Bytes(), os.ModeAppend); err != nil {
		return nil, err
	}

	return ack, nil
}

func TestGrpc(t *testing.T) {
	config := new(proto.ClusterConfig)
	config.Description = "this is description text"
	config.AlertingUrl = "this is alerting url"

	//start grpc server on localhost
	lis, err := net.Listen("tcp", GRPC_ADDR)
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	proto.RegisterSpecServer(s, &server{})
	reflection.Register(s)
	cerr := make(chan error)
	go func(c chan error) {
		c <- s.Serve(lis)
	}(cerr)

	select {
	case err := <-cerr:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second * 2): // wait 2 seconds for grpc server to kick in
		t.Log("grpc server probably up and running")
	}

	//dial to grpc server
	conn, err := grpc.Dial(GRPC_ADDR, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error)
	go func(c chan error) {
		client := osdconfig.NewGrpcConnection(conn)

		ack, err := client.SetClusterSpec(context.Background(), config)
		if err != nil {
			c <- err
			return
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
		t.Fatal("test timeout of 5 second")
	}

	go func(c chan error) {
		file, err := os.Open(ConfigFile)
		if err != nil {
			c <- err
		}
		defer file.Close()

		client := osdconfig.NewGrpcConnection(conn)
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
		t.Fatal("test timeout of 5 second")
	}
}
