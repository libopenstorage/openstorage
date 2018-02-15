package osdconfig

// kvdb keys
const (
	baseKey    = "osdconfig"   // root of the tree in kvdb
	clusterKey = "clusterConf" // cluster data is managed behind this key
	nodeKey    = "nodeConf"    // all nodeID's exist behind this key and all node data behind respective node ID
)

// these const indicates which type of kvdb changes callback is watching on
const (
	clusterWatcher Watcher = baseKey + "/" + clusterKey
	nodeWatcher    Watcher = baseKey + "/" + nodeKey
)
