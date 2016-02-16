// This file implements a cluster state machine.  It relies on a cluster
// wide key-value store for coordinating the state of the cluster.
// It also stores the state of the cluster in this key-value store.
package cluster

import (
	"container/list"
	"encoding/gob"
	"errors"
	"net"
	"time"

	"go.pedge.io/dlog"

	"github.com/fsouza/go-dockerclient"
	"github.com/libopenstorage/gossip"
	"github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/config"

	"github.com/libopenstorage/systemutils"
	"github.com/portworx/kvdb"
)

const (
	heartbeatKey = "heartbeat"
)

type ClusterManager struct {
	listeners *list.List
	config    config.ClusterConfig
	kv        kvdb.Kvdb
	status    api.Status
	nodeCache map[string]api.Node // Cached info on the nodes in the cluster.
	docker    *docker.Client
	g         gossip.Gossiper
	gEnabled  bool
	selfNode  api.Node
	system    systemutils.System
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
		return ip.String(), nil
	}

	return "", errors.New("Node not connected to the network.")
}

func ExternalIp(config *config.ClusterConfig) (string, error) {
	if config.MgtIface != "" {
		iface, err := net.InterfaceByName(config.MgtIface)
		if err != nil {
			return "", errors.New("Invalid network interface specified.")
		}
		return ifaceToIp(iface)
	}

	// No network interface specified, pick first default.
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		return ifaceToIp(&iface)
	}

	return "", errors.New("Node not connected to the network.")
}

func (c *ClusterManager) LocateNode(nodeID string) (api.Node, error) {
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

func (c *ClusterManager) UpdateData(dataKey string, value interface{}) {
	c.selfNode.NodeData[dataKey] = value
}

func (c *ClusterManager) GetData() map[string]*api.Node {
	nodes := make(map[string]*api.Node)
	for _, value := range c.nodeCache {
		nodes[value.Id] = &value
	}
	return nodes
}

func (c *ClusterManager) getCurrentState() *api.Node {
	c.selfNode.Timestamp = time.Now()

	c.selfNode.Cpu, _, _ = c.system.CpuUsage()
	c.selfNode.MemTotal, c.selfNode.MemUsed, c.selfNode.MemFree = c.system.MemUsage()
	c.selfNode.Luns = c.system.Luns()

	c.selfNode.Timestamp = time.Now()

	// Get containers running on this system.
	c.selfNode.Containers, _ = c.docker.ListContainers(docker.ListContainersOptions{All: true})

	return &c.selfNode
}

func (c *ClusterManager) getLatestNodeConfig(nodeId string) *NodeEntry {
	kvdb := kvdb.Instance()
	kvlock, err := kvdb.Lock("cluster/lock", 20)
	if err != nil {
		dlog.Warnln(" Unable to obtain cluster lock for updating config", err)
		return nil
	}
	defer kvdb.Unlock(kvlock)

	db, err := readDatabase()
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

func (c *ClusterManager) initNode(db *Database) (*api.Node, bool) {
	c.nodeCache[c.selfNode.Id] = *c.getCurrentState()

	_, exists := db.NodeEntries[c.selfNode.Id]

	// Add us into the database.
	db.NodeEntries[c.config.NodeId] = NodeEntry{Id: c.selfNode.Id,
		Ip: c.selfNode.Ip, GenNumber: c.selfNode.GenNumber}

	dlog.Infof("Node %s joining cluster...", c.config.NodeId)
	dlog.Infof("Cluster ID: %s", c.config.ClusterId)
	dlog.Infof("Node IP: %s", c.selfNode.Ip)

	return &c.selfNode, exists
}

func (c *ClusterManager) cleanupInit(db *Database, self *api.Node) error {
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

// Initialize node and alert listeners that we are joining the cluster.
func (c *ClusterManager) joinCluster(db *Database, self *api.Node, exist bool) error {
	var err error

	// If I am already in the cluster map, don't add me again.
	if exist {
		goto found
	}

	// Alert all listeners that we are a new node joining an existing cluster.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err = e.Value.(ClusterListener).Init(self, db)
		if err != nil {
			self.Status = api.Status_STATUS_ERROR
			dlog.Warnf("Failed to initialize Init %s: %v",
				e.Value.(ClusterListener).String(), err)
			c.cleanupInit(db, self)
			goto done
		}
	}

found:
	// Alert all listeners that we are joining the cluster.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err = e.Value.(ClusterListener).Join(self, db)
		if err != nil {
			self.Status = api.Status_STATUS_ERROR
			dlog.Warnf("Failed to initialize Join %s: %v",
				e.Value.(ClusterListener).String(), err)

			if exist == false {
				c.cleanupInit(db, self)
			}
			goto done
		}
	}

	for id, n := range db.NodeEntries {
		if id != c.config.NodeId {
			// Check to see if the IP is the same.  If it is, then we have a stale entry.
			if n.Ip == self.Ip {
				dlog.Warnf("Warning, Detected node %s with the same IP %s in the database.  Will not connect to this node.",
					id, n.Ip)
			} else {
				// Gossip with this node.
				dlog.Infof("Connecting to node %s with IP %s.", id, n.Ip)
				c.g.AddNode(n.Ip+":9002", types.NodeId(c.config.NodeId))
			}
		}
	}

done:
	return err
}

func (c *ClusterManager) initCluster(db *Database, self *api.Node, exist bool) error {
	err := error(nil)

	// Alert all listeners that we are initializing a new cluster.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err = e.Value.(ClusterListener).ClusterInit(self, db)
		if err != nil {
			self.Status = api.Status_STATUS_ERROR
			dlog.Printf("Failed to initialize %s",
				e.Value.(ClusterListener).String())
			goto done
		}
	}

	err = c.joinCluster(db, self, exist)
	if err != nil {
		dlog.Printf("Failed to join new cluster")
		goto done
	}

done:
	return err
}

func (c *ClusterManager) heartBeat() {
	gossipStoreKey := types.StoreKey(heartbeatKey + c.config.ClusterId)
	lastUpdateTs := time.Now()

	for {
		node := c.getCurrentState()
		c.nodeCache[node.Id] = *node

		currTime := time.Now()
		if currTime.Sub(lastUpdateTs) > 10*time.Second {
			dlog.Warnln("No gossip update for 10 seconds")
		}
		c.g.UpdateSelf(gossipStoreKey, *node)
		lastUpdateTs = currTime

		// Process heartbeats from other nodes...
		gossipValues := c.g.GetStoreKeyValue(gossipStoreKey)

		for id, nodeInfo := range gossipValues {
			if id == types.NodeId(node.Id) {
				continue
			}

			cachedNodeInfo, nodeFoundInCache := c.nodeCache[string(id)]
			n := cachedNodeInfo
			ok := false
			if nodeInfo.Value != nil {
				n, ok = nodeInfo.Value.(api.Node)
				if !ok {
					dlog.Errorln("Received a bad broadcast packet: %v", nodeInfo.Value)
					continue
				}
			}

			if nodeFoundInCache {
				if n.Status != api.Status_STATUS_OK {
					dlog.Warnln("Detected node ", n.Id, " to be unhealthy.")

					for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
						err := e.Value.(ClusterListener).Update(&n)
						if err != nil {
							dlog.Warnln("Failed to notify ", e.Value.(ClusterListener).String())
						}
					}

					delete(c.nodeCache, n.Id)
					continue
				} else if nodeInfo.Status == types.NODE_STATUS_DOWN {
					ne := c.getLatestNodeConfig(string(id))
					if ne != nil && nodeInfo.GenNumber < ne.GenNumber {
						dlog.Warnln("Detected stale update for node ", id,
							" going down, ignoring it")
						c.g.MarkNodeHasOldGen(id)
						delete(c.nodeCache, cachedNodeInfo.Id)
						continue
					}

					dlog.Warnln("Detected node ", id, " to be offline due to inactivity.")

					n.Status = api.Status_STATUS_OFFLINE
					for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
						err := e.Value.(ClusterListener).Update(&n)
						if err != nil {
							dlog.Warnln("Failed to notify ", e.Value.(ClusterListener).String())
						}
					}

					delete(c.nodeCache, cachedNodeInfo.Id)
				} else if nodeInfo.Status == types.NODE_STATUS_DOWN_WAITING_FOR_NEW_UPDATE {
					dlog.Warnln("Detected node ", n.Id, " to be offline due to inactivity.")

					n.Status = api.Status_STATUS_OFFLINE
					for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
						err := e.Value.(ClusterListener).Update(&n)
						if err != nil {
							dlog.Warnln("Failed to notify ", e.Value.(ClusterListener).String())
						}
					}

					delete(c.nodeCache, cachedNodeInfo.Id)
				} else {
					// node may be up or waiting for new update,
					// no need to tell listeners as yet.
					c.nodeCache[cachedNodeInfo.Id] = n
				}
			} else if nodeInfo.Status == types.NODE_STATUS_UP {
				// A node discovered in the cluster.
				dlog.Warnln("Detected node ", n.Id, " to be in the cluster.")

				c.nodeCache[n.Id] = n
				for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
					err := e.Value.(ClusterListener).Add(&n)
					if err != nil {
						dlog.Warnln("Failed to notify ", e.Value.(ClusterListener).String())
					}
				}
			}
		}

		time.Sleep(2 * time.Second)
	}
}

func (c *ClusterManager) DisableUpdates() {
	dlog.Warnln("Disabling gossip updates")
	c.gEnabled = false
}

func (c *ClusterManager) EnableUpdates() {
	dlog.Warnln("Enabling gossip updates")
	c.gEnabled = true
}

func (c *ClusterManager) GetState() *ClusterState {
	gossipStoreKey := types.StoreKey(heartbeatKey + c.config.ClusterId)
	nodeValue := c.g.GetStoreKeyValue(gossipStoreKey)
	nodes := make([]types.NodeValue, len(nodeValue), len(nodeValue))
	i := 0
	for _, value := range nodeValue {
		nodes[i] = value
		i++
	}

	history := c.g.GetGossipHistory()
	return &ClusterState{
		History: history, NodeStatus: nodes}
}

func (c *ClusterManager) Start() error {
	dlog.Infoln("Cluster manager starting...")

	c.gEnabled = true
	c.selfNode = api.Node{}
	c.selfNode.GenNumber = uint64(time.Now().UnixNano())
	c.selfNode.Id = c.config.NodeId
	c.selfNode.Status = api.Status_STATUS_OK
	c.selfNode.Ip, _ = ExternalIp(&c.config)
	c.selfNode.NodeData = make(map[string]interface{})
	c.system = systemutils.New()

	// Start the gossip protocol.
	// XXX Make the port configurable.
	gob.Register(api.Node{})
	c.g = gossip.New("0.0.0.0:9002", types.NodeId(c.config.NodeId),
		c.selfNode.GenNumber)
	c.g.SetGossipInterval(2 * time.Second)

	kvdb := kvdb.Instance()
	kvlock, err := kvdb.Lock("cluster/lock", 60)
	if err != nil {
		dlog.Panicln("Fatal, Unable to obtain cluster lock.", err)
	}
	defer kvdb.Unlock(kvlock)

	db, err := readDatabase()
	if err != nil {
		dlog.Panicln(err)
	}

	if db.Status == api.Status_STATUS_INIT {
		dlog.Infoln("Will initialize a new cluster.")

		c.status = api.Status_STATUS_OK
		db.Status = api.Status_STATUS_OK
		self, _ := c.initNode(&db)

		err = c.initCluster(&db, self, false)
		if err != nil {
			dlog.Errorln("Failed to initialize the cluster.", err)
			return err
		}

		// Update the new state of the cluster in the KV Database
		err := writeDatabase(&db)
		if err != nil {
			dlog.Errorln("Failed to save the database.", err)
			return err
		}

	} else if db.Status&api.Status_STATUS_OK > 0 {
		dlog.Infoln("Cluster state is OK... Joining the cluster.")

		c.status = api.Status_STATUS_OK
		self, exist := c.initNode(&db)

		err = c.joinCluster(&db, self, exist)
		if err != nil {
			dlog.Errorln("Failed to join cluster.", err)
			return err
		}

		err := writeDatabase(&db)
		if err != nil {
			return err
		}

	} else {
		return errors.New("Fatal, Cluster is in an unexpected state.")
	}

	// Start heartbeating to other nodes.
	c.g.Start()
	go c.heartBeat()

	return nil
}

// Enumerate lists all the nodes in the cluster.
func (c *ClusterManager) Enumerate() (api.Cluster, error) {
	i := 0

	cluster := api.Cluster{Id: c.config.ClusterId, Status: c.status}
	cluster.Nodes = make([]api.Node, len(c.nodeCache))
	for _, n := range c.nodeCache {
		cluster.Nodes[i] = n
		i++
	}

	return cluster, nil
}

// Remove node(s) from the cluster permanently.
func (c *ClusterManager) Remove(nodes []api.Node) error {
	// TODO
	return nil
}

// Shutdown can be called when THIS node is gracefully shutting down.
func (c *ClusterManager) Shutdown() error {
	db, err := readDatabase()
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
