package commands

import (
	"github.com/codegangsta/cli"
)

var Commands = []cli.Command{
	{
		Name:   "create",
		Usage:  "create",
		Action: createManifest,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cloud, c",
				Usage: "Cloud Provider account to select",
			},
			cli.StringFlag{
				Name:  "region, r",
				Usage: "Region to select",
			},
			cli.StringFlag{
				Name:  "environment, e",
				Usage: "Environment to use for processing",
			},
			cli.StringFlag{
				Name:  "application, a",
				Usage: "Application to use for processing",
			},
			cli.StringFlag{
				Name:  "path, p",
				Usage: "Local Chart path, if not specified pull from archive",
			},
		},
	},
}