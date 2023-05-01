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
	mu                   sync.Mutex
	stopCleanupCh        chan bool
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

func (rr *roundRobin) GetRemoteNodeConnection(ctx context.Context) (*grpc.ClientConn, bool, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	// Get all nodes and sort them
	cluster, err := rr.cluster.Enumerate()
	if err != nil {
		return nil, false, err
	}
	if len(cluster.Nodes) < 1 {
		return nil, false, errors.New("cluster nodes for remote connection not found")
	}
	sort.Slice(cluster.Nodes, func(i, j int) bool {
		return cluster.Nodes[i].Id < cluster.Nodes[j].Id
	})

	// Clean up connections for missing nodes
	rr.cleanupMissingNodeConnections(ctx, cluster.Nodes)

	// Get target node info and set next round robbin node.
	// nextNode is always lastNode + 1 mod (numOfNodes), to loop back to zero
	var (
		targetNodeNumber int
		isRemoteConn     bool
	)
	if rr.nextCreateNodeNumber != 0 {
		targetNodeNumber = rr.nextCreateNodeNumber
	}
	targetNode := cluster.Nodes[targetNodeNumber]
	if targetNode.Id != cluster.NodeId {
		// NodeID set on the cluster object is this node's ID.
		// Target NodeID does not match with our NodeID, so this will be a remote connection.
		isRemoteConn = true
	}
	targetNodeEndpoint := targetNode.MgmtIp
	rr.nextCreateNodeNumber = (targetNodeNumber + 1) % len(cluster.Nodes)

	// Get conn for this node, otherwise create new conn
	if len(rr.connMap) == 0 {
		rr.connMap = make(map[string]*TimedSDKConn)
	}
	if rr.connMap[targetNodeEndpoint] == nil {
		var err error
		rrlogger.WithContext(ctx).Infof("Round-robin connecting to node %v - %s:%s", targetNodeNumber, targetNodeEndpoint, rr.grpcServerPort)
		remoteConn, err := grpcserver.ConnectWithTimeout(
			fmt.Sprintf("%s:%s", targetNodeEndpoint, rr.grpcServerPort),
			[]grpc.DialOption{
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(correlation.ContextUnaryClientInterceptor),
			}, 10*time.Second)
		if err != nil {
			return nil, isRemoteConn, err
		}

		rr.connMap[targetNodeEndpoint] = &TimedSDKConn{
			Conn: remoteConn,
		}
	}

	// Keep track of when this conn was last accessed
	rrlogger.WithContext(ctx).Infof("Using remote connection to SDK node %v - %s:%s", targetNodeNumber, targetNodeEndpoint, rr.grpcServerPort)
	rr.connMap[targetNodeEndpoint].LastUsage = time.Now()
	return rr.connMap[targetNodeEndpoint].Conn, isRemoteConn, nil

}

func (rr *roundRobin) cleanupMissingNodeConnections(ctx context.Context, nodes []*api.Node) {
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
		}
	}
}

func (rr *roundRobin) cleanupConnections() {
	rr.stopCleanupCh = make(chan bool)
	ticker := time.NewTicker(connCleanupInterval)

	// Check every so often and delete/close connections
	for {
		select {
		case <-rr.stopCleanupCh:
			ticker.Stop()
			return

		case _ = <-ticker.C:
			ctx := correlation.WithCorrelationContext(context.Background(), correlation.ComponentRoundRobinBalancer)

			// Anonymous function for using defer to unlock mutex
			func() {
				rr.mu.Lock()
				defer rr.mu.Unlock()
				rrlogger.Tracef("Cleaning up open gRPC connections for CSI distributed provisioning")

				// Clean all expired connections
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

				// Get all nodes and cleanup conns for missing/deprovisioned nodes
				nodesResp, err := rr.cluster.Enumerate()
				if err != nil {
					rrlogger.Errorf("failed to get all nodes for connection cleanup: %v", err)
					return
				}
				if len(nodesResp.Nodes) < 1 {
					rrlogger.Errorf("no nodes available to cleanup: %v", err)
					return
				}
				rr.cleanupMissingNodeConnections(ctx, nodesResp.Nodes)

				if numConnsClosed > 0 {
					rrlogger.Infof("Cleaned up %v connections for CSI distributed provisioning. %v connections remaining", numConnsClosed, len(rr.connMap))
				}
			}()
		}
	}
}

func (rr *roundRobin) Stop() {
	if rr.stopCleanupCh != nil {
		close(rr.stopCleanupCh)
	}
}

func init() {
	rrlogger = correlation.NewPackageLogger(correlation.ComponentRoundRobinBalancer)
}
