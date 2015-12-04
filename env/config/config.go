package config

import (
	"github.com/docker/machine/libmachine"
)

type varsMap map[string]string

type optionsMap map[string]string

// Config is a configuration for an environment
type Config struct {
	Vars     varsMap          `yaml:"vars,omitempty"`
	Auth     *authConfig      `yaml:"auth,omitempty"`
	Engine   *engineConfig    `yaml:"engine,omitempty"`
	Driver   *driverConfig    `yaml:"driver,omitempty"`
	Swarm    *swarmConfig     `yaml:"swarm,omitempty"`
	Machines machineConfigMap `yaml:"machines,omitempty"`
}

type Populater interface {
	Populate(libmachine.API, *Config, *machineConfig) error
}

func (config *Config) Populate(api libmachine.API, _ *Config, _ *machineConfig) error {
	if config.Auth == nil {
		config.Auth = NewAuthConfig(api)
	}
	if config.Engine == nil {
		config.Engine = NewEngineConfig(api)
	}
	if config.Driver == nil {
		config.Driver = NewDriverConfig(api)
	}
	if config.Swarm == nil {
		config.Swarm = NewSwarmConfig(api)
	}

	// replace all the constants
	for _, p := range []Populater{config.Auth, config.Engine, config.Driver, config.Swarm, config.Machines} {
		if err := p.Populate(api, config, nil); err != nil {
			return err
		}
	}
	return nil
}
