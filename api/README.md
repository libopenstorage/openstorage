# OpenStorage API usage

Any storage product that uses the openstorage API can be managed via this API.  Below are some examples of using this API.

### Enumerate nodes in a cluster
```go

import (
    ...
    
    "github.com/libopenstorage/gossip/types"
    "github.com/libopenstorage/openstorage/api"
    "github.com/libopenstorage/openstorage/api/client/cluster"
)

type myapp struct {
    manager cluster.Cluster
}

func (c *myapp) init() {
    // Choose the default version.
    // Leave the host blank to use the local UNIX socket, or pass in an IP and a port at which the server is listening on.
    clnt, err := cluster.NewClusterClient("", cluster.APIVersion)
    if err != nil {
        fmt.Printf("Failed to initialize client library: %v\n", err)
        os.Exit(1)
    }
    c.manager = cluster.ClusterManager(clnt)
}

func (c *myapp) listNodes() {
    cluster, err := c.manager.Enumerate()
    if err != nil {
        cmdError(context, fn, err)
        return
    }
    
    // cluster is now a hashmap of nodes... do something useful with it:
    for _, n := range cluster.Nodes {
    
     }
}
```
