package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/inercia/docker-env/env/config"

	"github.com/docker/machine/commands"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
)

func Status(c commands.CommandLine, api libmachine.API, cfg *config.Config) error {
	hosts, err := cfg.Machines.LoadHosts(api)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tDRIVER\tSTATE\tURL\tERRORS")

	for _, host := range hosts {
		url := ""
		hostError := "(none)"
		currentState, err := host.Driver.GetState()
		if err != nil {
			log.Errorf("Error getting state for host %s: %s", host.Name, err)
			hostError = err.Error()
			if hostError == drivers.ErrHostIsNotRunning.Error() {
				hostError = ""
			}
		} else {
			if url, err = host.Driver.GetURL(); err != nil {
				log.Warnf("Error getting URL for host %s: %s", host.Name, err)
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			host.Name, host.DriverName, currentState, url, hostError)

	}
	w.Flush()

	return nil
}
