# osdconfig
osdconfig is a package allowing access to osd configuration parameters.

# connecting to backend
Access to configuration parameters is made available using one of the following
interfaces:
* io.ReadWriter
* kvdb connection
* grpc connection
* go package import

Assuming that the configuration parameters are stored in a backend that gets served
either using a kvdb, grpc or io.ReadWriter endpoint, an interface to that backend
can be used set/get configuration parameters at following levels:
* Global level allowing access to all data
* Cluster level (w/o any node level access)
* Node level using node-id

Below are some examples
 
## reading and writing against a file
Assuming that the configuration parameters are in a file, a read-writer can be
created on that file as follows:
```go
// read from file and create a new reader
bf, err := ioutil.ReadFile(fileName)
if err != nil {
	// do something
}
br := bufio.NewReader(bytes.NewReader(bf))

// create a new writer to bytes
var bb bytes.Buffer
bw := bufio.NewWriter(&bb)

// create a new read writer
brw := bufio.NewReadWriter(br, bw)
```

A client connection can then be obtained as follows:
```go
client := osdconfig.NewConnection(brw)
```

## reading and writing against kvdb connection
Assuming that the configuration parameters are in a kvdb database, a connection to
kvdb database can be obtained as follows:

```go
options := make(map[string]string)
options["KvUseInterface"] = ""
kv, err := kvdb.New("pwx/test", "", nil, options, nil)
if err != nil {
	// do something
}
```

A client connection can then be obtained as follows:
```go
client := osdconfig.NewConnection(kv)
```

## reading and writing against a grpc endpoint
Assuming that the configuration parameters are served via grpc server, a connection
to the server can be established as follows:
```go
//dial to grpc server
conn, err := grpc.Dial(GRPC_ADDR, grpc.WithInsecure())
if err != nil {
    t.Fatal(err)
}
```

A client connection can then be obtained as follows:
```go
client := osdconfig.NewConnection(conn)
```

# osdconfig interface
The protocol buffer definition serves as a "contract" and defines following
services on the configuration parameters:
```proto
service Spec {
    rpc GetGlobalSpec(Empty) returns (GlobalConfig) {}
    rpc SetGlobalSpec(GlobalConfig) returns (Ack) {}
    rpc GetClusterSpec(Empty) returns (ClusterConfig) {};
    rpc SetClusterSpec(ClusterConfig) returns (Ack) {};
    rpc GetNodeSpec(NodeID) returns (NodeConfig) {};
    rpc SetNodeSpec(NodeConfig) returns (Ack) {};
}
```

Correspondingly, the `conn` object created previously can be used to access
these methods as follows:

```go
ack, err = client.SetClusterSpec(ctx, globalConf.ClusterConf)
if err != nil {
	// do something
}
		
ack, err = client.GetNodeSpec(ctx, nodeID)
if err != nil {
	// do something 
}
		
// and so on
```

# package structure
* osdconfig/proto contains protocol buffer definitions and therefore contain struct defs
* osdconfig/api contains higher level abstraction and defines OsdConfigInterface
* osdconfig contains constructors to obtain connections