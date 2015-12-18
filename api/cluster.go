package api

import (
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/portworx/systemutils"
)

type Status int

const (
	StatusInit Status = 1 << iota
	StatusOk
	StatusOffline
	StatusError
)

type VolumeInfo struct {
	Path     string
	Storage  *VolumeSpec
	VolumeID VolumeID
}

// Node describes the state of a node.
// It includes the current physical state (CPU, memory, storage, network usage) as
// well as the containers running on the system.
type Node struct {
	Id         string
	Cpu        float64 // percentage.
	Memory     float64 // percentage.
	Luns       map[string]systemutils.Lun
	Avgload    int
	Ip         string
	Timestamp  time.Time
	Status     Status
	Containers []docker.APIContainers
	NodeData   map[string]interface{}
	GenNumber  uint64
}

// Cluster represents the state of the cluster.
type Cluster struct {
	Status Status
	Id     string
	Nodes  []Node
}

// VolumeActionParam desired action on volume
type ClusterActionParam int

// ClusterStateAction is the body of the REST request to specify desired actions
type ClusterStateAction struct {
	// Remove a node or a set of nodes
	Remove ClusterActionParam `json:"remove"`

	// Shutdown a node or a set of nodes
	Shutdown ClusterActionParam `json:"shutdown"`
}

// ClusterStateResponse is the body of the REST response
type ClusterStateResponse struct {
	// VolumeStateRequest the current state of the volume
	ClusterStateAction
	ClusterResponse
}

// VolumeResponse is embedded in all REST responses.
type ClusterResponse struct {
	// Error is "" on success or contains the error message on failure.
	Error string `json:"error"`
}
