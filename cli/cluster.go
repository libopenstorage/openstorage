package cli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"

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

func (c *clusterClient) enumerate(context *cli.Context) {
	c.clusterOptions(context)
	fn := "enumerate"

	cluster, err := c.manager.Enumerate()
	if err != nil {
		cmdError(context, fn, err)
		return
	}

	cmdOutput(context, cluster)
}

func (c *clusterClient) remove(context *cli.Context) {
}

func (c *clusterClient) shutdown(context *cli.Context) {
}

// ClusterCommands exports CLI comamnds for File VolumeDriver
func ClusterCommands(name string) []cli.Command {
	c := &clusterClient{name: name}

	commands := []cli.Command{
		{
			Name:    "inspect",
			Aliases: []string{"ci"},
			Usage:   "Inspect the cluster",
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
