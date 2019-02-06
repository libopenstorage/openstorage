package state

import (
	"fmt"

	"github.com/libopenstorage/gossip/types"
)

type up struct {
	nodeStatus          types.NodeStatus
	id                  types.NodeId
	numQuorumMembers    uint
	stateEvent          chan types.StateEvent
	activeFailureDomain string
}

func GetUp(
	numQuorumMembers uint,
	selfId types.NodeId,
	stateEvent chan types.StateEvent,
	activeFailureDomain string,
) State {
	return &up{
		nodeStatus:          types.NODE_STATUS_UP,
		numQuorumMembers:    numQuorumMembers,
		id:                  selfId,
		stateEvent:          stateEvent,
		activeFailureDomain: activeFailureDomain,
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
	down := GetDown(u.numQuorumMembers, u.id, u.stateEvent, u.activeFailureDomain)
	return down, nil
}

func isNodeInQuorum(
	localNodeInfoMap types.NodeInfoMap,
	selfId types.NodeId,
	quorumCount uint,
	activeFailureDomain string,
) bool {
	upNodes := uint(0)
	selfNodeInfo := localNodeInfoMap[selfId]
	selfDomain := selfNodeInfo.FailureDomain

	fmt.Println("selfDomain: ", selfDomain)
	if len(activeFailureDomain) > 0 && (selfDomain != activeFailureDomain) {
		// If there is an active failure domain, shoot ourselves down
		// if we are not part of that failure domain
		return false
	}

	// domainMap is a map of failure domain to no. of up nodes in that domain
	for _, nodeInfo := range localNodeInfoMap {
		if nodeInfo.QuorumMember {
			if nodeInfo.Status == types.NODE_STATUS_UP ||
				nodeInfo.Status == types.NODE_STATUS_NOT_IN_QUORUM ||
				nodeInfo.Status == types.NODE_STATUS_SUSPECT_NOT_IN_QUORUM {
				upNodes++
			}
		}
	}

	// Check if we are in quorum
	if upNodes >= quorumCount {
		return true
	}

	// Quorum nodes are not online

	// TODO: Check if there is a network split, only in that case
	// check for any active failure domain

	if len(activeFailureDomain) > 0 && (selfDomain == activeFailureDomain) {
		// This node is a part of the active failure domain
		return true
	}
	// We are out of quorum
	return false
}

func (u *up) NodeLeave(localNodeInfoMap types.NodeInfoMap) (State, error) {
	quorum := (u.numQuorumMembers / 2) + 1
	if !isNodeInQuorum(localNodeInfoMap, u.id, quorum, u.activeFailureDomain) {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.numQuorumMembers, u.id, u.stateEvent, u.activeFailureDomain), nil
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
	if !isNodeInQuorum(localNodeInfoMap, u.id, quorum, u.activeFailureDomain) {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.numQuorumMembers, u.id, u.stateEvent, u.activeFailureDomain), nil
	} else {
		return u, nil
	}
}

func (u *up) MarkActiveFailureDomain(
	activeFailureDomain string,
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	u.activeFailureDomain = activeFailureDomain
	quorum := (u.numQuorumMembers / 2) + 1
	if !isNodeInQuorum(localNodeInfoMap, u.id, quorum, u.activeFailureDomain) {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.numQuorumMembers, u.id, u.stateEvent, u.activeFailureDomain), nil
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
