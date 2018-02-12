package osdconfig

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// GetHTTPFunc wraps setters/getters of ConfigManager into http handler function.
func (manager *configManager) GetHTTPFunc(state interface{}, fn interface{}) (func(w http.ResponseWriter, r *http.Request), error) {
	switch v := fn.(type) {
	case func() (*ClusterConfig, error):
		f := func(w http.ResponseWriter, r *http.Request) {
			if config, err := v(); err != nil {
				http.Error(w, err.Error(), 0)
				return
			} else {
				dec := json.NewEncoder(w)
				if err := dec.Encode(config); err != nil {
					http.Error(w, err.Error(), 0)
				}
			}
		}
		return f, nil
	case func(id string) (*NodeConfig, error):
		f := func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			idTag, ok := state.(string)
			if !ok {
				http.Error(w, "invalid state input, need string", 0)
				return
			}
			id := string(vars[idTag])
			if config, err := v(id); err != nil {
				http.Error(w, err.Error(), 0)
				return
			} else {
				dec := json.NewEncoder(w)
				if err := dec.Encode(config); err != nil {
					http.Error(w, err.Error(), 0)
					return
				}
			}
		}
		return f, nil
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
