package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/codegangsta/cli"
	"github.com/seizadi/template-engine/cmd/template-engine/commands"
)

const version = "0.1"

func main() {
	app := cli.NewApp()
	app.Name = "Template Engine CLI"
	app.Usage = ""
	app.Author = "Soheil Eizadi"
	app.Email = "seizadi@gmail.com"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config_dir",
			Usage:  "path to the config directory",
			Value:  "config/",
			EnvVar: "CONFIG_SOURCE",
		},
		cli.StringFlag{
			Name:   "config_file",
			Usage:  "path to the config file",
			Value:  "config/",
			EnvVar: "CONFIG_FILE",
		},
		cli.StringFlag{
			Name:   "apikey",
			Usage:  "api key for CMDB access",
			Value:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBY2NvdW50SUQiOjF9.GsXyFDDARjXe1t9DPo2LIBKHEal3O7t3vLI3edA7dGU",
			EnvVar: "API_KEY",
		},
		cli.StringFlag{
			Name:   "cmdb",
			Usage:  "Host address to CMDB",
			Value:  "localhost:9090",
			EnvVar: "CMDB_HOST",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "enable debug logging",
		},
	}
	
	app.Before = func(c *cli.Context) error {
		return nil
	}
	app.Commands = commands.Commands
	app.CommandNotFound = cmdNotFound
	
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("%v", err)
	}
}


func cmdNotFound(c *cli.Context, command string) {
	fmt.Printf(
		"%s: '%s' is not a %s command. See '%s --help'.\n",
		c.App.Name,
		command,
		c.App.Name,
		os.Args[0],
	)
	os.Exit(1)
}