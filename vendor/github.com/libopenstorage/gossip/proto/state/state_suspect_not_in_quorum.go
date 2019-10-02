package state

import (
	"github.com/libopenstorage/gossip/types"
)

type suspectNotInQuorum struct {
	nodeStatus     types.NodeStatus
	stateEvent     chan types.StateEvent
	quorumProvider Quorum
}

var instanceSuspectNotInQuorum *suspectNotInQuorum

func GetSuspectNotInQuorum(
	stateEvent chan types.StateEvent,
	quorumProvider Quorum,
) State {
	return &suspectNotInQuorum{
		nodeStatus:     types.NODE_STATUS_SUSPECT_NOT_IN_QUORUM,
		stateEvent:     stateEvent,
		quorumProvider: quorumProvider,
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
	if !siq.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		return siq, nil
	} else {
		return GetUp(siq.stateEvent, siq.quorumProvider), nil
	}
}

func (siq *suspectNotInQuorum) SelfLeave() (State, error) {
	down := GetDown(siq.stateEvent, siq.quorumProvider)
	return down, nil
}

func (siq *suspectNotInQuorum) NodeLeave(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return siq, nil
}

func (siq *suspectNotInQuorum) UpdateClusterSize(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	if !siq.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		return siq, nil
	} else {
		return GetUp(siq.stateEvent, siq.quorumProvider), nil
	}
}

func (siq *suspectNotInQuorum) UpdateClusterDomainsActiveMap(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	if !siq.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		return siq, nil
	} else {
		return GetUp(siq.stateEvent, siq.quorumProvider), nil
	}
}

func (siq *suspectNotInQuorum) Timeout(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	if !siq.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		return GetNotInQuorum(siq.stateEvent, siq.quorumProvider), nil
	} else {
		return GetUp(siq.stateEvent, siq.quorumProvider), nil
	}
}
