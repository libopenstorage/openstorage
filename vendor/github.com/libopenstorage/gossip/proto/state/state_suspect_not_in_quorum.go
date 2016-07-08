package state

import (
	"github.com/libopenstorage/gossip/types"
)

type suspectNotInQuorum struct {
	nodeStatus  types.NodeStatus
	id          types.NodeId
	clusterSize int
	stateEvent  chan types.StateEvent
}

var instanceSuspectNotInQuorum *suspectNotInQuorum

func GetSuspectNotInQuorum(clusterSize int, selfId types.NodeId, stateEvent chan types.StateEvent) State {
	return &suspectNotInQuorum{
		nodeStatus:  types.NODE_STATUS_SUSPECT_NOT_IN_QUORUM,
		clusterSize: clusterSize,
		id:          selfId,
		stateEvent:  stateEvent,
	}
}

func (siq *suspectNotInQuorum) String() string {
	return "NODE_STATUS_SUSPECT_NOT_IN_QUORUM"
}

func (siq *suspectNotInQuorum) NodeStatus() types.NodeStatus {
	return siq.nodeStatus
}

func (siq *suspectNotInQuorum) SelfAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	return siq, nil
}

func (siq *suspectNotInQuorum) NodeAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	quorum := (siq.clusterSize / 2) + 1
	upNodes := calculateUpNodes(localNodeInfoMap)
	if upNodes < quorum {
		return siq, nil
	} else {
		up := GetUp(siq.clusterSize, siq.id, siq.stateEvent)
		return up, nil
	}
}

func (siq *suspectNotInQuorum) SelfLeave() (State, error) {
	down := GetDown(siq.clusterSize, siq.id, siq.stateEvent)
	return down, nil
}

func (siq *suspectNotInQuorum) NodeLeave(localNodeInfoMap types.NodeInfoMap) (State, error) {
	return siq, nil
}

func (siq *suspectNotInQuorum) UpdateClusterSize(clusterSize int, localNodeInfoMap types.NodeInfoMap) (State, error) {
	siq.clusterSize = clusterSize
	quorum := (siq.clusterSize / 2) + 1
	upNodes := calculateUpNodes(localNodeInfoMap)
	if upNodes < quorum {
		return siq, nil
	} else {
		up := GetUp(siq.clusterSize, siq.id, siq.stateEvent)
		return up, nil
	}
}

func (siq *suspectNotInQuorum) Timeout(clusterSize int, localNodeInfoMap types.NodeInfoMap) (State, error) {
	siq.clusterSize = clusterSize
	quorum := (siq.clusterSize / 2) + 1
	upNodes := calculateUpNodes(localNodeInfoMap)
	if upNodes < quorum {
		notInQuorum := GetNotInQuorum(siq.clusterSize, siq.id, siq.stateEvent)
		return notInQuorum, nil
	} else {
		up := GetUp(siq.clusterSize, siq.id, siq.stateEvent)
		return up, nil
	}
}
