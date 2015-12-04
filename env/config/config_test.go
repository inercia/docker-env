package config_test

import (
	"bytes"
	"testing"

	"github.com/inercia/docker-env/env/config"

	"github.com/davecgh/go-spew/spew"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var scs = spew.ConfigState{
	Indent:   "\t",
	SortKeys: true,
}

const test_config = `
# docker-env.yml
vars:
  NUM_DATABASES: 1
  NUM_WORKERS:   10

auth:
  tls-ca-cert:  /my/ca.pem

driver:
  openstack:
    flavor-name:        tiny
    image-name:         Ubuntu 14.04 LTS
    floatingip-pool:    myfloatingips

swarm:
  strategy:  spread
  discovery: token://1234

machines:
  master:
    instances: 1
    # this machine will not inherit anything from the global "awarm"
    swarm:
      master:  true
`

func TestConfigParsing(t *testing.T) {
	config := config.Config{}

	b := bytes.NewBufferString(test_config)
	err := yaml.Unmarshal(b.Bytes(), &config)
	require.NoError(t, err, "config parsing error (1)")

	//spew.Dump(config)

	_, found := config.Machines["master"]
	require.True(t, found, "machine not found")
	instances := config.Machines["master"].Instances
	require.Equal(t, instances, "1", "instances mismatch")
}

func TestConfigMultipleDriversError(t *testing.T) {
	config := config.Config{}

	const test_config_multi_driver = `
driver:
  openstack:
    flavor-name:        tiny
    image-name:         Ubuntu 14.04 LTS
    floatingip-pool:    myfloatingips
  virtualbox:
    memory:             2Gb
`

	b := bytes.NewBufferString(test_config_multi_driver)
	err := yaml.Unmarshal(b.Bytes(), &config)
	require.Error(t, err, "multiple drivers error")
}

func TestConfigMerge(t *testing.T) {
	config := config.Config{}
	const test_config_vars_new = `
# docker-env.yml
vars:
  NUM_DATABASES: 1
  NUM_WORKERS:   5
  NUM_QUEUES:    9
`

	const test_config_machines_new = `
machines:
  master:
    instances: 9
    swarm:
      discovery:  token://9999
`

	for _, s := range []string{test_config, test_config_vars_new, test_config_machines_new} {
		b := bytes.NewBufferString(s)
		err := yaml.Unmarshal(b.Bytes(), &config)
		require.NoError(t, err, "config parsing error (1)")
	}

	//scs.Dump(config)

	for _, constant := range []struct {
		Key   string
		Value string
	}{
		{"NUM_DATABASES", "1"},
		{"NUM_WORKERS", "5"},
		{"NUM_QUEUES", "9"},
	} {
		have, found := config.Vars[constant.Key]
		require.True(t, found, "constant not found")
		require.Equal(t, constant.Value, have, "constant mismatch")
	}

	_, found := config.Machines["master"]
	require.True(t, found, "machine not found")
	instances := config.Machines["master"].Instances
	require.Equal(t, instances, "9", "instances mismatch")
	discovery := config.Machines["master"].Swarm.Discovery
	require.Equal(t, discovery, "token://9999", "discovery mismatch")
}

func TestConfigPopulate(t *testing.T) {
	config := config.Config{}
	const (
		numMasters   = 1
		numDatabases = 1
		numWorkes    = 2
		numNone      = 0
	)

	const test_config_machines_databases = `
vars:
  NUM_NONE: 0
machines:
  database:
    instances: $(NUM_DATABASES)
  worker:
    instances: 2
  none:
    instances: $(NUM_NONE)
`

	for _, s := range []string{test_config, test_config_machines_databases} {
		b := bytes.NewBufferString(s)
		err := yaml.Unmarshal(b.Bytes(), &config)
		require.NoError(t, err, "config parsing error (1)")
	}

	api := libmachine.NewClient(mcndirs.GetBaseDir())
	config.Populate(api, &config, nil)

	// scs.Dump(config)

	require.Len(t, config.Machines, numMasters+numDatabases+numWorkes+numNone, "wrong number of machines")
	_, found := config.Machines["none"]
	require.False(t, found, "'none' in machines")

	// check the machine has inherited the global driver, engine, etc...
	for _, machine := range config.Machines {
		require.NotNil(t, machine.Driver, "nil driver section")
		require.NotNil(t, machine.Engine, "nil engine section")
		require.NotNil(t, machine.Auth, "nil auth section")
		require.NotNil(t, machine.Swarm, "nil swarm section")

		require.Equal(t, machine.Driver.Name, "openstack", "driver name mismatch")

		// the master machine has specified its own swarm section
		if machine.Name != "master" {
			require.Equal(t, machine.Swarm.Discovery, "token://1234", "swarm discovery mismatch")
		}
	}
}

func TestPopulateNoGlobal(t *testing.T) {
	const test_config_no_global = `
machines:
  master:
    instances: 1
    auth:
      tls-ca-cert:  /my/ca.pem
    driver:
      openstack:
        flavor-name:        tiny
        image-name:         Ubuntu 14.04 LTS
        floatingip-pool:    myfloatingips
        labels:             [aaa, bbb]
        ssh-enabled:        true
    swarm:
      strategy:  spread
      discovery: token://1234
      master:    true
`

	config := config.Config{}
	for _, s := range []string{test_config_no_global} {
		b := bytes.NewBufferString(s)
		err := yaml.Unmarshal(b.Bytes(), &config)
		require.NoError(t, err, "config parsing error")
	}

	api := libmachine.NewClient(mcndirs.GetBaseDir())
	config.Populate(api, &config, nil)

	machine := config.Machines["master"]
	require.Equal(t, machine.Swarm.Host, "tcp://0.0.0.0:3376", "swarm host mismatch")
	require.Equal(t, machine.Engine.InstallURL, "https://get.docker.com", "engine install URL mismatch")

	scs.Dump(config)
}
