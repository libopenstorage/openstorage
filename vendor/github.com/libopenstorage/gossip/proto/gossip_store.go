package proto

import (
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"

	"github.com/libopenstorage/gossip/types"
)

type GossipStoreImpl struct {
	sync.Mutex
	id    types.NodeId
	kvMap map[types.StoreKey]types.NodeInfoMap
}

func NewGossipStore(id types.NodeId) *GossipStoreImpl {
	n := &GossipStoreImpl{}
	n.InitStore(id)
	return n
}

func (s *GossipStoreImpl) NodeId() types.NodeId {
	return s.id
}

func (s *GossipStoreImpl) InitStore(id types.NodeId) {
	s.kvMap = make(map[types.StoreKey]types.NodeInfoMap)
	s.id = id
}

func (s *GossipStoreImpl) UpdateSelf(key types.StoreKey, val interface{}) {
	s.Lock()
	defer s.Unlock()

	nodeValue, ok := s.kvMap[key]
	if !ok {
		nodeValue = make(types.NodeInfoMap)
		s.kvMap[key] = nodeValue
	}

	nodeValue[s.id] = types.NodeInfo{Id: s.id,
		Value:        val,
		LastUpdateTs: time.Now(),
		Status:       types.NODE_STATUS_UP}
}

func (s *GossipStoreImpl) GetStoreKeyValue(key types.StoreKey) types.NodeInfoMap {
	s.Lock()
	defer s.Unlock()

	// we return an array, indexed by the node id.
	// Find the max node id.
	nodeInfoMap := make(types.NodeInfoMap)
	nodeInfos, ok := s.kvMap[key]
	if !ok || len(nodeInfos) == 0 {
		return nodeInfoMap
	}

	for id, nodeInfo := range nodeInfos {
		if nodeInfo.Status == types.NODE_STATUS_INVALID {
			continue
		}
		// this must create a copy
		nodeInfoMap[id] = nodeInfo
	}

	return nodeInfoMap
}

func (s *GossipStoreImpl) GetStoreKeys() []types.StoreKey {
	s.Lock()
	defer s.Unlock()

	storeKeys := make([]types.StoreKey, len(s.kvMap))
	i := 0
	for key, _ := range s.kvMap {
		storeKeys[i] = key
		i++
	}
	return storeKeys
}

func (s *GossipStoreImpl) MetaInfo() types.StoreMetaInfo {
	s.Lock()
	defer s.Unlock()

	mInfo := make(types.StoreMetaInfo, len(s.kvMap))

	for key, nodeValue := range s.kvMap {
		metaInfoList := make([]types.NodeMetaInfo, 0, len(nodeValue))

		for key, _ := range nodeValue {
			if nodeValue[key].Status != types.NODE_STATUS_INVALID {
				nodeMetaInfo := types.NodeMetaInfo{
					Id:           nodeValue[key].Id,
					LastUpdateTs: nodeValue[key].LastUpdateTs}
				metaInfoList = append(metaInfoList, nodeMetaInfo)
			}
		}

		if len(metaInfoList) > 0 {
			mInfo[key] = types.NodeMetaInfoList{List: metaInfoList}
		}
	}

	return mInfo
}

func (s *GossipStoreImpl) Diff(
	d types.StoreMetaInfo) (types.StoreNodes, types.StoreNodes) {
	s.Lock()
	defer s.Unlock()

	diffNewNodes := make(map[types.StoreKey][]types.NodeId)
	selfNewNodes := make(map[types.StoreKey][]types.NodeId)

	for key, metaInfoList := range d {
		selfNodeInfo, ok := s.kvMap[key]

		metaInfoLen := len(metaInfoList.List)
		if !ok {
			// we do not have info about this key
			newIds := make([]types.NodeId, metaInfoLen)
			for i := 0; i < metaInfoLen; i++ {
				newIds[i] = metaInfoList.List[i].Id
			}
			diffNewNodes[key] = newIds
			// nothing to add in selfNewNodes
			continue
		}

		diffNewIds := make([]types.NodeId, 0, metaInfoLen)
		selfNewIds := make([]types.NodeId, 0, metaInfoLen)
		for i := 0; i < metaInfoLen; i++ {
			metaId := metaInfoList.List[i].Id
			_, ok := selfNodeInfo[metaId]
			switch {
			case !ok:
				diffNewIds = append(diffNewIds, metaId)

			// avoid copying the whole node info
			// the diff has newer node if our status for node is invalid
			case selfNodeInfo[metaId].Status ==
				types.NODE_STATUS_INVALID:
				diffNewIds = append(diffNewIds, metaId)

			// or if its last update timestamp is newer than ours
			case selfNodeInfo[metaId].LastUpdateTs.Before(
				metaInfoList.List[i].LastUpdateTs):
				diffNewIds = append(diffNewIds, metaId)

			case selfNodeInfo[metaId].LastUpdateTs.After(
				metaInfoList.List[i].LastUpdateTs):
				selfNewIds = append(selfNewIds, metaId)
			}
		}

		if len(diffNewIds) > 0 {
			diffNewNodes[key] = diffNewIds
		}
		if len(selfNewIds) > 0 {
			selfNewNodes[key] = selfNewIds
		}
	}

	// go over keys present with us but not in the meta info
	for key, nodeInfoMap := range s.kvMap {
		_, ok := d[key]
		if ok {
			// we have handled this case above
			continue
		}

		// we do not have info about this key
		newIds := make([]types.NodeId, 0)
		for nodeId, _ := range nodeInfoMap {
			if nodeInfoMap[nodeId].Status != types.NODE_STATUS_INVALID {
				newIds = append(newIds, nodeId)
			}
		}
		selfNewNodes[key] = newIds
	}

	return diffNewNodes, selfNewNodes
}

func (s *GossipStoreImpl) Subset(nodes types.StoreNodes) types.StoreDiff {
	s.Lock()
	defer s.Unlock()

	subset := make(types.StoreDiff)

	for key, nodeIdList := range nodes {
		selfNodeInfos, ok := s.kvMap[key]
		if !ok {
			log.Info("No subset for key ", key)
			continue
		}

		// create a new map to hold the diff
		nodeInfoMap := make(types.NodeInfoMap)
		for _, id := range nodeIdList {
			_, ok := selfNodeInfos[id]
			if !ok {
				log.Info("Id missing from store, id: ", id, " for key: ", key)
				continue
			}
			nodeInfoMap[id] = selfNodeInfos[id]
		}
		// put it in the subset
		subset[key] = nodeInfoMap
	}

	return subset
}

func (s *GossipStoreImpl) Update(diff types.StoreDiff) {
	s.Lock()
	defer s.Unlock()

	for key, newValue := range diff {

		// XXX/gsangle: delete updates for self node, will this ever happen
		// given that we always have the most updated info ?
		delete(newValue, s.id)

		selfValue, ok := s.kvMap[key]
		if !ok {
			// create a copy
			nodeInfoMap := make(types.NodeInfoMap)
			for id, _ := range newValue {
				nodeInfoMap[id] = newValue[id]
			}
			s.kvMap[key] = nodeInfoMap
			continue
		}
		for id, info := range newValue {
			if selfValue[id].Status == types.NODE_STATUS_INVALID ||
				selfValue[id].LastUpdateTs.Before(info.LastUpdateTs) {
				selfValue[id] = info
			}
		}
	}
}

func (s *GossipStoreImpl) UpdateNodeStatuses(d time.Duration) {
	s.Lock()
	defer s.Unlock()

	for _, nodeValue := range s.kvMap {
		for id, _ := range nodeValue {
			currTime := time.Now()
			timeDiff := currTime.Sub(nodeValue[id].LastUpdateTs)
			if nodeValue[id].Status != types.NODE_STATUS_INVALID &&
				id != s.id && timeDiff >= d {
				log.Debugf("Marking node %s down since time diff %v is greater "+
					"than %v, its last update time was %v and current time is"+
					" %v", id, timeDiff, d, nodeValue[id].LastUpdateTs, currTime)
				nodeInfo := nodeValue[id]
				nodeInfo.Status = types.NODE_STATUS_DOWN
				nodeValue[id] = nodeInfo
			}
		}
	}
}
