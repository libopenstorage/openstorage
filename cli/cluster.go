package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"

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
	c.manager.DisableGossipUpdates()
}

func (c *clusterClient) enableGossip(context *cli.Context) {
	c.clusterOptions(context)
	c.manager.EnableGossipUpdates()
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
