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

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/libopenstorage/gossip"
	"github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"

	"github.com/portworx/kvdb"
	"github.com/portworx/systemutils"
)

const (
	heartbeatKey = "heartbeat"
)

type ClusterManager struct {
	listeners *list.List
	config    Config
	kv        kvdb.Kvdb
	status    api.Status
	nodeCache map[string]api.Node // Cached info on the nodes in the cluster.
	docker    *docker.Client
	g         gossip.Gossiper
	gEnabled  bool
	selfNode  api.Node
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

func externalIp(config *Config) (string, error) {

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
	logrus.Printf("Adding cluster event listener: %s", listener.String())
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
	s := systemutils.New()

	c.selfNode.Cpu, _, _ = s.CpuUsage()
	c.selfNode.Memory = s.MemUsage()
	c.selfNode.Luns = s.Luns()

	c.selfNode.Timestamp = time.Now()

	// Get containers running on this system.
	c.selfNode.Containers, _ = c.docker.ListContainers(docker.ListContainersOptions{All: true})

	return &c.selfNode
}

func (c *ClusterManager) getLatestNodeConfig(nodeId string) *NodeEntry {
	kvdb := kvdb.Instance()
	kvlock, err := kvdb.Lock("cluster/lock", 20)
	if err != nil {
		logrus.Warn(" Unable to obtain cluster lock for updating config", err)
		return nil
	}
	defer kvdb.Unlock(kvlock)

	db, err := readDatabase()
	if err != nil {
		logrus.Warn("Failed to read the database for updating config")
		return nil
	}

	ne, exists := db.NodeEntries[nodeId]
	if !exists {
		logrus.Warn("Could not find info for node with id ", nodeId)
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

	logrus.Infof("Node %s joining cluster... \n\tCluster ID: %s\n\tIP: %s",
		c.config.NodeId, c.config.ClusterId, c.selfNode.Ip)

	return &c.selfNode, exists
}

func (c *ClusterManager) cleanupInit(db *Database, self *api.Node) error {
	var resErr error
	var err error

	logrus.Infof("Cleanup Init services")

	for e := c.listeners.Front(); e != nil; e = e.Next() {
		logrus.Warnf("Cleanup Init for service %s.",
			e.Value.(ClusterListener).String())

		err = e.Value.(ClusterListener).CleanupInit(self, db)
		if err != nil {
			logrus.Warnf("Failed to Cleanup Init %s: %v",
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
			logrus.Warnf("Failed to initialize Init %s: %v",
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
			logrus.Warnf("Failed to initialize Join %s: %v",
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
				logrus.Warnf("Warning, Detected node %s with the same IP %s in the database.  Will not connect to this node.",
					id, n.Ip)
			} else {
				// Gossip with this node.
				logrus.Infof("Connecting to node %s with IP %s.", id, n.Ip)
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
			logrus.Printf("Failed to initialize %s",
				e.Value.(ClusterListener).String())
			goto done
		}
	}

	err = c.joinCluster(db, self, exist)
	if err != nil {
		logrus.Printf("Failed to join new cluster")
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
			logrus.Warn("No gossip update for 10 seconds")
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
					logrus.Error("Received a bad broadcast packet: %v", nodeInfo.Value)
					continue
				}
			}

			if nodeFoundInCache {
				if n.Status != api.StatusOk {
					logrus.Warn("Detected node ", n.Id, " to be unhealthy.")

					for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
						err := e.Value.(ClusterListener).Update(&n)
						if err != nil {
							logrus.Warn("Failed to notify ", e.Value.(ClusterListener).String())
						}
					}

					delete(c.nodeCache, n.Id)
					continue
				} else if nodeInfo.Status == types.NODE_STATUS_DOWN {
					ne := c.getLatestNodeConfig(string(id))
					if ne != nil && nodeInfo.GenNumber < ne.GenNumber {
						logrus.Warn("Detected stale update for node ", id,
							" going down, ignoring it")
						c.g.MarkNodeHasOldGen(id)
						delete(c.nodeCache, cachedNodeInfo.Id)
						continue
					}

					logrus.Warn("Detected node ", id, " to be offline due to inactivity.")

					n.Status = api.StatusOffline
					for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
						err := e.Value.(ClusterListener).Update(&n)
						if err != nil {
							logrus.Warn("Failed to notify ", e.Value.(ClusterListener).String())
						}
					}

					delete(c.nodeCache, cachedNodeInfo.Id)
				} else if nodeInfo.Status == types.NODE_STATUS_DOWN_WAITING_FOR_NEW_UPDATE {
					logrus.Warn("Detected node ", n.Id, " to be offline due to inactivity.")

					n.Status = api.StatusOffline
					for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
						err := e.Value.(ClusterListener).Update(&n)
						if err != nil {
							logrus.Warn("Failed to notify ", e.Value.(ClusterListener).String())
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
				logrus.Warn("Detected node ", n.Id, " to be in the cluster.")

				c.nodeCache[n.Id] = n
				for e := c.listeners.Front(); e != nil && c.gEnabled; e = e.Next() {
					err := e.Value.(ClusterListener).Add(&n)
					if err != nil {
						logrus.Warn("Failed to notify ", e.Value.(ClusterListener).String())
					}
				}
			}
		}

		time.Sleep(2 * time.Second)
	}
}

func (c *ClusterManager) DisableUpdates() {
	logrus.Warn("Disabling gossip updates")
	c.gEnabled = false
}

func (c *ClusterManager) EnableUpdates() {
	logrus.Warn("Enabling gossip updates")
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
	logrus.Info("Cluster manager starting...")

	c.gEnabled = true
	c.selfNode = api.Node{}
	c.selfNode.GenNumber = uint64(time.Now().UnixNano())
	c.selfNode.Id = c.config.NodeId
	c.selfNode.Status = api.StatusOk
	c.selfNode.Ip, _ = externalIp(&c.config)
	c.selfNode.NodeData = make(map[string]interface{})
	// Start the gossip protocol.
	// XXX Make the port configurable.
	gob.Register(api.Node{})
	c.g = gossip.New("0.0.0.0:9002", types.NodeId(c.config.NodeId),
		c.selfNode.GenNumber)
	c.g.SetGossipInterval(2 * time.Second)

	kvdb := kvdb.Instance()
	kvlock, err := kvdb.Lock("cluster/lock", 60)
	if err != nil {
		logrus.Panic("Fatal, Unable to obtain cluster lock.", err)
	}
	defer kvdb.Unlock(kvlock)

	db, err := readDatabase()
	if err != nil {
		logrus.Panic(err)
	}

	if db.Status == api.StatusInit {
		logrus.Info("Will initialize a new cluster.")

		c.status = api.StatusOk
		db.Status = api.StatusOk
		self, _ := c.initNode(&db)

		err = c.initCluster(&db, self, false)
		if err != nil {
			logrus.Error("Failed to initialize the cluster.", err)
			return err
		}

		// Update the new state of the cluster in the KV Database
		err := writeDatabase(&db)
		if err != nil {
			logrus.Error("Failed to save the database.", err)
			return err
		}

	} else if db.Status&api.StatusOk > 0 {
		logrus.Info("Cluster state is OK... Joining the cluster.")

		c.status = api.StatusOk
		self, exist := c.initNode(&db)

		err = c.joinCluster(&db, self, exist)
		if err != nil {
			logrus.Error("Failed to join cluster.", err)
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

func (c *ClusterManager) Remove(nodes []api.Node) error {
	// TODO
	return nil
}

func (c *ClusterManager) Shutdown(cluster bool, nodes []api.Node) error {
	// TODO
	return nil
}
