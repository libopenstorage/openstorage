package server

import (
	"encoding/json"
	"net/http"

	"github.com/sdeoras/openstorage/pxconfig"
)

type pxconfigAPI struct {
	restBase
}

func (p *pxconfigAPI) Routes() []*Route {
	return []*Route{
		{verb: "PUT", path: "/{buffer}", fn: p.set},
		{verb: "GET", path: "", fn: p.get},
	}
}

func (p *pxconfigAPI) set(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("")
}

func (p *pxconfigAPI) get(w http.ResponseWriter, r *http.Request) {
	if config, err := pxconfig.Get(); err != nil {
		// todo
	} else {
		json.NewEncoder(w).Encode(config)
	}
}
