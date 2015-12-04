package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/inercia/docker-env/env/config"
	cmd "github.com/inercia/docker-env/prog/commands"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"gopkg.in/yaml.v2"
)

const (
	defaultBasename = "docker-env"
)

var currDir = ""

func init() {
	var err error
	currDir, err = os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get current working directory: %s\n", err)
		os.Exit(1)
	}
}

var GlobalFlags = []cli.Flag{
	cli.StringFlag{
		EnvVar: "DOCKER_ENV_DIR",
		Name:   "d, dir",
		Value:  currDir,
		Usage:  "directory where to look for docker-env.yml files",
	},
	cli.StringSliceFlag{
		Name:  "X, var",
		Usage: "define a global variable (eg, '-X NUM_DATABASES=3')",
	},
	cli.StringFlag{
		EnvVar: "DOCKER_ENV_STORAGE_PATH",
		Name:   "s, storage-path",
		Value:  mcndirs.GetBaseDir(),
		Usage:  "configures storage path",
	},
	cli.StringFlag{
		EnvVar: "DOCKER_ENV_GITHUB_API_TOKEN",
		Name:   "github-api-token",
		Usage:  "token to use for requests to the Github API",
		Value:  "",
	},
	cli.BoolFlag{
		EnvVar: "DOCKER_ENV_NATIVE_SSH",
		Name:   "native-ssh",
		Usage:  "use the native (Go-based) SSH implementation.",
	},
	cli.BoolFlag{
		Name:  "debug, D",
		Usage: "enable debug mode",
	},
}

var Commands = []cli.Command{
	{
		Name:        "create",
		Usage:       "Create a Docker environment",
		Description: "Argument(s) are (optional) environment configuration files.",
		Action:      runCommand(cmd.Create),
	},
	{
		Name:        "rm",
		Usage:       "Remove all the hosts in an environment",
		Description: "Argument(s) are (optional) environment configuration files.",
		Action:      runCommand(cmd.Rm),
		Flags:       cmd.RmFlags,
	},
	{
		Name:        "start",
		Usage:       "Start all the hosts in an environment",
		Description: "Argument(s) are (optional) environment configuration files.",
		Action:      runCommand(cmd.Start),
	},
	{
		Name:        "stop",
		Usage:       "Stop all the hosts in an environment",
		Description: "Argument(s) are (optional) environment configuration files.",
		Action:      runCommand(cmd.Stop),
	},
	{
		Name:        "kill",
		Usage:       "Kill all the hosts in an environment",
		Description: "Argument(s) are (optional) environment configuration files.",
		Action:      runCommand(cmd.Kill),
	},
	{
		Name:        "status",
		Usage:       "Get the status of the hosts in an environment",
		Description: "Argument(s) are (optional) environment configuration files.",
		Action:      runCommand(cmd.Status),
	},
	{
		Name:   "version",
		Usage:  "Show the docker-env version information",
		Action: runCommand(cmd.Version),
	},
	{
		Name:   "info",
		Usage:  "Show some info",
		Action: runCommand(cmd.Info),
		Flags:  cmd.InfoFlags,
	},
}

type contextCommandLine struct {
	*cli.Context
}

func (c *contextCommandLine) ShowHelp()             { cli.ShowCommandHelp(c.Context, c.Command.Name) }
func (c *contextCommandLine) ShowVersion()          { cli.ShowVersion(c.Context) }
func (c *contextCommandLine) Application() *cli.App { return c.App }

type commandFun func(commandLine commands.CommandLine, api libmachine.API, cfg *config.Config) error

// runs a command
func runCommand(cmd commandFun) func(context *cli.Context) {
	return func(context *cli.Context) {
		log.Debugf("Creating API client")
		api := libmachine.NewClient(mcndirs.GetBaseDir())

		// set some things from the globals
		if context.GlobalBool("native-ssh") {
			log.Debugf("native SSH enabled")
			api.SSHClientType = ssh.Native
		}
		api.GithubAPIToken = context.GlobalString("github-api-token")
		api.Filestore.Path = context.GlobalString("storage-path")
		mcndirs.BaseDir = api.Filestore.Path
		mcnutils.GithubAPIToken = api.GithubAPIToken
		ssh.SetDefaultClient(api.SSHClientType)

		// load the configuration file(s)
		configDir := context.GlobalString("dir")
		argsStrings := ([]string)(context.Args())
		log.Debugf("Loading config from directory %s", configDir)
		config, err := loadConfig(configDir, argsStrings)
		if err != nil {
			log.Fatal(err)
		}

		// parse variable definitions in the form "var=some_value"
		for _, varDef := range context.GlobalStringSlice("var") {
			varDefComponents := strings.SplitN(varDef, "=", 2)
			if len(varDefComponents) != 2 {
				log.Fatalf("Could not parse variable definition '%s'", varDef)
			}
			key, value := varDefComponents[0], varDefComponents[1]
			log.Debugf("Command line variable: %s = %s", key, value)
			config.Vars[key] = value
		}
		err = config.Populate(api, nil, nil)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: verify the config

		if err := cmd(&contextCommandLine{context}, api, config); err != nil {
			log.Fatal(err)
		}
	}
}

// load the environment configuration from file(s)
func loadConfig(dir string, names []string) (*config.Config, error) {
	config := &config.Config{}

	load := func(basename string) error {
		for _, filename := range []string{
			filepath.Join(dir, fmt.Sprintf("%s.yml", basename)),
			filepath.Join(dir, fmt.Sprintf("%s.yaml", basename)),
		} {
			log.Debugf("Trying to load '%s'", filename)
			if b, err := ioutil.ReadFile(filename); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				if err = yaml.Unmarshal(b, &config); err != nil {
					return fmt.Errorf("Parse error when reading %s: %s", filename, err)
				} else {
					log.Debugf("... '%s' successfully loaded", filename)
					return nil
				}
			}
		}

		log.Debug("No config file found")
		return os.ErrNotExist
	}

	loaded := 0
	err := load(defaultBasename)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		loaded++
	}

	for _, name := range names {
		err = load(fmt.Sprintf("%s-%s", defaultBasename, name))
		if err != nil {
			return nil, err
		}
		loaded++
	}
	if loaded == 0 {
		return nil, fmt.Errorf("No configuration files found")
	}

	return config, nil
}
