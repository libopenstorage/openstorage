package cli

import (
	"github.com/codegangsta/cli"
)

const (
	DaemonAlias = "daemon, d"
	DaemonFlag  = "daemon"
	DriverFlag  = "driver"
)

func DaemonMode(c *cli.Context) bool {
	return c.GlobalBool(DaemonFlag)
}

func DriverName(c *cli.Context) string {
	return c.GlobalString(DriverFlag)
}
