package proto

import (
	"container/list"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/logutils"
	ml "github.com/hashicorp/memberlist"
	"github.com/libopenstorage/gossip/types"
)

type GossipHistory struct {
	// front is the latest, back is the last
	nodes  *list.List
	lock   sync.Mutex
	maxLen uint8
}

func NewGossipSessionInfo(node string,
	dir types.GossipDirection) *types.GossipSessionInfo {
	gs := new(types.GossipSessionInfo)
	gs.Node = node
	gs.Dir = dir
	gs.Ts = time.Now()
	gs.Err = ""
	return gs
}

func NewGossipHistory(maxLen uint8) *GossipHistory {
	s := new(GossipHistory)
	s.nodes = list.New()
	s.nodes.Init()
	s.maxLen = maxLen
	return s
}

func (s *GossipHistory) AddLatest(gs *types.GossipSessionInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if uint8(s.nodes.Len()) == s.maxLen {
		s.nodes.Remove(s.nodes.Back())
	}
	s.nodes.PushFront(gs)
}

func (s *GossipHistory) GetAllRecords() []*types.GossipSessionInfo {
	s.lock.Lock()
	defer s.lock.Unlock()
	records := make([]*types.GossipSessionInfo, s.nodes.Len(), s.nodes.Len())
	i := 0
	for element := s.nodes.Front(); element != nil; element = element.Next() {
		r, ok := element.Value.(*types.GossipSessionInfo)
		if !ok || r == nil {
			log.Error("gossip: Failed to convert element")
			continue
		}
		records[i] = &types.GossipSessionInfo{Node: r.Node,
			Ts: r.Ts, Dir: r.Dir, Err: r.Err}
		i++
	}
	return records
}

func (s *GossipHistory) LogRecords() {
	s.lock.Lock()
	defer s.lock.Unlock()
	status := make([]string, 2)
	status[types.GD_ME_TO_PEER] = "ME_TO_PEER"
	status[types.GD_PEER_TO_ME] = "PEER_TO_ME"

	for element := s.nodes.Front(); element != nil; element = element.Next() {
		r, ok := element.Value.(*types.GossipSessionInfo)
		if !ok || r == nil {
			continue
		}
		log.Infof("Node: %v LastTs: %v Dir: %v Error: %v",
			r.Node, r.Ts, status[r.Dir], r.Err)
	}
}

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
	nodes     GossipNodeList
	name      string
	nodesLock sync.Mutex
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
	mlConf.Name = string(selfNodeId)
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
	)
	mlConf.Delegate = ml.Delegate(g)
	mlConf.Events = ml.EventDelegate(g)
	mlConf.Alive = ml.AliveDelegate(g)
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
	g.InitCurrentState(len(knownIps) + 1)
	list, err := ml.Create(g.mlConf)
	if err != nil {
		log.Warnf("gossip: Unable to create memberlist: " + err.Error())
		return err
	}
	// Set the memberlist in gossiper object
	g.mlist = list

	if len(knownIps) != 0 {
		// Joining an existing cluster
		_, err := list.Join(knownIps)
		if err != nil {
			log.Infof("gossip: Unable to join other nodes at startup : %v", err)
			return err
		}
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

func (g *GossiperImpl) GetGossipHistory() []*types.GossipSessionInfo {
	return g.history.GetAllRecords()
}

func (g *GossiperImpl) GetNodes() []string {
	nodes := g.mlist.Members()
	nodeList := make([]string, len(nodes))
	for i, node := range nodes {
		nodeList[i] = node.Addr.String()
	}
	return nodeList

}

func (g *GossiperImpl) UpdateCluster(peers map[types.NodeId]string) {
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
		log.Infof("gossip: Our Status: %v. We should go down.", g.GetSelfStatus())
		return g.NodeId()
	}
}
