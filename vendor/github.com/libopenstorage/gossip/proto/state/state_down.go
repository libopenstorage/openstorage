package state

import (
	"github.com/libopenstorage/gossip/types"
)

type down struct {
	nodeStatus  types.NodeStatus
	id          types.NodeId
	clusterSize int
	stateEvent  chan types.StateEvent
}

func GetDown(clusterSize int, selfId types.NodeId, stateEvent chan types.StateEvent) State {
	return &down{
		nodeStatus:  types.NODE_STATUS_DOWN,
		clusterSize: clusterSize,
		id:          selfId,
		stateEvent:  stateEvent,
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

func (d *down) UpdateClusterSize(clusterSize int, localNodeInfoMap types.NodeInfoMap) (State, error) {
	return d, nil
}

func (d *down) Timeout(clusterSize int, localNodeInfoMap types.NodeInfoMap) (State, error) {
	return d, nil
}
