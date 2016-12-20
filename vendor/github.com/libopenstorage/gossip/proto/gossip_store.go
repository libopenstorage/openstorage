package proto

import (
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
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
	clusterSize int
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
	s.Lock()
	defer s.Unlock()

	nodeInfo, _ := s.nodeMap[s.id]
	nodeInfo.Status = status
	nodeInfo.LastUpdateTs = time.Now()
	s.nodeMap[s.id] = nodeInfo
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
			for key, _ := range nodeInfo.Value {
				keyMap[key] = true
			}
		}
	}
	storeKeys := make([]types.StoreKey, len(keyMap))
	i := 0
	for key, _ := range keyMap {
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

func (s *GossipStoreImpl) AddNode(id types.NodeId, status types.NodeStatus) {
	s.Lock()
	if nodeInfo, ok := s.nodeMap[id]; ok {
		nodeInfo.Status = status
		nodeInfo.LastUpdateTs = time.Now()
		s.nodeMap[id] = nodeInfo
		s.Unlock()
		return
	}

	newNodeInfo := types.NodeInfo{
		Id:                 id,
		GenNumber:          0,
		LastUpdateTs:       time.Now(),
		WaitForGenUpdateTs: time.Now(),
		Status:             status,
		Value:              make(types.StoreMap),
	}
	s.nodeMap[id] = newNodeInfo
	logrus.Infof("gossip: Adding Node to gossip map: %v", id)
	s.Unlock()
}

func (s *GossipStoreImpl) RemoveNode(id types.NodeId) error {
	s.Lock()
	if _, ok := s.nodeMap[id]; !ok {
		s.Unlock()
		return fmt.Errorf("Node %v does not exist in map", id)
	}
	logrus.Infof("gossip: Removing node from gossip map: %v", id)
	delete(s.nodeMap, id)
	s.Unlock()
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
		ClusterId: s.ClusterId,
	}
	return nodeMetaInfo
}

func (s *GossipStoreImpl) GetLocalState() types.NodeInfoMap {
	s.Lock()
	defer s.Unlock()
	localCopy := make(types.NodeInfoMap)
	for key, value := range s.nodeMap {
		localCopy[key] = value
	}
	return localCopy
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
			// We got an update for a node which we do not have in our map
			// Lets add it with an offline state
			continue
		}
		if !statusValid(selfValue.Status) ||
			selfValue.LastUpdateTs.Before(newNodeInfo.LastUpdateTs) {
			// Our view of Status of a Node, should only be determined by memberlist.
			// We should not update the Status field in our nodeInfo based on what other node's
			// value is.
			newNodeInfo.Status = selfValue.Status
			s.nodeMap[id] = newNodeInfo
		}
	}
}

func (s *GossipStoreImpl) updateCluster(peers map[types.NodeId]string) {
	removeNodeIds := []types.NodeId{}
	addNodeIds := []types.NodeId{}
	s.Lock()
	s.clusterSize = len(peers)
	// Lets check if a node was added or removed.
	if len(s.nodeMap) > len(peers) {
		// Node removed
		for id, _ := range s.nodeMap {
			if _, ok := peers[id]; !ok {
				removeNodeIds = append(removeNodeIds, id)
			}
		}
	} else if len(s.nodeMap) < len(peers) {
		// Node added
		for id, _ := range peers {
			if _, ok := s.nodeMap[id]; !ok {
				addNodeIds = append(addNodeIds, id)
			}
		}
	} else {
		// Nodes removed
		for id, _ := range s.nodeMap {
			if _, ok := peers[id]; !ok {
				removeNodeIds = append(removeNodeIds, id)
			}
		}
		// Nodes added
		for id, _ := range peers {
			if _, ok := s.nodeMap[id]; !ok {
				addNodeIds = append(addNodeIds, id)
			}
		}
	}
	s.Unlock()
	for _, nodeId := range removeNodeIds {
		s.RemoveNode(nodeId)
	}
	for _, nodeId := range addNodeIds {
		s.AddNode(nodeId, types.NODE_STATUS_DOWN)
	}
}

func (s *GossipStoreImpl) getClusterSize() int {
	return s.clusterSize
}
