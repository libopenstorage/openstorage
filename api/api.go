package api

import (
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/portworx/systemutils"
)

// Version API version
const Version = "v1"

// Strings for VolumeSpec
const (
	SpecEphemeral        = "ephemeral"
	SpecSize             = "size"
	SpecFilesystem       = "format"
	SpecBlockSize        = "blocksize"
	SpecHaLevel          = "ha_level"
	SpecCos              = "cos"
	SpecSnapshotInterval = "snapshot_interval"
	SpecDedupe           = "dedupe"
)

// OptionKey specifies a set of recognized query params
const (
	// OptName query parameter used to lookup volume by name
	OptName = "Name"
	// OptVolumeID query parameter used to lookup volume by ID.
	OptVolumeID = "VolumeID"
	// OptLabel query parameter used to lookup volume by set of labels.
	OptLabel = "Label"
	// OptConfigLabel query parameter used to lookup volume by set of labels.
	OptConfigLabel = "ConfigLabel"
)

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
