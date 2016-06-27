package state

import (
	"sync"

	"github.com/libopenstorage/gossip/types"
)

type down struct {
	nodeStatus  types.NodeStatus
	id          types.NodeId
	clusterSize int
	stateEvent  chan types.StateEvent
}

var instanceDown *down
var dOnce sync.Once

func GetDown(clusterSize int, selfId types.NodeId, stateEvent chan types.StateEvent) State {
	dOnce.Do(func() {
		instanceDown = &down{
			nodeStatus: types.NODE_STATUS_DOWN,
		}
	})
	instanceDown.clusterSize = clusterSize
	instanceDown.id = selfId
	instanceDown.stateEvent = stateEvent
	return instanceDown
}

func (d *down) String() string {
	return "NODE_STATUS_DOWN"
}

func (d *down) NodeStatus() types.NodeStatus {
	return d.nodeStatus
}

func (d *down) SelfAlive(localNodeInfoMap types.NodeInfoMap) (State, error) {
	notInQuorum := GetNotInQuorum(d.clusterSize, d.id, d.stateEvent)
	return notInQuorum, nil
}

func (d *down) NodeAlive(localNodeInfo types.NodeInfoMap) (State, error) {
	return d, nil
}

func (d *down) SelfLeave() (State, error) {
	return d, nil
}

func (d *down) NodeLeave(localNodeInfoMap types.NodeInfoMap) (State, error) {
	return d, nil
}

func (d *down) UpdateClusterSize(clusterSize int, localNodeInfoMap types.NodeInfoMap) (State, error) {
	return d, nil
}

func (d *down) Timeout() (State, error) {
	return d, nil
}
