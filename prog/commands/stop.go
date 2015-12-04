package commands

import (
	"github.com/inercia/docker-env/env/config"

	"github.com/docker/machine/commands"
	"github.com/docker/machine/libmachine"
)

func Stop(c commands.CommandLine, api libmachine.API, cfg *config.Config) error {
	return runForHosts("stop", api, cfg, true)
}
