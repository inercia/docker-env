package config

import (
	"strconv"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/swarm"
)

const (
	defaultSwarmHost     = "tcp://0.0.0.0:3376"
	defaultSwarmImage    = "swarm:latest"
	defaultSwarmStrategy = "spread"
)

type swarmConfig swarm.Options

func NewSwarmConfig(_ libmachine.API) *swarmConfig {
	return &swarmConfig{
		IsSwarm:  false, // swarm disabled by default
		Host:     defaultSwarmHost,
		Image:    defaultSwarmImage,
		Strategy: defaultSwarmStrategy,
	}
}

func (swarm swarmConfig) Copy() *swarmConfig {
	return &swarm
}

func (swarm *swarmConfig) Populate(api libmachine.API, root *Config, machine *machineConfig) error {
	// TODO
	return nil
}

func (swarm *swarmConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c := make(optionsMap)
	if err := unmarshal(c); err != nil {
		return err
	}

	// enable swarm if we are parsing a "swarm" section
	swarm.IsSwarm = true

	if v, found := c["master"]; found {
		vbool, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		swarm.Master = vbool
	} else {
		swarm.Master = false
	}
	if v, found := c["host"]; found {
		swarm.Host = v
	} else {
		swarm.Host = defaultSwarmHost
	}
	if v, found := c["discovery"]; found {
		swarm.Discovery = v
	}
	if v, found := c["image"]; found {
		swarm.Image = v
	} else {
		swarm.Image = defaultSwarmImage
	}
	if v, found := c["strategy"]; found {
		swarm.Strategy = v
	} else {
		swarm.Strategy = defaultSwarmStrategy
	}

	return nil
}
