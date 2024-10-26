package state

import (
	"github.com/libopenstorage/gossip/types"
)

// noopQuorumProvider is a no-op implementation of Quorum interface
// It disables quorum requirement for a cluster
type noopQuorumProvider struct {
}

func (d *noopQuorumProvider) IsNodeInQuorum(localNodeInfoMap types.NodeInfoMap) bool {
	// no quorum requirement
	return true
}

func (d *noopQuorumProvider) IsDomainActive(ipDomain string) bool {
	// domain agnostic
	return true
}

func (d *noopQuorumProvider) UpdateNumOfQuorumMembers(quorumMemberMap types.ClusterDomainsQuorumMembersMap) {
	// no op
}

func (d *noopQuorumProvider) UpdateClusterDomainsActiveMap(activeMap types.ClusterDomainsActiveMap) bool {
	// no op
	return false
}

func (d *noopQuorumProvider) Type() types.QuorumProvider {
	return types.QUORUM_PROVIDER_NOOP
}
