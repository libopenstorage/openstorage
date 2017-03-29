package state

import (
	"github.com/libopenstorage/gossip/types"
)

type down struct {
	nodeStatus       types.NodeStatus
	id               types.NodeId
	numQuorumMembers uint
	stateEvent       chan types.StateEvent
}

func GetDown(
	numQuorumMembers uint,
	selfId types.NodeId,
	stateEvent chan types.StateEvent,
) State {
	return &down{
		nodeStatus:       types.NODE_STATUS_DOWN,
		numQuorumMembers: numQuorumMembers,
		id:               selfId,
		stateEvent:       stateEvent,
	}
}

func (d *down) String() string {
	return "NODE_STATUS_DOWN"
}

func (d *down) NodeStatus() types.NodeStatus {
	return d.nodeStatus
}

func (d *down) SelfAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	return d, nil
}

func (d *down) NodeAlive(localNodeInfo types.NodeInfoMap) (State, error) {
	return d, nil
}

func (d *down) SelfLeave() (State, error) {
	return d, nil
}

func (d *down) NodeLeave(localNodeInfoMap types.NodeInfoMap) (State, error) {
	return d, nil
}

func (d *down) UpdateClusterSize(
	numQuorumMembers uint,
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return d, nil
}

func (d *down) Timeout(
	numQuorumMembers uint,
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return d, nil
}
