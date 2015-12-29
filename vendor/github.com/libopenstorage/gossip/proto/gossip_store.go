package proto

import (
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"

	"github.com/libopenstorage/gossip/types"
)

const (
	INVALID_GEN_NUMBER = 0
)

type GossipStoreImpl struct {
	sync.Mutex
	id          types.NodeId
	GenNumber   uint64
	nodeMap     types.NodeInfoMap
	selfCorrect bool
}

func NewGossipStore(id types.NodeId) *GossipStoreImpl {
	n := &GossipStoreImpl{}
	n.InitStore(id)
	n.selfCorrect = false
	return n
}

func (s *GossipStoreImpl) NodeId() types.NodeId {
	return s.id
}

func (s *GossipStoreImpl) InitStore(id types.NodeId) {
	s.nodeMap = make(types.NodeInfoMap)
	s.id = id
	s.selfCorrect = true
}

func (s *GossipStoreImpl) UpdateSelf(key types.StoreKey, val interface{}) {
	s.Lock()
	defer s.Unlock()

	nodeInfo, ok := s.nodeMap[s.id]
	if !ok {
		nodeInfo = types.NodeInfo{Id: s.id,
			GenNumber:    s.GenNumber,
			Value:        make(types.StoreMap),
			LastUpdateTs: time.Now(),
			Status:       types.NODE_STATUS_UP}
		s.nodeMap[s.id] = nodeInfo
	}

	nodeInfo.Value[key] = val
	nodeInfo.LastUpdateTs = time.Now()
	s.nodeMap[s.id] = nodeInfo
}

func (s *GossipStoreImpl) MarkNodeHasOldGen(nodeId types.NodeId) bool {
	s.Lock()
	defer s.Unlock()

	nodeInfo, ok := s.nodeMap[nodeId]
	if !ok {
		return false
	}

	nodeInfo.Status = types.NODE_STATUS_WAITING_FOR_NEW_UPDATE
	nodeInfo.WaitForGenUpdateTs = time.Now()
	s.nodeMap[nodeId] = nodeInfo
	return true
}

func (s *GossipStoreImpl) GetStoreKeyValue(key types.StoreKey) types.NodeValueMap {
	s.Lock()
	defer s.Unlock()

	nodeInfoMap := make(types.NodeValueMap)
	for id, nodeInfo := range s.nodeMap {
		if statusValid(nodeInfo.Status) && nodeInfo.Value != nil {
			if val, ok := nodeInfo.Value[key]; ok {
				n := types.NodeValue{Id: nodeInfo.Id,
					GenNumber:    nodeInfo.GenNumber,
					LastUpdateTs: nodeInfo.LastUpdateTs,
					Status:       nodeInfo.Status}
				n.Value = val
				nodeInfoMap[id] = n
			}
		}
	}

	return nodeInfoMap
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

func statusValid(s types.NodeStatus) bool {
	return (s != types.NODE_STATUS_INVALID &&
		s != types.NODE_STATUS_NEVER_GOSSIPED)
}

func (s *GossipStoreImpl) MetaInfo() types.StoreMetaInfo {
	s.Lock()
	defer s.Unlock()

	mInfo := make(types.StoreMetaInfo)

	for nodeId, nodeValue := range s.nodeMap {
		if statusValid(nodeValue.Status) {
			nodeMetaInfo := types.NodeMetaInfo{
				Id:           nodeId,
				LastUpdateTs: nodeValue.LastUpdateTs}
			mInfo[nodeId] = nodeMetaInfo
		}
	}

	return mInfo
}

func (s *GossipStoreImpl) Diff(
	d types.StoreMetaInfo) (types.StoreNodes, types.StoreNodes) {
	s.Lock()
	defer s.Unlock()

	diffNewNodes := make([]types.NodeId, 0)
	selfNewNodes := make([]types.NodeId, 0)

	for nodeId, nodeMetaInfo := range d {

		selfNodeInfo, ok := s.nodeMap[nodeId]
		if !ok {
			// we do not have info about this node
			diffNewNodes = append(diffNewNodes, nodeId)
			// nothing to add in selfNewNodes
			continue
		}

		// the diff has newer node if our status for node is invalid
		if !statusValid(selfNodeInfo.Status) ||
			selfNodeInfo.LastUpdateTs.Before(nodeMetaInfo.LastUpdateTs) {
			diffNewNodes = append(diffNewNodes, nodeId)
		} else if selfNodeInfo.LastUpdateTs.After(nodeMetaInfo.LastUpdateTs) {
			selfNewNodes = append(selfNewNodes, nodeId)
		}
	}

	// go over nodes present with us but not in the given meta info
	for nodeId, nodeInfo := range s.nodeMap {
		if _, ok := d[nodeId]; ok {
			// we have handled this case above
			continue
		}

		// peer does not have info about this node
		if statusValid(nodeInfo.Status) {
			selfNewNodes = append(selfNewNodes, nodeId)
		}
	}

	return diffNewNodes, selfNewNodes
}

func (s *GossipStoreImpl) Subset(nodes types.StoreNodes) types.NodeInfoMap {
	s.Lock()
	defer s.Unlock()

	subset := make(types.NodeInfoMap)

	for _, nodeId := range nodes {
		nodeInfo, ok := s.nodeMap[nodeId]
		if !ok || !statusValid(nodeInfo.Status) {
			continue
		}
		status := nodeInfo.Status
		if status == types.NODE_STATUS_WAITING_FOR_NEW_UPDATE ||
			status == types.NODE_STATUS_DOWN_WAITING_FOR_NEW_UPDATE {
			status = types.NODE_STATUS_DOWN
		}
		n := types.NodeInfo{Id: nodeInfo.Id,
			GenNumber:    nodeInfo.GenNumber,
			Value:        make(types.StoreMap),
			LastUpdateTs: nodeInfo.LastUpdateTs,
			Status:       status}
		for key, value := range nodeInfo.Value {
			n.Value[key] = value
		}
		subset[nodeId] = n
	}

	return subset
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
			if selfValue.Status == types.NODE_STATUS_WAITING_FOR_NEW_UPDATE {
				newNodeInfo.Status = selfValue.Status
			}
			s.nodeMap[id] = newNodeInfo
		}
	}
}

func (s *GossipStoreImpl) UpdateNodeStatuses(d time.Duration, sd time.Duration) {
	s.Lock()
	defer s.Unlock()

	for id, nodeInfo := range s.nodeMap {
		currTime := time.Now()
		timeDiff := currTime.Sub(nodeInfo.LastUpdateTs)
		if id == s.id {
			if (timeDiff > d/2) && s.selfCorrect {
				log.Warnf("No self update for long, updating self update ts "+
					"to not be marked down, time diff: %v, limit: %v, "+
					"Last TS: %v,Current TS: %v",
					timeDiff, d, nodeInfo.LastUpdateTs, currTime)
				nodeInfo.LastUpdateTs = currTime
				s.nodeMap[id] = nodeInfo
			}
			continue
		}
		waitGenTimeDiff := currTime.Sub(nodeInfo.WaitForGenUpdateTs)
		nodeStatus := nodeInfo.Status
		switch {
		case nodeInfo.Status == types.NODE_STATUS_WAITING_FOR_NEW_UPDATE,
			nodeInfo.Status == types.NODE_STATUS_DOWN_WAITING_FOR_NEW_UPDATE:
			if nodeInfo.LastUpdateTs.After(nodeInfo.WaitForGenUpdateTs) {
				// new update has happened since we marked us
				// waiting for new update
				if timeDiff >= d {
					// mark the node down
					nodeStatus = types.NODE_STATUS_DOWN_WAITING_FOR_NEW_UPDATE
				} else {
					// mark the node up
					nodeStatus = types.NODE_STATUS_UP
				}
			} else {
				// no new update has happened
				if waitGenTimeDiff >= d {
					nodeStatus = types.NODE_STATUS_DOWN_WAITING_FOR_NEW_UPDATE
				} // else maintain the current status
			}
		case nodeInfo.Status == types.NODE_STATUS_NEVER_GOSSIPED:
			if timeDiff >= sd {
				log.Warnf("Node ", id, " never gossiped, marking it down")
				nodeStatus = types.NODE_STATUS_DOWN
			} // else node is now marked up
		case nodeInfo.Status != types.NODE_STATUS_INVALID:
			if timeDiff >= d {
				nodeStatus = types.NODE_STATUS_DOWN
			} // else node is marked up
		}
		if nodeInfo.Status != nodeStatus {
			log.Warnf("Gossip Status change: for node: %v newStatus: %v, "+
				"time diff: %v, limit: %v, Last TS: %v,Current TS: %v", id,
				nodeStatus, timeDiff, d, nodeInfo.LastUpdateTs, currTime)
			nodeInfo.Status = nodeStatus
			s.nodeMap[id] = nodeInfo
		}
	}
}
