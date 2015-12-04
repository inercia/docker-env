package config

import (
	"fmt"
	"strconv"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/swarm"
)

const (
	defaultMachineCPUs   = 2
	defaultMachineMemory = 2048
)

// Config is a configuration for an environment
type machineConfig struct {
	Name      string        `yaml:"-"`
	Instances string        `yaml:"instances,omitempty"`
	Auth      *authConfig   `yaml:"auth,omitempty"`
	Engine    *engineConfig `yaml:"engine,omitempty"`
	Driver    *driverConfig `yaml:"driver,omitempty"`
	Swarm     *swarmConfig  `yaml:"swarm,omitempty"`
}

func (machine machineConfig) Copy() *machineConfig {
	machine.Auth = machine.Auth.Copy()
	machine.Engine = machine.Engine.Copy()
	machine.Driver = machine.Driver.Copy()
	machine.Swarm = machine.Swarm.Copy()
	return &machine
}

func (machine *machineConfig) Populate(api libmachine.API, root *Config, _ *machineConfig) error {
	// take missing sections from the global config
	if machine.Auth == nil {
		machine.Auth = root.Auth.Copy()
	}
	if machine.Engine == nil {
		machine.Engine = root.Engine.Copy()
	}
	if machine.Driver == nil {
		machine.Driver = root.Driver.Copy()
	}
	if machine.Swarm == nil {
		machine.Swarm = root.Swarm.Copy()
	}

	// populate the sections
	for _, p := range []Populater{machine.Auth, machine.Engine, machine.Driver, machine.Swarm} {
		if err := p.Populate(api, root, machine); err != nil {
			return err
		}
	}

	return nil
}

func (machine *machineConfig) NewHost(api libmachine.API) (*host.Host, error) {
	driver, err := machine.Driver.Get(api)
	if err != nil {
		return nil, fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}
	driverName := driver.DriverName()

	h, err := api.NewHost(driver)
	if err != nil {
		return nil, fmt.Errorf("Error getting new host: %s", err)
	}

	h.HostOptions = &host.Options{
		Driver:        driverName,
		Memory:        defaultMachineCPUs,
		Disk:          defaultMachineMemory,
		EngineOptions: (*engine.Options)(machine.Engine),
		SwarmOptions:  (*swarm.Options)(machine.Swarm),
		AuthOptions:   (*auth.Options)(machine.Auth),
	}

	exists, err := api.Exists(h.Name)
	if err != nil {
		return nil, fmt.Errorf("Error checking if host exists: %s", err)
	}
	if exists {
		return nil, mcnerror.ErrHostAlreadyExists{
			Name: h.Name,
		}
	}

	return h, nil
}

// Load the machine configuration from the storage
func (machine *machineConfig) LoadHost(api libmachine.API) (*host.Host, error) {
	host, err := api.Load(machine.Name)
	if err != nil {
		return nil, err
	}
	return host, nil
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type machineConfigMap map[string]*machineConfig

func (m machineConfigMap) Populate(api libmachine.API, root *Config, _ *machineConfig) error {
	for name, machine := range m {
		// fix the name
		machine.Name = name

		// replace all the vars
		machine, ok := ReplaceAllStringMap(machine, map[string]string(root.Vars)).(*machineConfig)
		if !ok {
			panic("could not cast to machineConfig")
		}

		// populate the machine
		if err := machine.Populate(api, root, machine); err != nil {
			return err
		}

		// check if there are multiple instances of the machine...
		numInstances, err := strconv.Atoi(machine.Instances)
		if err != nil {
			if hasVars(machine.Instances) {
				return fmt.Errorf("undefined variable(s) in the instances number, '%s'", machine.Instances)
			}
			return fmt.Errorf("cannot parse the number of instances from '%s'", machine.Instances)
		}

		if numInstances == 0 {
			delete(m, name)
		} else if numInstances > 1 {
			if !hasVar(machine.Name, "#") {
				machine.Name = fmt.Sprintf("%s-$(#)", machine.Name)
			}

			// create the additional machines
			for i := 2; i <= numInstances; i++ {
				additionalMachine := machine.Copy()
				newName, err := replaceVars(machine.Name, map[string]string{"#": strconv.Itoa(i)})
				if err != nil {
					return fmt.Errorf("could not replace the $(#) variable in '%s': %s", additionalMachine.Name, err)
				}
				additionalMachine.Name = newName
				additionalMachine.Instances = "1"

				// populate the new machine
				if err := additionalMachine.Populate(api, root, machine); err != nil {
					return err
				}

				m[additionalMachine.Name] = additionalMachine
			}

			newName, err := replaceVars(machine.Name, map[string]string{"#": "1"})
			if err != nil {
				return fmt.Errorf("could not replace some vars in '%s': %s", machine.Name, err)
			}
			machine.Name = newName
			machine.Instances = "1"

			// we have inserted "machine-1", "machine-2"..., so remove the old "machine"
			delete(m, name)
			m[machine.Name] = machine
		} else {
			// update the machine with the populated version
			m[machine.Name] = machine
		}
	}
	return nil
}

func (m machineConfigMap) NewHosts(api libmachine.API) ([]*host.Host, error) {
	res := []*host.Host{}
	for _, machine := range m {
		host, err := machine.NewHost(api)
		if err != nil {
			return nil, err
		}
		res = append(res, host)
	}
	return res, nil
}

// Load all the existing hosts
func (m machineConfigMap) LoadExistingHosts(api libmachine.API, f func(string)) ([]*host.Host, error) {
	res := []*host.Host{}
	for _, machine := range m {
		host, err := machine.LoadHost(api)
		if err != nil {
			switch err := err.(type) {
			case mcnerror.ErrHostDoesNotExist:
				f(machine.Name)
				continue
			default:
				return nil, err
			}
		}
		res = append(res, host)
	}
	return res, nil
}

// Load all the hosts (they must exist in the store)
func (m machineConfigMap) LoadHosts(api libmachine.API) ([]*host.Host, error) {
	notFound := 0
	hosts, err := m.LoadExistingHosts(api, func(name string) {
		log.Infof("Host '%s' does not exist", name)
		notFound++
	})
	if notFound > 0 {
		return nil, fmt.Errorf("Could not load %d hosts", notFound)
	}
	return hosts, err
}
