package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/libopenstorage/openstorage/osdconfig"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/etcd/v2"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var clusterManager osdconfig.ConfigManager

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "json, j",
			Usage: "Print JSON output",
		},
		cli.StringFlag{
			Name:  "host",
			Usage: "etcd http://host:port",
		},
		cli.StringFlag{
			Name:  "prefix",
			Usage: "base prefix for kvdb keys",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:        "config",
			Usage:       "Configure cluster",
			Description: "Configure cluster and nodes",
			Hidden:      false,
			Action:      setConfigValues,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "description",
					Usage:  "(Str)\tCluster description",
					Hidden: false,
				},
				cli.StringFlag{
					Name:   "mode",
					Usage:  "(Str)\tMode for cluster",
					Hidden: false,
				},
				cli.StringFlag{
					Name:   "version",
					Usage:  "(Str)\tVersion info for cluster",
					Hidden: false,
				},
				cli.StringFlag{
					Name:   "cluster_id",
					Usage:  "(Str)\tCluster ID info",
					Hidden: false,
				},
				cli.StringFlag{
					Name:   "domain",
					Usage:  "(Str)\tusage to be added",
					Hidden: false,
				},
			},
			Subcommands: []cli.Command{
				{
					Name:        "show",
					Usage:       "Show values",
					Description: "Show values",
					Action:      showConfigValues,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:   "all, a",
							Usage:  "(Bool)\tShow all data",
							Hidden: false,
						},
						cli.BoolFlag{
							Name:   "description",
							Usage:  "(Bool)\tCluster description",
							Hidden: false,
						},
						cli.BoolFlag{
							Name:   "mode",
							Usage:  "(Bool)\tMode for cluster",
							Hidden: false,
						},
						cli.BoolFlag{
							Name:   "version",
							Usage:  "(Bool)\tVersion info for cluster",
							Hidden: false,
						},
						cli.BoolFlag{
							Name:   "created",
							Usage:  "(Bool)\tCreation info for cluster",
							Hidden: false,
						},
						cli.BoolFlag{
							Name:   "cluster_id",
							Usage:  "(Bool)\tCluster ID info",
							Hidden: false,
						},
						cli.BoolFlag{
							Name:   "domain",
							Usage:  "(Bool)\tusage to be added",
							Hidden: false,
						},
					},
				},
				{
					Name:        "node",
					Usage:       "node usage",
					Description: "node description",
					Hidden:      false,
					Action:      setNodeValues,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "all, a",
							Usage: "(Bool)\tFor all nodes on cluster",
						},
						cli.StringFlag{
							Name:   "node_id",
							Usage:  "(Str)\tID for the node",
							Hidden: false,
						},
						cli.StringFlag{
							Name:   "csi_endpoint",
							Usage:  "(Str)\tCSI endpoint",
							Hidden: false,
						},
					},
					Subcommands: []cli.Command{
						{
							Name:        "show",
							Usage:       "Show values",
							Description: "Show values",
							Action:      showNodeValues,
							Flags: []cli.Flag{
								cli.BoolFlag{
									Name:   "all, a",
									Usage:  "(Bool)\tShow all data",
									Hidden: false,
								},
								cli.BoolFlag{
									Name:   "node_id",
									Usage:  "(Bool)\tID for the node",
									Hidden: false,
								},
								cli.BoolFlag{
									Name:   "csi_endpoint",
									Usage:  "(Bool)\tCSI endpoint",
									Hidden: false,
								},
							},
						},
						{
							Name:        "network",
							Usage:       "Network configuration",
							Description: "Configure network values for a node",
							Hidden:      false,
							Action:      setNetworkValues,
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:   "mgt_interface",
									Usage:  "(Str)\tManagement interface",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "data_interface",
									Usage:  "(Str)\tData interface",
									Hidden: false,
								},
							},
							Subcommands: []cli.Command{
								{
									Name:        "show",
									Usage:       "Show values",
									Description: "Show values",
									Action:      showNetworkValues,
									Flags: []cli.Flag{
										cli.BoolFlag{
											Name:   "all, a",
											Usage:  "(Bool)\tShow all data",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "mgt_interface",
											Usage:  "(Bool)\tManagement interface",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "data_interface",
											Usage:  "(Bool)\tData interface",
											Hidden: false,
										},
									},
								},
							},
						},
						{
							Name:        "storage",
							Usage:       "Storage configuration",
							Description: "Configure storage values for a node",
							Hidden:      false,
							Action:      setStorageValues,
							Flags: []cli.Flag{
								cli.StringSliceFlag{
									Name:   "devices_md",
									Usage:  "(Str...)\tDevices MD",
									Hidden: false,
								},
								cli.StringSliceFlag{
									Name:   "devices",
									Usage:  "(Str...)\tDevices list",
									Hidden: false,
								},
								cli.UintFlag{
									Name:   "max_count",
									Usage:  "(Uint)\tMaximum count",
									Hidden: false,
								},
								cli.UintFlag{
									Name:   "max_drive_set_count",
									Usage:  "(Uint)\tMax drive set count",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "raid_level",
									Usage:  "(Str)\tRAID level info",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "raid_level_md",
									Usage:  "(Str)\tRAID level MD",
									Hidden: false,
								},
							},
							Subcommands: []cli.Command{
								{
									Name:        "show",
									Usage:       "Show values",
									Description: "Show values",
									Action:      showStorageValues,
									Flags: []cli.Flag{
										cli.BoolFlag{
											Name:   "all, a",
											Usage:  "(Bool)\tShow all data",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "devices_md",
											Usage:  "(Bool)\tDevices MD",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "devices",
											Usage:  "(Bool)\tDevices list",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "max_count",
											Usage:  "(Bool)\tMaximum count",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "max_drive_set_count",
											Usage:  "(Bool)\tMax drive set count",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "raid_level",
											Usage:  "(Bool)\tRAID level info",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "raid_level_md",
											Usage:  "(Bool)\tRAID level MD",
											Hidden: false,
										},
									},
								},
							},
						},
						{
							Name:        "geo",
							Usage:       "Geographic configuration",
							Description: "Stores geo info for node",
							Hidden:      false,
							Action:      setGeoValues,
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:   "rack",
									Usage:  "(Str)\tRack info",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "zone",
									Usage:  "(Str)\tZone info",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "region",
									Usage:  "(Str)\tRegion info",
									Hidden: false,
								},
							},
							Subcommands: []cli.Command{
								{
									Name:        "show",
									Usage:       "Show values",
									Description: "Show values",
									Action:      showGeoValues,
									Flags: []cli.Flag{
										cli.BoolFlag{
											Name:   "all, a",
											Usage:  "(Bool)\tShow all data",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "rack",
											Usage:  "(Bool)\tRack info",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "zone",
											Usage:  "(Bool)\tZone info",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "region",
											Usage:  "(Bool)\tRegion info",
											Hidden: false,
										},
									},
								},
							},
						},
					},
				},
				{
					Name:        "secrets",
					Usage:       "usage to be added",
					Description: "description to be added",
					Hidden:      false,
					Action:      setSecretsValues,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:   "secret_type",
							Usage:  "(Str)\tSecret type",
							Hidden: false,
						},
						cli.StringFlag{
							Name:   "cluster_secret_key",
							Usage:  "(Str)\tSecret key",
							Hidden: false,
						},
					},
					Subcommands: []cli.Command{
						{
							Name:        "show",
							Usage:       "Show values",
							Description: "Show values",
							Action:      showSecretsValues,
							Flags: []cli.Flag{
								cli.BoolFlag{
									Name:   "all, a",
									Usage:  "(Bool)\tShow all data",
									Hidden: false,
								},
								cli.BoolFlag{
									Name:   "secret_type",
									Usage:  "(Bool)\tSecret type",
									Hidden: false,
								},
								cli.BoolFlag{
									Name:   "cluster_secret_key",
									Usage:  "(Bool)\tSecret key",
									Hidden: false,
								},
							},
						},
						{
							Name:        "vault",
							Usage:       "Vault configuration",
							Description: "none yet",
							Hidden:      false,
							Action:      setVaultValues,
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:   "token",
									Usage:  "(Str)\tVault token",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "address",
									Usage:  "(Str)\tVault address",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "ca_cert",
									Usage:  "(Str)\tVault CA certificate",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "ca_path",
									Usage:  "(Str)\tVault CA path",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "client_cert",
									Usage:  "(Str)\tVault client certificate",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "client_key",
									Usage:  "(Str)\tVault client key",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "skip_verify",
									Usage:  "(Str)\tVault skip verification",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "tls_server_name",
									Usage:  "(Str)\tVault TLS server name",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "base_path",
									Usage:  "(Str)\tVault base path",
									Hidden: false,
								},
							},
							Subcommands: []cli.Command{
								{
									Name:        "show",
									Usage:       "Show values",
									Description: "Show values",
									Action:      showVaultValues,
									Flags: []cli.Flag{
										cli.BoolFlag{
											Name:   "all, a",
											Usage:  "(Bool)\tShow all data",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "token",
											Usage:  "(Bool)\tVault token",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "address",
											Usage:  "(Bool)\tVault address",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "ca_cert",
											Usage:  "(Bool)\tVault CA certificate",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "ca_path",
											Usage:  "(Bool)\tVault CA path",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "client_cert",
											Usage:  "(Bool)\tVault client certificate",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "client_key",
											Usage:  "(Bool)\tVault client key",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "skip_verify",
											Usage:  "(Bool)\tVault skip verification",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "tls_server_name",
											Usage:  "(Bool)\tVault TLS server name",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "base_path",
											Usage:  "(Bool)\tVault base path",
											Hidden: false,
										},
									},
								},
							},
						},
						{
							Name:        "aws",
							Usage:       "AWS configuration",
							Description: "none yet",
							Hidden:      false,
							Action:      setAwsValues,
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:   "aws_access_key_id",
									Usage:  "(Str)\tAWS access key ID",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "aws_secret_access_key",
									Usage:  "(Str)\tAWS secret access key",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "aws_secret_token_key",
									Usage:  "(Str)\tAWS secret token key",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "aws_cmk",
									Usage:  "(Str)\tAWS CMK",
									Hidden: false,
								},
								cli.StringFlag{
									Name:   "aws_region",
									Usage:  "(Str)\tAWS region",
									Hidden: false,
								},
							},
							Subcommands: []cli.Command{
								{
									Name:        "show",
									Usage:       "Show values",
									Description: "Show values",
									Action:      showAwsValues,
									Flags: []cli.Flag{
										cli.BoolFlag{
											Name:   "all, a",
											Usage:  "(Bool)\tShow all data",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "aws_access_key_id",
											Usage:  "(Bool)\tAWS access key ID",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "aws_secret_access_key",
											Usage:  "(Bool)\tAWS secret access key",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "aws_secret_token_key",
											Usage:  "(Bool)\tAWS secret token key",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "aws_cmk",
											Usage:  "(Bool)\tAWS CMK",
											Hidden: false,
										},
										cli.BoolFlag{
											Name:   "aws_region",
											Usage:  "(Bool)\tAWS region",
											Hidden: false,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	app.Before = func(c *cli.Context) error {
		host := c.String("host")
		prefix := c.String("prefix")
		kv, err := kvdb.New(etcdv2.Name, prefix, []string{host}, nil, nil)
		if err != nil {
			return err
		}
		manager, err := osdconfig.NewManager(kv)
		if err != nil {
			return err
		}

		clusterManager = manager
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
func setConfigValues(c *cli.Context) error {
	config, err := clusterManager.GetClusterConf()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if config == nil {
		err := errors.New("config" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}

	if c.IsSet("description") {
		config.Description = c.String("description")
	}
	if c.IsSet("mode") {
		config.Mode = c.String("mode")
	}
	if c.IsSet("version") {
		config.Version = c.String("version")
	}
	if c.IsSet("cluster_id") {
		config.ClusterId = c.String("cluster_id")
	}
	if c.IsSet("domain") {
		config.Domain = c.String("domain")
	}
	if err := clusterManager.SetClusterConf(config); err != nil {
		logrus.Error("Set config for cluster")
		return err
	}
	logrus.Info("Set config for cluster")
	return nil
}

func showConfigValues(c *cli.Context) error {
	config, err := clusterManager.GetClusterConf()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if config == nil {
		err := errors.New("config" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}

	if c.GlobalBool("json") {
		return printJson(config)
	}
	if c.IsSet("all") || c.IsSet("description") {
		fmt.Println("description:", config.Description)
	}
	if c.IsSet("all") || c.IsSet("mode") {
		fmt.Println("mode:", config.Mode)
	}
	if c.IsSet("all") || c.IsSet("version") {
		fmt.Println("version:", config.Version)
	}
	if c.IsSet("all") || c.IsSet("created") {
		fmt.Println("created:", config.Created)
	}
	if c.IsSet("all") || c.IsSet("cluster_id") {
		fmt.Println("cluster_id:", config.ClusterId)
	}
	if c.IsSet("all") || c.IsSet("domain") {
		fmt.Println("domain:", config.Domain)
	}
	return nil
}

func setNodeValues(c *cli.Context) error {
	if !c.IsSet("node_id") && !c.IsSet("all") {
		err := errors.New("--node_id must be provided or --all must be set")
		logrus.Error(err)
		return err
	}
	configs := new(osdconfig.NodesConfig)
	var err error
	if c.IsSet("all") {
		configs, err = clusterManager.EnumerateNodeConf()
		if err != nil {
			logrus.Error(err)
			return err
		}
	} else {
		config, err := clusterManager.GetNodeConf(c.String("node_id"))
		if err != nil {
			logrus.Error(err)
			return err
		}
		*configs = append(*configs, config)
	}
	for _, config := range *configs {
		config := config
		if config == nil {
			err := errors.New("config" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}

		if c.IsSet("node_id") {
			config.NodeId = c.String("node_id")
		}
		if c.IsSet("csi_endpoint") {
			config.CSIEndpoint = c.String("csi_endpoint")
		}
		if err := clusterManager.SetNodeConf(config); err != nil {
			logrus.Error("Set config for node: ", config.NodeId)
			return err
		}
		logrus.Info("Set config for node: ", config.NodeId)
	}
	return nil
}

func showNodeValues(c *cli.Context) error {
	if !c.Parent().IsSet("node_id") && !c.Parent().IsSet("all") {
		err := errors.New("--node_id must be provided or --all must be set")
		logrus.Error(err)
		return err
	}
	configs := new(osdconfig.NodesConfig)
	var err error
	if c.Parent().IsSet("all") {
		configs, err = clusterManager.EnumerateNodeConf()
		if err != nil {
			logrus.Error(err)
			return err
		}
	} else {
		config, err := clusterManager.GetNodeConf(c.Parent().String("node_id"))
		if err != nil {
			logrus.Error(err)
			return err
		}
		*configs = append(*configs, config)
	}
	for _, config := range *configs {
		if config == nil {
			err := errors.New("config" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}

		if c.GlobalBool("json") {
			if err := printJson(struct {
				NodeId string      `json:"node_id"`
				Config interface{} `json:"config"`
			}{config.NodeId, config}); err != nil {
				return err
			}
		} else {
			fmt.Println("node_id:", config.NodeId)
			if c.IsSet("all") || c.IsSet("node_id") {
				fmt.Println("node_id:", config.NodeId)
			}
			if c.IsSet("all") || c.IsSet("csi_endpoint") {
				fmt.Println("csi_endpoint:", config.CSIEndpoint)
			}
			fmt.Println()
		}
	}
	return nil
}

func setNetworkValues(c *cli.Context) error {
	if !c.Parent().IsSet("node_id") && !c.Parent().IsSet("all") {
		err := errors.New("--node_id must be provided or --all must be set")
		logrus.Error(err)
		return err
	}
	configs := new(osdconfig.NodesConfig)
	var err error
	if c.Parent().IsSet("all") {
		configs, err = clusterManager.EnumerateNodeConf()
		if err != nil {
			logrus.Error(err)
			return err
		}
	} else {
		config, err := clusterManager.GetNodeConf(c.Parent().String("node_id"))
		if err != nil {
			logrus.Error(err)
			return err
		}
		*configs = append(*configs, config)
	}
	for _, config := range *configs {
		config := config
		if config == nil {
			err := errors.New("config" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}
		if config.Network == nil {
			err := errors.New("config.Network" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}

		if c.IsSet("mgt_interface") {
			config.Network.MgtIface = c.String("mgt_interface")
		}
		if c.IsSet("data_interface") {
			config.Network.DataIface = c.String("data_interface")
		}
		if err := clusterManager.SetNodeConf(config); err != nil {
			logrus.Error("Set config for node: ", config.NodeId)
			return err
		}
		logrus.Info("Set config for node: ", config.NodeId)
	}
	return nil
}

func showNetworkValues(c *cli.Context) error {
	if !c.Parent().Parent().IsSet("node_id") && !c.Parent().Parent().IsSet("all") {
		err := errors.New("--node_id must be provided or --all must be set")
		logrus.Error(err)
		return err
	}
	configs := new(osdconfig.NodesConfig)
	var err error
	if c.Parent().Parent().IsSet("all") {
		configs, err = clusterManager.EnumerateNodeConf()
		if err != nil {
			logrus.Error(err)
			return err
		}
	} else {
		config, err := clusterManager.GetNodeConf(c.Parent().Parent().String("node_id"))
		if err != nil {
			logrus.Error(err)
			return err
		}
		*configs = append(*configs, config)
	}
	for _, config := range *configs {
		if config == nil {
			err := errors.New("config" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}
		if config.Network == nil {
			err := errors.New("config.Network" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}

		if c.GlobalBool("json") {
			if err := printJson(struct {
				NodeId string      `json:"node_id"`
				Config interface{} `json:"config"`
			}{config.NodeId, config.Network}); err != nil {
				return err
			}
		} else {
			fmt.Println("node_id:", config.NodeId)
			if c.IsSet("all") || c.IsSet("mgt_interface") {
				fmt.Println("mgt_interface:", config.Network.MgtIface)
			}
			if c.IsSet("all") || c.IsSet("data_interface") {
				fmt.Println("data_interface:", config.Network.DataIface)
			}
			fmt.Println()
		}
	}
	return nil
}

func setStorageValues(c *cli.Context) error {
	if !c.Parent().IsSet("node_id") && !c.Parent().IsSet("all") {
		err := errors.New("--node_id must be provided or --all must be set")
		logrus.Error(err)
		return err
	}
	configs := new(osdconfig.NodesConfig)
	var err error
	if c.Parent().IsSet("all") {
		configs, err = clusterManager.EnumerateNodeConf()
		if err != nil {
			logrus.Error(err)
			return err
		}
	} else {
		config, err := clusterManager.GetNodeConf(c.Parent().String("node_id"))
		if err != nil {
			logrus.Error(err)
			return err
		}
		*configs = append(*configs, config)
	}
	for _, config := range *configs {
		config := config
		if config == nil {
			err := errors.New("config" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}
		if config.Storage == nil {
			err := errors.New("config.Storage" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}

		if c.IsSet("devices_md") {
			config.Storage.DevicesMd = c.StringSlice("devices_md")
		}
		if c.IsSet("devices") {
			config.Storage.Devices = c.StringSlice("devices")
		}
		if c.IsSet("max_count") {
			config.Storage.MaxCount = uint32(c.Uint("max_count"))
		}
		if c.IsSet("max_drive_set_count") {
			config.Storage.MaxDriveSetCount = uint32(c.Uint("max_drive_set_count"))
		}
		if c.IsSet("raid_level") {
			config.Storage.RaidLevel = c.String("raid_level")
		}
		if c.IsSet("raid_level_md") {
			config.Storage.RaidLevelMd = c.String("raid_level_md")
		}
		if err := clusterManager.SetNodeConf(config); err != nil {
			logrus.Error("Set config for node: ", config.NodeId)
			return err
		}
		logrus.Info("Set config for node: ", config.NodeId)
	}
	return nil
}

func showStorageValues(c *cli.Context) error {
	if !c.Parent().Parent().IsSet("node_id") && !c.Parent().Parent().IsSet("all") {
		err := errors.New("--node_id must be provided or --all must be set")
		logrus.Error(err)
		return err
	}
	configs := new(osdconfig.NodesConfig)
	var err error
	if c.Parent().Parent().IsSet("all") {
		configs, err = clusterManager.EnumerateNodeConf()
		if err != nil {
			logrus.Error(err)
			return err
		}
	} else {
		config, err := clusterManager.GetNodeConf(c.Parent().Parent().String("node_id"))
		if err != nil {
			logrus.Error(err)
			return err
		}
		*configs = append(*configs, config)
	}
	for _, config := range *configs {
		if config == nil {
			err := errors.New("config" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}
		if config.Storage == nil {
			err := errors.New("config.Storage" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}

		if c.GlobalBool("json") {
			if err := printJson(struct {
				NodeId string      `json:"node_id"`
				Config interface{} `json:"config"`
			}{config.NodeId, config.Storage}); err != nil {
				return err
			}
		} else {
			fmt.Println("node_id:", config.NodeId)
			if c.IsSet("all") || c.IsSet("devices_md") {
				fmt.Println("devices_md:", config.Storage.DevicesMd)
			}
			if c.IsSet("all") || c.IsSet("devices") {
				fmt.Println("devices:", config.Storage.Devices)
			}
			if c.IsSet("all") || c.IsSet("max_count") {
				fmt.Println("max_count:", config.Storage.MaxCount)
			}
			if c.IsSet("all") || c.IsSet("max_drive_set_count") {
				fmt.Println("max_drive_set_count:", config.Storage.MaxDriveSetCount)
			}
			if c.IsSet("all") || c.IsSet("raid_level") {
				fmt.Println("raid_level:", config.Storage.RaidLevel)
			}
			if c.IsSet("all") || c.IsSet("raid_level_md") {
				fmt.Println("raid_level_md:", config.Storage.RaidLevelMd)
			}
			fmt.Println()
		}
	}
	return nil
}

func setGeoValues(c *cli.Context) error {
	if !c.Parent().IsSet("node_id") && !c.Parent().IsSet("all") {
		err := errors.New("--node_id must be provided or --all must be set")
		logrus.Error(err)
		return err
	}
	configs := new(osdconfig.NodesConfig)
	var err error
	if c.Parent().IsSet("all") {
		configs, err = clusterManager.EnumerateNodeConf()
		if err != nil {
			logrus.Error(err)
			return err
		}
	} else {
		config, err := clusterManager.GetNodeConf(c.Parent().String("node_id"))
		if err != nil {
			logrus.Error(err)
			return err
		}
		*configs = append(*configs, config)
	}
	for _, config := range *configs {
		config := config
		if config == nil {
			err := errors.New("config" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}
		if config.Geo == nil {
			err := errors.New("config.Geo" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}

		if c.IsSet("rack") {
			config.Geo.Rack = c.String("rack")
		}
		if c.IsSet("zone") {
			config.Geo.Zone = c.String("zone")
		}
		if c.IsSet("region") {
			config.Geo.Region = c.String("region")
		}
		if err := clusterManager.SetNodeConf(config); err != nil {
			logrus.Error("Set config for node: ", config.NodeId)
			return err
		}
		logrus.Info("Set config for node: ", config.NodeId)
	}
	return nil
}

func showGeoValues(c *cli.Context) error {
	if !c.Parent().Parent().IsSet("node_id") && !c.Parent().Parent().IsSet("all") {
		err := errors.New("--node_id must be provided or --all must be set")
		logrus.Error(err)
		return err
	}
	configs := new(osdconfig.NodesConfig)
	var err error
	if c.Parent().Parent().IsSet("all") {
		configs, err = clusterManager.EnumerateNodeConf()
		if err != nil {
			logrus.Error(err)
			return err
		}
	} else {
		config, err := clusterManager.GetNodeConf(c.Parent().Parent().String("node_id"))
		if err != nil {
			logrus.Error(err)
			return err
		}
		*configs = append(*configs, config)
	}
	for _, config := range *configs {
		if config == nil {
			err := errors.New("config" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}
		if config.Geo == nil {
			err := errors.New("config.Geo" + ": no data found, received nil pointer")
			logrus.Error(err)
			return err
		}

		if c.GlobalBool("json") {
			if err := printJson(struct {
				NodeId string      `json:"node_id"`
				Config interface{} `json:"config"`
			}{config.NodeId, config.Geo}); err != nil {
				return err
			}
		} else {
			fmt.Println("node_id:", config.NodeId)
			if c.IsSet("all") || c.IsSet("rack") {
				fmt.Println("rack:", config.Geo.Rack)
			}
			if c.IsSet("all") || c.IsSet("zone") {
				fmt.Println("zone:", config.Geo.Zone)
			}
			if c.IsSet("all") || c.IsSet("region") {
				fmt.Println("region:", config.Geo.Region)
			}
			fmt.Println()
		}
	}
	return nil
}

func setSecretsValues(c *cli.Context) error {
	config, err := clusterManager.GetClusterConf()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if config == nil {
		err := errors.New("config" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets == nil {
		err := errors.New("config.Secrets" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}

	if c.IsSet("secret_type") {
		config.Secrets.SecretType = c.String("secret_type")
	}
	if c.IsSet("cluster_secret_key") {
		config.Secrets.ClusterSecretKey = c.String("cluster_secret_key")
	}
	if err := clusterManager.SetClusterConf(config); err != nil {
		logrus.Error("Set config for cluster")
		return err
	}
	logrus.Info("Set config for cluster")
	return nil
}

func showSecretsValues(c *cli.Context) error {
	config, err := clusterManager.GetClusterConf()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if config == nil {
		err := errors.New("config" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets == nil {
		err := errors.New("config.Secrets" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}

	if c.GlobalBool("json") {
		return printJson(config.Secrets)
	}
	if c.IsSet("all") || c.IsSet("secret_type") {
		fmt.Println("secret_type:", config.Secrets.SecretType)
	}
	if c.IsSet("all") || c.IsSet("cluster_secret_key") {
		fmt.Println("cluster_secret_key:", config.Secrets.ClusterSecretKey)
	}
	return nil
}

func setVaultValues(c *cli.Context) error {
	config, err := clusterManager.GetClusterConf()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if config == nil {
		err := errors.New("config" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets == nil {
		err := errors.New("config.Secrets" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets.Vault == nil {
		err := errors.New("config.Secrets.Vault" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}

	if c.IsSet("token") {
		config.Secrets.Vault.Token = c.String("token")
	}
	if c.IsSet("address") {
		config.Secrets.Vault.Address = c.String("address")
	}
	if c.IsSet("ca_cert") {
		config.Secrets.Vault.CACert = c.String("ca_cert")
	}
	if c.IsSet("ca_path") {
		config.Secrets.Vault.CAPath = c.String("ca_path")
	}
	if c.IsSet("client_cert") {
		config.Secrets.Vault.ClientCert = c.String("client_cert")
	}
	if c.IsSet("client_key") {
		config.Secrets.Vault.ClientKey = c.String("client_key")
	}
	if c.IsSet("skip_verify") {
		config.Secrets.Vault.TLSSkipVerify = c.String("skip_verify")
	}
	if c.IsSet("tls_server_name") {
		config.Secrets.Vault.TLSServerName = c.String("tls_server_name")
	}
	if c.IsSet("base_path") {
		config.Secrets.Vault.BasePath = c.String("base_path")
	}
	if err := clusterManager.SetClusterConf(config); err != nil {
		logrus.Error("Set config for cluster")
		return err
	}
	logrus.Info("Set config for cluster")
	return nil
}

func showVaultValues(c *cli.Context) error {
	config, err := clusterManager.GetClusterConf()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if config == nil {
		err := errors.New("config" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets == nil {
		err := errors.New("config.Secrets" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets.Vault == nil {
		err := errors.New("config.Secrets.Vault" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}

	if c.GlobalBool("json") {
		return printJson(config.Secrets.Vault)
	}
	if c.IsSet("all") || c.IsSet("token") {
		fmt.Println("token:", config.Secrets.Vault.Token)
	}
	if c.IsSet("all") || c.IsSet("address") {
		fmt.Println("address:", config.Secrets.Vault.Address)
	}
	if c.IsSet("all") || c.IsSet("ca_cert") {
		fmt.Println("ca_cert:", config.Secrets.Vault.CACert)
	}
	if c.IsSet("all") || c.IsSet("ca_path") {
		fmt.Println("ca_path:", config.Secrets.Vault.CAPath)
	}
	if c.IsSet("all") || c.IsSet("client_cert") {
		fmt.Println("client_cert:", config.Secrets.Vault.ClientCert)
	}
	if c.IsSet("all") || c.IsSet("client_key") {
		fmt.Println("client_key:", config.Secrets.Vault.ClientKey)
	}
	if c.IsSet("all") || c.IsSet("skip_verify") {
		fmt.Println("skip_verify:", config.Secrets.Vault.TLSSkipVerify)
	}
	if c.IsSet("all") || c.IsSet("tls_server_name") {
		fmt.Println("tls_server_name:", config.Secrets.Vault.TLSServerName)
	}
	if c.IsSet("all") || c.IsSet("base_path") {
		fmt.Println("base_path:", config.Secrets.Vault.BasePath)
	}
	return nil
}

func setAwsValues(c *cli.Context) error {
	config, err := clusterManager.GetClusterConf()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if config == nil {
		err := errors.New("config" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets == nil {
		err := errors.New("config.Secrets" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets.Aws == nil {
		err := errors.New("config.Secrets.Aws" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}

	if c.IsSet("aws_access_key_id") {
		config.Secrets.Aws.AccessKeyId = c.String("aws_access_key_id")
	}
	if c.IsSet("aws_secret_access_key") {
		config.Secrets.Aws.SecretAccessKey = c.String("aws_secret_access_key")
	}
	if c.IsSet("aws_secret_token_key") {
		config.Secrets.Aws.SecretTokenKey = c.String("aws_secret_token_key")
	}
	if c.IsSet("aws_cmk") {
		config.Secrets.Aws.Cmk = c.String("aws_cmk")
	}
	if c.IsSet("aws_region") {
		config.Secrets.Aws.Region = c.String("aws_region")
	}
	if err := clusterManager.SetClusterConf(config); err != nil {
		logrus.Error("Set config for cluster")
		return err
	}
	logrus.Info("Set config for cluster")
	return nil
}

func showAwsValues(c *cli.Context) error {
	config, err := clusterManager.GetClusterConf()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if config == nil {
		err := errors.New("config" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets == nil {
		err := errors.New("config.Secrets" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}
	if config.Secrets.Aws == nil {
		err := errors.New("config.Secrets.Aws" + ": no data found, received nil pointer")
		logrus.Error(err)
		return err
	}

	if c.GlobalBool("json") {
		return printJson(config.Secrets.Aws)
	}
	if c.IsSet("all") || c.IsSet("aws_access_key_id") {
		fmt.Println("aws_access_key_id:", config.Secrets.Aws.AccessKeyId)
	}
	if c.IsSet("all") || c.IsSet("aws_secret_access_key") {
		fmt.Println("aws_secret_access_key:", config.Secrets.Aws.SecretAccessKey)
	}
	if c.IsSet("all") || c.IsSet("aws_secret_token_key") {
		fmt.Println("aws_secret_token_key:", config.Secrets.Aws.SecretTokenKey)
	}
	if c.IsSet("all") || c.IsSet("aws_cmk") {
		fmt.Println("aws_cmk:", config.Secrets.Aws.Cmk)
	}
	if c.IsSet("all") || c.IsSet("aws_region") {
		fmt.Println("aws_region:", config.Secrets.Aws.Region)
	}
	return nil
}

func printJson(obj interface{}) error {
	if b, err := json.MarshalIndent(obj, "", "  "); err != nil {
		return err
	} else {
		fmt.Println(string(b))
		return nil
	}
}
