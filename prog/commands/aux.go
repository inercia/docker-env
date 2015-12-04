package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/inercia/docker-env/env"
	"github.com/inercia/docker-env/env/config"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
)

// runs an action for a list of (already existing) hosts provided in the command line
func runForHosts(actionName string, api libmachine.API, cfg *config.Config, ignoreMissing bool) error {
	var hosts []*host.Host
	var err error

	consolidateErrs := func(errs []error) error {
		finalErr := ""
		for _, err := range errs {
			finalErr = fmt.Sprintf("%s\n%s", finalErr, err)
		}

		return errors.New(strings.TrimSpace(finalErr))
	}

	if ignoreMissing {
		hosts, err = cfg.Machines.LoadExistingHosts(api, func(name string) {
			log.Infof("Host '%s' does not exist", name)
		})
	} else {
		hosts, err = cfg.Machines.LoadHosts(api)
	}
	if err != nil {
		return err
	}

	if errs := env.RunActionForeachMachine(actionName, hosts); len(errs) > 0 {
		return consolidateErrs(errs)
	}
	for _, h := range hosts {
		if err := api.Save(h); err != nil {
			return fmt.Errorf("Error saving host to store: %s", err)
		}
	}
	return nil
}
