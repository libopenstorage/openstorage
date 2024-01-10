package loadbalancer

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/pkg/sched"
	"github.com/stretchr/testify/require"
)

var (
	ips = []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5", "127.0.0.6"}
	ids = []string{"1", "2", "3", "4", "5", "6"}
)

func getMockClusterResponse(enableClusterDomain bool) *api.Cluster {
	nodes := make([]*api.Node, 0)
	for i := 0; i < len(ips); i++ {
		node := &api.Node{
			MgmtIp: ips[i],
			Id:     ids[i],
		}
		if enableClusterDomain {
			if i%2 == 0 {
				node.DomainID = "domain1"
			} else {
				node.DomainID = "domain2"
			}
		}
		nodes = append(nodes, node)
	}
	return &api.Cluster{
		NodeId: ids[0], // self node
		Nodes:  nodes,
	}
}

func TestGetRemoteNodeWithoutDomains(t *testing.T) {
	if sched.Instance() == nil {
		sched.Init(time.Second)
	}
	ctrl := gomock.NewController(t)
	cc := mock.NewMockCluster(ctrl)
	rr, err := NewRoundRobinBalancer(cc, "1234")
	require.NoError(t, err, "failed to create round robin balancer")

	cc.EXPECT().Enumerate().Return(*getMockClusterResponse(false), nil).AnyTimes()
	cc.EXPECT().Inspect(ids[0]).Return(api.Node{MgmtIp: ips[0], Id: ids[0]}, nil).AnyTimes()

	for loop := 0; loop < 2; loop++ {
		targetNode, isRemoteConn, err := rr.GetRemoteNode()
		require.NoError(t, err, "failed to get remote node")
		require.Equal(t, targetNode, ips[0], "target node is not as expected")
		require.False(t, isRemoteConn, "isRemoteConn is not as expected")

		for i := 1; i < len(ips); i++ {
			targetNode, isRemoteConn, err := rr.GetRemoteNode()
			require.NoError(t, err, "failed to get remote node")
			require.Equal(t, targetNode, ips[i], "target node is not as expected")
			require.True(t, isRemoteConn, "isRemoteConn is not as expected")

		}
	}
}

func TestGetRemoteNodeWithDomains(t *testing.T) {
	if sched.Instance() == nil {
		sched.Init(time.Second)
	}
	ctrl := gomock.NewController(t)
	cc := mock.NewMockCluster(ctrl)
	rr, err := NewRoundRobinBalancer(cc, "1234")
	require.NoError(t, err, "failed to create round robin balancer")

	cc.EXPECT().Enumerate().Return(*getMockClusterResponse(true), nil).AnyTimes()
	cc.EXPECT().Inspect(ids[0]).Return(api.Node{MgmtIp: ips[0], Id: ids[0], DomainID: "domain1"}, nil).AnyTimes()

	for loop := 0; loop < 2; loop++ {
		targetNode, isRemoteConn, err := rr.GetRemoteNode()
		require.NoError(t, err, "failed to get remote node")
		require.Equal(t, targetNode, ips[0], "target node is not as expected")
		require.False(t, isRemoteConn, "isRemoteConn is not as expected")

		for i := 1; i < 3; i++ {
			targetNode, isRemoteConn, err := rr.GetRemoteNode()
			require.NoError(t, err, "failed to get remote node")
			require.Equal(t, targetNode, ips[i*2], "target node is not as expected")
			require.True(t, isRemoteConn, "isRemoteConn is not as expected")

		}
	}
}

// TestGetTargetAndIncrementNullPointer tests the case when getTargetAndIncrement
// causes null pointer. This happens when the round robin index is not checked against node array
// length before accessing it.
//
// https://portworx.atlassian.net/browse/PWX-35601
func TestGetTargetAndIncrementNullPointer(t *testing.T) {
	filteredNodes := []*api.Node{
		{
			Id:     "1",
			MgmtIp: "1",
		},
		{
			Id:     "2",
			MgmtIp: "2",
		},
	}
	rr := &roundRobin{nextCreateNodeNumber: len(filteredNodes)}
	endpoint, isRemote := rr.getTargetAndIncrement(filteredNodes, "")
	require.True(t, isRemote, "isRemote is not as expected")
	require.Equal(t, "1", endpoint, "target endpoint is not as expected")

	filteredNodes = []*api.Node{}
	rr.nextCreateNodeNumber = 0
	endpoint, isRemote = rr.getTargetAndIncrement(filteredNodes, "")
	require.False(t, isRemote, "isRemote is not as expected")
	require.Equal(t, "", endpoint, "target endpoint is not as expected")
}
