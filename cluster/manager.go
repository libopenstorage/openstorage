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

	"github.com/samalba/dockerclient"

	"github.com/libopenstorage/gossip"
	gossiptypes "github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"

	log "github.com/Sirupsen/logrus"

	kv "github.com/portworx/kvdb"
	"github.com/portworx/systemutils"
)

const (
	dockerHost   = "unix:///var/run/docker.sock"
	heartbeatKey = "heartbeat"
)

type ClusterManager struct {
	listeners *list.List
	config    Config
	kv        kv.Kvdb
	status    api.Status
	nodeCache map[string]api.Node // Cached info on the nodes in the cluster.
	docker    *dockerclient.DockerClient
	g         gossip.Gossiper
}

func externalIp() (string, error) {
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
	}

	return "", errors.New("Node not connected to the network.")
}

func (c *ClusterManager) AddEventListener(listener ClusterListener) error {
	log.Printf("Adding cluster event listener: %s", listener.String())
	c.listeners.PushBack(listener)
	return nil
}

func (c *ClusterManager) getSelf() *api.Node {
	var node = api.Node{}
	s := systemutils.New()

	// Get physical node info.
	node.Id = c.config.NodeId
	node.Status = api.StatusOk
	node.Ip, _ = externalIp()

	node.Cpu, _, _ = s.CpuUsage()
	node.Memory = s.MemUsage()
	node.Luns = s.Luns()

	node.Timestamp = time.Now()

	// Get containers running on this system.
	node.Containers, _ = c.docker.ListContainers(true, true, "")

	return &node
}

func (c *ClusterManager) initNode(db *Database) (*api.Node, bool) {
	node := c.getSelf()
	c.nodeCache[node.Id] = *node

	_, exists := db.NodeEntries[node.Id]

	// Add us into the database.
	db.NodeEntries[c.config.NodeId] = NodeEntry{Id: node.Id, Ip: node.Ip}

	log.Infof("Node %s joining cluster... \n\tCluster ID: %s\n\tIP: %s",
		c.config.NodeId, c.config.ClusterId, node.Ip)

	return node, exists
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
			log.Warnf("Failed to initialize %s: %v\n",
				e.Value.(ClusterListener).String(), err)
			goto done
		}
	}

found:
	// Alert all listeners that we are joining the cluster.
	for e := c.listeners.Front(); e != nil; e = e.Next() {
		err = e.Value.(ClusterListener).Join(self, db)
		if err != nil {
			log.Warnf("Failed to initialize %s: %v\n",
				e.Value.(ClusterListener).String(), err)
			goto done
		}
	}

	for id, n := range db.NodeEntries {
		if id != c.config.NodeId {
			// Check to see if the IP is the same.  If it is, then we have a stale entry.
			if n.Ip == self.Ip {
				log.Warnf("Warning, Detected node %s with the same IP %s in the database.  Will not connect to this node.",
					id, n.Ip)
			} else {
				log.Infof("Connecting to node %s with IP %s.", id, n.Ip)
				// Gossip with this node.
				c.g.AddNode(n.Ip + ":9002")

				// Assume this node is OK.  We will catch any problems during heartbeating.
				c.nodeCache[id] = api.Node{Status: api.StatusOk, Timestamp: time.Now()}
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
			log.Printf("Failed to initialize %s\n",
				e.Value.(ClusterListener).String())
			goto done
		}
	}

	err = c.joinCluster(db, self, exist)
	if err != nil {
		log.Printf("Failed to join new cluster\n")
		goto done
	}

done:
	return err
}

func (c *ClusterManager) heartBeat() {
	for {
		node := c.getSelf()
		c.nodeCache[node.Id] = *node

		c.g.UpdateSelf(heartbeatKey, *node)

		// Process heartbeats from other nodes...
		gossipValues := c.g.GetStoreKeyValue(heartbeatKey)

		for _, nodeInfo := range gossipValues {
			n, ok := nodeInfo.Value.(api.Node)

			if !ok {
				log.Warn("Received a bad broadcast packet: %v", nodeInfo.Value)
				continue
			}

			if n.Id == node.Id {
				continue
			}

			_, ok = c.nodeCache[n.Id]
			if ok {
				if n.Status != api.StatusOk {
					log.Warn("Detected node ", n.Id, " to be unhealthy.")

					for e := c.listeners.Front(); e != nil; e = e.Next() {
						err := e.Value.(ClusterListener).Update(&n)
						if err != nil {
							log.Warn("Failed to notify ", e.Value.(ClusterListener).String())
						}
					}

					delete(c.nodeCache, n.Id)
				} else if time.Since(n.Timestamp) > 60*time.Second {
					log.Warn("Detected node ", n.Id, " to be offline due to inactivity.")

					n.Status = api.StatusOffline

					for e := c.listeners.Front(); e != nil; e = e.Next() {
						err := e.Value.(ClusterListener).Update(&n)
						if err != nil {
							log.Warn("Failed to notify ", e.Value.(ClusterListener).String())
						}
					}

					delete(c.nodeCache, n.Id)
				}
			} else {
				// A node joined the cluster.
				log.Warn("Detected node ", n.Id, " to be online.")

				c.nodeCache[n.Id] = n
				for e := c.listeners.Front(); e != nil; e = e.Next() {
					err := e.Value.(ClusterListener).Update(&n)
					if err != nil {
						log.Warn("Failed to notify ", e.Value.(ClusterListener).String())
					}
				}
			}
		}

		// Process stale entries in our local cache.
		for id, n := range c.nodeCache {
			if id == node.Id {
				continue
			}

			if time.Since(n.Timestamp) > 60*time.Second {
				log.Warn("Detected node ", n.Id, " to be offline due to a lack of heartbeat.")

				n.Status = api.StatusOffline

				for e := c.listeners.Front(); e != nil; e = e.Next() {
					err := e.Value.(ClusterListener).Update(&n)
					if err != nil {
						log.Warn("Failed to notify ", e.Value.(ClusterListener).String())
					}
				}

				delete(c.nodeCache, id)
			}
		}

		time.Sleep(2 * time.Second)
	}
}

func (c *ClusterManager) Start() error {
	log.Info("Cluster manager starting...")
	kvdb := kv.Instance()

	// Start the gossip protocol.
	// XXX make the port configurable.
	gob.Register(api.Node{})
	c.g = gossip.New("0.0.0.0:9002", gossiptypes.NodeId(c.config.NodeId))
	c.g.SetGossipInterval(2 * time.Second)

	kvlock, err := kvdb.Lock("cluster/lock", 60)
	if err != nil {
		log.Panic("Fatal, Unable to obtain cluster lock.", err)
	}

	db, err := readDatabase()
	if err != nil {
		log.Panic(err)
	}

	if db.Status == api.StatusInit {
		log.Info("Will initialize a new cluster.")

		c.status = api.StatusOk
		db.Status = api.StatusOk
		self, _ := c.initNode(&db)

		// Update the new state of the cluster in the KV Database
		err = writeDatabase(&db)
		if err != nil {
			log.Panic(err)
		}

		err = kvdb.Unlock(kvlock)
		if err != nil {
			log.Panic("Fatal, unable to unlock cluster... Did something take too long to initialize?", err)
		}

		err = c.initCluster(&db, self, false)
		if err != nil {
			log.Panic(err)
		}
	} else if db.Status&api.StatusOk > 0 {
		log.Info("Cluster state is OK... Joining the cluster.")

		c.status = api.StatusOk
		self, exist := c.initNode(&db)

		err = writeDatabase(&db)
		if err != nil {
			log.Panic(err)
		}

		err = kvdb.Unlock(kvlock)
		if err != nil {
			log.Panic("Fatal, unable to unlock cluster... Did something take too long to initialize?", err)
		}

		err = c.joinCluster(&db, self, exist)
		if err != nil {
			log.Panic(err)
		}
	} else {
		err = kvdb.Unlock(kvlock)
		err = errors.New("Fatal, Cluster is in an unexpected state.")
		log.Panic(err)
	}

	// Start heartbeating to other nodes.
	go c.heartBeat()

	return nil
}

func (c *ClusterManager) Init() error {
	docker, err := dockerclient.NewDockerClient(dockerHost, nil)
	if err != nil {
		log.Printf("Fatal, could not connect to Docker.")
		return err
	}

	c.listeners = list.New()
	c.nodeCache = make(map[string]api.Node)
	c.docker = docker

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
