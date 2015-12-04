package commands

import (
	"github.com/inercia/docker-env/env/config"

	"github.com/docker/machine/commands"
	"github.com/docker/machine/libmachine"
)

func Version(c commands.CommandLine, api libmachine.API, cfg *config.Config) error {
	c.ShowVersion()
	return nil
}
