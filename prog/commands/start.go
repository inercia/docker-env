package commands

import (
	"github.com/inercia/docker-env/env/config"

	"github.com/docker/machine/commands"
	"github.com/docker/machine/libmachine"
)

func Start(c commands.CommandLine, api libmachine.API, cfg *config.Config) error {
	if err := runForHosts("start", api, cfg, false); err != nil {
		return err
	}
	return nil
}
