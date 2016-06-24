package proto

import (
	"fmt"
	"sync"
	"time"

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
}

func NewGossipStore(id types.NodeId, version string) *GossipStoreImpl {
	n := &GossipStoreImpl{}
	n.InitStore(id, version)
	n.selfCorrect = false
	return n
}

func (s *GossipStoreImpl) NodeId() types.NodeId {
	return s.id
}

func (s *GossipStoreImpl) InitStore(id types.NodeId, version string) {
	s.nodeMap = make(types.NodeInfoMap)
	s.id = id
	s.selfCorrect = true
	s.GossipVersion = version
	nodeInfo := types.NodeInfo{
		Id:           s.id,
		GenNumber:    s.GenNumber,
		Value:        make(types.StoreMap),
		LastUpdateTs: time.Now(),
		Status:       types.NODE_STATUS_UP,
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

func statusValid(s types.NodeStatus) bool {
	return (s != types.NODE_STATUS_INVALID &&
		s != types.NODE_STATUS_NEVER_GOSSIPED)
}

func (s *GossipStoreImpl) NewNode(id types.NodeId) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.nodeMap[id]; ok {
		return
	}

	newNodeInfo := types.NodeInfo{
		Id:                 id,
		GenNumber:          0,
		LastUpdateTs:       time.Now(),
		WaitForGenUpdateTs: time.Now(),
		Status:             types.NODE_STATUS_UP,
		Value:              make(types.StoreMap),
	}

	s.nodeMap[id] = newNodeInfo
}

func (s *GossipStoreImpl) MetaInfo() types.NodeMetaInfo {
	s.Lock()
	defer s.Unlock()

	selfNodeInfo, _ := s.nodeMap[s.id]
	nodeMetaInfo := types.NodeMetaInfo{
		Id:            selfNodeInfo.Id,
		LastUpdateTs:  selfNodeInfo.LastUpdateTs,
		GenNumber:     selfNodeInfo.GenNumber,
		GossipVersion: s.GossipVersion,
	}
	return nodeMetaInfo
}

func (s *GossipStoreImpl) GetLocalState() types.NodeInfoMap {
	s.Lock()
	defer s.Unlock()
	return s.nodeMap
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
		if !ok || !statusValid(selfValue.Status) ||
			selfValue.LastUpdateTs.Before(newNodeInfo.LastUpdateTs) {
			// Our view of Status of a Node, should only be determined by memberlist.
			// We should not update the Status field in our nodeInfo based on what other node's
			// value is.
			newNodeInfo.Status = selfValue.Status
			s.nodeMap[id] = newNodeInfo
		}
	}
}
