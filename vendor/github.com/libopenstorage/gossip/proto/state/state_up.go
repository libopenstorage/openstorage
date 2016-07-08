package state

import (
	"github.com/libopenstorage/gossip/types"
)

type up struct {
	nodeStatus  types.NodeStatus
	id          types.NodeId
	clusterSize int
	stateEvent  chan types.StateEvent
}

func GetUp(clusterSize int, selfId types.NodeId, stateEvent chan types.StateEvent) State {
	return &up{
		nodeStatus:  types.NODE_STATUS_UP,
		clusterSize: clusterSize,
		id:          selfId,
		stateEvent:  stateEvent,
	}
}

func (u *up) String() string {
	return "NODE_STATUS_UP"
}

func (u *up) NodeStatus() types.NodeStatus {
	return u.nodeStatus
}

func (u *up) SelfAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	return u, nil
}

func (u *up) NodeAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	return u, nil
}

func (u *up) SelfLeave() (State, error) {
	down := GetDown(u.clusterSize, u.id, u.stateEvent)
	return down, nil
}

func calculateUpNodes(localNodeInfoMap types.NodeInfoMap) int {
	upNodes := 0
	for _, nodeInfo := range localNodeInfoMap {
		if nodeInfo.Status == types.NODE_STATUS_UP ||
			nodeInfo.Status == types.NODE_STATUS_NOT_IN_QUORUM ||
			nodeInfo.Status == types.NODE_STATUS_SUSPECT_NOT_IN_QUORUM {
			upNodes++
		}
	}
	return upNodes
}

func (u *up) NodeLeave(localNodeInfoMap types.NodeInfoMap) (State, error) {
	quorum := (u.clusterSize / 2) + 1
	upNodes := calculateUpNodes(localNodeInfoMap)
	if upNodes < quorum {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.clusterSize, u.id, u.stateEvent), nil
	} else {
		return u, nil
	}
}

func (u *up) UpdateClusterSize(clusterSize int, localNodeInfoMap types.NodeInfoMap) (State, error) {
	u.clusterSize = clusterSize
	quorum := (u.clusterSize / 2) + 1
	upNodes := calculateUpNodes(localNodeInfoMap)
	if upNodes < quorum {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.clusterSize, u.id, u.stateEvent), nil
	} else {
		return u, nil
	}
}

func (u *up) Timeout(clusterSize int, localNodeInfoMap types.NodeInfoMap) (State, error) {
	return u, nil
}
