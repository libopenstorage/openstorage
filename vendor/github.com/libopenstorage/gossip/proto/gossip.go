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
			log.Error("Failed to convert element")
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
	// to signal exit gossip loop
	gossipInterval time.Duration
	// quorum timeout to change the quorum status of a node
	quorumTimeout time.Duration
	//nodeDeathInterval time.Duration
	shutDown bool
	// stopQuorumCheck
	stopQuorumCheck chan bool
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
	g.quorumTimeout = gossipIntervals.QuorumTimeout
	g.stopQuorumCheck = make(chan bool)

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
	g.InitGossipDelegate(genNumber, selfNodeId, gossipVersion)
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
	list, err := ml.Create(g.mlConf)
	if err != nil {
		log.Warnf("Unable to create memberlist: " + err.Error())
		return err
	}
	// Set the memberlist in gossiper object
	g.mlist = list
	// We are started gossiping
	g.UpdateSelfStatus(types.NODE_STATUS_WAITING_FOR_QUORUM)

	if len(knownIps) != 0 {
		// Joining an existing cluster
		_, err := list.Join(knownIps)
		if err != nil {
			log.Infof("Unable to join other nodes at startup : %v", err)
			return err
		}
	}

	// Check and update quorum periodically
	go func() {
		for {
			g.CheckAndUpdateQuorum()
			select {
			case stop := <-g.stopQuorumCheck:
				if stop {
					return
				}
			default:
				time.Sleep(g.GossipInterval())
			}
		}
	}()

	return nil
}

func (g *GossiperImpl) Stop(leaveTimeout time.Duration) error {
	if g.shutDown == true {
		return fmt.Errorf("Gossiper already stopped")
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
	g.stopQuorumCheck <- true
	return nil
}

func (g *GossiperImpl) CheckAndUpdateQuorum() {
	clusterSize := g.GetClusterSize()

	quorum := (clusterSize / 2) + 1
	selfNodeId := g.NodeId()

	upNodes := 0
	var selfStatus types.NodeStatus
	localNodeInfoMap := g.GetLocalState()
	for _, nodeInfo := range localNodeInfoMap {
		if nodeInfo.Id == selfNodeId {
			selfStatus = nodeInfo.Status
		}
		if nodeInfo.Status == types.NODE_STATUS_UP ||
			nodeInfo.Status == types.NODE_STATUS_WAITING_FOR_QUORUM ||
			nodeInfo.Status == types.NODE_STATUS_UP_AND_WAITING_FOR_QUORUM {
			upNodes++
		}
	}

	if upNodes < quorum {
		if selfStatus == types.NODE_STATUS_DOWN {
			// We are already down. No need of updating the status based on quorum.
			return
		}
		// We do not have quorum
		if selfStatus == types.NODE_STATUS_UP {
			// We were up, but now we have lost quorum
			log.Warnf("Node %v : %v with status: (UP) lost quorum. "+
				"New Status: (UP_AND_WAITING_FOR_QUORUM)", g.NodeId(), g.mlConf.BindAddr)
			g.UpdateLostQuorumTs()
			g.UpdateSelfStatus(types.NODE_STATUS_UP_AND_WAITING_FOR_QUORUM)
		} else {
			if selfStatus == types.NODE_STATUS_UP_AND_WAITING_FOR_QUORUM {
				waitTime := g.quorumTimeout
				diffTime := time.Since(g.GetLostQuorumTs())
				// Check the difference between the current time and the time
				// when we found out we were up and not in quorum
				if diffTime > waitTime {
					// Change the status to waiting for quorum
					log.Warnf("Quorum Timeout for Node %v : %v with status:"+
						" (UP_AND_WAITING_FOR_QUORUM). "+
						"New Status: (WAITING_FOR_QUORUM)", g.NodeId(), g.mlConf.BindAddr)
					g.UpdateSelfStatus(types.NODE_STATUS_WAITING_FOR_QUORUM)
				}
			} else {
				g.UpdateSelfStatus(types.NODE_STATUS_WAITING_FOR_QUORUM)
			}
		}
	} else {
		if selfStatus != types.NODE_STATUS_UP {
			g.UpdateSelfStatus(types.NODE_STATUS_UP)
		} else {
			// No need to update status, we are already up
		}
	}

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
