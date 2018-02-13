package osdconfig

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// httpFunc wraps setters/getters of ConfigManager into http handler func
func (manager *configManager) httpFunc(fn interface{}) (func(w http.ResponseWriter, r *http.Request), error) {
	switch v := fn.(type) {
	// get cluster config
	case func() (*ClusterConfig, error):
		f := func(w http.ResponseWriter, r *http.Request) {
			if config, err := v(); err != nil {
				http.Error(w, err.Error(), 0)
				return
			} else {
				jb, err := json.Marshal(config)
				if err != nil {
					http.Error(w, err.Error(), 0)
				}
				w.Write(jb)
			}
		}
		return f, nil

	// get node config
	case func(id string) (*NodeConfig, error):
		f := func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			id := string(vars[Id])
			if config, err := v(id); err != nil {
				http.Error(w, err.Error(), 0)
				return
			} else {
				jb, err := json.Marshal(config)
				if err != nil {
					http.Error(w, err.Error(), 0)
				}
				w.Write(jb)
			}
		}
		return f, nil

	// set cluster config
	case func(config *ClusterConfig) error:
		f := func(w http.ResponseWriter, r *http.Request) {
			config := new(ClusterConfig)
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), 0)
				return
			}
			if err := json.Unmarshal(b, config); err != nil {
				http.Error(w, err.Error(), 0)
				return
			}
			if err := v(config); err != nil {
				http.Error(w, err.Error(), 0)
				return
			}
		}
		return f, nil

	// set node config
	case func(config *NodeConfig) error:
		f := func(w http.ResponseWriter, r *http.Request) {
			config := new(NodeConfig)
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), 0)
				return
			}
			if err := json.Unmarshal(b, config); err != nil {
				http.Error(w, err.Error(), 0)
				return
			}
			if err := v(config); err != nil {
				http.Error(w, err.Error(), 0)
				return
			}
		}
		return f, nil
	default:
		return nil, errors.New("Unsupported input type")
	}
}
