package discovery

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.pedge.io/dlog"
)

var LocalIPs = GetLocalIPList()

// GetLocalIPList returns the list of local IP addresses
func GetLocalIPList() (ipList []string) {
	//dlog.SetLevel(dlog.LevelDebug)

	ifaces, err := net.Interfaces()
	if err != nil {
		dlog.WithField("err", err).Warnf("Could not enumerate network interfaces (some tests will be disabled)")
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			// Let's make it non-fatal warning
			dlog.WithField("err", err).Warnf("Error inspecting addr %v\n", addrs, err)
			continue
		}
		// TODO: Exclude IPv6?
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP.To4()
			case *net.IPAddr:
				ip = v.IP.To4()
			}
			// process IP address
			if ip != nil && !ip.IsLoopback() && !ip.IsUnspecified() {
				ipList = append(ipList, ip.String())
			}
		}
	}
	return
}

// TestMulticastSingleNode tests the mechanics of a single node in KVDB cluster.
// WARNING: the test assumes 19010 TCP/UDP port is free (will not check if it is)
func TestMulticastSingleNode(t *testing.T) {
	serviceA, err := NewDiscoveryMcast("myCluster", "224.0.0.1:19010", "127.0.0.1:19010")
	defer serviceA.Shutdown()
	assert.Nil(t, err)

	// Basic sanity check -- add one node, expect to find one node

	node1 := NodeEntry{
		Id:            "node1",
		Ip:            "127.0.0.1",
		GossipVersion: "99",
	}
	ci, err := serviceA.AddNode(node1)

	assert.Nil(t, err)
	assert.NotEmpty(t, ci)
	assert.Equal(t, 1, ci.Size)
	assert.Equal(t, 1, len(ci.Nodes))
	assert.Equal(t, uint64(2), ci.Version)
	assert.NotNil(t, ci.Nodes[node1.Id])
	assert.Equal(t, node1, ci.Nodes[node1.Id])

	// Activate watch
	var watchdCi *ClusterInfo
	numWatchCalled := 0
	anonWcb := func(ci *ClusterInfo, err error) error {
		numWatchCalled++
		assert.NotNil(t, ci)
		watchdCi = ci
		return nil
	}
	serviceA.WatchCluster(anonWcb, uint64(123))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 0, numWatchCalled)
	assert.Nil(t, watchdCi)

	// Add another node ...
	node2 := NodeEntry{
		Id:            "node2",
		Ip:            "127.0.0.2",
		GossipVersion: "88",
	}
	ci, err = serviceA.AddNode(node2)

	assert.Nil(t, err)
	assert.NotEmpty(t, ci)
	assert.Equal(t, 2, ci.Size)
	assert.Equal(t, 2, len(ci.Nodes))
	assert.Equal(t, uint64(124), ci.Version)
	assert.NotNil(t, ci.Nodes[node1.Id])
	assert.NotNil(t, ci.Nodes[node2.Id])
	assert.Equal(t, node2, ci.Nodes[node2.Id])

	// Repeat the same checks w/ watched CI
	assert.NotEmpty(t, watchdCi)
	assert.Equal(t, 2, watchdCi.Size)
	assert.Equal(t, 2, len(watchdCi.Nodes))
	assert.Equal(t, uint64(124), watchdCi.Version)
	assert.NotNil(t, watchdCi.Nodes[node1.Id])
	assert.Equal(t, node1, watchdCi.Nodes[node1.Id])
	assert.NotNil(t, watchdCi.Nodes[node2.Id])
	assert.Equal(t, node2, watchdCi.Nodes[node2.Id])

	// Add first node again
	//
	ci, err = serviceA.AddNode(node1)

	// make sure version and number of elems didn't change
	assert.Nil(t, err)
	assert.Equal(t, 2, ci.Size)
	assert.Equal(t, 2, len(ci.Nodes))
	assert.Equal(t, uint64(124), ci.Version)

	// Change then re-add second
	node2.Ip = "198.162.1.2"
	ci, err = serviceA.AddNode(node2)

	// expect same number of nodes, new version ID
	assert.Nil(t, err)
	assert.Equal(t, 2, ci.Size)
	assert.Equal(t, 2, len(ci.Nodes))
	assert.Equal(t, uint64(125), ci.Version)
	assert.Equal(t, "198.162.1.2", watchdCi.Nodes[node2.Id].Ip)

}

func TestMulticastInvalidConstructors(t *testing.T) {
	var err error
	_, err = NewDiscoveryMcast("test-cluster", "224.0.0.1", "127.0.0.1:19011")
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "missing port in address"))

	_, err = NewDiscoveryMcast("test-cluster", "", "127.0.0.1")
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "missing port in address"))

	_, err = NewDiscoveryMcast("test-cluster", "224.0.0.1:19010", "127.0.0.1:19011")
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "must match "))
}

func TestMulticastConfigPayloads(t *testing.T) {
	restAddr := "127.0.0.1:19013"
	svc, err := NewDiscoveryMcast("test-cluster", "224.0.0.1:19013", restAddr)
	defer svc.Shutdown()
	assert.Nil(t, err)
	assert.NotNil(t, svc)

	// GET should fail because REST is not up, until we set up a watch
	restUrl := "http://" + restAddr
	resp, err := http.Get(restUrl)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "connection refused"))

	// set up a watch
	svc.WatchCluster(func(ci *ClusterInfo, err error) error {
		assert.NotNil(t, ci)
		return nil
	}, uint64(223))
	time.Sleep(1 * time.Second)

	// GET rest call fails now because it is not supported
	resp, err = http.Get(restUrl)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	// POST is supported, but let's pass empty request
	resp, err = http.Post(restUrl, "application/json", strings.NewReader(""))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// .. also non-json object
	resp, err = http.Post(restUrl, "application/json", strings.NewReader("^test/this_"))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestMulticastMultiNodes emulates the multi-node environment, by binding to different NICs and ports
// WARNING: the test assumes 19010 TCP/UDP ports are free (will not check if it is)
func TestMulticastMultiNodes(t *testing.T) {
	// For this test, we'll need at least 1 more NIC/IP apart from localhost/127.0.0.1
	if len(LocalIPs) == 0 {
		dlog.Warnf("No extra IPs found on this system (how did you log in?!) -- skipping this test")
		assert.False(t, false)
		return
	}

	clusterName := "testCluster2"
	serviceA, err := NewDiscoveryMcast(clusterName, "224.0.0.1:19010", "127.0.0.1:19010")
	assert.Nil(t, err)
	assert.NotNil(t, serviceA)
	defer serviceA.Shutdown()

	serviceB, err := NewDiscoveryMcast(clusterName, "224.0.0.1:19010", LocalIPs[0]+":19010")
	assert.Nil(t, err)
	assert.NotNil(t, serviceA)
	defer serviceB.Shutdown()

	neA1 := NodeEntry{
		Id:            "nodeA1",
		Ip:            "127.0.0.1",
		GossipVersion: "2",
	}

	serviceA.AddNode(neA1)

	// Activate watch A
	var watchdCiA *ClusterInfo
	serviceA.WatchCluster(func(ci *ClusterInfo, err error) error {
		assert.NotNil(t, ci)
		watchdCiA = ci
		return nil
	}, uint64(123))
	time.Sleep(1 * time.Second)

	neBs := make([]NodeEntry, 5)
	for i := 0; i < len(neBs); i++ {
		strI := strconv.Itoa(i + 1)
		neBs[i] = NodeEntry{
			Id:            "nodeB" + strI,
			Ip:            "192.168.1.10" + strI,
			GossipVersion: "2",
		}
		serviceB.AddNode(neBs[i])
	}

	// Activate watch B
	var watchdCiB *ClusterInfo
	serviceB.WatchCluster(func(ci *ClusterInfo, err error) error {
		assert.NotNil(t, ci)
		watchdCiB = ci
		return nil
	}, uint64(223))
	time.Sleep(1 * time.Second)

	// Validate the configs are in sync between 2 services
	ciA, _ := serviceA.Enumerate()
	ciB, _ := serviceB.Enumerate()

	assert.Equal(t, 6, ciA.Size)
	assert.Equal(t, 6, len(ciA.Nodes))
	assert.Equal(t, 6, ciB.Size)
	assert.Equal(t, 6, len(ciB.Nodes))

	assert.NotNil(t, ciB.Nodes[neA1.Id])
	assert.Equal(t, neA1, ciB.Nodes[neA1.Id])
	assert.NotNil(t, ciA.Nodes[neA1.Id])
	assert.Equal(t, neA1, ciA.Nodes[neA1.Id])

	for i := 0; i < len(neBs); i++ {
		assert.NotNil(t, ciB.Nodes[neBs[i].Id])
		assert.Equal(t, neBs[i], ciB.Nodes[neBs[i].Id])
		assert.NotNil(t, ciA.Nodes[neBs[i].Id])
		assert.Equal(t, neBs[i], ciA.Nodes[neBs[i].Id])
	}
}
