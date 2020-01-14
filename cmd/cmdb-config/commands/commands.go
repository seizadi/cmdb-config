package commands

import (
	"github.com/urfave/cli"
)

var Commands = []cli.Command{
	{
		Name:   "config",
		Usage:  "config",
		Action: configCmdb,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "repo, r",
				Usage: "GitHub repo for CMDB Config",
			},
			cli.StringFlag{
				Name:  "charts, c",
				Usage: "Charts repo for CMDB Config",
			},
		},
	},
}
