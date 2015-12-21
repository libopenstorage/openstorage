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
