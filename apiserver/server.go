package apiserver

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
	notFound(w http.ResponseWriter, r *http.Request)
	volNotFound(request string, id string, e error, w http.ResponseWriter) error
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

func (rest *restBase) notFound(w http.ResponseWriter, r *http.Request) {
	log.Warnf("[%s] Not found: %+v ", rest.name, r.URL)
	http.NotFound(w, r)
}

func (rest *restBase) volNotFound(request string, id string, e error, w http.ResponseWriter) error {
	err := fmt.Errorf("Failed to locate volume: " + e.Error())
	rest.logReq(request, id).Warn(http.StatusNotFound, " ", err.Error())
	return err
}

func startServer(name string, sockBase string, port int, rest restServer) error {

	var (
		listener net.Listener
		err      error
	)
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(rest.notFound)
	routes := rest.Routes()

	for _, v := range routes {
		router.Methods(v.verb).Path(v.path).HandlerFunc(v.fn)
	}
	socket := path.Join(sockBase, name+".sock")
	os.Remove(socket)
	os.MkdirAll(path.Dir(socket), 0755)

	log.Printf("Starting REST service on %+v", socket)
	listener, err = net.Listen("unix", socket)
	if err != nil {
		return err
	}
	go http.Serve(listener, router)
	if port != 0 {
		go http.ListenAndServe(fmt.Sprintf(":%v", port), router)
	}
	return err
}

// StartDriverAPI starts a REST server to receive driver configuration commands
// from the CLI/UX.
func StartDriverAPI(name string, port int, restBase string) error {
	rest := newVolumeDriver(name)
	return startServer(name, restBase, port, rest)
}

// StartPluginAPI starts a REST server to receive volume commands from the
// Linux container engine.
func StartPluginAPI(name string, pluginBase string) error {
	rest := newVolumePlugin(name)
	return startServer(name, pluginBase, 0, rest)
}
