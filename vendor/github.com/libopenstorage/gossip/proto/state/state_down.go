package state

import (
	"github.com/libopenstorage/gossip/types"
)

type down struct {
	nodeStatus     types.NodeStatus
	stateEvent     chan types.StateEvent
	quorumProvider Quorum
}

func GetDown(
	stateEvent chan types.StateEvent,
	quorumProvider Quorum,
) State {
	return &down{
		nodeStatus:     types.NODE_STATUS_DOWN,
		stateEvent:     stateEvent,
		quorumProvider: quorumProvider,
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
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return d, nil
}

func (d *down) UpdateClusterDomainsActiveMap(
	localNodeInfo types.NodeInfoMap,
) (State, error) {
	return d, nil
}

func (d *down) Timeout(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return d, nil
}
