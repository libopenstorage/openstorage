package state

import (
	"sync"

	"github.com/libopenstorage/gossip/types"
)

// Quorum provides a set of APIs to determine and modify a node's Quorum state
type Quorum interface {
	// IsNodeInQuorum returns a boolean indicating whether a node is in quorum
	IsNodeInQuorum(localNodeInfoMap types.NodeInfoMap) bool
	// UpdateNumOfQuorumMembers updates the number of members
	// participating in quorum calculationgs
	UpdateNumOfQuorumMembers(quorumMemberMap types.ClusterDomainsQuorumMembersMap)
	// UpdateClusterDomainsActiveMap updates the map of active and inactive failure
	// domains. It returns a boolean value indicating if an update was done
	UpdateClusterDomainsActiveMap(activeMap types.ClusterDomainsActiveMap) bool
	// Type returns the type of quorum implementation
	Type() types.QuorumProvider
}

// NewQuorumProvider returns tan implementation of Quorum interface based on
// the input type
func NewQuorumProvider(
	selfId types.NodeId,
	provider types.QuorumProvider,
) Quorum {
	if provider == types.QUORUM_PROVIDER_DEFAULT {
		return &defaultQuorum{
			selfId: selfId,
		}
	}
	return &failureDomainsQuorum{
		selfId: selfId,
	}
}

type defaultQuorum struct {
	numQuorumMembers uint
	selfId           types.NodeId
	lock             sync.Mutex
}

func (d *defaultQuorum) IsNodeInQuorum(localNodeInfoMap types.NodeInfoMap) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	upNodes := uint(0)
	for _, nodeInfo := range localNodeInfoMap {
		if nodeInfo.QuorumMember &&
			(nodeInfo.Status == types.NODE_STATUS_UP ||
				nodeInfo.Status == types.NODE_STATUS_NOT_IN_QUORUM ||
				nodeInfo.Status == types.NODE_STATUS_SUSPECT_NOT_IN_QUORUM) {
			upNodes++
		}
	}
	quorum := (d.numQuorumMembers / 2) + 1
	return upNodes >= quorum
}

func (d *defaultQuorum) UpdateNumOfQuorumMembers(quorumMemberMap types.ClusterDomainsQuorumMembersMap) {
	d.lock.Lock()
	defer d.lock.Unlock()
	numOfQuorumMembers := uint(0)
	for _, quorumMembersInDomain := range quorumMemberMap {
		numOfQuorumMembers = numOfQuorumMembers + uint(quorumMembersInDomain)
	}
	d.numQuorumMembers = numOfQuorumMembers
}

func (d *defaultQuorum) UpdateClusterDomainsActiveMap(activeMap types.ClusterDomainsActiveMap) bool {
	// no op
	return false
}

func (d *defaultQuorum) Type() types.QuorumProvider {
	return types.QUORUM_PROVIDER_DEFAULT
}
