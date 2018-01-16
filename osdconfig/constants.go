package osdconfig

// kvdb keys
const (
	base_key    = "osdconfig"   // root of the tree in kvdb
	cluster_key = "clusterConf" // cluster data is managed behind this key
	nodes_key   = "nodesConf"   // all nodeID's exist behind this key and all node data behind respective node ID
)

// error constants
const (
	base_err     osdconfigError = "osdconfig:"
	INPUT_ERR    osdconfigError = base_err + "input is nil"
	REGISTER_ERR osdconfigError = base_err + "callback name already registered"
	DATA_ERR     osdconfigError = base_err + "no data fetched from kvdb"
	EXEC_ERR     osdconfigError = base_err + "callback exec error"
)
