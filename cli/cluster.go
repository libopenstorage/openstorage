package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"

	"github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/cluster"
)

type clusterClient struct {
	manager cluster.Cluster
	name    string
}

func (c *clusterClient) clusterOptions(context *cli.Context) {
	clnt, err := client.NewClient("http://localhost:9001", "v1")
	if err != nil {
		fmt.Printf("Failed to initialize client library: %v\n", err)
		os.Exit(1)
	}
	c.manager = clnt.ClusterManager()
}

func (c *clusterClient) inspect(context *cli.Context) {
	c.clusterOptions(context)
	jsonOut := context.GlobalBool("json")
	outFd := os.Stdout
	fn := "inspect"

	cluster, err := c.manager.Enumerate()
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	if jsonOut {
		fmtOutput(context, &Format{Cluster: &cluster})
	} else {
		fmt.Fprintf(outFd, "ID %s: Status: %v\n",
			cluster.Id, cluster.Status)

		w := new(tabwriter.Writer)
		w.Init(outFd, 12, 12, 1, ' ', 0)

		fmt.Fprintln(w, "ID\t IP\t STATUS\t CPU\t MEMORY\t CONTAINERS")
		for _, n := range cluster.Nodes {
			status := ""
			if n.Status == api.StatusInit {
				status = "Initializing"
			} else if n.Status == api.StatusOk {
				status = "OK"
			} else if n.Status == api.StatusOffline {
				status = "Off Line"
			} else {
				status = "Error"
			}

			fmt.Fprintln(w, n.Id, "\t", n.Ip, "\t", status, "\t",
				n.Cpu, "\t", n.Memory, "\t", len(n.Containers))
		}

		fmt.Fprintln(w)
		w.Flush()
	}
}

func (c *clusterClient) enumerate(context *cli.Context) {
	c.clusterOptions(context)
	jsonOut := context.GlobalBool("json")
	outFd := os.Stdout
	fn := "enumerate"

	cluster, err := c.manager.Enumerate()
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	if jsonOut {
		fmtOutput(context, &Format{Cluster: &cluster})
	} else {
		w := new(tabwriter.Writer)
		w.Init(outFd, 12, 12, 1, ' ', 0)

		fmt.Fprintln(w, "ID\t IMAGE\t STATUS\t NAMES\t NODE")
		for _, n := range cluster.Nodes {
			for _, c := range n.Containers {
				fmt.Fprintln(w, c.ID, "\t", c.Image, "\t", c.Status, "\t",
					c.Names, "\t", n.Ip)
			}
		}

		fmt.Fprintln(w)
		w.Flush()
	}
}

func (c *clusterClient) remove(context *cli.Context) {
}

func (c *clusterClient) shutdown(context *cli.Context) {
}

func (c *clusterClient) disableGossip(context *cli.Context) {
	c.clusterOptions(context)
	c.manager.DisableUpdates()
}

func (c *clusterClient) enableGossip(context *cli.Context) {
	c.clusterOptions(context)
	c.manager.EnableUpdates()
}

func (c *clusterClient) displayGossipStatus(context *cli.Context) {
	c.clusterOptions(context)
	jsonOut := context.GlobalBool("json")
	outFd := os.Stdout
	fn := "displayGossipStatus"

	s := c.manager.GetState()
	if s == nil {
		cmdError(context, fn, fmt.Errorf("Failed to get status"))
		return
	}

	if jsonOut {
		fmtOutput(context, &Format{Result: s})
	} else {
		w := new(tabwriter.Writer)
		w.Init(outFd, 12, 12, 1, ' ', 0)

		fmt.Fprintln(w, "ID\t LAST CONTACT TS\t DIR\t Errors")
		for _, n := range s.History {
			dirStr := "From Peer"
			if n.Dir == types.GD_ME_TO_PEER {
				dirStr = "To Peer"
			}
			fmt.Fprintln(w, n.Node, "\t", n.Ts, "\t", dirStr, "\t", n.Err)
		}

		fmt.Fprintln(w)
		w.Flush()

		fmt.Println("Individual Node Status")
		w = new(tabwriter.Writer)
		w.Init(outFd, 12, 12, 1, ' ', 0)

		fmt.Fprintln(w, "ID\t LAST UPDATE TS\t STATUS")
		for _, n := range s.NodeStatus {
			statusStr := "Up"
			switch {
			case n.Status == types.NODE_STATUS_DOWN,
				n.Status == types.NODE_STATUS_DOWN_WAITING_FOR_NEW_UPDATE:
				statusStr = "Down"
			case n.Status == types.NODE_STATUS_INVALID:
				statusStr = "Invalid"
			case n.Status == types.NODE_STATUS_NEVER_GOSSIPED:
				statusStr = "Node not yet gossiped"
			case n.Status == types.NODE_STATUS_WAITING_FOR_NEW_UPDATE:
				statusStr = "Waiting for new data with new generation"
			}
			fmt.Fprintln(w, n.Id, "\t", n.LastUpdateTs, "\t", statusStr)
		}

		fmt.Fprintln(w)
		w.Flush()
	}
}

// ClusterCommands exports CLI comamnds for File VolumeDriver
func ClusterCommands(name string) []cli.Command {
	c := &clusterClient{name: name}

	commands := []cli.Command{
		{
			Name:    "inspect",
			Aliases: []string{"i"},
			Usage:   "Inspect the cluster",
			Action:  c.inspect,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "machine,m",
					Usage: "Comma separated machine ids, e.g uuid1,uuid2",
					Value: "",
				},
			},
		},
		{
			Name:    "enumerate",
			Aliases: []string{"e"},
			Usage:   "Enumerate containers in the cluster",
			Action:  c.enumerate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "machine,m",
					Usage: "Comma separated machine ids, e.g uuid1,uuid2",
					Value: "",
				},
			},
		},
		{
			Name:    "disable-gossip",
			Aliases: []string{"dg"},
			Usage:   "Disable gossip updates",
			Action:  c.disableGossip,
		},
		{
			Name:    "enable-gossip",
			Aliases: []string{"eg"},
			Usage:   "Enable gossip updates",
			Action:  c.enableGossip,
		},
		{
			Name:    "gossip-status",
			Aliases: []string{"gs"},
			Usage:   "Display gossip status",
			Action:  c.displayGossipStatus,
		},
		{
			Name:    "remove",
			Aliases: []string{"r"},
			Usage:   "Remove a machine from the cluster",
			Action:  c.remove,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "machine,m",
					Usage: "Comma separated machine ids, e.g uuid1,uuid2",
					Value: "",
				},
			},
		},
		{
			Name:   "shutdown",
			Usage:  "Shutdown a cluster or a specific machine",
			Action: c.shutdown,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "machine,m",
					Usage: "Comma separated machine ids, e.g uuid1,uuid2",
					Value: "",
				},
			},
		},
	}
	return commands
}
