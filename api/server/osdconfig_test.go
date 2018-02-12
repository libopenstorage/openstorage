package server

import (
	"testing"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/osdconfig"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

func TestOsdConfigAPI(t *testing.T) {
	// create in memory kvdb
	kv, err := kvdb.New(mem.Name, "", []string{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// get new config manager using handle to kvdb
	ctx := context.Background()

	// start http service first
	api, err := newOsdconfigAPI("osdconfig", "v1", ctx, kv)
	if err != nil {
		t.Fatal(err)
	}
	defer api.close()

	// start http service
	r := mux.NewRouter()
	r.HandleFunc(filepath.Join("/", cluster), api.getHTTPFunc(filepath.Join(get, cluster))).Methods("GET")
	r.HandleFunc(filepath.Join("/", node, "{"+id+"}"), api.getHTTPFunc(filepath.Join(get, node))).Methods("GET")
	r.HandleFunc(filepath.Join("/", cluster), api.getHTTPFunc(filepath.Join(post, cluster))).Methods("POST")
	r.HandleFunc(filepath.Join("/", node), api.getHTTPFunc(filepath.Join(post, node))).Methods("POST")

	http.Handle("/", r)
	go http.ListenAndServe(":8080", nil)

	conf := new(osdconfig.ClusterConfig)
	conf.ClusterId = "cluster_id_1"
	conf.AlertingUrl = "altering.url"
	jb, err := json.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}

	url := "localhost:/cluster"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jb))
	req.Header.Set("X-Custom-Header", "myClusterInfo")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	re, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(re.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))
}
