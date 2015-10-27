package proto

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libopenstorage/gossip/types"
)

const (
	DEFAULT_GOSSIP_INTERVAL     = 2 * time.Second
	DEFAULT_NODE_DEATH_INTERVAL = 15 * DEFAULT_GOSSIP_INTERVAL
)

// Implements the UnreliableBroadcast interface
type GossiperImpl struct {
	// GossipstoreImpl implements the GossipStoreInterface
	GossipStoreImpl

	// node list, maintained separately
	nodes     []string
	name      string
	nodesLock sync.Mutex
	// to signal exit gossip loop
	send_done         chan bool
	rcv_done          chan bool
	update_done       chan bool
	gossipInterval    time.Duration
	nodeDeathInterval time.Duration
	peerSelector      PeerSelector
}

// Utility methods
func logAndGetError(msg string) error {
	log.Error(msg)
	return errors.New(msg)
}

type PeerSelector interface {
	SetMaxLen(uint32)
	NextPeer() int32
	SetStartHint(m uint32)
}

type RoundRobinPeerSelector struct {
	maxLen       uint32
	lastSelected uint32
}

func (r *RoundRobinPeerSelector) Init() {
	r.maxLen = 0
	r.lastSelected = 0
}

func (r *RoundRobinPeerSelector) SetStartHint(m uint32) {
	maxLen := atomic.LoadUint32(&r.maxLen)
	var lastSelected uint32
	lastSelected = 0
	if m != maxLen {
		lastSelected = uint32((m + 1) % maxLen)
	}
	atomic.StoreUint32(&r.lastSelected, lastSelected)
}

func (r *RoundRobinPeerSelector) SetMaxLen(m uint32) {
	if m > math.MaxUint16 {
		log.Panicf("Number of peers %v greater than those suported %v",
			m, math.MaxUint16)
	}
	atomic.StoreUint32(&r.maxLen, m)
}

func (r *RoundRobinPeerSelector) NextPeer() int32 {
	maxLen := atomic.LoadUint32(&r.maxLen)
	lastSelected := atomic.LoadUint32(&r.lastSelected)
	if maxLen < 1 {
		return -1
	}

	atomic.StoreUint32(&r.lastSelected, (lastSelected+1)%maxLen)
	return int32(r.lastSelected)
}

func NewPeerSelector() PeerSelector {
	s := new(RoundRobinPeerSelector)
	s.Init()
	return s
}

func (g *GossiperImpl) Init(ip string, selfNodeId types.NodeId) {
	g.InitStore(selfNodeId)
	g.name = ip
	g.nodes = make([]string, 0)
	g.send_done = make(chan bool, 1)
	g.rcv_done = make(chan bool, 1)
	g.update_done = make(chan bool, 1)
	g.gossipInterval = DEFAULT_GOSSIP_INTERVAL
	g.nodeDeathInterval = DEFAULT_NODE_DEATH_INTERVAL
	g.peerSelector = NewPeerSelector()
	rand.Seed(time.Now().UnixNano())
}

func (g *GossiperImpl) Start() {
	// start gossiping
	go g.sendLoop()
	go g.receiveLoop()
	go g.updateStatusLoop()
}

func (g *GossiperImpl) Stop() {
	if g.send_done != nil {
		g.send_done <- true
		g.send_done = nil
	}
	if g.rcv_done != nil {
		g.rcv_done <- true
		g.rcv_done = nil
	}
	if g.update_done != nil {
		g.update_done <- true
		g.update_done = nil
	}
}

func (g *GossiperImpl) SetGossipInterval(t time.Duration) {
	g.gossipInterval = t
}

func (g *GossiperImpl) GossipInterval() time.Duration {
	return g.gossipInterval
}

func (g *GossiperImpl) SetNodeDeathInterval(t time.Duration) {
	g.nodeDeathInterval = t
}

func (g *GossiperImpl) NodeDeathInterval() time.Duration {
	return g.nodeDeathInterval
}

func (g *GossiperImpl) AddNode(ip string) error {
	g.nodesLock.Lock()
	defer g.nodesLock.Unlock()

	for _, node := range g.nodes {
		if node == ip {
			return logAndGetError("Node being added already exists:" + ip)
		}
	}
	g.nodes = append(g.nodes, ip)
	sort.Strings(g.nodes)
	g.peerSelector.SetMaxLen(uint32(len(g.nodes)))
	if len(g.nodes) >= 2 {
		// In order to make sure that not all of the
		// nodes go in the same order, try to reset the order
		// by sorting the nodes by name and starting at the position
		// next to this node
		temp := make([]string, len(g.nodes))
		copy(temp, g.nodes)
		temp = append(temp, g.name)
		sort.Strings(temp)
		for i, n := range temp {
			if n == g.name {
				g.peerSelector.SetStartHint(uint32(i % len(g.nodes)))
			}
		}
	}

	return nil
}

func (g *GossiperImpl) RemoveNode(ip string) error {
	g.nodesLock.Lock()
	defer g.nodesLock.Unlock()

	for i, node := range g.nodes {
		if node == ip {
			// not sure if this is the most efficient way
			g.nodes = append(g.nodes[:i], g.nodes[i+1:]...)
			g.peerSelector.SetMaxLen(uint32(len(g.nodes)))
			return nil
		}
	}
	return logAndGetError("Node being added already exists:" + ip)
}

func (g *GossiperImpl) GetNodes() []string {
	g.nodesLock.Lock()
	defer g.nodesLock.Unlock()

	nodeList := make([]string, len(g.nodes))
	copy(nodeList, g.nodes)
	return nodeList
}

// getUpdatesFromPeer receives node data from the peer
// for which the peer has more latest information available
func (g *GossiperImpl) getUpdatesFromPeer(conn types.MessageChannel) error {

	var newPeerData types.StoreDiff
	err := conn.RcvData(&newPeerData)
	if err != nil {
		log.Error("Error fetching the latest peer data", err)
		return err
	}

	g.Update(newPeerData)

	return nil
}

// sendNodeMetaInfo sends a list of meta info for all
// the nodes in the nodes's store to the peer
func (g *GossiperImpl) sendNodeMetaInfo(conn types.MessageChannel) error {
	msg := g.MetaInfo()
	err := conn.SendData(&msg)
	return err
}

// sendUpdatesToPeer sends the information about the given
// nodes to the peer
func (g *GossiperImpl) sendUpdatesToPeer(diff *types.StoreNodes,
	conn types.MessageChannel) error {
	dataToSend := g.Subset(*diff)
	return conn.SendData(&dataToSend)
}

func (g *GossiperImpl) handleGossip(conn types.MessageChannel) {
	log.Debug(g.id, " Servicing gossip request")
	var peerMetaInfo types.StoreMetaInfo
	err := error(nil)

	// Get the info about the node data that the sender has
	err = conn.RcvData(&peerMetaInfo)
	log.Debug(g.id, " Got meta data: \n", peerMetaInfo)
	if err != nil {
		return
	}

	// 2. Compare with current data that this node has and get
	//    the names of the nodes for which this node has stale info
	//    as compared to the sender
	diffNew, selfNew := g.Diff(peerMetaInfo)
	log.Debug(g.id, " The diff is: diffNew: \n", diffNew, " \nselfNew:\n", selfNew)

	// Send this list to the peer, and get the latest data
	// for them
	err = conn.SendData(diffNew)
	if err != nil {
		log.Error("Error sending list of nodes to fetch: ", err)
		return
	}

	// get the data for nodes sent above from the peer
	err = g.getUpdatesFromPeer(conn)
	if err != nil {
		log.Error("Failed to get data for nodes from the peer: ", err)
		return
	}

	// Since you know which data is stale on the sender side,
	// send him the data for the updated nodes
	err = g.sendUpdatesToPeer(&selfNew, conn)
	if err != nil {
		return
	}
	log.Debug(g.id, " Finished Servicing gossip request")
}

func (g *GossiperImpl) receiveLoop() {
	var handler types.OnMessageRcv = func(c types.MessageChannel) { g.handleGossip(c) }
	c := NewRunnableMessageChannel(g.name, handler)
	go c.RunOnRcvData()
	// block waiting for the done signal
	<-g.rcv_done
	c.Close()
}

// sendLoop periodically connects to a random peer
// and gossips about the state of the cluster
func (g *GossiperImpl) sendLoop() {
	tick := time.Tick(g.gossipInterval)
	for {
		select {
		case <-tick:
			g.gossip()
		case <-g.send_done:
			return
		}
	}
}

// updateStatusLoop updates the status of each node
// depending on when it was last updated
func (g *GossiperImpl) updateStatusLoop() {
	tick := time.Tick(g.gossipInterval)
	for {
		select {
		case <-tick:
			g.UpdateNodeStatuses(g.nodeDeathInterval)
		case <-g.update_done:
			return
		}
	}
}

// selectGossipPeer randomly selects a peer
// to gossip with from the list of nodes added
func (g *GossiperImpl) selectGossipPeer() string {
	g.nodesLock.Lock()
	defer g.nodesLock.Unlock()

	nodesLen := len(g.nodes)
	if nodesLen == 0 {
		return ""
	}

	peer := g.peerSelector.NextPeer()
	if peer < 0 {
		return ""
	}
	return g.nodes[peer]
}

func (g *GossiperImpl) gossip() {

	// select a node to gossip with
	peerNode := g.selectGossipPeer()
	if len(peerNode) == 0 {
		return
	}
	log.Debug("Starting gossip with ", peerNode)

	conn := NewMessageChannel(peerNode)
	if conn == nil {
		//XXX: FIXME : note that the peer is down
		return
	}

	// send meta data info about the node to the peer
	err := g.sendNodeMetaInfo(conn)
	if err != nil {
		log.Error("Failed to send meta info to the peer: ", err)
		//XXX: FIXME : note that the peer is down
		return
	}

	// get a list of requested nodes from the peer and
	var diff types.StoreNodes
	err = conn.RcvData(&diff)
	if err != nil {
		log.Error("Failed to get request info to the peer: ", err)
		//XXX: FIXME : note that the peer is down
		return
	}

	// send back the data
	err = g.sendUpdatesToPeer(&diff, conn)
	if err != nil {
		log.Error("Failed to send newer data to the peer: ", err)
		//XXX: FIXME : note that the peer is down
		return
	}

	// receive any updates the send has for us
	err = g.getUpdatesFromPeer(conn)
	if err != nil {
		log.Error("Failed to get newer data from the peer: ", err)
		//XXX: FIXME : note that the peer is down
		return
	}
	log.Debug("Ending gossip with ", peerNode)

}
