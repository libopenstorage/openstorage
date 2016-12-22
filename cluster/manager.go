// This file implements a cluster state machine.  It relies on a cluster
// wide key-value store for coordinating the state of the cluster.
// It also stores the state of the cluster in this key-value store.
package cluster

import (
	"container/list"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.pedge.io/dlog"

	"github.com/libopenstorage/gossip"
	"github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/config"

	"github.com/Sirupsen/logrus"
	"github.com/libopenstorage/systemutils"
	"github.com/portworx/kvdb"
)

const (
	heartbeatKey     = "heartbeat"
	clusterLockKey   = "/cluster/lock"
	gossipVersionKey = "Gossip Version"
)

var (
	ErrNodeRemovePending = errors.New("Node remove is pending")
	stopHeartbeat        = make(chan bool)
)

type ClusterManager struct {
	size          int
	listeners     *list.List
	config        config.ClusterConfig
	kv            kvdb.Kvdb
	status        api.Status
	nodeCache     map[string]api.Node   // Cached info on the nodes in the cluster.
	nodeStatuses  map[string]api.Status // Set of nodes currently marked down.
	gossip        gossip.Gossiper
	gossipVersion string
	gEnabled      bool
	selfNode      api.Node
	system        systemutils.System
}

func ifaceToIp(iface *net.Interface) (string, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip = ip.To4()
		if ip == nil {
			continue // not an ipv4 address
		}
		if ip.String() == "" {
			continue // address is empty string
		}
		return ip.String(), nil
	}

	return "", errors.New("Node not connected to the network.")
}

func ifaceNameToIp(ifaceName string) (string, error) {
	stdout, err := exec.Command("/usr/sbin/ip", "a", "show", ifaceName, "label", ifaceName).Output()
	if err != nil {
		return "", err
	}
	ipOp := string(stdout)
	// Parse the output of command /usr/bin/ip a show eth0 label eth0:0
	ipOpParts := strings.Fields(ipOp)
	for i, tokens := range ipOpParts {
		// Only check for ipv4 addresses
		if tokens == "inet" {
			ip := ipOpParts[i+1]
			// Remove the mask
			ipAddr := strings.Split(ip, "/")
			if strings.Contains(ipAddr[0], "127") {
				// Loopback address
				continue
			}
			if ipAddr[0] == "" {
				// Address is empty string
				continue
			}
			return ipAddr[0], nil
		}
	}
	return "", fmt.Errorf("Unable to find Ip address for given interface")
}

func ExternalIp(config *config.ClusterConfig) (string, string, error) {
	mgmtIp := ""
	dataIp := ""

	var err error
	if config.MgmtIp == "" && config.MgtIface != "" {
		mgmtIp, err = ifaceNameToIp(config.MgtIface)
		if err != nil {
			return "", "", errors.New("Invalid data network interface " +
				"specified.")
		}
	} else if config.MgmtIp != "" {
		mgmtIp = config.MgmtIp
	}

	if config.DataIp == "" && config.DataIface != "" {
		dataIp, err = ifaceNameToIp(config.DataIface)
		if err != nil {
			return "", "", errors.New("Invalid data network interface " +
				"specified.")
		}
	} else if config.DataIp != "" {
		dataIp = config.DataIp
	}

	if mgmtIp != "" && dataIp != "" {
		return mgmtIp, dataIp, nil
	} else if mgmtIp != "" { // dataIp is empty
		return mgmtIp, mgmtIp, nil
	} else if dataIp != "" { // mgmtIp is empty
		return dataIp, dataIp, nil
	} // both are empty, try to pick first available interface for both

	// No network interface specified, pick first default.
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		mgmtIp, err = ifaceToIp(&iface)
		if err != nil {
			dlog.Printf("Skipping interface without IP: %v: %v",
				iface, err)
			continue
		}
		return mgmtIp, mgmtIp, err
	}

	return "", "", errors.New("Node not connected to the network.")
}

func (c *ClusterManager) Inspect(nodeID string) (api.Node, error) {
	n, ok := c.nodeCache[nodeID]

	if !ok {
		return api.Node{}, errors.New("Unable to locate node with provided UUID.")
	} else {
		return n, nil
	}
}

func (c *ClusterManager) AddEventListener(listener ClusterListener) error {
	dlog.Printf("Adding cluster event listener: %s", listener.String())
	c.listeners.PushBack(listener)
	return nil
}

func (c *ClusterManager) UpdateData(dataKey string, value interface{}) error {
	c.selfNode.NodeData[dataKey] = value
	return nil
}

func (c *ClusterManager) GetData() (map[string]*api.Node, error) {
	nodes := make(map[string]*api.Node)
	for _, value := range c.nodeCache {
		var copyValue api.Node
		copyValue = value
		nodes[value.Id] = &copyValue
	}
	return nodes, nil
}

func (c *ClusterManager) getCurrentState() *api.Node {
	c.selfNode.Timestamp = time.Now()

	c.selfNode.Cpu, _, _ = c.system.CpuUsage()
	c.selfNode.MemTotal, c.selfNode.MemUsed, c.selfNode.MemFree = c.system.MemUsage()

	c.selfNode.Timestamp = time.Now()

	for e := c.listeners.Front(); e != nil; e = e.Next() {
		listenerDataMap := e.Value.(ClusterListener).ListenerData()
		if listenerDataMap == nil {
			continue
		}
		for key, val := range listenerDataMap {
			c.selfNode.NodeData[key] = val
		}
	}

	return &c.selfNode
}

func (c *ClusterManager) getNonDecommisionedPeers(db ClusterInfo) map[types.NodeId]string {
	peers := make(map[types.NodeId]string)
	for _, nodeEntry := range db.NodeEntries {
		if nodeEntry.Status == api.Status_STATUS_DECOMMISSION {
			continue
		}
		ip := nodeEntry.DataIp + ":9002"
		peers[types.NodeId(nodeEntry.Id)] = ip
	}
	return peers
}

// Get the latest config.
func (c *ClusterManager) watchDB(key string, opaque interface{},
	kvp *kvdb.KVPair, err error) error {

	db, err := readClusterInfo()
	if err != nil {
		dlog.Warnln("Failed to read database after update ", err)
		return nil
	}

	for _, nodeEntry := range db.NodeEntries {
		if nodeEntry.Status == api.Status_STATUS_DECOMMISSION {
			logrus.Infof("ClusterManager watchDB, node ID "+
				"%s state is Decommission.",
				nodeEntry.Id)

			n, found := c.nodeCache[nodeEntry.Id]
			if !found {
				logrus.Errorf("ClusterManager watchDB, "+
					"node ID %s not in node cache",
					nodeEntry.Id)
				continue
			}

			if n.Status == api.Status_STATUS_DECOMMISSION {
				logrus.Infof("ClusterManager watchDB, "+
					"node ID %s is already decommission "+
					"on this node",
					nodeEntry.Id)
				continue
			}

			logrus.Infof("ClusterManager watchDB, "+
				"decommsission node ID %s on this node",
				nodeEntry.Id)

			n.Status = api.Status_STATUS_DECOMMISSION
			c.nodeCache[nodeEntry.Id] = n
			// We are getting decommissioned!!
			if nodeEntry.Id == c.selfNode.Id {
				// We are getting decommissioned.
				// Stop the heartbeat and stop the watch
				stopHeartbeat <- true
				c.gossip.Stop(time.Duration(10 * time.Second))
				return fmt.Errorf("stop watch")
			}
		}
	}

	c.size = db.Size

	// Update the peers. A node might have been removed or added
	peers := c.getNonDecommisionedPeers(db)
	c.gossip.UpdateCluster(peers)
	for _, n := range c.nodeCache {
		_, found := peers[types.NodeId(n.Id)]
		if !found {
			delete(c.nodeCache, n.Id)
		}
	}
	return nil
}

func (c *ClusterManager) getLatestNodeConfig(nodeId string) *NodeEntry {
	db, err := readClusterInfo()
	if err != nil {
		dlog.Warnln("Failed to read the database for updating config")
		return nil
	}

	ne, exists := db.NodeEntries[nodeId]
	if !exists {
		dlog.Warnln("Could not find info for node with id ", nodeId)
		return nil
	}

	return &ne
}

func (c *ClusterManager) initNode(db *ClusterInfo) (*api.Node, bool) {
	_, exists := db.NodeEntries[c.selfNode.Id]

	// Add us into the database.
	labels := make(map[string]string)
	labels[gossipVersionKey] = c.gossipVersion
	nodeEntry := NodeEntry{
		Id:         c.selfNode.Id,
		MgmtIp:     c.selfNode.MgmtIp,
		DataIp:     c.selfNode.DataIp,
		GenNumber:  c.selfNode.GenNumber,
		StartTime:  c.selfNode.StartTime,
		MemTotal:   c.selfNode.MemTotal,
		Hostname:   c.selfNode.Hostname,
		NodeLabels: labels,
	}

	db.NodeEntries[c.config.NodeId] = nodeEntry

	dlog.Infof("Node %s joining cluster...", c.config.NodeId)
	dlog.Infof("Cluster ID: %s", c.config.ClusterId)
	dlog.Infof("Node Mgmt IP: %s", c.selfNode.MgmtIp)
	dlog.Infof("Node Data IP: %s", c.selfNode.DataIp)

	return &c.selfNode, exists
}

func (c *ClusterManager) cleanupInit(db *ClusterInfo, self *api.Node) error {
	var resErr error
	var err error

	dlog.Infof("Cleanup Init services")

	for e := c.listeners.Front(); e != nil; e = e.Next() {
		dlog.Warnf("Cleanup Init for service %s.",
			e.Value.(ClusterListener).String())

		err = e.Value.(ClusterListener).CleanupInit(self, db)
		if err != nil {
			dlog.Warnf("Failed to Cleanup Init %s: %v",
				e.Value.(ClusterListener).String(), err)
			resErr = err
		}

	}

	return resErr
}

// Initialize node and alert listeners that we are initializing a node in the cluster.
func (c *ClusterManager) initNodeInCluster(
	clusterInfo *ClusterInfo,
	self *api.Node,
	exist bool,
) error {
	var err error

	// If I am already in the cluster map, don't add me again.
	if exist {
		err = nil
		goto done
	}

	// Alert all listeners that we are a new node and we are initializing.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err = e.Value.(ClusterListener).Init(self, clusterInfo)
		if err != nil {
			if self.Status != api.Status_STATUS_MAINTENANCE {
				self.Status = api.Status_STATUS_ERROR
			}
			dlog.Warnf("Failed to initialize Init %s: %v",
				e.Value.(ClusterListener).String(), err)
			c.cleanupInit(clusterInfo, self)
			goto done
		}
	}
done:
	return err
}

// Alert all listeners that we are joining the cluster
func (c *ClusterManager) joinCluster(
	self *api.Node,
	exist bool,
) error {
	// Listeners may update initial state, so snap again.
	// The cluster db may have diverged since we waited for quorum
	// in between.
	initState, err := snapAndReadClusterInfo()
	defer func() {
		if initState.Collector != nil {
			initState.Collector.Stop()
		}
	}()
	if err != nil {
		return err
	}

	// Alert all listeners that we are joining the cluster.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err := e.Value.(ClusterListener).Join(self, initState, c.HandleNotifications)
		if err != nil {
			if self.Status != api.Status_STATUS_MAINTENANCE {
				self.Status = api.Status_STATUS_ERROR
			}
			dlog.Warnf("Failed to initialize Join %s: %v",
				e.Value.(ClusterListener).String(), err)

			if exist == false {
				c.cleanupInit(initState.ClusterInfo, self)
			}
			dlog.Errorln("Failed to join cluster.", err)
			return err
		}
	}
	selfNodeEntry, ok := initState.ClusterInfo.NodeEntries[c.config.NodeId]
	if !ok {
		dlog.Panicln("Fatal, Unable to find self node entry in local cache")
	}

	_, _, err = c.updateNodeEntryDB(selfNodeEntry)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClusterManager) initClusterForListeners(
	self *api.Node,
) error {
	err := error(nil)

	// Alert all listeners that we are initializing a new cluster.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err = e.Value.(ClusterListener).ClusterInit(self)
		if err != nil {
			if self.Status != api.Status_STATUS_MAINTENANCE {
				self.Status = api.Status_STATUS_ERROR
			}
			dlog.Printf("Failed to initialize %s",
				e.Value.(ClusterListener).String())
			goto done
		}
	}
done:
	return err
}

func (c *ClusterManager) startClusterDBWatch(lastIndex uint64,
	kv kvdb.Kvdb) error {

	go kv.WatchKey(ClusterDBKey, lastIndex, nil, c.watchDB)
	return nil
}

func (c *ClusterManager) startHeartBeat(clusterInfo *ClusterInfo) {
	gossipStoreKey := types.StoreKey(heartbeatKey + c.config.ClusterId)

	node := c.getCurrentState()
	c.nodeCache[c.selfNode.Id] = *node
	c.gossip.UpdateSelf(gossipStoreKey, *node)
	var nodeIps []string
	for nodeId, nodeEntry := range clusterInfo.NodeEntries {
		if nodeId == node.Id {
			continue
		}
		labels := nodeEntry.NodeLabels
		version, ok := labels[gossipVersionKey]
		if !ok || version != c.gossipVersion {
			// Do not add nodes with mismatched version
			continue
		}

		nodeIps = append(nodeIps, nodeEntry.DataIp+":9002")
	}
	if len(nodeIps) > 0 {
		dlog.Infof("Starting Gossip... Gossiping to these nodes : %v", nodeIps)
	} else {
		dlog.Infof("Starting Gossip...")
	}
	c.gossip.Start(nodeIps)
	c.gossip.UpdateCluster(c.getNonDecommisionedPeers(*clusterInfo))

	lastUpdateTs := time.Now()
	for {
		select {
		case <-stopHeartbeat:
			return
		default:
			node = c.getCurrentState()

			currTime := time.Now()
			diffTime := currTime.Sub(lastUpdateTs)
			if diffTime > 10*time.Second {
				dlog.Warnln("No gossip update for ", diffTime.Seconds(), "s")
			}
			c.gossip.UpdateSelf(gossipStoreKey, *node)
			lastUpdateTs = currTime
		}
		time.Sleep(2 * time.Second)
	}
}

func (c *ClusterManager) updateClusterStatus() {
	gossipStoreKey := types.StoreKey(heartbeatKey + c.config.ClusterId)
	for {
		node := c.getCurrentState()
		c.nodeCache[node.Id] = *node

		// Process heartbeats from other nodes...
		gossipValues := c.gossip.GetStoreKeyValue(gossipStoreKey)

		numNodes := 0
		for id, gossipNodeInfo := range gossipValues {
			numNodes = numNodes + 1

			// Check to make sure we are not exceeding the size of the cluster.
			if c.size > 0 && numNodes > c.size {
				dlog.Fatalf("Fatal, number of nodes in the cluster has"+
					"exceeded the cluster size: %d > %d", numNodes, c.size)
				os.Exit(1)
			}

			// Special handling for self node
			if id == types.NodeId(node.Id) {
				if c.selfNode.Status == api.Status_STATUS_OK &&
					(gossipNodeInfo.Status == types.NODE_STATUS_NOT_IN_QUORUM ||
						gossipNodeInfo.Status == types.NODE_STATUS_DOWN) {
					// We have lost quorum
					dlog.Warnf("Not in quorum. Gracefully shutting down...")
					c.gossip.UpdateSelfStatus(types.NODE_STATUS_DOWN)
					c.selfNode.Status = api.Status_STATUS_OFFLINE
					c.status = api.Status_STATUS_NOT_IN_QUORUM
					c.Shutdown()
					os.Exit(1)
				}
				//else Ignore this update
				continue
			}

			// Notify node status change if required.
			peerNodeInCache := api.Node{}
			peerNodeInCache.Id = string(id)
			peerNodeInCache.Status = api.Status_STATUS_OK

			switch {
			case gossipNodeInfo.Status == types.NODE_STATUS_DOWN:
				// Replace the status of this node in cache to offline
				peerNodeInCache.Status = api.Status_STATUS_OFFLINE
				lastStatus, ok := c.nodeStatuses[string(id)]
				if ok && lastStatus == peerNodeInCache.Status {
					break
				}

				c.nodeStatuses[string(id)] = peerNodeInCache.Status

				dlog.Warnln("Detected node ", id,
					" to be offline due to inactivity.")

				for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
					err := e.Value.(ClusterListener).Update(&peerNodeInCache)
					if err != nil {
						dlog.Warnln("Failed to notify ",
							e.Value.(ClusterListener).String())
					}
				}

			case gossipNodeInfo.Status == types.NODE_STATUS_UP:
				peerNodeInCache.Status = api.Status_STATUS_OK
				lastStatus, ok := c.nodeStatuses[string(id)]
				if ok && lastStatus == peerNodeInCache.Status {
					break
				}
				c.nodeStatuses[string(id)] = peerNodeInCache.Status

				// A node discovered in the cluster.
				dlog.Infoln("Detected node", peerNodeInCache.Id,
					" to be in the cluster.")

				for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
					err := e.Value.(ClusterListener).Add(&peerNodeInCache)
					if err != nil {
						dlog.Warnln("Failed to notify ",
							e.Value.(ClusterListener).String())
					}
				}
			}

			// Update cache with gossip data
			if gossipNodeInfo.Value != nil {
				peerNodeInGossip, ok := gossipNodeInfo.Value.(api.Node)
				if ok {
					if peerNodeInCache.Status == api.Status_STATUS_OFFLINE {
						// Overwrite the status of Node in Gossip data with Down
						peerNodeInGossip.Status = peerNodeInCache.Status
					} else {
						if peerNodeInGossip.Status == api.Status_STATUS_MAINTENANCE {
							// If the node sent its status as Maintenance
							// do not overwrite it with online
						} else {
							peerNodeInGossip.Status = peerNodeInCache.Status
						}
					}
					c.nodeCache[peerNodeInGossip.Id] = peerNodeInGossip
				} else {
					dlog.Errorln("Unable to get node info from gossip")
					c.nodeCache[peerNodeInCache.Id] = peerNodeInCache
				}
			} else {
				c.nodeCache[peerNodeInCache.Id] = peerNodeInCache
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func (c *ClusterManager) DisableUpdates() error {
	dlog.Warnln("Disabling gossip updates")
	c.gEnabled = false

	return nil
}

func (c *ClusterManager) EnableUpdates() error {
	dlog.Warnln("Enabling gossip updates")
	c.gEnabled = true

	return nil
}

func (c *ClusterManager) GetGossipState() *ClusterState {
	gossipStoreKey := types.StoreKey(heartbeatKey + c.config.ClusterId)
	nodeValue := c.gossip.GetStoreKeyValue(gossipStoreKey)
	nodes := make([]types.NodeValue, len(nodeValue), len(nodeValue))
	i := 0
	for _, value := range nodeValue {
		nodes[i] = value
		i++
	}

	history := c.gossip.GetGossipHistory()
	return &ClusterState{
		History: history, NodeStatus: nodes}
}

func (c *ClusterManager) waitForQuorum(exist bool) error {
	// Max quorum retries allowed = 600
	// 600 * 2 seconds (gossip interval) = 20 minutes before it restarts
	quorumRetries := 0
	for {
		gossipSelfStatus := c.gossip.GetSelfStatus()
		if c.selfNode.Status == api.Status_STATUS_NOT_IN_QUORUM &&
			gossipSelfStatus == types.NODE_STATUS_UP {
			// Node not initialized yet
			// Achieved quorum in the cluster.
			// Lets start the node
			c.selfNode.Status = api.Status_STATUS_INIT
			err := c.joinCluster(&c.selfNode, exist)
			if err != nil {
				if c.selfNode.Status != api.Status_STATUS_MAINTENANCE {
					c.selfNode.Status = api.Status_STATUS_ERROR
				}
				return err
			}
			c.status = api.Status_STATUS_OK
			c.selfNode.Status = api.Status_STATUS_OK
			break
		} else {
			c.status = api.Status_STATUS_NOT_IN_QUORUM
			if quorumRetries == 600 {
				err := fmt.Errorf("Unable to achieve Quorum." +
					" Timeout 20 minutes exceeded.")
				dlog.Warnln("Failed to join cluster: ", err)
				c.status = api.Status_STATUS_NOT_IN_QUORUM
				c.selfNode.Status = api.Status_STATUS_OFFLINE
				c.gossip.UpdateSelfStatus(types.NODE_STATUS_DOWN)
				return err
			}
			if quorumRetries == 0 {
				dlog.Infof("Waiting for the cluster to reach quorum...")
			}
			time.Sleep(types.DEFAULT_GOSSIP_INTERVAL)
			quorumRetries++
		}
	}
	return nil
}

func (c *ClusterManager) initializeCluster(db kvdb.Kvdb) (
	*ClusterInfo,
	error,
) {
	kvlock, err := db.LockWithID(clusterLockKey, c.config.NodeId)
	if err != nil {
		dlog.Panicln("Fatal, Unable to obtain cluster lock.", err)
	}
	defer func() {
		db.Unlock(kvlock)
	}()

	clusterInfo, err := readClusterInfo()
	if err != nil {
		dlog.Panicln(err)
	}

	selfNodeEntry, ok := clusterInfo.NodeEntries[c.config.NodeId]
	if ok && selfNodeEntry.Status == api.Status_STATUS_DECOMMISSION {
		msg := fmt.Sprintf("Node is in decommision state, Node ID %s.",
			c.selfNode.Id)
		dlog.Errorln(msg)
		return nil, errors.New(msg)
	}
	// Set the clusterID in db
	clusterInfo.Id = c.config.ClusterId

	if clusterInfo.Status == api.Status_STATUS_INIT {
		dlog.Infoln("Initializing a new cluster.")
		// Initialize self node
		clusterInfo.Status = api.Status_STATUS_OK

		err = c.initClusterForListeners(&c.selfNode)
		if err != nil {
			dlog.Errorln("Failed to initialize the cluster.", err)
			return nil, err
		}
		// While we hold the lock write the cluster info
		// to kvdb.
		_, err := writeClusterInfo(&clusterInfo)
		if err != nil {
			dlog.Errorln("Failed to initialize the cluster.", err)
			return nil, err
		}
	} else if clusterInfo.Status&api.Status_STATUS_OK > 0 {
		dlog.Infoln("Cluster state is OK... Joining the cluster.")
	} else {
		return nil, errors.New("Fatal, Cluster is in an unexpected state.")
	}
	// Cluster database max size... 0 if unlimited.
	c.size = clusterInfo.Size
	c.status = api.Status_STATUS_OK
	return &clusterInfo, nil
}

func (c *ClusterManager) initListeners(db kvdb.Kvdb, nodeExists *bool) (
	uint64,
	*ClusterInfo,
	error,
) {
	// Initialize the cluster if required
	clusterInfo, err := c.initializeCluster(db)
	if err != nil {
		return 0, nil, err
	}

	// Initialize the node in cluster
	self, exist := c.initNode(clusterInfo)
	*nodeExists = exist
	err = c.initNodeInCluster(clusterInfo, self, *nodeExists)
	if err != nil {
		dlog.Errorln("Failed to initialize node in cluster.", err)
		return 0, nil, err
	}

	selfNodeEntry, ok := clusterInfo.NodeEntries[c.config.NodeId]
	if !ok {
		dlog.Panicln("Fatal, Unable to find self node entry in local cache")
	}

	kvp, kvClusterInfo, err := c.updateNodeEntryDB(selfNodeEntry)
	if err != nil {
		dlog.Errorln("Failed to save the database.", err)
		return 0, nil, err
	}
	if kvClusterInfo.Status == api.Status_STATUS_INIT {
		dlog.Panicln("Cluster in an unexpected state: ", kvClusterInfo.Status)
	}
	return kvp.ModifiedIndex, kvClusterInfo, nil
}

func (c *ClusterManager) initializeAndStartHeartbeat(kvdb kvdb.Kvdb, exist *bool) (uint64, error) {
	lastIndex, clusterInfo, err := c.initListeners(kvdb, exist)
	if err != nil {
		return 0, err
	}

	// Set the status to NOT_IN_QUORUM to start the node.
	// Once we achieve quorum then we actually join the cluster
	// and change the status to OK
	c.selfNode.Status = api.Status_STATUS_NOT_IN_QUORUM
	// Start heartbeating to other nodes.
	go c.startHeartBeat(clusterInfo)
	return lastIndex, nil
}

func (c *ClusterManager) Start() error {
	var err error

	dlog.Infoln("Cluster manager starting...")

	c.gEnabled = true
	c.selfNode = api.Node{}
	c.selfNode.GenNumber = uint64(time.Now().UnixNano())
	c.selfNode.Id = c.config.NodeId
	c.selfNode.Status = api.Status_STATUS_INIT
	c.selfNode.MgmtIp, c.selfNode.DataIp, err = ExternalIp(&c.config)
	c.selfNode.StartTime = time.Now()
	c.selfNode.Hostname, _ = os.Hostname()
	if err != nil {
		dlog.Errorf("Failed to get external IP address for mgt/data interfaces: %s.",
			err)
		return err
	}

	c.selfNode.NodeData = make(map[string]interface{})
	c.system = systemutils.New()

	// Start the gossip protocol.
	// XXX Make the port configurable.
	gob.Register(api.Node{})
	gossipIntervals := types.GossipIntervals{
		GossipInterval:   types.DEFAULT_GOSSIP_INTERVAL,
		PushPullInterval: types.DEFAULT_PUSH_PULL_INTERVAL,
		ProbeInterval:    types.DEFAULT_PROBE_INTERVAL,
		ProbeTimeout:     types.DEFAULT_PROBE_TIMEOUT,
		QuorumTimeout:    types.DEFAULT_QUORUM_TIMEOUT,
	}
	c.gossip = gossip.New(
		c.selfNode.DataIp+":9002",
		types.NodeId(c.config.NodeId),
		c.selfNode.GenNumber,
		gossipIntervals,
		types.GOSSIP_VERSION_2,
		c.config.ClusterId,
	)
	c.gossipVersion = types.GOSSIP_VERSION_2

	var exist bool
	kvdb := kvdb.Instance()

	lastIndex, err := c.initializeAndStartHeartbeat(kvdb, &exist)
	if err != nil {
		return err
	}

	c.startClusterDBWatch(lastIndex, kvdb)

	err = c.waitForQuorum(exist)
	if err != nil {
		return err
	}

	go c.updateClusterStatus()
	go c.replayNodeDecommission()

	return nil
}

func (c *ClusterManager) NodeStatus(listenerName string) (api.Status, error) {
	clusterNodeStatus := c.selfNode.Status
	if clusterNodeStatus != api.Status_STATUS_OK {
		// Status of this node as seen by Cluster Manager is not OK
		// This takes highest precedence over other listener statuses.
		// Returning our status
		return clusterNodeStatus, nil
	}
	if listenerName == "" {
		return clusterNodeStatus, nil
	}

	listenerStatus := api.Status_STATUS_NONE
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		if e.Value.(ClusterListener).String() == listenerName {
			listenerStatus = e.Value.(ClusterListener).ListenerStatus()
			break
		}
	}
	if int(listenerStatus.StatusKind()) >= int(clusterNodeStatus.StatusKind()) {
		return listenerStatus, nil
	}
	return clusterNodeStatus, nil
}

func (c *ClusterManager) PeerStatus(listenerName string) (map[string]api.Status, error) {
	statusMap := make(map[string]api.Status)
	var listenerStatusMap map[string]api.Status
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		if e.Value.(ClusterListener).String() == listenerName {
			listenerStatusMap = e.Value.(ClusterListener).ListenerPeerStatus()
			break
		}
	}
	// Listener failed to provide peer status
	if listenerStatusMap == nil || len(listenerStatusMap) == 0 {
		for _, n := range c.nodeCache {
			if n.Id == c.selfNode.Id {
				// skip self
				continue
			}
			statusMap[n.Id] = n.Status
		}
		return statusMap, nil
	}
	// Compare listener's peer statuses and cluster provider's peer statuses
	for _, n := range c.nodeCache {
		if n.Id == c.selfNode.Id {
			// Skip self
			continue
		}
		clusterNodeStatus := n.Status
		listenerNodeStatus, ok := listenerStatusMap[n.Id]
		if !ok {
			// Could not find listener's peer status
			// Using cluster provider's peer status
			statusMap[n.Id] = clusterNodeStatus
		}
		if int(listenerNodeStatus.StatusKind()) >= int(clusterNodeStatus.StatusKind()) {
			// Use listener's peer status
			statusMap[n.Id] = listenerNodeStatus
		} else {
			// Use the cluster provider's peer status
			statusMap[n.Id] = clusterNodeStatus
		}
	}
	return statusMap, nil
}

// Enumerate lists all the nodes in the cluster.
func (c *ClusterManager) Enumerate() (api.Cluster, error) {
	i := 0

	cluster := api.Cluster{
		Id:     c.config.ClusterId,
		Status: c.status,
		NodeId: c.selfNode.Id,
	}
	var clusterDB ClusterInfo
	clusterDBSet := false
	cluster.Nodes = make([]api.Node, len(c.nodeCache))
	for _, n := range c.nodeCache {
		if n.Id == c.selfNode.Id {
			cluster.Nodes[i] = *c.getCurrentState()
		} else {
			if n.Status == api.Status_STATUS_OFFLINE &&
				(n.DataIp == "" || n.MgmtIp == "") {
				if !clusterDBSet {
					clusterDB, _ = readClusterInfo()
					clusterDBSet = true
				}
				// Gossip does not have essential information of
				// an offline node. Provide the essential data
				// that we have in the cluster db
				nodeValueDB, ok := clusterDB.NodeEntries[n.Id]
				if ok {
					n.MgmtIp = nodeValueDB.MgmtIp
					n.DataIp = nodeValueDB.DataIp
					n.Hostname = nodeValueDB.Hostname
					n.NodeLabels = nodeValueDB.NodeLabels
				}
			}
			cluster.Nodes[i] = n
		}
		i++
	}

	return cluster, nil
}

func (c *ClusterManager) updateNodeEntryDB(nodeEntry NodeEntry) (*kvdb.KVPair, *ClusterInfo, error) {
	kvdb := kvdb.Instance()
	kvlock, err := kvdb.LockWithID(clusterLockKey, c.config.NodeId)
	if err != nil {
		dlog.Warnln("Unable to obtain cluster lock for updating cluster DB.",
			err)
		return nil, nil, err
	}
	defer kvdb.Unlock(kvlock)

	currentState, err := readClusterInfo()
	if err != nil {
		return nil, nil, err
	}
	currentState.NodeEntries[nodeEntry.Id] = nodeEntry
	kvp, err := writeClusterInfo(&currentState)
	if err != nil {
		dlog.Errorln("Failed to save the database.", err)
	}
	return kvp, &currentState, err
}

// SetSize sets the maximum number of nodes in a cluster.
func (c *ClusterManager) SetSize(size int) error {
	kvdb := kvdb.Instance()
	kvlock, err := kvdb.LockWithID(clusterLockKey, c.config.NodeId)
	if err != nil {
		dlog.Warnln("Unable to obtain cluster lock for updating config", err)
		return nil
	}
	defer kvdb.Unlock(kvlock)

	db, err := readClusterInfo()
	if err != nil {
		return err
	}

	db.Size = size

	_, err = writeClusterInfo(&db)

	return err
}

func (c *ClusterManager) markNodeDecommission(node api.Node) error {
	kvdb := kvdb.Instance()
	kvlock, err := kvdb.LockWithID(clusterLockKey, c.config.NodeId)
	if err != nil {
		dlog.Warnln("Unable to obtain cluster lock for marking "+
			"node decommission",
			err)
		return err
	}
	defer kvdb.Unlock(kvlock)

	db, err := readClusterInfo()
	if err != nil {
		return err
	}

	nodeEntry, ok := db.NodeEntries[node.Id]
	if !ok {
		msg := fmt.Sprintf("Node entry does not exist, Node ID %s",
			node.Id)
		return errors.New(msg)
	}

	nodeEntry.Status = api.Status_STATUS_DECOMMISSION
	db.NodeEntries[node.Id] = nodeEntry

	if c.selfNode.Id == node.Id {
		c.selfNode.Status = api.Status_STATUS_DECOMMISSION
	}
	_, err = writeClusterInfo(&db)

	return err
}

func (c *ClusterManager) deleteNodeFromDB(nodeID string) error {
	// Delete node from cluster DB
	kvdb := kvdb.Instance()
	kvlock, err := kvdb.LockWithID(clusterLockKey, c.config.NodeId)
	if err != nil {
		dlog.Panicln("fatal, unable to obtain cluster lock. ", err)
	}
	defer kvdb.Unlock(kvlock)

	currentState, err := readClusterInfo()
	if err != nil {
		dlog.Errorln("Failed to read cluster info. ", err)
		return err
	}

	delete(currentState.NodeEntries, nodeID)

	_, err = writeClusterInfo(&currentState)
	if err != nil {
		dlog.Errorln("Failed to save the database.", err)
	}
	return err
}

// Remove node(s) from the cluster permanently.
func (c *ClusterManager) Remove(nodes []api.Node) error {
	logrus.Infof("ClusterManager Remove node.")

	var resultErr error

	for _, n := range nodes {
		if _, exist := c.nodeCache[n.Id]; !exist {
			msg := fmt.Sprintf("Node does not exist in cluster, "+
				"Node ID %s.",
				n.Id)
			dlog.Errorf(msg)
			return errors.New(msg)
		}

		// If removing node is self and node is not in maintenance mode,
		// disallow node remove.
		if n.Id == c.selfNode.Id &&
			c.selfNode.Status != api.Status_STATUS_MAINTENANCE {
			msg := fmt.Sprintf("Cannot remove self from cluster, "+
				"Node ID %s.",
				n.Id)
			dlog.Errorf(msg)
			return errors.New(msg)
		} else if n.Id != c.selfNode.Id {
			nodeCacheStatus := c.nodeCache[n.Id].Status
			// If node is not down, do not remove it
			if nodeCacheStatus != api.Status_STATUS_OFFLINE &&
				nodeCacheStatus != api.Status_STATUS_MAINTENANCE &&
				nodeCacheStatus != api.Status_STATUS_DECOMMISSION {

				msg := fmt.Sprintf("Cannot remove node that is not "+
					"offline, Node ID %s, "+
					"Status %s.",
					n.Id, nodeCacheStatus)
				dlog.Errorf(msg)
				return errors.New(msg)
			}

		}

		// Ask listeners, can we remove this node?
		for e := c.listeners.Front(); e != nil; e = e.Next() {
			dlog.Infof("Remove node: ask cluster listener: "+
				"can we remove node ID %s, %s",
				n.Id, e.Value.(ClusterListener).String())

			err := e.Value.(ClusterListener).CanNodeRemove(&n)
			if err != nil {
				msg := fmt.Sprintf("Cannot remove node ID %s: %s",
					n.Id, err)
				dlog.Warnf(msg)
				return errors.New(msg)
			}

		}

		err := c.markNodeDecommission(n)
		if err != nil {
			msg := fmt.Sprintf("Failed to mark node as "+
				"decommision, error %s",
				err)
			dlog.Errorf(msg)
			return errors.New(msg)
		}

		// Alert all listeners that we are removing this node.
		for e := c.listeners.Front(); e != nil; e = e.Next() {
			dlog.Infof("Remove node: notify cluster listener: %s",
				e.Value.(ClusterListener).String())
			err := e.Value.(ClusterListener).Remove(&n)
			if err != nil {
				if err != ErrNodeRemovePending {
					dlog.Warnf("Cluster listener failed to "+
						"remove node: %s: %s",
						e.Value.(ClusterListener).String(),
						err)
					return err
				} else {
					resultErr = err
				}
			}
		}
	}

	return resultErr
}

func (c *ClusterManager) NodeRemoveDone(nodeID string, result error) {
	// XXX: only storage will make callback right now
	if result != nil {
		msg := fmt.Sprintf("Storage failed to decommission node %s, "+
			"error %s",
			nodeID,
			result)
		logrus.Errorf(msg)
		return
	}

	logrus.Infof("Cluster manager node remove done: node ID %s", nodeID)

	err := c.deleteNodeFromDB(nodeID)
	if err != nil {
		msg := fmt.Sprintf("Failed to delete node %s "+
			"from cluster database, error %s",
			nodeID, err)
		dlog.Errorf(msg)
	}
}

func (c *ClusterManager) replayNodeDecommission() {
	currentState, err := readClusterInfo()
	if err != nil {
		dlog.Infof("Failed to read cluster db for node decommissions: %v", err)
		return
	}

	for _, nodeEntry := range currentState.NodeEntries {
		if nodeEntry.Status == api.Status_STATUS_DECOMMISSION {
			dlog.Infof("Replay Node Remove for node ID %s", nodeEntry.Id)

			var n api.Node
			n.Id = nodeEntry.Id
			nodes := make([]api.Node, 0)
			nodes = append(nodes, n)
			err := c.Remove(nodes)
			if err != nil {
				dlog.Warnf("Failed to replay node remove: "+
					"node ID %s, error %s",
					nodeEntry.Id, err)
			}
		}
	}
}

// Shutdown can be called when THIS node is gracefully shutting down.
func (c *ClusterManager) Shutdown() error {
	db, err := readClusterInfo()
	if err != nil {
		dlog.Warnf("Could not read cluster database (%v).", err)
		return err
	}

	// Alert all listeners that we are shutting this node down.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		dlog.Infof("Shutting down %s", e.Value.(ClusterListener).String())
		if err := e.Value.(ClusterListener).Halt(&c.selfNode, &db); err != nil {
			dlog.Warnf("Failed to shutdown %s",
				e.Value.(ClusterListener).String())
		}
	}
	return nil
}

func (c *ClusterManager) HandleNotifications(culpritNodeId string, notification api.ClusterNotify) (string, error) {
	if notification == api.ClusterNotify_CLUSTER_NOTIFY_DOWN {
		killNodeId := c.gossip.ExternalNodeLeave(types.NodeId(culpritNodeId))
		return string(killNodeId), nil
	} else {
		return "", fmt.Errorf("Error in Handle Notifications. Unknown Notification : %v", notification)
	}
}
