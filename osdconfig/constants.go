package osdconfig

// kvdb keys
const (
	baseKey    = "osdconfig"   // root of the tree in kvdb
	clusterKey = "clusterConf" // cluster data is managed behind this key
	nodesKey   = "nodesConf"   // all nodeID's exist behind this key and all node data behind respective node ID
)

// error constants
const (
	baseErr  osdconfigError = "osdconfig:"
	InputErr                = baseErr + "input is nil"
	RegErr                  = baseErr + "callback name already registered"
	DataErr                 = baseErr + "no data fetched from kvdb"
	ExecErr                 = baseErr + "callback exec error"
)

// these const indicates which type of kvdb changes callback is watching on
const (
	ClusterWatcher Watcher = baseKey + "/" + clusterKey
	NodeWatcher    Watcher = baseKey + "/" + nodesKey
)
