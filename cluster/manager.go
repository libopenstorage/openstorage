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
	heartbeatKey   = "heartbeat"
	clusterLockKey = "/cluster/lock"
)

type ClusterManager struct {
	size         int
	listeners    *list.List
	config       config.ClusterConfig
	kv           kvdb.Kvdb
	status       api.Status
	nodeCache    map[string]api.Node   // Cached info on the nodes in the cluster.
	nodeStatuses map[string]api.Status // Set of nodes currently marked down.
	gossip       gossip.Gossiper
	gEnabled     bool
	selfNode     api.Node
	system       systemutils.System
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

func ExternalIp(config *config.ClusterConfig) (string, string, error) {
	mgmtIp := ""
	dataIp := ""

	if config.MgtIface != "" {
		iface, err := net.InterfaceByName(config.MgtIface)
		if err != nil {
			return "", "", errors.New("Invalid management network " +
				"interface specified.")
		}
		mgmtIp, err = ifaceToIp(iface)
		if err != nil {
			return "", "", err
		}
	}

	if config.DataIface != "" {
		iface, err := net.InterfaceByName(config.DataIface)
		if err != nil {
			return "", "", errors.New("Invalid data network interface " +
				"specified.")
		}
		dataIp, err = ifaceToIp(iface)
		if err != nil {
			return "", "", err
		}
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
		nodes[value.Id] = &value
	}
	return nodes, nil
}

func (c *ClusterManager) getCurrentState() *api.Node {
	c.selfNode.Timestamp = time.Now()

	c.selfNode.Cpu, _, _ = c.system.CpuUsage()
	c.selfNode.MemTotal, c.selfNode.MemUsed, c.selfNode.MemFree = c.system.MemUsage()
	c.selfNode.Luns = c.system.Luns()

	c.selfNode.Timestamp = time.Now()

	return &c.selfNode
}

func (c *ClusterManager) getPeers(db ClusterInfo) map[types.NodeId]string {
	peers := make(map[types.NodeId]string)
	for _, nodeEntry := range db.NodeEntries {
		ip := nodeEntry.MgmtIp + ":9002"
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

	// The only value we rely on during an update is the cluster size.
	c.size = db.Size

	// Probably new node was added into the cluster db
	peers := c.getPeers(db)
	c.gossip.UpdateCluster(peers)
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
	c.nodeCache[c.selfNode.Id] = *c.getCurrentState()

	_, exists := db.NodeEntries[c.selfNode.Id]

	// Add us into the database.
	db.NodeEntries[c.config.NodeId] = NodeEntry{
		Id:        c.selfNode.Id,
		MgmtIp:    c.selfNode.MgmtIp,
		DataIp:    c.selfNode.DataIp,
		GenNumber: c.selfNode.GenNumber,
		StartTime: c.selfNode.StartTime,
		MemTotal:  c.selfNode.MemTotal,
		Hostname:  c.selfNode.Hostname,
	}

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
	initState *ClusterInitState,
	self *api.Node,
	exist bool,
) error {
	var err error
	var newInitState *ClusterInitState

	// If I am already in the cluster map, don't add me again.
	if exist {
		err = nil
		goto done
	}

	// Alert all listeners that we are a new node and we are initializing.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err = e.Value.(ClusterListener).Init(self, initState)
		if err != nil {
			self.Status = api.Status_STATUS_ERROR
			dlog.Warnf("Failed to initialize Init %s: %v",
				e.Value.(ClusterListener).String(), err)
			c.cleanupInit(initState.ClusterInfo, self)
			goto done
		}
	}

	// Listeners may update initial state, so snap again
	newInitState, err = snapAndReadClusterInfo()
	initState.InitDb = newInitState.InitDb
	initState.Version = newInitState.Version

done:
	return err
}

// Alert all listeners that we are joining the cluster
func (c *ClusterManager) joinCluster(
	initState *ClusterInitState,
	self *api.Node,
	exist bool,
) error {
	// Alert all listeners that we are joining the cluster.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err := e.Value.(ClusterListener).Join(self, initState, c.HandleNotifications)
		if err != nil {
			self.Status = api.Status_STATUS_ERROR
			dlog.Warnf("Failed to initialize Join %s: %v",
				e.Value.(ClusterListener).String(), err)

			if exist == false {
				c.cleanupInit(initState.ClusterInfo, self)
			}
			dlog.Errorln("Failed to join cluster.", err)
			return err
		}
	}
	return nil
}

func (c *ClusterManager) initCluster(
	initState *ClusterInitState,
	self *api.Node,
	exist bool,
) error {
	err := error(nil)

	// Alert all listeners that we are initializing a new cluster.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err = e.Value.(ClusterListener).ClusterInit(self, initState)
		if err != nil {
			self.Status = api.Status_STATUS_ERROR
			dlog.Printf("Failed to initialize %s",
				e.Value.(ClusterListener).String())
			goto done
		}
	}

	err = c.initNodeInCluster(initState, self, exist)
	if err != nil {
		dlog.Printf("Failed to join new cluster")
		goto done
	}

done:
	return err
}

func (c *ClusterManager) startHeartBeat(clusterInfo *ClusterInfo) {
	gossipStoreKey := types.StoreKey(heartbeatKey + c.config.ClusterId)

	node := c.getCurrentState()
	c.gossip.UpdateSelf(gossipStoreKey, *node)
	var nodeIps []string
	for nodeId, nodeEntry := range clusterInfo.NodeEntries {
		if nodeId == node.Id {
			continue
		}
		nodeIps = append(nodeIps, nodeEntry.MgmtIp+":9002")
	}
	if len(nodeIps) > 0 {
		dlog.Infof("Starting Gossip... Gossiping to these nodes : %v", nodeIps)
	} else {
		dlog.Infof("Starting Gossip...")
	}
	c.gossip.Start(nodeIps)
	peers := c.getPeers(*clusterInfo)
	c.gossip.UpdateCluster(peers)

	lastUpdateTs := time.Now()
	for {
		node = c.getCurrentState()

		currTime := time.Now()
		diffTime := currTime.Sub(lastUpdateTs)
		if diffTime > 10*time.Second {
			dlog.Warnln("No gossip update for ", diffTime.Seconds(), "s")
		}
		c.gossip.UpdateSelf(gossipStoreKey, *node)
		lastUpdateTs = currTime

		time.Sleep(2 * time.Second)
	}
}

func (c *ClusterManager) updateClusterStatus(initState *ClusterInitState, exist bool) {
	gossipStoreKey := types.StoreKey(heartbeatKey + c.config.ClusterId)

	for {
		node := c.getCurrentState()
		c.nodeCache[node.Id] = *node

		// Process heartbeats from other nodes...
		gossipValues := c.gossip.GetStoreKeyValue(gossipStoreKey)

		numNodes := 0
		for id, nodeInfo := range gossipValues {
			numNodes = numNodes + 1

			// Check to make sure we are not exceeding the size of the cluster.
			if c.size > 0 && numNodes > c.size {
				dlog.Fatalf("Fatal, number of nodes in the cluster has"+
					"exceeded the cluster size: %d > %d", numNodes, c.size)
				os.Exit(-1)
			}

			// Special handling for self node
			if id == types.NodeId(node.Id) {
				if c.selfNode.Status == api.Status_STATUS_OK &&
					(nodeInfo.Status == types.NODE_STATUS_NOT_IN_QUORUM ||
						nodeInfo.Status == types.NODE_STATUS_DOWN) {
					// We have lost quorum
					dlog.Warnf("Not in quorum. Gracefully shutting down...")
					c.gossip.UpdateSelfStatus(types.NODE_STATUS_DOWN)
					c.selfNode.Status = api.Status_STATUS_OFFLINE
					c.status = api.Status_STATUS_NOT_IN_QUORUM
					c.Shutdown()
					os.Exit(-1)
				} else if c.selfNode.Status == api.Status_STATUS_OFFLINE &&
					nodeInfo.Status == types.NODE_STATUS_UP {
					dlog.Infof("Back in quorum..")
					err := c.joinCluster(initState, &c.selfNode, exist)
					if err == nil {
						c.selfNode.Status = api.Status_STATUS_OK
					}
				}
				//else Ignore this update
				continue
			}

			// Notify node status change if required.
			newNodeInfo := api.Node{}
			newNodeInfo.Id = string(id)
			newNodeInfo.Status = api.Status_STATUS_OK

			switch {
			case nodeInfo.Status == types.NODE_STATUS_DOWN:
				newNodeInfo.Status = api.Status_STATUS_OFFLINE
				lastStatus, ok := c.nodeStatuses[string(id)]
				if ok && lastStatus == newNodeInfo.Status {
					break
				}

				c.nodeStatuses[string(id)] = newNodeInfo.Status

				dlog.Warnln("Detected node ", id,
					" to be offline due to inactivity.")

				for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
					err := e.Value.(ClusterListener).Update(&newNodeInfo)
					if err != nil {
						dlog.Warnln("Failed to notify ",
							e.Value.(ClusterListener).String())
					}
				}

			case nodeInfo.Status == types.NODE_STATUS_UP:
				newNodeInfo.Status = api.Status_STATUS_OK
				lastStatus, ok := c.nodeStatuses[string(id)]
				if ok && lastStatus == newNodeInfo.Status {
					break
				}
				c.nodeStatuses[string(id)] = newNodeInfo.Status

				// A node discovered in the cluster.
				dlog.Warnln("Detected node ", newNodeInfo.Id,
					" to be in the cluster.")

				for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
					err := e.Value.(ClusterListener).Add(&newNodeInfo)
					if err != nil {
						dlog.Warnln("Failed to notify ",
							e.Value.(ClusterListener).String())
					}
				}
			}

			// Update cache.
			if nodeInfo.Value != nil {
				n, ok := nodeInfo.Value.(api.Node)
				if ok {
					n.Status = newNodeInfo.Status
					c.nodeCache[n.Id] = n
				} else {
					dlog.Errorln("Unable to get node info from gossip")
					c.nodeCache[newNodeInfo.Id] = newNodeInfo
				}
			} else {
				c.nodeCache[newNodeInfo.Id] = newNodeInfo
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

func (c *ClusterManager) GetState() (*ClusterState, error) {
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
		History: history, NodeStatus: nodes}, nil
}

func (c *ClusterManager) Start() error {
	var err error

	dlog.Infoln("Cluster manager starting...")

	c.gEnabled = true
	c.selfNode = api.Node{}
	c.selfNode.GenNumber = uint64(time.Now().UnixNano())
	c.selfNode.Id = c.config.NodeId
	// Set the status to NOT_IN_QUORUM to start the node.
	// Once we achieve quorum then we actually join the cluster
	// and change the status to OK
	c.selfNode.Status = api.Status_STATUS_NOT_IN_QUORUM
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
		c.selfNode.MgmtIp+":9002",
		types.NodeId(c.config.NodeId),
		c.selfNode.GenNumber,
		gossipIntervals,
		types.DEFAULT_GOSSIP_VERSION,
	)

	kvdb := kvdb.Instance()
	kvlock, err := kvdb.Lock(clusterLockKey)
	if err != nil {
		dlog.Panicln("Fatal, Unable to obtain cluster lock.", err)
	}

	initState, err := snapAndReadClusterInfo()
	if err != nil {
		dlog.Panicln(err)
	}
	kvdb.Unlock(kvlock)

	// Cluster database max size... 0 if unlimited.
	c.size = initState.ClusterInfo.Size
	// Set the clusterID in db
	initState.ClusterInfo.Id = c.config.ClusterId

	var exist bool
	if initState.ClusterInfo.Status == api.Status_STATUS_INIT {
		dlog.Infoln("Will initialize a new cluster.")

		c.status = api.Status_STATUS_OK
		initState.ClusterInfo.Status = api.Status_STATUS_OK
		self, _ := c.initNode(initState.ClusterInfo)
		err = c.initCluster(initState, self, false)
		if err != nil {
			dlog.Errorln("Failed to initialize the cluster.", err)
			return err
		}
	} else if initState.ClusterInfo.Status&api.Status_STATUS_OK > 0 {
		dlog.Infoln("Cluster state is OK... Joining the cluster.")

		c.status = api.Status_STATUS_OK

		self, exist := c.initNode(initState.ClusterInfo)

		err = c.initNodeInCluster(initState, self, exist)
		if err != nil {
			dlog.Errorln("Failed to initialize node in cluster.", err)
			return err
		}
	} else {
		return errors.New("Fatal, Cluster is in an unexpected state.")
	}

	kvlock, err = kvdb.Lock(clusterLockKey)
	if err != nil {
		dlog.Panicln("Fatal, Unable to obtain cluster lock. ", err)
	}
	selfNodeEntry, ok := initState.ClusterInfo.NodeEntries[c.config.NodeId]
	if !ok {
		kvdb.Unlock(kvlock)
		dlog.Panicln("Fatal, Unable to find self node entry in local cache")
	}
	// Add ourselves into the cluster DB and release the lock
	currentState, err := readClusterInfo()
	if err != nil {
		kvdb.Unlock(kvlock)
		dlog.Errorln("Failed to read cluster info. ", err)
		return err
	}
	currentState.NodeEntries[c.config.NodeId] = selfNodeEntry
	if currentState.Status == api.Status_STATUS_INIT {
		// We are the first node to join the cluster.
		currentState.Status = api.Status_STATUS_OK
	}
	err = writeClusterInfo(&currentState)
	if err != nil {
		dlog.Errorln("Failed to save the database.", err)
		kvdb.Unlock(kvlock)
		return err
	}
	kvdb.Unlock(kvlock)

	// Start heartbeating to other nodes.
	go c.startHeartBeat(&currentState)

	// Max quorum retries allowed = 30
	// 30 * 2 seconds (gossip interval) = 1 minute (quorum timeout)
	quorumRetries := 0
	for {
		gossipSelfStatus := c.gossip.GetSelfStatus()
		if c.selfNode.Status == api.Status_STATUS_NOT_IN_QUORUM &&
			gossipSelfStatus == types.NODE_STATUS_UP {
			// Node not initialized yet
			// Achieved quorum in the cluster.
			// Lets start the node
			err := c.joinCluster(initState, &c.selfNode, exist)
			if err != nil {
				return err
			}
			c.status = api.Status_STATUS_OK
			c.selfNode.Status = api.Status_STATUS_OK
			break
		} else {
			c.status = api.Status_STATUS_NOT_IN_QUORUM
			if quorumRetries == 30 {
				err := fmt.Errorf("Unable to achieve Quorum."+
					" Timeout (%v) exceeded.",
					types.DEFAULT_QUORUM_TIMEOUT)
				dlog.Warnln("Failed to join cluster: ", err)
				c.status = api.Status_STATUS_NOT_IN_QUORUM
				c.selfNode.Status = api.Status_STATUS_OFFLINE
				c.gossip.UpdateSelfStatus(types.NODE_STATUS_DOWN)
				return err
			}
			dlog.Infof("Sleeping as no quorum")
			time.Sleep(types.DEFAULT_GOSSIP_INTERVAL)
			quorumRetries++
		}
	}
	go c.updateClusterStatus(initState, exist)

	kvdb.WatchKey(ClusterDBKey, 0, nil, c.watchDB)

	return nil
}

// Enumerate lists all the nodes in the cluster.
func (c *ClusterManager) Enumerate() (api.Cluster, error) {
	i := 0

	cluster := api.Cluster{
		Id:     c.config.ClusterId,
		Status: c.status,
		NodeId: c.selfNode.Id,
	}
	cluster.Nodes = make([]api.Node, len(c.nodeCache))
	for _, n := range c.nodeCache {
		cluster.Nodes[i] = n
		i++
	}

	return cluster, nil
}

// SetSize sets the maximum number of nodes in a cluster.
func (c *ClusterManager) SetSize(size int) error {
	kvdb := kvdb.Instance()
	kvlock, err := kvdb.Lock(clusterLockKey)
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

	err = writeClusterInfo(&db)

	return err
}

// Remove node(s) from the cluster permanently.
func (c *ClusterManager) Remove(nodes []api.Node) error {
	logrus.Infof("ClusterManager Remove node.")

	for _, n := range nodes {

		if _, exist := c.nodeCache[n.Id]; !exist {
			msg := fmt.Sprintf("Node does not exist in cluster, Node ID %s.", n.Id)
			dlog.Errorf(msg)
			return errors.New(msg)
		}

		// If removing node is self, return error
		if n.Id == c.selfNode.Id {
			msg := fmt.Sprintf("Cannot remove self from cluster, Node ID %s.", n.Id)
			dlog.Errorf(msg)
			return errors.New(msg)
		}

		// If node is not down, do not remove it
		if c.nodeCache[n.Id].Status != api.Status_STATUS_OFFLINE {
			msg := fmt.Sprintf("Cannot remove node that is not offline, Node ID %s.", n.Id)
			dlog.Errorf(msg)
			return errors.New(msg)
		}

		// Alert all listeners that we are removing this node.
		for e := c.listeners.Front(); e != nil; e = e.Next() {
			dlog.Infof("Remove node: notify cluster listener: %s",
				e.Value.(ClusterListener).String())
			if err := e.Value.(ClusterListener).Remove(&n); err != nil {
				dlog.Warnf("Cluster listener failed to remove node: %s: %s",
					e.Value.(ClusterListener).String(), err)
				return err
			}
		}
	}
	return nil
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
