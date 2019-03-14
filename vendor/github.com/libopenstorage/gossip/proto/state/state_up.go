package state

import (
	"github.com/libopenstorage/gossip/types"
)

type up struct {
	nodeStatus     types.NodeStatus
	stateEvent     chan types.StateEvent
	quorumProvider Quorum
}

func GetUp(
	stateEvent chan types.StateEvent,
	quorumProvider Quorum,
) State {
	return &up{
		nodeStatus:     types.NODE_STATUS_UP,
		stateEvent:     stateEvent,
		quorumProvider: quorumProvider,
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
	down := GetDown(u.stateEvent, u.quorumProvider)
	return down, nil
}

func (u *up) NodeLeave(localNodeInfoMap types.NodeInfoMap) (State, error) {
	if !u.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.stateEvent, u.quorumProvider), nil
	} else {
		return u, nil
	}
}

func (u *up) UpdateClusterSize(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	if !u.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.stateEvent, u.quorumProvider), nil
	} else {
		return u, nil
	}
}

func (u *up) UpdateClusterDomainsActiveMap(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	if !u.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		// Caller of this function should start a timer
		return GetSuspectNotInQuorum(u.stateEvent, u.quorumProvider), nil
	} else {
		return u, nil
	}
}

func (u *up) Timeout(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return u, nil
}
