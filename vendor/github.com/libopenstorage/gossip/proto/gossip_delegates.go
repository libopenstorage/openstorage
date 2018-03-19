package proto

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/hashicorp/memberlist"

	"github.com/libopenstorage/gossip/proto/state"
	"github.com/libopenstorage/gossip/types"
)

type GossipDelegate struct {
	// GossipstoreImpl implements the GossipStoreInterface
	GossipStoreImpl
	nodeId string
	// last gossip time
	lastGossipTsLock sync.Mutex
	lastGossipTs     time.Time
	// channel to receive state change events
	stateEvent chan types.StateEvent
	// current State object
	currentState state.State
	// quorum timeout to change the quorum status of a node
	quorumTimeout      time.Duration
	timeoutVersion     uint64
	timeoutVersionLock sync.Mutex
}

func (gd *GossipDelegate) InitGossipDelegate(
	genNumber uint64,
	selfNodeId types.NodeId,
	gossipVersion string,
	quorumTimeout time.Duration,
	clusterId string,
) {
	gd.GenNumber = genNumber
	gd.nodeId = string(selfNodeId)
	gd.stateEvent = make(chan types.StateEvent)
	// We start with a NOT_IN_QUORUM status
	gd.InitStore(
		selfNodeId,
		gossipVersion,
		types.NODE_STATUS_NOT_IN_QUORUM,
		clusterId,
	)
	gd.quorumTimeout = quorumTimeout
}

func (gd *GossipDelegate) InitCurrentState(clusterSize uint) {
	// Our initial state is NOT_IN_QUORUM
	gd.currentState = state.GetNotInQuorum(
		uint(clusterSize), types.NodeId(gd.nodeId), gd.stateEvent)
	// Start the go routine which handles all the events
	// and changes state of the node
	go gd.handleStateEvents()
}

func (gd *GossipDelegate) updateGossipTs() {
	gd.lastGossipTsLock.Lock()
	defer gd.lastGossipTsLock.Unlock()
	gd.lastGossipTs = time.Now()
}

func (gd *GossipDelegate) gossipChecks(node *memberlist.Node) error {
	// Check the gossip version of other node
	var nodeMeta types.NodeMetaInfo
	nodeName := gd.parseMemberlistNodeName(node.Name)
	err := gd.convertFromBytes(node.Meta, &nodeMeta)
	if err != nil {
		err = fmt.Errorf("gossip: Error in unmarshalling peer's meta data. Error : %v", err.Error())
	} else {
		if nodeMeta.GossipVersion != gd.GetGossipVersion() {
			// Version Mismatch
			// We do not add this node in our memberlist
			err = fmt.Errorf("Version mismatch with "+
				"Node (%v):(%v). Our version: (%v). Their version: (%v)",
				nodeName, node.Addr, gd.GetGossipVersion(), nodeMeta.GossipVersion)
		} else {
			// Version Match
			// Check for ClusterId match
			if nodeMeta.ClusterId != gd.GetClusterId() {
				// ClusterId Mismatch
				// We do not add this node in our memberlist
				err = fmt.Errorf("(%v) ClusterId mismatch with"+
					" Node (%v):(%v). Our clusterId: (%v). Their clusterId: (%v)",
					gd.nodeId, nodeName, node.Addr, gd.GetClusterId(), nodeMeta.ClusterId)
			} else {
				// ClusterId Match
				// Add this new node in our node map
				err = nil
			}
		}
	}
	return err
}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (gd *GossipDelegate) NodeMeta(limit int) []byte {
	msg := gd.MetaInfo()
	msgBytes, _ := gd.convertToBytes(msg)
	return msgBytes
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed.
// Note: Currently, we do not use broadcasts and hence this function does nothing
func (gd *GossipDelegate) NotifyMsg(data []byte) {
	var nodeId string
	json.Unmarshal(data, &nodeId)
	return
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit. Care should be taken that this method does not block,
// since doing so would block the entire UDP packet receive loop.
// Note: Currently, we do not use broadcasts and hence this function does nothing
func (gd *GossipDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	var test [][]byte
	s1, _ := json.Marshal(gd.nodeId)
	s2, _ := json.Marshal("test_string")
	test = append(test, s1)
	test = append(test, s2)
	return test
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (gd *GossipDelegate) LocalState(join bool) []byte {
	gd.updateSelfTs()

	// We send our local state of nodeMap
	// The receiver will decide which nodes to merge and which to ignore
	byteLocalState, err := gd.GetLocalStateInBytes()
	if err != nil {
		byteLocalState = []byte{}
	}
	gd.updateGossipTs()
	return byteLocalState
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (gd *GossipDelegate) MergeRemoteState(buf []byte, join bool) {
	var remoteState types.NodeInfoMap
	if join == true {
		// NotifyJoin will take care of this info
		return
	}
	gd.updateSelfTs()

	err := gd.convertFromBytes(buf, &remoteState)
	if err != nil {
		logrus.Infof("gossip: Error in unmarshalling peer's local data. "+
			"Error : %v", err.Error())
	}

	gd.Update(remoteState)
	gd.updateGossipTs()
	return
}

// NotifyJoin is invoked when a node is detected to have joined.
// The Node argument must not be modified.
func (gd *GossipDelegate) NotifyJoin(node *memberlist.Node) {
	// Ignore self NotifyJoin
	nodeName := gd.parseMemberlistNodeName(node.Name)
	if nodeName == gd.nodeId {
		return
	}

	gd.updateGossipTs()

	// NotifyAlive should remove a node from memberlist if the
	// gossip version mismatches.
	// Nevertheless we are doing an extra check here.
	if err := gd.gossipChecks(node); err != nil {
		gd.RemoveNode(types.NodeId(nodeName))
	}
}

// NotifyLeave is invoked when a node is detected to have left.
// The Node argument must not be modified.
func (gd *GossipDelegate) NotifyLeave(node *memberlist.Node) {
	nodeName := gd.parseMemberlistNodeName(node.Name)
	if nodeName == gd.nodeId {
		gd.triggerStateEvent(types.SELF_LEAVE)
	} else {
		err := gd.UpdateNodeStatus(types.NodeId(nodeName), types.NODE_STATUS_DOWN)
		if err != nil {
			logrus.Infof("gossip: Could not update status on NotifyLeave : %v", err.Error())
			return
		}
		gd.triggerStateEvent(types.NODE_LEAVE)
	}

	gd.updateGossipTs()
	return
}

// NotifyUpdate is invoked when a node is detected to have
// updated, usually involving the meta data. The Node argument
// must not be modified.
// Note: Currently we do not use memberlists Node meta or modify it.
// Probably future use ?
func (gd *GossipDelegate) NotifyUpdate(node *memberlist.Node) {
	nodeName := gd.parseMemberlistNodeName(node.Name)
	logrus.Infof("gossip: Update Notification from %v %v", nodeName, node.Addr)
}

func (gd *GossipDelegate) NotifyMerge(peers []*memberlist.Node) error {
	for _, peer := range peers {
		err := gd.gossipChecks(peer)
		if err != nil {
			return err
		}
	}
	return nil
}

// AliveDelegate is used to involve a client in processing a node "alive" message.
// TODO/Future-use : Check if we want to add this node in memberlist
func (gd *GossipDelegate) NotifyAlive(node *memberlist.Node) error {
	nodeName := gd.parseMemberlistNodeName(node.Name)
	if nodeName == gd.nodeId {
		gd.triggerStateEvent(types.SELF_ALIVE)
		return nil
	}

	gd.updateGossipTs()

	err := gd.gossipChecks(node)
	if err != nil {
		gd.RemoveNode(types.NodeId(nodeName))
		// Do not add this node to the memberlist.
		// Returning a non-nil err value
		return err
	}

	diffNode, err := gd.GetLocalNodeInfo(types.NodeId(nodeName))
	if err == nil && diffNode.Status != types.NODE_STATUS_UP {
		gd.UpdateNodeStatus(types.NodeId(nodeName), types.NODE_STATUS_UP)
		gd.triggerStateEvent(types.NODE_ALIVE)
	} // else if err != nil -> A new node sending us data. We do not add node unless it is added
	// in our local map externally
	return nil
}

func (gd *GossipDelegate) triggerStateEvent(event types.StateEvent) {
	gd.stateEvent <- event
	return
}

func (gd *GossipDelegate) startQuorumTimer() {
	gd.timeoutVersionLock.Lock()
	localVersion := gd.timeoutVersion + 1
	gd.timeoutVersion = localVersion
	gd.timeoutVersionLock.Unlock()

	logrus.Infof("gossip: Starting Quorum Timer with version v%v. Waiting for quorum timeout of (%v)", localVersion, gd.quorumTimeout)
	time.Sleep(gd.quorumTimeout)

	gd.timeoutVersionLock.Lock()
	if localVersion == gd.timeoutVersion {
		gd.timeoutVersionLock.Unlock()
		gd.stateEvent <- types.TIMEOUT
		return
	} // else do not send an event. Another timer started
	gd.timeoutVersionLock.Unlock()
}

func (gd *GossipDelegate) parseMemberlistNodeName(nodeName string) string {
	return strings.TrimSuffix(nodeName, gd.GetGossipVersion())
}

func (gd *GossipDelegate) handleStateEvents() {
	for {
		// We block here until we get an event
		event := <-gd.stateEvent
		previousStatus := gd.currentState.NodeStatus()
		switch event {
		case types.SELF_ALIVE:
			gd.currentState, _ = gd.currentState.SelfAlive(gd.GetLocalState())
		case types.NODE_ALIVE:
			gd.currentState, _ = gd.currentState.NodeAlive(gd.GetLocalState())
		case types.SELF_LEAVE:
			gd.currentState, _ = gd.currentState.SelfLeave()
		case types.NODE_LEAVE:
			gd.currentState, _ = gd.currentState.NodeLeave(gd.GetLocalState())
		case types.UPDATE_CLUSTER_SIZE:
			gd.currentState, _ = gd.currentState.UpdateClusterSize(
				gd.getNumQuorumMembers(), gd.GetLocalState())
		case types.TIMEOUT:
			newState, _ := gd.currentState.Timeout(
				gd.getNumQuorumMembers(), gd.GetLocalState())
			if newState.NodeStatus() != gd.currentState.NodeStatus() {
				logrus.Infof("gossip: Quorum Timeout. Waited for (%v)",
					gd.quorumTimeout)
			}
			gd.currentState = newState
		}
		newStatus := gd.currentState.NodeStatus()
		if previousStatus == types.NODE_STATUS_UP &&
			newStatus == types.NODE_STATUS_SUSPECT_NOT_IN_QUORUM {
			// Start a timer
			go gd.startQuorumTimer()
		}
		gd.UpdateSelfStatus(gd.currentState.NodeStatus())
	}
}
