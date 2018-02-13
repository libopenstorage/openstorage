package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/osdconfig"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

const (
	get     = "get"
	post    = "post"
	id      = "id"
	cluster = "cluster"
	node    = "node"
)

func main() {
	// create in memory kvdb
	kv, err := kvdb.New(mem.Name, "", []string{}, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	// get new config manager using handle to kvdb
	ctx := context.Background()

	manager, err := osdconfig.NewManager(ctx, kv)
	if err != nil {
		log.Fatal(err)
	}
	defer manager.Close()

	conf := new(osdconfig.ClusterConfig)
	conf.ClusterId = "cluster_id_1"
	conf.AlertingUrl = "altering.url"
	manager.SetClusterConf(conf)

	// prepare expected cluster config
	nodeConf := new(osdconfig.NodeConfig)
	nodeConf.NodeId = "node1"
	nodeConf.Storage = new(osdconfig.StorageConfig)
	nodeConf.Storage.Devices = []string{"dev1", "dev2"}

	// set the expected cluster config value
	if err := manager.SetNodeConf(nodeConf); err != nil {
		log.Fatal(err)
	}

	// prepare expected cluster config
	nodeConf = new(osdconfig.NodeConfig)
	nodeConf.NodeId = "node2"
	nodeConf.Storage = new(osdconfig.StorageConfig)
	nodeConf.Storage.Devices = []string{"dev3", "dev4"}

	// set the expected cluster config value
	if err := manager.SetNodeConf(nodeConf); err != nil {
		log.Fatal(err)
	}

	// start http service
	r := mux.NewRouter()
	routes := manager.GetRoutes()
	for _, route := range routes {
		r.HandleFunc(route.Path, route.Fn).Methods(route.Method)
	}

	done := make(chan error)
	http.Handle("/", r)
	go func(c chan error) {
		c <- http.ListenAndServe(":8080", nil)
	}(done)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
	case <-done:
	}
}
