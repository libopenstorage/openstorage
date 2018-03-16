package proto

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/libopenstorage/gossip/types"
)

const (
	INVALID_GEN_NUMBER = 0
)

type GossipStoreImpl struct {
	sync.Mutex
	id            types.NodeId
	GenNumber     uint64
	nodeMap       types.NodeInfoMap
	selfCorrect   bool
	GossipVersion string
	ClusterId     string
	// This cluster size is updated from an external source
	// such as a kv database. This is an extra measure to find the
	// number of nodes in the cluster other than just relying on
	// memberlist and the length of nodeMap. It is used in
	// determining the cluster quorum
	clusterSize uint
	// numQuorumMembers is the number of members which participate in quorum
	numQuorumMembers uint
	// Ts at which we lost quorum
	lostQuorumTs time.Time
}

func NewGossipStore(id types.NodeId, version, clusterId string) *GossipStoreImpl {
	n := &GossipStoreImpl{}
	n.InitStore(id, version, types.NODE_STATUS_NOT_IN_QUORUM, clusterId)
	n.selfCorrect = false
	return n
}

func (s *GossipStoreImpl) NodeId() types.NodeId {
	return s.id
}

func (s *GossipStoreImpl) UpdateLostQuorumTs() {
	s.Lock()
	defer s.Unlock()

	s.lostQuorumTs = time.Now()
}

func (s *GossipStoreImpl) GetLostQuorumTs() time.Time {
	return s.lostQuorumTs
}

func (s *GossipStoreImpl) InitStore(
	id types.NodeId,
	version string,
	status types.NodeStatus,
	clusterId string,
) {
	s.nodeMap = make(types.NodeInfoMap)
	s.id = id
	s.selfCorrect = true
	s.GossipVersion = version
	s.ClusterId = clusterId
	nodeInfo := types.NodeInfo{
		Id:           s.id,
		GenNumber:    s.GenNumber,
		Value:        make(types.StoreMap),
		LastUpdateTs: time.Now(),
		Status:       status,
	}
	s.nodeMap[s.id] = nodeInfo
}

func (s *GossipStoreImpl) updateSelfTs() {
	s.Lock()
	defer s.Unlock()

	nodeInfo, _ := s.nodeMap[s.id]
	nodeInfo.LastUpdateTs = time.Now()
	s.nodeMap[s.id] = nodeInfo
}

func (s *GossipStoreImpl) UpdateSelf(key types.StoreKey, val interface{}) {
	s.Lock()
	defer s.Unlock()

	nodeInfo, _ := s.nodeMap[s.id]
	nodeInfo.Value[key] = val
	nodeInfo.LastUpdateTs = time.Now()
	s.nodeMap[s.id] = nodeInfo
}

func (s *GossipStoreImpl) UpdateSelfStatus(status types.NodeStatus) {
	s.UpdateNodeStatus(s.id, status)
}

func (s *GossipStoreImpl) GetSelfStatus() types.NodeStatus {
	s.Lock()
	defer s.Unlock()

	nodeInfo, _ := s.nodeMap[s.id]
	return nodeInfo.Status
}

func (s *GossipStoreImpl) UpdateNodeStatus(nodeId types.NodeId, status types.NodeStatus) error {
	s.Lock()
	defer s.Unlock()

	nodeInfo, ok := s.nodeMap[nodeId]
	if !ok {
		return fmt.Errorf("Node with id (%v) not found", nodeId)
	}
	nodeInfo.Status = status
	nodeInfo.LastUpdateTs = time.Now()
	s.nodeMap[nodeId] = nodeInfo
	return nil
}

func (s *GossipStoreImpl) GetStoreKeyValue(key types.StoreKey) types.NodeValueMap {
	s.Lock()
	defer s.Unlock()

	nodeValueMap := make(types.NodeValueMap)
	for id, nodeInfo := range s.nodeMap {
		if statusValid(nodeInfo.Status) && nodeInfo.Value != nil {
			ok := len(nodeInfo.Value) == 0
			val, exists := nodeInfo.Value[key]
			if ok || exists {
				n := types.NodeValue{Id: nodeInfo.Id,
					GenNumber:    nodeInfo.GenNumber,
					LastUpdateTs: nodeInfo.LastUpdateTs,
					Status:       nodeInfo.Status}
				n.Value = val
				nodeValueMap[id] = n
			}
		}
	}
	return nodeValueMap
}

func (s *GossipStoreImpl) GetStoreKeys() []types.StoreKey {
	s.Lock()
	defer s.Unlock()

	keyMap := make(map[types.StoreKey]bool)
	for _, nodeInfo := range s.nodeMap {
		if nodeInfo.Value != nil {
			for key := range nodeInfo.Value {
				keyMap[key] = true
			}
		}
	}
	storeKeys := make([]types.StoreKey, len(keyMap))
	i := 0
	for key := range keyMap {
		storeKeys[i] = key
		i++
	}
	return storeKeys
}

func (s *GossipStoreImpl) GetGossipVersion() string {
	return s.GossipVersion
}

func (s *GossipStoreImpl) GetClusterId() string {
	return s.ClusterId
}

func statusValid(s types.NodeStatus) bool {
	return (s != types.NODE_STATUS_INVALID &&
		s != types.NODE_STATUS_NEVER_GOSSIPED)
}

func (s *GossipStoreImpl) AddNode(
	id types.NodeId,
	status types.NodeStatus,
	quorumMember bool,
) {
	s.Lock()
	defer s.Unlock()
	s.addNodeUnlocked(id, status, quorumMember)
}

func (s *GossipStoreImpl) addNodeUnlocked(
	id types.NodeId,
	status types.NodeStatus,
	quorumMember bool,
) {
	if nodeInfo, ok := s.nodeMap[id]; ok {
		nodeInfo.Status = status
		nodeInfo.LastUpdateTs = time.Now()
		nodeInfo.QuorumMember = quorumMember
		s.nodeMap[id] = nodeInfo
		return
	}

	s.nodeMap[id] = types.NodeInfo{
		Id:                 id,
		GenNumber:          0,
		LastUpdateTs:       time.Now(),
		WaitForGenUpdateTs: time.Now(),
		Status:             status,
		Value:              make(types.StoreMap),
		QuorumMember:       quorumMember,
	}
	logrus.Infof("gossip: Adding Node to gossip map: %v", id)
}

func (s *GossipStoreImpl) RemoveNode(id types.NodeId) error {
	s.Lock()
	defer s.Unlock()
	return s.removeNodeUnlocked(id)
}

func (s *GossipStoreImpl) removeNodeUnlocked(id types.NodeId) error {
	if _, ok := s.nodeMap[id]; !ok {
		return fmt.Errorf("Node %v does not exist in map", id)
	}
	logrus.Infof("gossip: Removing node from gossip map: %v", id)
	delete(s.nodeMap, id)
	return nil
}

func (s *GossipStoreImpl) MetaInfo() types.NodeMetaInfo {
	s.Lock()
	defer s.Unlock()

	selfNodeInfo, _ := s.nodeMap[s.id]
	nodeMetaInfo := types.NodeMetaInfo{
		Id:            selfNodeInfo.Id,
		LastUpdateTs:  selfNodeInfo.LastUpdateTs,
		GossipVersion: s.GossipVersion,
		ClusterId:     s.ClusterId,
	}
	return nodeMetaInfo
}

func (s *GossipStoreImpl) GetLocalState() types.NodeInfoMap {
	s.Lock()
	defer s.Unlock()
	return s.getLocalState()
}

func (s *GossipStoreImpl) GetLocalStateInBytes() ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	return s.convertToBytes(s.getLocalState())
}

func (s *GossipStoreImpl) GetLocalNodeInfo(id types.NodeId) (types.NodeInfo, error) {
	s.Lock()
	defer s.Unlock()

	nodeInfo, ok := s.nodeMap[id]
	if !ok {
		return types.NodeInfo{}, fmt.Errorf("Node with id (%v) not found", id)
	}
	return nodeInfo, nil
}

func (s *GossipStoreImpl) Update(diff types.NodeInfoMap) {
	s.Lock()
	defer s.Unlock()

	for id, newNodeInfo := range diff {
		if id == s.id {
			continue
		}
		selfValue, ok := s.nodeMap[id]
		if !ok {
			// Ignore updates for a node which we do not know about.
			continue
		}
		if !statusValid(selfValue.Status) ||
			selfValue.LastUpdateTs.Before(newNodeInfo.LastUpdateTs) {
			// Our view of Status of a Node, should only be determined by
			// memberlist. We should not update the Status field in our
			// nodeInfo based on what other node's value is.
			newNodeInfo.Status = selfValue.Status
			s.nodeMap[id] = newNodeInfo
		}
	}
}

func (s *GossipStoreImpl) updateCluster(
	peers map[types.NodeId]types.NodeUpdate,
) {
	removeNodeIds := []types.NodeId{}
	addNodeIds := []types.NodeId{}
	s.Lock()
	defer s.Unlock()
	s.clusterSize = uint(len(peers))
	// Nodes removed
	for id := range s.nodeMap {
		if _, ok := peers[id]; !ok {
			removeNodeIds = append(removeNodeIds, id)
		}
	}
	// Nodes added
	for id := range peers {
		if _, ok := s.nodeMap[id]; !ok {
			addNodeIds = append(addNodeIds, id)
		}
	}

	for _, nodeId := range removeNodeIds {
		s.removeNodeUnlocked(nodeId)
	}
	for _, nodeId := range addNodeIds {
		update, _ := peers[nodeId]
		s.addNodeUnlocked(nodeId, types.NODE_STATUS_DOWN, update.QuorumMember)
	}

	// Update quorum members
	s.numQuorumMembers = 0
	for id, nodeInfo := range s.nodeMap {
		if update, ok := peers[id]; ok {
			nodeInfo.QuorumMember = update.QuorumMember
			s.nodeMap[id] = nodeInfo
		}
		if nodeInfo.QuorumMember {
			s.numQuorumMembers++
		}
	}
}

func (s *GossipStoreImpl) getNumQuorumMembers() uint {
	return s.numQuorumMembers
}

func (s *GossipStoreImpl) convertToBytes(obj interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(obj)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func (s *GossipStoreImpl) convertFromBytes(buf []byte, msg interface{}) error {
	msgBuffer := bytes.NewBuffer(buf)
	dec := gob.NewDecoder(msgBuffer)
	err := dec.Decode(msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *GossipStoreImpl) getLocalState() types.NodeInfoMap {
	localCopy := make(types.NodeInfoMap)
	for key, value := range s.nodeMap {
		localCopy[key] = value
	}
	return localCopy
}
