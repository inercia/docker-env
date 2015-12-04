package commands

import (
	"github.com/inercia/docker-env/env/config"

	"github.com/codegangsta/cli"

	"github.com/davecgh/go-spew/spew"
	"github.com/docker/machine/commands"
	"github.com/docker/machine/libmachine"
)

var InfoFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "tree, t",
		Usage: "dump the current environment tree",
	},
}

func Info(c commands.CommandLine, api libmachine.API, cfg *config.Config) error {
	showTree := c.Bool("tree")

	if showTree {
		scs := spew.ConfigState{
			Indent:   "\t",
			SortKeys: true,
			SpewKeys: true,
		}

		scs.Dump(cfg)
	}

	return nil
}
