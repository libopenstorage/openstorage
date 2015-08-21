// This file implements a cluster state machine.  It relies on a cluster
// wide key-value store for coordinating the state of the cluster.
// It also stores the state of the cluster in this key-value store.
package cluster

import (
	_ "bytes"
	"container/list"
	_ "encoding/json"
	"errors"
	"net"
	_ "strings"
	_ "time"

	log "github.com/Sirupsen/logrus"

	kv "github.com/portworx/kvdb"
	"github.com/portworx/systemutils"
)

type ClusterManager struct {
	listeners *list.List
	config    Config
	kv        kv.Kvdb
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

func (c *ClusterManager) getInfo() *NodeInfo {
	var info = NodeInfo{}
	s := systemutils.New()

	info.Config = c.config
	info.Cpu, _, _ = s.CpuUsage()
	info.Memory = s.MemUsage()
	info.Luns = s.Luns()

	return &info
}

func (c *ClusterManager) AddEventListener(ClusterListener) error {
	return nil
}

func (c *ClusterManager) Start() error {
	log.Info("cluster manager starting...")

	return nil
}
