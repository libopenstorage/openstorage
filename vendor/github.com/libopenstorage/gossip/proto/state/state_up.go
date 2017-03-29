package state

import (
	"github.com/libopenstorage/gossip/types"
)

type up struct {
	nodeStatus       types.NodeStatus
	id               types.NodeId
	numQuorumMembers uint
	stateEvent       chan types.StateEvent
}

func GetUp(
	numQuorumMembers uint,
	selfId types.NodeId,
	stateEvent chan types.StateEvent,
) State {
	return &up{
		nodeStatus:       types.NODE_STATUS_UP,
		numQuorumMembers: numQuorumMembers,
		id:               selfId,
		stateEvent:       stateEvent,
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
	down := GetDown(u.numQuorumMembers, u.id, u.stateEvent)
	return down, nil
}

func numQuorumMembersUp(localNodeInfoMap types.NodeInfoMap) uint {
	upNodes := uint(0)
	for _, nodeInfo := range localNodeInfoMap {
		if nodeInfo.QuorumMember &&
			(nodeInfo.Status == types.NODE_STATUS_UP ||
				nodeInfo.Status == types.NODE_STATUS_NOT_IN_QUORUM ||
				nodeInfo.Status == types.NODE_STATUS_SUSPECT_NOT_IN_QUORUM) {
			upNodes++
		}
	}
	return upNodes
}

func (u *up) NodeLeave(localNodeInfoMap types.NodeInfoMap) (State, error) {
	quorum := (u.numQuorumMembers / 2) + 1
	upNodes := numQuorumMembersUp(localNodeInfoMap)
	if upNodes < quorum {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.numQuorumMembers, u.id, u.stateEvent), nil
	} else {
		return u, nil
	}
}

func (u *up) UpdateClusterSize(
	numQuorumMembers uint,
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	u.numQuorumMembers = numQuorumMembers
	quorum := (u.numQuorumMembers / 2) + 1
	upNodes := numQuorumMembersUp(localNodeInfoMap)
	if upNodes < quorum {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.numQuorumMembers, u.id, u.stateEvent), nil
	} else {
		return u, nil
	}
}

func (u *up) Timeout(
	numQuorumMembers uint,
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return u, nil
}
