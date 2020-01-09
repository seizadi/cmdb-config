package commands

import (
	"github.com/seizadi/cmdb-config/engine"
	"github.com/urfave/cli"
)

func configCmdb(c *cli.Context) {
	engine.ConfigCmdb(getHost(c), getApiKey(c), c.String("repo"))
}
