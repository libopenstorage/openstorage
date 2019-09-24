package state

import (
	"sync"

	"github.com/libopenstorage/gossip/types"
	"github.com/sirupsen/logrus"
)

// failureDomainsQuorum is an implementation of Quorum that incorporates
// failure domain information to determine whether a node is in quorum
type failureDomainsQuorum struct {
	selfId           types.NodeId
	activeMap        types.ClusterDomainsActiveMap
	quorumMembersMap types.ClusterDomainsQuorumMembersMap
	lock             sync.Mutex
}

func (f *failureDomainsQuorum) IsNodeInQuorum(localNodeInfoMap types.NodeInfoMap) bool {
	f.lock.Lock()
	defer f.lock.Unlock()

	selfNodeInfo := localNodeInfoMap[f.selfId]
	selfDomain := selfNodeInfo.ClusterDomain

	if !f.isDomainActive(selfDomain) {
		// This node is a part of deactivated failure domain
		// Shoot ourselves down as we are not in quorum
		return false
	}

	totalNodesInActiveDomains := uint(0)
	for domainName, isActive := range f.activeMap {
		if isActive != types.CLUSTER_DOMAIN_STATE_ACTIVE {
			continue
		}
		quorumCount := f.quorumMembersMap[domainName]
		totalNodesInActiveDomains = totalNodesInActiveDomains + uint(quorumCount)
	}
	upNodesInActiveDomains := uint(0)

	for _, nodeInfo := range localNodeInfoMap {
		if nodeInfo.QuorumMember {
			if !f.isDomainActive(nodeInfo.ClusterDomain) {
				// node is not a part of active domain
				// do not consider in quorum calculations
				continue
			}

			if nodeInfo.Status == types.NODE_STATUS_UP ||
				nodeInfo.Status == types.NODE_STATUS_NOT_IN_QUORUM ||
				nodeInfo.Status == types.NODE_STATUS_SUSPECT_NOT_IN_QUORUM {
				upNodesInActiveDomains++
			}
		}
	}

	// Check if we are in quorum
	quorumCount := (totalNodesInActiveDomains / 2) + 1
	return upNodesInActiveDomains >= quorumCount
}

// isDomainActive returns true if the input failure domains is a part of
// of an active domain list
func (f *failureDomainsQuorum) isDomainActive(ipDomain string) bool {
	isActive, _ := f.activeMap[ipDomain]
	if isActive == types.CLUSTER_DOMAIN_STATE_ACTIVE {
		return true
	}
	return false
}

func (f *failureDomainsQuorum) IsDomainActive(inputDomain string) bool {
	return f.isDomainActive(inputDomain)
}

func (f *failureDomainsQuorum) UpdateNumOfQuorumMembers(quorumMembersMap types.ClusterDomainsQuorumMembersMap) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.quorumMembersMap = make(types.ClusterDomainsQuorumMembersMap)
	for k, v := range quorumMembersMap {
		f.quorumMembersMap[k] = v
	}
	return
}

func (f *failureDomainsQuorum) UpdateClusterDomainsActiveMap(activeMap types.ClusterDomainsActiveMap) bool {
	f.lock.Lock()
	defer f.lock.Unlock()

	prevMap := f.activeMap
	f.activeMap = make(types.ClusterDomainsActiveMap)

	var stateChanged bool
	for domain, isActive := range activeMap {
		f.activeMap[domain] = isActive
		prevState := prevMap[domain]
		if prevState != isActive {
			stateChanged = true
			// State has changed
			if isActive == types.CLUSTER_DOMAIN_STATE_ACTIVE {
				logrus.Infof("gossip: Marking %v domain as active", domain)
			} else {
				logrus.Infof("gossip: Marking %v domain as inactive", domain)
			}
		}
	}
	return stateChanged
}

func (f *failureDomainsQuorum) Type() types.QuorumProvider {
	return types.QUORUM_PROVIDER_FAILURE_DOMAINS
}
