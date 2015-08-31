// This file implements a cluster state machine.  It relies on a cluster
// wide key-value store for coordinating the state of the cluster.
// It also stores the state of the cluster in this key-value store.
package cluster

import (
	"container/list"
	"errors"
	"net"
	"time"

	"github.com/libopenstorage/openstorage/api"

	log "github.com/Sirupsen/logrus"

	kv "github.com/portworx/kvdb"
	"github.com/portworx/systemutils"
)

type ClusterManager struct {
	listeners *list.List
	config    Config
	kv        kv.Kvdb
	status    api.Status
	nodes     map[string]api.Node // Info on the nodes in the cluster
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

	node.Id = c.config.NodeId
	node.Status = api.StatusOk
	node.Ip, _ = externalIp()

	node.Cpu, _, _ = s.CpuUsage()
	node.Memory = s.MemUsage()
	node.Luns = s.Luns()

	return &node
}

func (c *ClusterManager) initNode(db *Database) (*api.Node, bool) {
	node := c.getSelf()

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
				log.Warn("Warning, Detected node %s with the same IP %s in the database.  Will not connect to this node.",
					id, n.Ip)
			} else {
				// err = ubcast.AddNode(n.Ip)
				if err != nil {
					log.Infof("Node %d is OFFLINE in the cluster %s... \n\tUUID: %s\n\tIP: %s\n\t",
						id, c.config.ClusterId, n.Ip)
					err = nil
				} else {
					log.Infof("Node %d is ONLINE in the cluster %s... \n\tUUID: %s\n\tIP: %s\n\t",
						id, c.config.ClusterId, n.Ip)
				}
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

func (c *ClusterManager) processHeartbeat(err error, ip string, t interface{}) {
	var node *api.Node = t.(*api.Node)

	last, ok := c.nodes[node.Id]
	c.nodes[node.Id] = *node

	// Allert listeners if status changed significantly...
	if !ok || last.Status != node.Status {
		log.Info("Node ", node.Id, " changed status\n\tIP: ",
			node.Ip, "\n\tTime: ", node.Timestamp, "\n\tStatus: ", node.Status)

		for e := c.listeners.Front(); e != nil; e = e.Next() {
			err = e.Value.(ClusterListener).Update(node)
			if err != nil {
				log.Warn("Failed to notify ", e.Value.(ClusterListener).String())
			}
		}
	}
}

func (c *ClusterManager) heartBeat() {
	for {
		time.Sleep(2 * time.Second)

		// myInfo := c.getInfo()
		// ubcast.Push(NodeUpdate, &myInfo)

		// Process heartbeats from other nodes...
		for id, n := range c.nodes {
			if n.Status == api.StatusOk && time.Since(n.Timestamp) > 10000*time.Millisecond {
				log.Warn("Detected node ", id, " to be offline.")

				n.Status = api.StatusOffline
				c.nodes[id] = n

				for e := c.listeners.Front(); e != nil; e = e.Next() {
					err := e.Value.(ClusterListener).Leave(&n)
					if err != nil {
						log.Warn("Failed to notify ",
							e.Value.(ClusterListener).String())
					}
				}
			}
		}
	}
}

func (c *ClusterManager) Start() error {
	log.Info("Cluster manager starting...")
	kvdb := kv.Instance()

	c.listeners = list.New()

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

	// Join the clusterwide heartbeat mesh.
	go c.heartBeat()

	return nil
}

func (c *ClusterManager) Enumerate() (api.Cluster, error) {
	i := 0

	cluster := api.Cluster{Id: c.config.ClusterId, Status: c.status}
	cluster.Nodes = make([]api.Node, len(c.nodes))
	for _, n := range c.nodes {
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
