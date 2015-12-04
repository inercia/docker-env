package config

import (
	"bytes"
	"strconv"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/engine"
	"gopkg.in/yaml.v2"
)

const (
	defaultEngineInstallURL    = "https://get.docker.com"
	defaultEngineStorageDriver = "aufs"
	defaultEngineTLSVerify     = true
)

type engineConfig engine.Options

func NewEngineConfig(_ libmachine.API) *engineConfig {
	return &engineConfig{
		StorageDriver:    defaultEngineStorageDriver,
		TLSVerify:        defaultEngineTLSVerify,
		InstallURL:       defaultEngineInstallURL,
		ArbitraryFlags:   []string{},
		DNS:              []string{},
		Env:              []string{},
		InsecureRegistry: []string{},
		Labels:           []string{},
		RegistryMirror:   []string{},
	}
}

func parseStringList(s string) ([]string, error) {
	lst := []string{}
	err := yaml.Unmarshal(bytes.NewBufferString(s).Bytes(), &lst)
	if err != nil {
		return nil, err
	}
	return lst, nil
}

func (engine engineConfig) Copy() *engineConfig {
	return &engine
}

func (engine *engineConfig) Populate(api libmachine.API, root *Config, machine *machineConfig) error {
	// TODO: replace vars

	return nil
}

func (engine *engineConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c := make(map[string]string)
	if err := unmarshal(c); err != nil {
		return err
	}

	if v, found := c["opt"]; found {
		lst, err := parseStringList(v)
		if err != nil {
			return err
		}
		engine.ArbitraryFlags = lst
	}
	if v, found := c["environment"]; found {
		lst, err := parseStringList(v)
		if err != nil {
			return err
		}
		engine.Env = lst
	}
	if v, found := c["dns"]; found {
		lst, err := parseStringList(v)
		if err != nil {
			return err
		}
		engine.DNS = lst
	}
	if v, found := c["labels"]; found {
		lst, err := parseStringList(v)
		if err != nil {
			return err
		}
		engine.Labels = lst
	}
	if v, found := c["storage-driver"]; found {
		engine.StorageDriver = v
	} else {
		engine.StorageDriver = defaultEngineStorageDriver
	}

	if v, found := c["se-linux-enabled"]; found {
		vbool, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		engine.SelinuxEnabled = vbool
	}
	if v, found := c["tls-verify"]; found {
		vbool, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		engine.TLSVerify = vbool
	} else {
		engine.TLSVerify = defaultEngineTLSVerify
	}

	if v, found := c["ipv6"]; found {
		vbool, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		engine.Ipv6 = vbool
	}
	if v, found := c["install-url"]; found {
		engine.InstallURL = v
	} else {
		engine.InstallURL = defaultEngineInstallURL
	}

	if v, found := c["log-level"]; found {
		engine.LogLevel = v
	}

	return nil
}
