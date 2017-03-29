package state

import (
	"github.com/libopenstorage/gossip/types"
)

type suspectNotInQuorum struct {
	nodeStatus       types.NodeStatus
	id               types.NodeId
	numQuorumMembers uint
	stateEvent       chan types.StateEvent
}

var instanceSuspectNotInQuorum *suspectNotInQuorum

func GetSuspectNotInQuorum(
	numQuorumMembers uint,
	selfId types.NodeId,
	stateEvent chan types.StateEvent,
) State {
	return &suspectNotInQuorum{
		nodeStatus:       types.NODE_STATUS_SUSPECT_NOT_IN_QUORUM,
		numQuorumMembers: numQuorumMembers,
		id:               selfId,
		stateEvent:       stateEvent,
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
	quorum := (siq.numQuorumMembers / 2) + 1
	upNodes := numQuorumMembersUp(localNodeInfoMap)
	if upNodes < quorum {
		return siq, nil
	} else {
		up := GetUp(siq.numQuorumMembers, siq.id, siq.stateEvent)
		return up, nil
	}
}

func (siq *suspectNotInQuorum) SelfLeave() (State, error) {
	down := GetDown(siq.numQuorumMembers, siq.id, siq.stateEvent)
	return down, nil
}

func (siq *suspectNotInQuorum) NodeLeave(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return siq, nil
}

func (siq *suspectNotInQuorum) UpdateClusterSize(
	numQuorumMembers uint,
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	siq.numQuorumMembers = numQuorumMembers
	quorum := (siq.numQuorumMembers / 2) + 1
	upNodes := numQuorumMembersUp(localNodeInfoMap)
	if upNodes < quorum {
		return siq, nil
	} else {
		up := GetUp(siq.numQuorumMembers, siq.id, siq.stateEvent)
		return up, nil
	}
}

func (siq *suspectNotInQuorum) Timeout(
	numQuorumMembers uint,
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	siq.numQuorumMembers = numQuorumMembers
	quorum := (siq.numQuorumMembers / 2) + 1
	upNodes := numQuorumMembersUp(localNodeInfoMap)
	if upNodes < quorum {
		notInQuorum := GetNotInQuorum(siq.numQuorumMembers, siq.id, siq.stateEvent)
		return notInQuorum, nil
	} else {
		up := GetUp(siq.numQuorumMembers, siq.id, siq.stateEvent)
		return up, nil
	}
}
