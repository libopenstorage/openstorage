package osdconfig

// kvdb keys
const (
	rootKey    = "osdconfig"   // root of the tree in kvdb
	clusterKey = "clusterConf" // cluster data is managed behind this key
	nodeKey    = "nodeConf"    // all nodeID's exist behind this key and all node data behind respective node ID
)

// error constants
const (
	baseErr       osdconfigError = "osdconfig:"
	ErrorInput                   = baseErr + "input is nil"
	ErrorRegister                = baseErr + "callback name already registered"
	ErrorData                    = baseErr + "error in data fetched from kvdb"
	ErrorExec                    = baseErr + "callback exec error"
)

// these const indicates which type of kvdb changes callback is watching on
const (
	TuneCluster Band = rootKey + "/" + clusterKey
	TuneNode    Band = rootKey + "/" + nodeKey
)

// logrus key values
const (
	pkg                      = "osdconfig"
	mgr                      = "manager"
	kv                       = "kv"
	sourceManagerWatch       = pkg + "/" + mgr + "/" + "watch"
	sourceManagerPrintStatus = pkg + "/" + mgr + "/" + "printStatus"
	sourceManagerAbort       = pkg + "/" + mgr + "/" + "abort"
	sourceManagerRun         = pkg + "/" + mgr + "/" + "run"
	sourceManagerClose       = pkg + "/" + mgr + "/" + "close"
	sourceKV                 = pkg + "/" + kv + "/" + "callback"
	sourceCallback           = pkg + "/" + mgr + "/" + "callback"
)

// logrus messages
const (
	msgCtxCancelled = "context cancelled"     // all moving parts will respond to context cancellation
	msgTrigArrived  = "trigger arrived"       // trigger generally refers to callback execution by kvdb
	msgExecSuccess  = "executed successfully" // execution succeeded
	msgExecError    = "executed with error"   // there were errors in execution but it may not necessarily stop service
	msgMemRelease   = "releasing memory"      // releasing memory during cleanup
	msgCleanup      = "cleanup"               // performing cleanup
	msgAborting     = "aborting"              // aborting by sending cancellation
	msgAbortSuccess = "abort successful"      // abort was successful
	msgSpawned      = "spawned successfully"  // callback function container was spawned
	msgDataError    = "data transfer error"   // when transfer of data to callback function container did not succeed
	msgDataSuccess  = "data transferred"      // when transfer of data to callback function container succeeded
	msgStatusError  = "status error"          // when callback function did not return with a status
	msgHeapStatus   = "heap status"           // job que is stored in a max-heap. Status returns length (there is limit of len)
	msgWatchStopped = "watch stopped"         // watch on kvdb has stopped
	msgTrigOnChan   = "triggered on channel"  // when execution of callback function container triggered on channel
	msgTrigOnPoll   = "triggered on poll"     // when execution of callback function container triggered on timed poll
)
