package proto

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/hashicorp/logutils"
	ml "github.com/hashicorp/memberlist"
	"github.com/libopenstorage/gossip/types"
)

type GossipNode struct {
	Id types.NodeId
	Ip string
}

type GossipNodeList []GossipNode

func (nodes GossipNodeList) Len() int {
	return len(nodes)
}

func (nodes GossipNodeList) Less(i, j int) bool {
	return nodes[i].Id < nodes[i].Id
}

func (nodes GossipNodeList) Swap(i, j int) {
	nodes[i], nodes[j] = nodes[j], nodes[i]
}

// Implements the UnreliableBroadcast interface
type GossiperImpl struct {
	// GossipDelegate implements the GossipStoreInterface
	// as well as the memberlist Delegates
	GossipDelegate

	mlConf *ml.Config
	mlist  *ml.Memberlist

	// node list, maintained separately
	nodes          GossipNodeList
	name           string
	nodesLock      sync.Mutex
	gossipInterval time.Duration
	//nodeDeathInterval time.Duration
	shutDown bool
}

// Utility methods
func logAndGetError(msg string) error {
	log.Error(msg)
	return errors.New(msg)
}

func (g *GossiperImpl) Init(
	ipPort string,
	selfNodeId types.NodeId,
	genNumber uint64,
	gossipIntervals types.GossipIntervals,
	gossipVersion string,
	clusterId string,
) {
	g.name = ipPort
	g.shutDown = false

	g.nodes = make(GossipNodeList, 0)
	g.gossipInterval = gossipIntervals.GossipInterval

	// Memberlist Config setup
	mlConf := ml.DefaultLANConfig()

	s := strings.Split(ipPort, ":")
	ip, port := s[0], s[1]
	port64, _ := strconv.ParseInt(port, 10, 64)

	// Memberlist conf Name is the name of the node
	// and it should be unique in the cluster
	nodeName := string(selfNodeId) + gossipVersion
	mlConf.Name = nodeName
	mlConf.BindAddr = ip
	mlConf.BindPort = int(port64)

	// This should be twice the RTT of the network
	mlConf.ProbeTimeout = gossipIntervals.ProbeTimeout
	mlConf.PushPullInterval = gossipIntervals.PushPullInterval
	mlConf.GossipInterval = g.gossipInterval
	// ProbeInterval used for broadcasts and decides probing behavior
	mlConf.ProbeInterval = gossipIntervals.ProbeInterval

	// MemberDelegates
	g.InitGossipDelegate(
		genNumber,
		selfNodeId,
		gossipVersion,
		gossipIntervals.QuorumTimeout,
		clusterId,
	)
	mlConf.Delegate = ml.Delegate(g)
	mlConf.Events = ml.EventDelegate(g)
	mlConf.Alive = ml.AliveDelegate(g)
	mlConf.Merge = ml.MergeDelegate(g)
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stderr,
	}
	mlConf.LogOutput = filter

	g.mlConf = mlConf
	rand.Seed(time.Now().UnixNano())
}

func (g *GossiperImpl) Start(knownIps []string) error {
	g.InitCurrentState(uint(len(knownIps) + 1))
	list, err := ml.Create(g.mlConf)
	if err != nil {
		log.Warnf("gossip: Unable to create memberlist: " + err.Error())
		return err
	}
	// Set the memberlist in gossiper object
	g.mlist = list

	if len(knownIps) != 0 {
		// Joining an existing cluster
		joinedNodes, err := list.Join(knownIps)
		if err != nil {
			log.Infof("gossip: Unable to join other nodes at startup : %v", err)
			return err
		}
		log.Infof("gossip: Successfully joined with %v node(s)", joinedNodes)
	}
	return nil
}

func (g *GossiperImpl) Stop(leaveTimeout time.Duration) error {
	if g.shutDown == true {
		return fmt.Errorf("gossip: Gossiper already stopped")
	}
	err := g.mlist.Leave(leaveTimeout)
	if err != nil {
		return err
	}
	err = g.mlist.Shutdown()
	if err != nil {
		return err
	}
	g.shutDown = true
	return nil
}

func (g *GossiperImpl) GossipInterval() time.Duration {
	return g.gossipInterval
}

func (g *GossiperImpl) GetNodes() []string {
	nodes := g.mlist.Members()
	nodeList := make([]string, len(nodes))
	for i, node := range nodes {
		nodeList[i] = node.Addr.String()
	}
	return nodeList
}

func (g *GossiperImpl) UpdateCluster(peers map[types.NodeId]types.NodeUpdate) {
	g.updateCluster(peers)
	g.triggerStateEvent(types.UPDATE_CLUSTER_SIZE)
}

func (g *GossiperImpl) ExternalNodeLeave(nodeId types.NodeId) types.NodeId {
	log.Infof("gossip: Request for a Node Leave operation on Node %v", nodeId)
	if g.GetSelfStatus() == types.NODE_STATUS_UP {
		log.Infof("gossip: Node %v should go down.", nodeId)
		return nodeId
	} else {
		// We are the culprit as we are not in quorum
		log.Infof("gossip: Our Status: %v. We should go down.",
			g.GetSelfStatus())
		return g.NodeId()
	}
}
