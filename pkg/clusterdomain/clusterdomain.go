package clusterdomain

import (
	"errors"

	"github.com/libopenstorage/gossip/types"
)

var (
	ErrNotImplemented        = errors.New("Not implemented")
	ErrClusterDomainNotFound = errors.New("Cluster Domain not found")
)

// ClusterDomainInfo identifies a cluster domain in a cluster
type ClusterDomainInfo struct {
	Name  string
	State types.ClusterDomainState
}

// ClusterDomainProvider interface
type ClusterDomainProvider interface {
	// GetSelfDomain returns the cluster domain for this node
	GetSelfDomain() (*ClusterDomainInfo, error)

	// EnumerateDomains returns all the cluster domains in the cluster
	EnumerateDomains() ([]*ClusterDomainInfo, error)

	// Inspect returns the cluster domain info for the provided argument.
	InspectDomain(name string) (*ClusterDomainInfo, error)

	// DeleteDomain deletes a cluster domain entry
	DeleteDomain(name string) error

	// UpdateDomainState updates the state of cluster domain
	UpdateDomainState(name string, state types.ClusterDomainState) error
}

func NewDefaultClusterDomainPorvider() ClusterDomainProvider {
	return &NullClusterDomainManager{}
}

type NullClusterDomainManager struct{}

// GetSelfDomain returns the cluster domain for this node
func (n *NullClusterDomainManager) GetSelfDomain() (*ClusterDomainInfo, error) {
	return nil, ErrNotImplemented
}

// EnumerateDomains returns all the cluster domains in the cluster
func (n *NullClusterDomainManager) EnumerateDomains() ([]*ClusterDomainInfo, error) {
	return nil, ErrNotImplemented
}

// InspectDomain returns the cluster domain info for the provided argument.
func (n *NullClusterDomainManager) InspectDomain(name string) (*ClusterDomainInfo, error) {
	return nil, ErrNotImplemented
}

// DeleteDomain deletes a cluster domain entry
func (n *NullClusterDomainManager) DeleteDomain(name string) error {
	return ErrNotImplemented
}

// UpdateDomainState sets the cluster domain info object into kvdb
func (n *NullClusterDomainManager) UpdateDomainState(name string, state types.ClusterDomainState) error {
	return ErrNotImplemented
}

func (c *ClusterDomainInfo) Copy() *ClusterDomainInfo {
	cp := &ClusterDomainInfo{
		Name:  c.Name,
		State: c.State,
	}
	return cp
}

// GetActiveMapFromClusterDomainInfos is a helper function that converts a list of ClusterDomainInfo
// objects into a gossip cluster domain active map
func GetActiveMapFromClusterDomainInfos(clusterDomainInfos []*ClusterDomainInfo) types.ClusterDomainsActiveMap {
	activeMap := make(types.ClusterDomainsActiveMap)
	for _, clusterDomainInfo := range clusterDomainInfos {
		activeMap[clusterDomainInfo.Name] = clusterDomainInfo.State
	}
	return activeMap
}
