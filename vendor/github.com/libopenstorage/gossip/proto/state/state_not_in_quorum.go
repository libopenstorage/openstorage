package state

import (
	"github.com/libopenstorage/gossip/types"
)

type notInQuorum struct {
	nodeStatus  types.NodeStatus
	id          types.NodeId
	clusterSize int
	stateEvent  chan types.StateEvent
}

var instanceNotInQuorum *notInQuorum

func GetNotInQuorum(clusterSize int, selfId types.NodeId, stateEvent chan types.StateEvent) State {
	return &notInQuorum{
		nodeStatus:  types.NODE_STATUS_NOT_IN_QUORUM,
		clusterSize: clusterSize,
		id:          selfId,
		stateEvent:  stateEvent,
	}
}

func (niq *notInQuorum) String() string {
	return "NODE_STATUS_NOT_IN_QUORUM"
}

func (niq *notInQuorum) NodeStatus() types.NodeStatus {
	return niq.nodeStatus
}

func (niq *notInQuorum) SelfAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	quorum := (niq.clusterSize / 2) + 1
	upNodes := calculateUpNodes(localNodeInfoMap)
	if upNodes < quorum {
		return niq, nil
	} else {
		up := GetUp(niq.clusterSize, niq.id, niq.stateEvent)
		return up, nil
	}
}

func (niq *notInQuorum) NodeAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	quorum := (niq.clusterSize / 2) + 1
	upNodes := calculateUpNodes(localNodeInfoMap)
	if upNodes < quorum {
		return niq, nil
	} else {
		up := GetUp(niq.clusterSize, niq.id, niq.stateEvent)
		return up, nil
	}
}

func (niq *notInQuorum) SelfLeave() (State, error) {
	down := GetDown(niq.clusterSize, niq.id, niq.stateEvent)
	return down, nil
}

func (niq *notInQuorum) NodeLeave(localNodeInfoMap types.NodeInfoMap) (State, error) {
	return niq, nil
}

func (niq *notInQuorum) UpdateClusterSize(clusterSize int, localNodeInfoMap types.NodeInfoMap) (State, error) {
	niq.clusterSize = clusterSize
	quorum := (niq.clusterSize / 2) + 1
	upNodes := calculateUpNodes(localNodeInfoMap)
	if upNodes < quorum {
		return niq, nil
	} else {
		up := GetUp(niq.clusterSize, niq.id, niq.stateEvent)
		return up, nil
	}
}

func (niq *notInQuorum) Timeout(clusterSize int, localNodeInfoMap types.NodeInfoMap) (State, error) {
	return niq, nil
}
