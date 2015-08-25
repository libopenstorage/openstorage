// This file implements a cluster state machine.  It relies on a cluster
// wide key-value store for coordinating the state of the cluster.
// It also stores the state of the cluster in this key-value store.
package cluster

import (
	"container/list"
	"errors"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"

	kv "github.com/portworx/kvdb"
	"github.com/portworx/systemutils"
)

type ClusterManager struct {
	listeners *list.List
	config    Config
	kv        kv.Kvdb
	nodeInfo  map[string]NodeInfo // Info on the nodes in the cluster
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

func (c *ClusterManager) getInfo() *NodeInfo {
	var info = NodeInfo{}
	s := systemutils.New()

	info.Cpu, _, _ = s.CpuUsage()
	info.Memory = s.MemUsage()
	info.Luns = s.Luns()
	info.NodeId = c.config.NodeId
	info.Ip, _ = externalIp()
	info.Status = StatusOk

	return &info
}

func (c *ClusterManager) initNode(db *Database) (*NodeInfo, bool) {
	info := c.getInfo()

	node := Node{
		Ip:     info.Ip,
		Status: info.Status}

	_, exists := db.Nodes[c.config.NodeId]

	// Add us into the database.
	db.Nodes[c.config.NodeId] = node

	log.Infof("Node %d joining cluster %s... \n\tIP: %s",
		c.config.NodeId, c.config.ClusterId, node.Ip)

	return info, exists
}

// Initialize node and alert listeners that we are joining the cluster.
func (c *ClusterManager) joinCluster(db *Database, self *NodeInfo, exist bool) error {
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

	for id, n := range db.Nodes {
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

func (c *ClusterManager) initCluster(db *Database, self *NodeInfo, exist bool) error {
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
	var info *NodeInfo = t.(*NodeInfo)

	last, ok := c.nodeInfo[info.NodeId]
	c.nodeInfo[info.NodeId] = *info

	// Allert listeners if status changed significantly...
	if !ok || last.Status != info.Status {
		log.Info("Node ", info.NodeId, " changed status\n\tIP: ",
			info.Ip, "\n\tTime: ", info.Timestamp, "\n\tStatus: ", info.Status)

		for e := c.listeners.Front(); e != nil; e = e.Next() {
			err = e.Value.(ClusterListener).Update(info)
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
		for id, info := range c.nodeInfo {
			if info.Status == StatusOk && time.Since(info.Timestamp) > 10000*time.Millisecond {
				log.Warn("Detected node ", id, " to be offline.")

				info.Status = StatusOffline
				c.nodeInfo[id] = info

				for e := c.listeners.Front(); e != nil; e = e.Next() {
					err := e.Value.(ClusterListener).Leave(&info)
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

	kvlock, err := kvdb.Lock("cluster/lock", 60)
	if err != nil {
		log.Panic("Fatal, Unable to obtain cluster lock.", err)
	}

	db, err := readDatabase()
	if err != nil {
		log.Panic(err)
	}

	if db.Cluster.Status == StatusInit {
		log.Info("Will initialize a new cluster.")

		db.Cluster.Status = StatusOk
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
	} else if db.Cluster.Status&StatusOk > 0 {
		log.Info("Cluster state is OK... Joining the cluster.")

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
