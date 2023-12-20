package loadbalancer

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/pkg/correlation"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/pkg/sched"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// TimedSDKConn represents a gRPC connection and the last time it was used
type TimedSDKConn struct {
	Conn      *grpc.ClientConn
	LastUsage time.Time
}

type roundRobin struct {
	cluster              cluster.Cluster
	connMap              map[string]*TimedSDKConn
	nextCreateNodeNumber int
	mu                   sync.RWMutex
	grpcServerPort       string
}

var (
	rrlogger *logrus.Logger
)

const (
	connCleanupInterval = 15 * time.Minute
	connIdleConnLength  = 30 * time.Minute
)

// NewRoundRobinBalancer returns an implementation of the RoundRobin interface
// for getting a remote grpc client connection to one of the nodes in the cluster.
func NewRoundRobinBalancer(
	cluster cluster.Cluster,
	grpcServerPort string,
) (Balancer, error) {
	if cluster == nil {
		return nil, fmt.Errorf("cluster cannot be nil")
	}
	rr := &roundRobin{cluster: cluster, grpcServerPort: grpcServerPort}
	if sched.Instance() == nil {
		return nil, fmt.Errorf("sched instance is not initialized")
	}
	if _, err := sched.Instance().Schedule(
		func(interval sched.Interval) { rr.cleanupConnections() },
		sched.Periodic(connCleanupInterval),
		time.Now().Add(connCleanupInterval),
		false,
	); err != nil {
		return nil, fmt.Errorf("failed to schedule round robin cleanup routine: %v", err)
	}
	return rr, nil
}

func (rr *roundRobin) GetRemoteNode() (string, bool, error) {
	// Get all nodes and sort them
	cluster, err := rr.cluster.Enumerate()
	if err != nil {
		return "", false, err
	}
	if len(cluster.Nodes) < 1 {
		return "", false, errors.New("cluster nodes for remote connection not found")
	}
	// Get our node object
	selfNode, err := rr.cluster.Inspect(cluster.NodeId)
	if err != nil {
		return "", false, err
	}
	var filteredNodes []*api.Node

	if selfNode.DomainID != "" {
		// Filter out nodes from a different cluster domain.
		for _, node := range cluster.Nodes {
			if selfNode.DomainID == node.DomainID {
				filteredNodes = append(filteredNodes, node)
			}
		}
	} else {
		filteredNodes = cluster.Nodes
	}

	sort.Slice(filteredNodes, func(i, j int) bool {
		return filteredNodes[i].Id < filteredNodes[j].Id
	})

	// Get target node info and set next round robbin node.
	// nextNode is always lastNode + 1 mod (numOfNodes), to loop back to zero
	targetNodeEndpoint, isRemoteConn := rr.getTargetAndIncrement(filteredNodes, selfNode.Id)

	return targetNodeEndpoint, isRemoteConn, nil
}

func (rr *roundRobin) GetRemoteNodeConnection(ctx context.Context) (*grpc.ClientConn, bool, error) {
	targetNodeEndpoint, isRemoteConn, err := rr.GetRemoteNode()
	if err != nil {
		return nil, false, err
	}
	// Get conn for this node, otherwise create new conn
	timedSDKConn, ok := rr.getNodeConnection(targetNodeEndpoint)
	if !ok {
		var err error
		rrlogger.WithContext(ctx).Infof("Round-robin connecting to node %s:%s", targetNodeEndpoint, rr.grpcServerPort)
		remoteConn, err := grpcserver.ConnectWithTimeout(
			fmt.Sprintf("%s:%s", targetNodeEndpoint, rr.grpcServerPort),
			[]grpc.DialOption{
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(correlation.ContextUnaryClientInterceptor),
			}, 10*time.Second)
		if err != nil {
			return nil, isRemoteConn, err
		}
		timedSDKConn = &TimedSDKConn{
			Conn: remoteConn,
		}

		rr.setNodeConnection(targetNodeEndpoint, timedSDKConn)
	}

	// Keep track of when this conn was last accessed
	rrlogger.WithContext(ctx).Infof("Using remote connection to SDK node %s:%s", targetNodeEndpoint, rr.grpcServerPort)
	timedSDKConn.LastUsage = time.Now()
	return timedSDKConn.Conn, isRemoteConn, nil

}

func (rr *roundRobin) getTargetAndIncrement(filteredNodes []*api.Node, selfNodeID string) (string, bool) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	var (
		targetNodeNumber int
		isRemoteConn     bool
	)
	if rr.nextCreateNodeNumber != 0 {
		targetNodeNumber = rr.nextCreateNodeNumber
	}
	targetNode := filteredNodes[targetNodeNumber]
	if targetNode.Id != selfNodeID {
		// NodeID set on the cluster object is this node's ID.
		// Target NodeID does not match with our NodeID, so this will be a remote connection.
		isRemoteConn = true
	}
	targetNodeEndpoint := targetNode.MgmtIp
	rr.nextCreateNodeNumber = (targetNodeNumber + 1) % len(filteredNodes)

	return targetNodeEndpoint, isRemoteConn
}

func (rr *roundRobin) getNodeConnection(targetNodeEndpoint string) (*TimedSDKConn, bool) {
	if len(rr.connMap) == 0 {
		rr.mu.Lock()
		rr.connMap = make(map[string]*TimedSDKConn)
		rr.mu.Unlock()
	}

	rr.mu.RLock()
	timedSDKConn, ok := rr.connMap[targetNodeEndpoint]
	rr.mu.RUnlock()

	return timedSDKConn, ok
}

func (rr *roundRobin) setNodeConnection(targetNodeEndpoint string, tsc *TimedSDKConn) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(rr.connMap) == 0 {
		rr.connMap = make(map[string]*TimedSDKConn)
	}
	rr.connMap[targetNodeEndpoint] = tsc
}

func (rr *roundRobin) cleanupMissingNodeConnections(ctx context.Context, nodes []*api.Node) int {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	numConnsClosed := 0
	nodesMap := make(map[string]bool)
	for _, node := range nodes {
		nodesMap[node.MgmtIp] = true
	}
	for ip, timedConn := range rr.connMap {
		if ok := nodesMap[ip]; !ok {
			// If key in connmap is not in current nodes, close and remove it
			if err := timedConn.Conn.Close(); err != nil {
				rrlogger.WithContext(ctx).Errorf("failed to close conn to %s: %v", ip, err)
			}
			delete(rr.connMap, ip)
			numConnsClosed++
		}
	}

	return numConnsClosed
}

func (rr *roundRobin) cleanupExpiredConnections() int {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	numConnsClosed := 0

	for ip, timedConn := range rr.connMap {
		expiryTime := timedConn.LastUsage.Add(connIdleConnLength)

		// Connection has expired after 1hr of no usage.
		// Close connection and remove from connMap
		if expiryTime.Before(time.Now()) {
			rrlogger.Infof("SDK gRPC connection to %s is has expired after %v minutes of no usage. Closing this connection", ip, connIdleConnLength.Minutes())
			if err := timedConn.Conn.Close(); err != nil {
				rrlogger.Errorf("failed to close connection to %s: %v", ip, timedConn.Conn)
			}
			delete(rr.connMap, ip)
			numConnsClosed++
		}
	}

	return numConnsClosed
}

func (rr *roundRobin) cleanupConnections() {
	ctx := correlation.WithCorrelationContext(context.Background(), correlation.ComponentRoundRobinBalancer)
	rrlogger.Tracef("Cleaning up open gRPC connections created for round-robin balancing.")

	// Clean all expired connections
	expiredConnsClosed := rr.cleanupExpiredConnections()
	if expiredConnsClosed > 0 {
		rrlogger.Infof("Cleaned up %v expired node connections created for round-robin balancing. %v connections remaining", expiredConnsClosed, len(rr.connMap))
	}

	// Get all nodes and cleanup conns for missing/decommissioned nodes
	nodesResp, err := rr.cluster.Enumerate()
	if err != nil {
		rrlogger.Errorf("failed to get all nodes for connection cleanup: %v", err)
		return
	}
	if len(nodesResp.Nodes) < 1 {
		rrlogger.Errorf("no nodes available to cleanup: %v", err)
		return
	}
	missingNodeConnsClosed := rr.cleanupMissingNodeConnections(ctx, nodesResp.Nodes)
	if missingNodeConnsClosed > 0 {
		rrlogger.Infof("Cleaned up %v connections for missing nodes created for round-robin balancing. %v connections remaining", missingNodeConnsClosed, len(rr.connMap))
	}
}

func init() {
	rrlogger = correlation.NewPackageLogger(correlation.ComponentRoundRobinBalancer)
}
