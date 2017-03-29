package state

import (
	"github.com/libopenstorage/gossip/types"
)

type notInQuorum struct {
	nodeStatus       types.NodeStatus
	id               types.NodeId
	numQuorumMembers uint
	stateEvent       chan types.StateEvent
}

var instanceNotInQuorum *notInQuorum

func GetNotInQuorum(
	numQuorumMembers uint,
	selfId types.NodeId,
	stateEvent chan types.StateEvent,
) State {
	return &notInQuorum{
		nodeStatus:       types.NODE_STATUS_NOT_IN_QUORUM,
		numQuorumMembers: numQuorumMembers,
		id:               selfId,
		stateEvent:       stateEvent,
	}
}

func (niq *notInQuorum) String() string {
	return "NODE_STATUS_NOT_IN_QUORUM"
}

func (niq *notInQuorum) NodeStatus() types.NodeStatus {
	return niq.nodeStatus
}

func (niq *notInQuorum) SelfAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	quorum := (niq.numQuorumMembers / 2) + 1
	upNodes := numQuorumMembersUp(localNodeInfoMap)
	if upNodes < quorum {
		return niq, nil
	} else {
		up := GetUp(niq.numQuorumMembers, niq.id, niq.stateEvent)
		return up, nil
	}
}

func (niq *notInQuorum) NodeAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	quorum := (niq.numQuorumMembers / 2) + 1
	upNodes := numQuorumMembersUp(localNodeInfoMap)
	if upNodes < quorum {
		return niq, nil
	} else {
		up := GetUp(niq.numQuorumMembers, niq.id, niq.stateEvent)
		return up, nil
	}
}

func (niq *notInQuorum) SelfLeave() (State, error) {
	down := GetDown(niq.numQuorumMembers, niq.id, niq.stateEvent)
	return down, nil
}

func (niq *notInQuorum) NodeLeave(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return niq, nil
}

func (niq *notInQuorum) UpdateClusterSize(
	numQuorumMembers uint,
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	niq.numQuorumMembers = numQuorumMembers
	quorum := (niq.numQuorumMembers / 2) + 1
	upNodes := numQuorumMembersUp(localNodeInfoMap)
	if upNodes < quorum {
		return niq, nil
	} else {
		up := GetUp(niq.numQuorumMembers, niq.id, niq.stateEvent)
		return up, nil
	}
}

func (niq *notInQuorum) Timeout(
	numQuorumMembers uint,
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return niq, nil
}
