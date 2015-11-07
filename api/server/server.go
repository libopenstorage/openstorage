package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// Route is a specification and  handler for a REST endpoint.
type Route struct {
	verb string
	path string
	fn   func(http.ResponseWriter, *http.Request)
}

type restServer interface {
	Routes() []*Route
	String() string
	logReq(request string, id string) *log.Entry
	sendError(request string, id string, w http.ResponseWriter, msg string, code int)
}

type restBase struct {
	restServer
	version string
	name    string
}

func (rest *restBase) logReq(request string, id string) *log.Entry {
	return log.WithFields(log.Fields{
		"Driver":  rest.name,
		"Request": request,
		"ID":      id,
	})
}
func (rest *restBase) sendError(request string, id string, w http.ResponseWriter, msg string, code int) {
	rest.logReq(request, id).Warn(code, " ", msg)
	http.Error(w, msg, code)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	log.Warnf("Not found: %+v ", r.URL)
	http.NotFound(w, r)
}

func startServer(name string, sockBase string, port int, routes []*Route) error {
	var (
		listener net.Listener
		err      error
	)
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFound)

	for _, v := range routes {
		router.Methods(v.verb).Path(v.path).HandlerFunc(v.fn)
	}
	socket := path.Join(sockBase, name+".sock")
	os.Remove(socket)
	os.MkdirAll(path.Dir(socket), 0755)

	log.Printf("Starting REST service on %+v", socket)
	listener, err = net.Listen("unix", socket)
	if err != nil {
		log.Warn("Cannot listen on UNIX socket: ", err)
		return err
	}
	go http.Serve(listener, router)
	if port != 0 {
		go http.ListenAndServe(fmt.Sprintf(":%v", port), router)
	}
	return nil
}

// StartGraphAPI starts a REST server to receive GraphDriver commands
func StartGraphAPI(name string, port int, restBase string) error {
	graphPlugin := newGraphPlugin(name)
	routes := append(graphPlugin.Routes())
	return startServer(name, restBase, port, routes)
}

// StartServerAPI starts a REST server to receive driver configuration commands
// from the CLI/UX.
func StartServerAPI(name string, port int, restBase string) error {
	volApi := newVolumeAPI(name)
	clusterApi := newClusterAPI(name)
	routes := append(volApi.Routes(), clusterApi.Routes()...)
	return startServer(name, restBase, port, routes)
}

// StartPluginAPI starts a REST server to receive volume commands from the
// Linux container engine.
func StartPluginAPI(name string, pluginBase string) error {
	rest := newVolumePlugin(name)
	return startServer(name, pluginBase, 0, rest.Routes())
}
