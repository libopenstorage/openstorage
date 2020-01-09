//go:generate swagger generate spec -m -o ../../api/swagger/swagger.json

// Package classification OSD API.
//
// OpenStorage is a clustered implementation of the Open Storage specification and relies on the OCI runtime.
// It allows you to run stateful services in containers in a multi-host clustered environment.
// This document represents the API documentaton of Openstorage, for the GO client please visit:
// https://github.com/libopenstorage/openstorage
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /v1
//     Version: 2.0.0
//     License: APACHE2 https://opensource.org/licenses/Apache-2.0
//     Contact: https://github.com/libopenstorage/openstorage
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/reexec"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/flexvolume"
	"github.com/libopenstorage/openstorage/api/server"
	"github.com/libopenstorage/openstorage/api/server/sdk"
	osdcli "github.com/libopenstorage/openstorage/cli"
	"github.com/libopenstorage/openstorage/cluster"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/csi"
	graphdrivers "github.com/libopenstorage/openstorage/graph/drivers"
	"github.com/libopenstorage/openstorage/objectstore"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/auth/systemtoken"
	"github.com/libopenstorage/openstorage/pkg/role"
	policy "github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/libopenstorage/openstorage/schedpolicy"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/consul"
	etcd "github.com/portworx/kvdb/etcd/v2"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"
)

var (
	datastores = []string{mem.Name, etcd.Name, consul.Name}
)

func main() {
	if reexec.Init() {
		return
	}
	app := cli.NewApp()
	app.Name = "osd"
	app.Usage = "Open Storage CLI"
	app.Version = config.Version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "json,j",
			Usage: "output in json",
		},
		cli.BoolFlag{
			Name:  osdcli.DaemonAlias,
			Usage: "Start OSD in daemon mode",
		},
		cli.StringSliceFlag{
			Name:  "driver",
			Usage: "driver name and options: name=btrfs,home=/var/openstorage/btrfs",
			Value: new(cli.StringSlice),
		},
		cli.StringFlag{
			Name:  "kvdb,k",
			Usage: "uri to kvdb e.g. kv-mem://localhost, etcd-kv://localhost:4001, consul-kv://localhost:8500",
			Value: "kv-mem://localhost",
		},
		cli.StringFlag{
			Name:  "file,f",
			Usage: "file to read the OSD configuration from.",
			Value: "",
		},
		cli.StringFlag{
			Name:  "mgmtport,m",
			Usage: "Management Port for REST server. Example: 9001",
			Value: "9001",
		},
		cli.StringFlag{
			Name:  "sdkport",
			Usage: "gRPC port for SDK. Example: 9100",
			Value: "9100",
		},
		cli.StringFlag{
			Name:  "sdkrestport",
			Usage: "gRPC REST Gateway port for SDK. Example: 9110",
			Value: "9110",
		},
		cli.StringFlag{
			Name:  "nodeid",
			Usage: "Name of this node",
			Value: "1",
		},
		cli.StringFlag{
			Name:  "clusterid",
			Usage: "Cluster id",
			Value: "openstorage.cluster",
		},
		cli.StringFlag{
			Name:  "tls-cert-file",
			Usage: "TLS Cert file path",
		},
		cli.StringFlag{
			Name:  "tls-key-file",
			Usage: "TLS Key file path",
		},
		cli.StringFlag{
			Name: "username-claim",
			Usage: "Claim key from the token to use as the unique id of the ." +
				"user. Values can be only 'sub', 'email', or 'name'",
			Value: "sub",
		},
		cli.StringFlag{
			Name:  "oidc-issuer",
			Usage: "OIDC Issuer,e.g. https://accounts.google.com",
		},
		cli.StringFlag{
			Name:  "oidc-client-id",
			Usage: "OIDC Client ID provided by issuer",
		},
		cli.StringFlag{
			Name:  "oidc-custom-claim-namespace",
			Usage: "OIDC namespace for custom claims if needed",
		},
		cli.BoolFlag{
			Name:  "oidc-skip-client-id-check",
			Usage: "OIDC skip verification of client id in the token",
		},
		cli.StringFlag{
			Name:  "jwt-issuer",
			Usage: "JSON Web Token issuer",
			Value: "openstorage.io",
		},
		cli.StringFlag{
			Name:  "jwt-shared-secret",
			Usage: "JSON Web Token shared secret",
		},
		cli.StringFlag{
			Name:  "jwt-rsa-pubkey-file",
			Usage: "JSON Web Token RSA Public file path",
		},
		cli.StringFlag{
			Name:  "jwt-ecds-pubkey-file",
			Usage: "JSON Web Token ECDS Public file path",
		},
		cli.StringFlag{
			Name:  "jwt-system-shared-secret",
			Usage: "JSON Web Token system shared secret used by clusters to create tokens for internal cluster communication",
			Value: "non-secure-secret",
		},
		cli.StringFlag{
			Name:  "clusterdomain",
			Usage: "Cluster Domain Name",
			Value: "",
		},
	}
	app.Action = wrapAction(start)
	app.Commands = []cli.Command{
		{
			Name:        "driver",
			Aliases:     []string{"d"},
			Usage:       "Manage drivers",
			Subcommands: osdcli.DriverCommands(),
		},
		{
			Name:        "cluster",
			Aliases:     []string{"c"},
			Usage:       "Manage cluster",
			Subcommands: osdcli.ClusterCommands(),
		},
		{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "Display version",
			Action:  wrapAction(showVersion),
		},
	}

	// Register all volume drivers with the CLI.
	for _, v := range volumedrivers.AllDrivers {
		// TODO(pedge): was an and, but we have no drivers that have two types
		switch v.DriverType {
		case api.DriverType_DRIVER_TYPE_BLOCK:
			bCmds := osdcli.BlockVolumeCommands(v.Name)
			cmds := append(bCmds)
			c := cli.Command{
				Name:        v.Name,
				Usage:       fmt.Sprintf("Manage %s storage", v.Name),
				Subcommands: cmds,
			}
			app.Commands = append(app.Commands, c)
		case api.DriverType_DRIVER_TYPE_FILE:
			fCmds := osdcli.FileVolumeCommands(v.Name)
			cmds := append(fCmds)
			c := cli.Command{
				Name:        v.Name,
				Usage:       fmt.Sprintf("Manage %s volumes", v.Name),
				Subcommands: cmds,
			}
			app.Commands = append(app.Commands, c)
		}
	}

	// Register all graph drivers with the CLI.
	for _, v := range graphdrivers.AllDrivers {
		// TODO(pedge): was an and, but we have no drivers that have two types
		switch v.DriverType {
		case api.DriverType_DRIVER_TYPE_GRAPH:
			cmds := osdcli.GraphDriverCommands(v.Name)
			c := cli.Command{
				Name:        v.Name,
				Usage:       fmt.Sprintf("Manage %s graph storage", v.Name),
				Subcommands: cmds,
			}
			app.Commands = append(app.Commands, c)
		}
	}

	app.Run(os.Args)
}

func start(c *cli.Context) error {
	if !osdcli.DaemonMode(c) {
		cli.ShowAppHelp(c)
		return nil
	}

	var (
		cfg *config.Config
	)

	// We are in daemon mode.
	file := c.String("file")
	if len(file) != 0 {
		// Read from file
		var err error
		cfg, err = config.Parse(file)
		if err != nil {
			return err
		}
	} else {
		cfg = &config.Config{}
	}

	// Check if values are set
	if len(cfg.Osd.ClusterConfig.ClusterId) == 0 {
		cfg.Osd.ClusterConfig.ClusterId = c.String("clusterid")
	}
	if len(cfg.Osd.ClusterConfig.NodeId) == 0 {
		cfg.Osd.ClusterConfig.NodeId = c.String("nodeid")
	}

	// Get driver information
	driverInfoList := c.StringSlice("driver")
	if len(driverInfoList) != 0 {
		if cfg.Osd.Drivers == nil {
			cfg.Osd.Drivers = make(map[string]map[string]string)
		}
		params := make(map[string]string)
		var name string

		// many driver infos provided as a []string
		for _, driverInfo := range driverInfoList {

			// driverInfo of the format name=xxx,opt1=val1,opt2=val2
			for _, pair := range strings.Split(driverInfo, ",") {
				kv := strings.Split(pair, "=")
				if len(kv) != 2 {
					return fmt.Errorf("driver option has a an invalid pair %s", kv)
				}
				k := kv[0]
				v := kv[1]
				if len(k) == 0 || len(v) == 0 {
					return fmt.Errorf("driver option '%s' is invalid", pair)
				}
				if k == "name" {
					// Driver name
					name = v
				} else {
					// Options for driver
					params[k] = v
				}
			}
			if len(name) == 0 {
				return fmt.Errorf("driver option is missing driver name")
			}
			cfg.Osd.Drivers[name] = params
		}
	}
	if len(cfg.Osd.Drivers) == 0 {
		return fmt.Errorf("Must supply driver information")
	}

	kvdbURL := c.String("kvdb")
	u, err := url.Parse(kvdbURL)
	scheme := u.Scheme
	u.Scheme = "http"

	kv, err := kvdb.New(scheme, "openstorage", []string{u.String()}, nil, kvdb.LogFatalErrorCB)
	if err != nil {
		return fmt.Errorf("Failed to initialize KVDB: %v (%v)\nSupported datastores: %v", scheme, err, datastores)
	}
	if err := kvdb.SetInstance(kv); err != nil {
		return fmt.Errorf("Failed to initialize KVDB: %v", err)
	}

	// Start the cluster state machine, if enabled.
	clusterInit := false
	if cfg.Osd.ClusterConfig.NodeId != "" && cfg.Osd.ClusterConfig.ClusterId != "" {
		logrus.Infof("OSD enabling cluster mode.")
		if err := clustermanager.Init(cfg.Osd.ClusterConfig); err != nil {
			return fmt.Errorf("Unable to init cluster server: %v", err)
		}
		if err := server.StartClusterAPI(cluster.APIBase, 0); err != nil {
			return fmt.Errorf("Unable to start cluster API server: %v", err)
		}
		clusterInit = true
	}

	isDefaultSet := false
	// Start the volume drivers.
	for d, v := range cfg.Osd.Drivers {
		logrus.Infof("Starting volume driver: %v", d)
		if err := volumedrivers.Register(d, v); err != nil {
			return fmt.Errorf("Unable to start volume driver: %v, %v", d, err)
		}

		var mgmtPort, pluginPort uint64
		if port, ok := v[config.MgmtPortKey]; ok {
			mgmtPort, err = strconv.ParseUint(port, 10, 16)
			if err != nil {
				return fmt.Errorf("Invalid OSD Config File. Invalid Mgmt Port number for Driver : %s", d)
			}
		} else if c.String("mgmtport") != "" {
			mgmtPort, err = strconv.ParseUint(c.String("mgmtport"), 10, 16)
			if err != nil {
				return fmt.Errorf("Invalid Mgmt Port number for Driver : %s", d)
			}
		} else {
			mgmtPort = 0
		}

		if port, ok := v[config.PluginPortKey]; ok {
			pluginPort, err = strconv.ParseUint(port, 10, 16)
			if err != nil {
				return fmt.Errorf("Invalid OSD Config File. Invalid Plugin Port number for Driver : %s", d)
			}
		} else {
			pluginPort = 0
		}

		sdksocket := fmt.Sprintf("/var/lib/osd/driver/%s-sdk.sock", d)

		if err := server.StartVolumePluginAPI(
			d, sdksocket,
			volume.PluginAPIBase,
			uint16(pluginPort),
		); err != nil {
			return fmt.Errorf("Unable to start plugin api server: %v", err)
		}

		if _, _, err := server.StartVolumeMgmtAPI(
			d, sdksocket,
			volume.DriverAPIBase,
			uint16(mgmtPort),
			false,
		); err != nil {
			return fmt.Errorf("Unable to start volume mgmt api server: %v", err)
		}

		if d != "" && cfg.Osd.ClusterConfig.DefaultDriver == d {
			isDefaultSet = true
		}

		// Start CSI Server for this driver
		csisock := os.Getenv("CSI_ENDPOINT")
		if len(csisock) == 0 {
			csisock = fmt.Sprintf("/var/lib/osd/driver/%s-csi.sock", d)
		}
		os.Remove(csisock)
		cm, err := clustermanager.Inst()
		if err != nil {
			return fmt.Errorf("Unable to find cluster instance: %v", err)
		}
		csiServer, err := csi.NewOsdCsiServer(&csi.OsdCsiServerConfig{
			Net:        "unix",
			Address:    csisock,
			DriverName: d,
			Cluster:    cm,
			SdkUds:     sdksocket,
		})
		if err != nil {
			return fmt.Errorf("Failed to start CSI server for driver %s: %v", d, err)
		}
		csiServer.Start()

		// Create a role manager
		rm, err := role.NewSdkRoleManager(kv)
		if err != nil {
			return fmt.Errorf("Failed to create a role manager")
		}

		// Get authenticators
		authenticators := make(map[string]auth.Authenticator)
		selfSigned, err := selfSignedAuth(c)
		if err != nil {
			logrus.Fatalf("Failed to create self signed config: %v", err)
		} else if selfSigned != nil {
			authenticators[c.String("jwt-issuer")] = selfSigned
		}

		oidcAuth, err := oidcAuth(c)
		if err != nil {
			logrus.Fatalf("Failed to create self signed config: %v", err)
		} else if oidcAuth != nil {
			authenticators[c.String("oidc-issuer")] = oidcAuth
		}

		tlsConfig, err := setupSdkTls(c)
		if err != nil {
			logrus.Fatalf("Failed to access TLS file information: %v", err)
		}

		// Auth is enabled, setup system token manager for inter-cluster communication
		if len(authenticators) > 0 {
			if c.String("jwt-system-shared-secret") == "" {
				return fmt.Errorf("Must provide a jwt-system-shared-secret if auth with oidc or shared-secret is enabled")
			}

			if len(cfg.Osd.ClusterConfig.SystemSharedSecret) == 0 {
				cfg.Osd.ClusterConfig.SystemSharedSecret = c.String("jwt-system-shared-secret")
			}

			// Initialize system token manager if an authenticator is setup
			stm, err := systemtoken.NewManager(&systemtoken.Config{
				ClusterId:    cfg.Osd.ClusterConfig.ClusterId,
				NodeId:       cfg.Osd.ClusterConfig.NodeId,
				SharedSecret: cfg.Osd.ClusterConfig.SystemSharedSecret,
			})
			if err != nil {
				return fmt.Errorf("Failed to create system token manager: %v\n", err)
			}
			auth.InitSystemTokenManager(stm)
		}

		sp, err := policy.Init()
		if err != nil {
			return fmt.Errorf("Unable to Initialise Storage Policy Manager Instances %v", err)
		}

		// Start SDK Server for this driver
		os.Remove(sdksocket)
		sdkServer, err := sdk.New(&sdk.ServerConfig{
			Net:           "tcp",
			Address:       ":" + c.String("sdkport"),
			RestPort:      c.String("sdkrestport"),
			Socket:        sdksocket,
			DriverName:    d,
			Cluster:       cm,
			StoragePolicy: sp,
			Security: &sdk.SecurityConfig{
				Role:                         rm,
				Tls:                          tlsConfig,
				Authenticators:               authenticators,
				PublicVolumeCreationDisabled: !c.Bool("public-volume-create-allowed"),
			},
		})
		if err != nil {
			return fmt.Errorf("Failed to start SDK server for driver %s: %v", d, err)
		}
		sdkServer.Start()
	}

	if cfg.Osd.ClusterConfig.DefaultDriver != "" && !isDefaultSet {
		return fmt.Errorf("Invalid OSD config file: Default Driver specified but driver not initialized")
	}

	if err := flexvolume.StartFlexVolumeAPI(config.FlexVolumePort, cfg.Osd.ClusterConfig.DefaultDriver); err != nil {
		return fmt.Errorf("Unable to start flexvolume API: %v", err)
	}

	// Start the graph drivers.
	for d := range cfg.Osd.GraphDrivers {
		logrus.Infof("Starting graph driver: %v", d)
		if err := server.StartGraphAPI(d, volume.PluginAPIBase); err != nil {
			return fmt.Errorf("Unable to start graph plugin: %v", err)
		}
	}

	if clusterInit {
		cm, err := clustermanager.Inst()
		if err != nil {
			return fmt.Errorf("Unable to find cluster instance: %v", err)
		}
		if err := cm.StartWithConfiguration(
			false,
			"9002",
			[]string{},
			c.String("clusterdomain"),
			&cluster.ClusterServerConfiguration{
				ConfigSchedManager:       schedpolicy.NewFakeScheduler(),
				ConfigObjectStoreManager: objectstore.NewfakeObjectstore(),
				ConfigSystemTokenManager: auth.SystemTokenManagerInst(),
			},
		); err != nil {
			return fmt.Errorf("Unable to start cluster manager: %v", err)
		}
	}

	// Daemon does not exit.
	select {}
}

func showVersion(c *cli.Context) error {
	fmt.Println("OSD Version:", config.Version)
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("OS:", runtime.GOOS)
	fmt.Println("Arch:", runtime.GOARCH)
	return nil
}

func wrapAction(f func(*cli.Context) error) func(*cli.Context) {
	return func(c *cli.Context) {
		if err := f(c); err != nil {
			logrus.Warnln(err.Error())
			os.Exit(1)
		}
	}
}

func selfSignedAuth(c *cli.Context) (*auth.JwtAuthenticator, error) {
	var err error

	rsaFile := getConfigVar(os.Getenv("OPENSTORAGE_AUTH_RSA_PUBKEY"),
		c.String("jwt-rsa-pubkey-file"))
	ecdsFile := getConfigVar(os.Getenv("OPENSTORAGE_AUTH_ECDS_PUBKEY"),
		c.String("jwt-ecds-pubkey-file"))
	sharedsecret := getConfigVar(os.Getenv("OPENSTORAGE_AUTH_SHAREDSECRET"),
		c.String("jwt-shared-secret"))

	if len(rsaFile) == 0 &&
		len(ecdsFile) == 0 &&
		len(sharedsecret) == 0 {
		return nil, nil
	}

	authConfig := &auth.JwtAuthConfig{
		SharedSecret:  []byte(sharedsecret),
		UsernameClaim: auth.UsernameClaimType(c.String("username-claim")),
	}

	// Read RSA file
	if len(rsaFile) != 0 {
		authConfig.RsaPublicPem, err = ioutil.ReadFile(rsaFile)
		if err != nil {
			logrus.Errorf("Failed to read %s", rsaFile)
		}
	}

	// Read Ecds file
	if len(ecdsFile) != 0 {
		authConfig.ECDSPublicPem, err = ioutil.ReadFile(ecdsFile)
		if err != nil {
			logrus.Errorf("Failed to read %s", ecdsFile)
		}
	}

	return auth.NewJwtAuth(authConfig)
}

func oidcAuth(c *cli.Context) (*auth.OIDCAuthenticator, error) {

	if len(c.String("oidc-issuer")) == 0 ||
		len(c.String("oidc-client-id")) == 0 {
		return nil, nil
	}

	return auth.NewOIDC(&auth.OIDCAuthConfig{
		Issuer:            c.String("oidc-issuer"),
		ClientID:          c.String("oidc-client-id"),
		Namespace:         c.String("oidc-custom-claim-namespace"),
		SkipClientIDCheck: c.Bool("oidc-skip-client-id-check"),
		UsernameClaim:     auth.UsernameClaimType(c.String("username-claim")),
	})
}

func setupSdkTls(c *cli.Context) (*sdk.TLSConfig, error) {

	certFile := getConfigVar(os.Getenv("OPENSTORAGE_CERTFILE"),
		c.String("tls-cert-file"))
	keyFile := getConfigVar(os.Getenv("OPENSTORAGE_KEYFILE"),
		c.String("tls-key-file"))

	if len(certFile) != 0 && len(keyFile) != 0 {
		logrus.Infof("TLS %s and %s", certFile, keyFile)
		return &sdk.TLSConfig{
			CertFile: certFile,
			KeyFile:  keyFile,
		}, nil
	}

	return nil, nil
}

func getConfigVar(envVar, cliVar string) string {
	if len(envVar) != 0 {
		return envVar
	}
	return cliVar
}
