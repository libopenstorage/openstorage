package cli

import (
	"github.com/codegangsta/cli"
)

const (
	DaemonAlias = "daemon, d"
	DaemonFlag  = "daemon"
)

func DaemonMode(c *cli.Context) bool {
	return c.GlobalBool(DaemonFlag)
}
