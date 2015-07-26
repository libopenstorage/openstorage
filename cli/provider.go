package cli

import (
	"github.com/codegangsta/cli"
)

func providerList(c *cli.Context) {
}

func providerAdd(c *cli.Context) {
}

func ProviderCommands() []cli.Command {
	commands := []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add a new provider",
			Action:  providerAdd,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name,n",
					Usage: "Provider Name",
				},
				cli.StringFlag{
					Name:  "options,o",
					Usage: "Comma separated name=value pairs, e.g disk=/dev/xvdg,mount=/var/openstorage/btrfs",
				},
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List providers",
			Action:  providerList,
		},
	}
	return commands
}
