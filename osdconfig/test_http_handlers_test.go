package osdconfig

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

// TestHttpHandlers will start a http server and query against it
func TestHttpHandlers(t *testing.T) {
	// create in memory kvdb
	kv, err := kvdb.New(mem.Name, "", []string{}, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	// get new config manager using handle to kvdb
	ctx := context.Background()

	manager, err := NewManager(ctx, kv)
	if err != nil {
		log.Fatal(err)
	}
	defer manager.Close()

	// start http service
	r := mux.NewRouter()
	routes := manager.GetRoutes()
	for _, route := range routes {
		r.HandleFunc(route.Path, route.Fn).Methods(route.Method)
	}

	done := make(chan error)
	http.Handle("/", r)
	go func(c chan error) {
		c <- http.ListenAndServe(":8085", nil)
	}(done)

	url := "http://localhost:8085"

	conf := new(ClusterConfig)
	conf.ClusterId = "cluster_id_1"
	conf.AlertingUrl = "altering.url"

	jb, err := json.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PUT", url+"/cluster", bytes.NewBuffer(jb))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	// prepare expected cluster config
	nodeConf := new(NodeConfig)
	nodeConf.NodeId = "node1"
	nodeConf.Storage = new(StorageConfig)
	nodeConf.Storage.Devices = []string{"dev1", "dev2"}

	jb, err = json.Marshal(nodeConf)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("PUT", url+"/node", bytes.NewBuffer(jb))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	default:
	}

	// check for cluster value
	resp, err = http.Get(url + "/cluster")
	if err != nil {
		t.Fatal(err)
	}

	conf2 := new(ClusterConfig)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(b, conf2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(conf, conf2) {
		t.Fatal("did not get expected http response")
	}

	// now check for node
	resp, err = http.Get(url + "/node/node1")
	if err != nil {
		t.Fatal(err)
	}

	nodeConf2 := new(NodeConfig)
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(b, nodeConf2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(nodeConf, nodeConf2) {
		t.Fatal("did not get expected http response")
	}
}
