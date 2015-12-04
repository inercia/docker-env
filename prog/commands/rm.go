package commands

import (
	"github.com/inercia/docker-env/env/config"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

var RmFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "force, f",
		Usage: "remove local configuration even if machine cannot be removed",
	},
}

func Rm(c commands.CommandLine, api libmachine.API, cfg *config.Config) error {
	force := c.Bool("force")

	hosts, err := cfg.Machines.LoadExistingHosts(api, func(name string) {
		log.Infof("Nothing to do on '%s': host does not exist", name)
	})
	if err != nil {
		return err
	}
	for _, h := range hosts {
		if err := h.Driver.Remove(); err != nil {
			if !force {
				log.Errorf("Provider error removing machine %q: %s", h.Name, err)
				continue
			}
		}

		if err := api.Remove(h.Name); err != nil {
			log.Errorf("Error removing machine %q from store: %s", h.Name, err)
		} else {
			log.Infof("Successfully removed %s", h.Name)
		}
	}

	return nil
}
