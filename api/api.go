package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/portworx/systemutils"
)

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

func StatusSimpleValueOf(s string) (Status, error) {
	obj, err := simpleValueOf("status", Status_value, s)
	return Status(obj), err
}

func (x Status) SimpleString() string {
	return simpleString("status", Status_name, int32(x))
}

func DriverTypeSimpleValueOf(s string) (DriverType, error) {
	obj, err := simpleValueOf("driver_type", DriverType_value, s)
	return DriverType(obj), err
}

func (x DriverType) SimpleString() string {
	return simpleString("driver_type", DriverType_name, int32(x))
}

func FSTypeSimpleValueOf(s string) (FSType, error) {
	obj, err := simpleValueOf("fs_type", FSType_value, s)
	return FSType(obj), err
}

func (x FSType) SimpleString() string {
	return simpleString("fs_type", FSType_name, int32(x))
}

func GraphDriverChangeTypeSimpleValueOf(s string) (GraphDriverChangeType, error) {
	obj, err := simpleValueOf("graph_driver_change_type", GraphDriverChangeType_value, s)
	return GraphDriverChangeType(obj), err
}

func (x GraphDriverChangeType) SimpleString() string {
	return simpleString("graph_driver_change_type", GraphDriverChangeType_name, int32(x))
}

func VolumeActionParamSimpleValueOf(s string) (VolumeActionParam, error) {
	obj, err := simpleValueOf("volume_action_param", VolumeActionParam_value, s)
	return VolumeActionParam(obj), err
}

func (x VolumeActionParam) SimpleString() string {
	return simpleString("volume_action_param", VolumeActionParam_name, int32(x))
}

func VolumeStateSimpleValueOf(s string) (VolumeState, error) {
	obj, err := simpleValueOf("volume_state", VolumeState_value, s)
	return VolumeState(obj), err
}

func (x VolumeState) SimpleString() string {
	return simpleString("volume_state", VolumeState_name, int32(x))
}

func VolumeStatusSimpleValueOf(s string) (VolumeStatus, error) {
	obj, err := simpleValueOf("volume_status", VolumeStatus_value, s)
	return VolumeStatus(obj), err
}

func (x VolumeStatus) SimpleString() string {
	return simpleString("volume_status", VolumeStatus_name, int32(x))
}

func simpleValueOf(typeString string, valueMap map[string]int32, s string) (int32, error) {
	obj, ok := valueMap[strings.ToUpper(fmt.Sprintf("%s_%s", typeString, s))]
	if !ok {
		return 0, fmt.Errorf("no openstorage.%s for %s", strings.ToUpper(typeString), s)
	}
	return obj, nil
}

func simpleString(typeString string, nameMap map[int32]string, v int32) string {
	s, ok := nameMap[v]
	if !ok {
		return strconv.Itoa(int(v))
	}
	return strings.TrimPrefix(strings.ToLower(s), fmt.Sprintf("%s_", strings.ToLower(typeString)))
}
