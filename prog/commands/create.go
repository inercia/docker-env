package commands

import (
	"fmt"

	"github.com/inercia/docker-env/env/config"

	"github.com/docker/machine/commands"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

func Create(c commands.CommandLine, api libmachine.API, cfg *config.Config) error {

	hosts, err := cfg.Machines.NewHosts(api)
	if err != nil {
		return err
	}

	for _, h := range hosts {
		log.Infof("Bringing %s up", h.Name)
		if err := api.Create(h); err != nil {
			return fmt.Errorf("Error attempting to create %s: %s", h.Name, err)
		}
		if err := api.Save(h); err != nil {
			return fmt.Errorf("Error attempting to save store: %s", err)
		}
	}

	return nil
}
