package state

import (
	"github.com/libopenstorage/gossip/types"
)

type notInQuorum struct {
	nodeStatus     types.NodeStatus
	stateEvent     chan types.StateEvent
	quorumProvider Quorum
}

var instanceNotInQuorum *notInQuorum

func GetNotInQuorum(
	stateEvent chan types.StateEvent,
	quorumProvider Quorum,
) State {
	return &notInQuorum{
		nodeStatus:     types.NODE_STATUS_NOT_IN_QUORUM,
		stateEvent:     stateEvent,
		quorumProvider: quorumProvider,
	}
}

func (niq *notInQuorum) String() string {
	return "NODE_STATUS_NOT_IN_QUORUM"
}

func (niq *notInQuorum) NodeStatus() types.NodeStatus {
	return niq.nodeStatus
}

func (niq *notInQuorum) SelfAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	if !niq.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		return niq, nil
	} else {
		return GetUp(niq.stateEvent, niq.quorumProvider), nil
	}
}

func (niq *notInQuorum) NodeAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	if !niq.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		return niq, nil
	} else {
		return GetUp(niq.stateEvent, niq.quorumProvider), nil
	}
}

func (niq *notInQuorum) SelfLeave() (State, error) {
	return GetDown(niq.stateEvent, niq.quorumProvider), nil
}

func (niq *notInQuorum) NodeLeave(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return niq, nil
}

func (niq *notInQuorum) UpdateClusterSize(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	if !niq.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		return niq, nil
	} else {
		return GetUp(niq.stateEvent, niq.quorumProvider), nil
	}
}

func (niq *notInQuorum) UpdateClusterDomainsActiveMap(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	if !niq.quorumProvider.IsNodeInQuorum(localNodeInfoMap) {
		return niq, nil
	} else {
		return GetUp(niq.stateEvent, niq.quorumProvider), nil
	}
}

func (niq *notInQuorum) Timeout(
	localNodeInfoMap types.NodeInfoMap,
) (State, error) {
	return niq, nil
}
