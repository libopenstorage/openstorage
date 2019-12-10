package server

import (
	"encoding/base64"
	"encoding/json"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/proto/time"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/osdconfig"
)

// swagger:operation GET /config/cluster config getClusterConfig
//
// Get cluster configuration.
//
// This will return the requested cluster configuration object
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: a cluster config
//      schema:
//       $ref: '#/definitions/ClusterConfig'
func (c *clusterApi) getClusterConf(w http.ResponseWriter, r *http.Request) {
	method := "getClusterConf"

	ctx, err := c.annotateContext(r)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		osdConfigClient := api.NewOpenStorageOsdConfigClient(conn)

		resp, err := osdConfigClient.GetClusterConf(ctx, &api.SdkOsdGetClusterConfigRequest{})

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		config := osdconfig.ClusterConfig{
			Description: resp.Description,
			Mode:        resp.Mode,
			Version:     resp.Version,
			Created:     prototime.TimestampToTime(resp.Created),
			ClusterId:   resp.ClusterId,
			Domain:      resp.Domain,
		}

		if err := json.NewEncoder(w).Encode(config); err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// swagger:operation GET /config/node/{id} config getNodeConfig
//
// Get node configuration.
//
// This will return the requested node configuration object
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get node with
//   required: true
//   type: string
// responses:
//   '200':
//      description: a node
//      schema:
//       $ref: '#/definitions/NodeConfig'
func (c *clusterApi) getNodeConf(w http.ResponseWriter, r *http.Request) {
	method := "getNodeConf"

	vars := mux.Vars(r)
	ctx, err := c.annotateContext(r)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		osdConfigClient := api.NewOpenStorageOsdConfigClient(conn)

		resp, err := osdConfigClient.GetNodeConf(ctx, &api.SdkOsdGetNodeConfigRequest{
			Id: vars["id"],
		})

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		config := sdkNodeConfigToOsdNodeConfig(resp)

		if err := json.NewEncoder(w).Encode(config); err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// swagger:operation GET /config/enumerate config enumerate
//
// Get configuration for all nodes.
//
// This will return the node configuration for all nodes
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: node config enumeration
//      schema:
//       $ref: '#/definitions/NodesConfig'
func (c *clusterApi) enumerateConf(w http.ResponseWriter, r *http.Request) {
	method := "enumerateConf"
	ctx, err := c.annotateContext(r)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		osdConfigClient := api.NewOpenStorageOsdConfigClient(conn)

		resp, err := osdConfigClient.EnumerateNodeConf(ctx, &api.SdkOsdEnumerateNodeConfigRequest{})

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(resp.NodeConfig); err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// swagger:operation DELETE /config/node/{id} config deleteNodeConfig
//
// Delete node configuration.
//
// This will delete the requested node configuration object
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to reference node
//   required: true
//   type: string
// responses:
//   '200':
//      description: success
func (c *clusterApi) delNodeConf(w http.ResponseWriter, r *http.Request) {
	method := "delNodeConf"
	vars := mux.Vars(r)

	ctx, err := c.annotateContext(r)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		osdConfigClient := api.NewOpenStorageOsdConfigClient(conn)

		_, err := osdConfigClient.DeleteNodeConf(ctx, &api.SdkOsdDeleteNodeConfigRequest{
			Id: vars["id"],
		})

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// swagger:operation POST /config/cluster config setClusterConfig
//
// Set cluster configuration.
//
// This will set the requested cluster configuration
//
// ---
// produces:
// - application/json
// parameters:
// - name: config
//   in: body
//   description: cluster config json
//   required: true
//   schema:
//    $ref: '#/definitions/ClusterConfig'
// responses:
//   '200':
//     description: success
//     schema:
//       type: string
func (c *clusterApi) setClusterConf(w http.ResponseWriter, r *http.Request) {
	method := "setClusterConf"

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(data) > 2 {
		data = data[1 : len(data)-1]
	} else {
		c.sendError(c.name, method, w, "incorrect form input", http.StatusInternalServerError)
		return
	}

	data, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	config := new(osdconfig.ClusterConfig)
	if err := json.Unmarshal(data, config); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, err := c.annotateContext(r)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		osdConfigClient := api.NewOpenStorageOsdConfigClient(conn)

		resp, err := osdConfigClient.SetClusterConf(ctx, &api.SdkOsdSetClusterConfigRequest{
			Description: config.Description,
			Mode:        config.Mode,
			Version:     config.Version,
			Created:     prototime.TimeToTimestamp(config.Created),
			ClusterId:   config.ClusterId,
		})

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// swagger:operation POST /config/node config setNodeConfig
//
// Set node configuration.
//
// This will set the requested node configuration
//
// ---
// produces:
// - application/json
// parameters:
// - name: config
//   in: body
//   description: node config json
//   required: true
//   schema:
//     $ref: '#/definitions/NodeConfig'
// responses:
//   '200':
//      description: success
func (c *clusterApi) setNodeConf(w http.ResponseWriter, r *http.Request) {
	method := "setNodeConf"
	vars := mux.Vars(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(data) > 2 {
		data = data[1 : len(data)-1]
	} else {
		c.sendError(c.name, method, w, "incorrect form input", http.StatusInternalServerError)
		return
	}

	data, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	config := new(osdconfig.NodeConfig)
	if err := json.Unmarshal(data, config); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, err := c.annotateContext(r)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		sdkNodeConf := osdNodeConfigSdkNodeConfig(config)
		sdkNodeConf.Id = vars["id"]

		osdConfigClient := api.NewOpenStorageOsdConfigClient(conn)
		resp, err := osdConfigClient.SetNodeConf(ctx, sdkNodeConf)

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func osdNodeConfigSdkNodeConfig(nodeConfig *osdconfig.NodeConfig) *api.SdkOsdNodeConfig {
	var networkConfig *api.NetworkConfig
	if nodeConfig.Network != nil {
		networkConfig = &api.NetworkConfig{
			ManagementInterface: nodeConfig.Network.MgtIface,
			DataInterface:       nodeConfig.Network.DataIface,
		}
	}

	var storageConfig *api.StorageConfig

	if nodeConfig.Storage != nil {
		storageConfig = &api.StorageConfig{
			DevicesMode:      nodeConfig.Storage.DevicesMd,
			Devices:          nodeConfig.Storage.Devices,
			MaxCount:         nodeConfig.Storage.MaxCount,
			MaxDriveSetCount: nodeConfig.Storage.MaxDriveSetCount,
			RaidLevel:        nodeConfig.Storage.RaidLevel,
			RaidLevelMode:    nodeConfig.Storage.RaidLevelMd,
		}
	}

	var geoConfig *api.GeoConfig

	if nodeConfig.Geo != nil {
		geoConfig = &api.GeoConfig{
			Rack:   nodeConfig.Geo.Rack,
			Zone:   nodeConfig.Geo.Zone,
			Region: nodeConfig.Geo.Region,
		}
	}

	config := &api.SdkOsdNodeConfig{
		Id:            nodeConfig.NodeId,
		CSIEndpoint:   nodeConfig.CSIEndpoint,
		Network:       networkConfig,
		Storage:       storageConfig,
		Geo:           geoConfig,
		ClusterDomain: nodeConfig.ClusterDomain,
	}

	return config
}

func sdkNodeConfigToOsdNodeConfig(sdkOsdNodeConfig *api.SdkOsdNodeConfig) *osdconfig.NodeConfig {
	var networkConfig *osdconfig.NetworkConfig

	if sdkOsdNodeConfig.Network != nil {
		networkConfig = &osdconfig.NetworkConfig{
			MgtIface:  sdkOsdNodeConfig.Network.ManagementInterface,
			DataIface: sdkOsdNodeConfig.Network.DataInterface,
		}
	}

	var storageConfig *osdconfig.StorageConfig

	if sdkOsdNodeConfig.Storage != nil {
		storageConfig = &osdconfig.StorageConfig{
			DevicesMd:        sdkOsdNodeConfig.Storage.DevicesMode,
			Devices:          sdkOsdNodeConfig.Storage.Devices,
			MaxCount:         sdkOsdNodeConfig.Storage.MaxCount,
			MaxDriveSetCount: sdkOsdNodeConfig.Storage.MaxDriveSetCount,
			RaidLevel:        sdkOsdNodeConfig.Storage.RaidLevel,
			RaidLevelMd:      sdkOsdNodeConfig.Storage.RaidLevelMode,
		}
	}

	var geoConfig *osdconfig.GeoConfig

	if sdkOsdNodeConfig.Geo != nil {
		geoConfig = &osdconfig.GeoConfig{
			Rack:   sdkOsdNodeConfig.Geo.Rack,
			Zone:   sdkOsdNodeConfig.Geo.Zone,
			Region: sdkOsdNodeConfig.Geo.Region,
		}
	}

	config := &osdconfig.NodeConfig{
		NodeId:        sdkOsdNodeConfig.Id,
		CSIEndpoint:   sdkOsdNodeConfig.CSIEndpoint,
		Network:       networkConfig,
		Storage:       storageConfig,
		Geo:           geoConfig,
		ClusterDomain: sdkOsdNodeConfig.ClusterDomain,
	}

	return config
}
