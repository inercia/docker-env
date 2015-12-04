package config

import (
	"encoding/json"
	"fmt"

	"strings"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/drivers/errdriver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/log"
)

type driverOptions map[string]interface{}

func (d driverOptions) String(key string) string        { return d[key].(string) }
func (d driverOptions) StringSlice(key string) []string { return d[key].([]string) }
func (d driverOptions) Int(key string) int              { return d[key].(int) }
func (d driverOptions) Bool(key string) bool            { return d[key].(bool) }

type driverConfig struct {
	Name    string
	Options driverOptions

	machine *machineConfig `yaml:"-"`
}

func NewDriverConfig(_ libmachine.API) *driverConfig {
	return &driverConfig{
		Options: make(map[string]interface{}),
	}
}

func (driver driverConfig) Copy() *driverConfig {
	return &driver
}

func (driver *driverConfig) Populate(api libmachine.API, root *Config, machine *machineConfig) error {
	driver.machine = machine
	// TODO: fill things that are missing...
	return nil
}

func (driver *driverConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c := make(map[string]driverOptions)
	if err := unmarshal(c); err != nil {
		return err
	}

	if len(c) > 1 {
		return fmt.Errorf("Only one driver can be specified")
	}

	for k, v := range c {
		driver.Name = k
		driver.Options = v
		break
	}

	return nil
}

// Get a plugin driver
func (driver *driverConfig) Get(api libmachine.API) (drivers.Driver, error) {
	bareDriverData, err := json.Marshal(&drivers.BaseDriver{
		MachineName: driver.machine.Name,
		StorePath:   mcndirs.GetBaseDir(),
	})
	if err != nil {
		return nil, fmt.Errorf("Error attempting to marshal data: %s", err)
	}

	log.Debugf("Creating plugin driver for '%s'", driver.Name)
	d, err := api.NewPluginDriver(driver.Name, bareDriverData)
	if err != nil {
		return nil, fmt.Errorf("Error loading '%s' driver: %s", driver.Name, err)
	}
	if _, ok := d.(*errdriver.Driver); ok {
		return nil, errdriver.NotLoadable{driver.Name}
	}

	// We need it so that we can actually send the flags for creating
	// a machine over the wire (cli.Context is a no go since there is so
	// much stuff in it).
	driverOpts := rpcdriver.RPCFlags{
		Values: make(map[string]interface{}),
	}

	mcnFlags := d.GetCreateFlags()
	for _, f := range mcnFlags {
		flagName := f.String()
		prefix := fmt.Sprintf("%s-", driver.Name)
		if !strings.HasPrefix(flagName, prefix) {
			panic(fmt.Sprintf("Flag '%s' does not start with '%s'", f, prefix))
		}

		flagWithoutDriver := flagName[len(prefix):]
		value, found := driver.Options[flagWithoutDriver]
		if found {
			log.Debugf("Setting %s = %s", flagName, value)
			driverOpts.Values[flagName] = value
		} else {
			value = f.Default()
			log.Debugf("Setting %s = %v (default)", flagName, value)
			driverOpts.Values[flagName] = value

			// Hardcoded logic for boolean... :(
			if f.Default() == nil {
				driverOpts.Values[f.String()] = false
			}
		}
	}

	if err := d.SetConfigFromFlags(driverOpts); err != nil {
		return nil, fmt.Errorf("Error setting machine configuration from flags provided: %s", err)
	}

	return d, nil
}
