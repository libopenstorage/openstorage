package main

import (
	"github.com/libopenstorage/openstorage/pkg/flexvolume"
	"github.com/spf13/cobra"
	"go.pedge.io/env"
	"go.pedge.io/lion/env"
	"go.pedge.io/pkg/cobra"
	"google.golang.org/grpc"
)

type appEnv struct {
	OpenstorageAddress string `env:"OPENSTORAGE_ADDRESS,default=0.0.0.0:2345"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	if err := envlion.Setup(); err != nil {
		return err
	}

	initCmd := &cobra.Command{
		Use: "init",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			client, err := getClient(appEnv)
			if err != nil {
				return err
			}
			return client.Init()
		}),
	}

	rootCmd := &cobra.Command{
		Use: "app",
	}
	rootCmd.AddCommand(initCmd)
	return rootCmd.Execute()
}

func getClient(appEnv *appEnv) (flexvolume.Client, error) {
	clientConn, err := grpc.Dial(appEnv.OpenstorageAddress)
	if err != nil {
		return nil, err
	}
	return flexvolume.NewClient(flexvolume.NewAPIClient(clientConn)), nil
}
