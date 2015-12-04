package main

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/log"
)

const (
	// the application version
	Version = "0.1-dev"
)

var AppHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

{{.Usage}}

Version: {{.Version}}{{if or .Author .Email}}

Author:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

var CommandHelpTemplate = `Usage: docker-env {{.Name}}{{if .Flags}} [OPTIONS]{{end}} [arg...]

{{.Usage}}{{if .Description}}

Description:
   {{.Description}}{{end}}{{if .Flags}}

Options:
   {{range .Flags}}
   {{.}}{{end}}{{ end }}
`

func setDebugOutputLevel() {
	for _, f := range os.Args {
		if f == "-D" || f == "--debug" || f == "-debug" {
			log.IsDebug = true
		}
	}

	debugEnv := os.Getenv("DOCKER_ENV_DEBUG")
	if debugEnv != "" {
		showDebug, err := strconv.ParseBool(debugEnv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing boolean value from DOCKER_ENV_DEBUG: %s\n", err)
			os.Exit(1)
		}
		log.IsDebug = showDebug
	}
}

func main() {
	setDebugOutputLevel()

	cli.AppHelpTemplate = AppHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Author = "Alvaro Saurin"
	app.Email = "https://github.com/inercia/docker-env"
	app.Commands = Commands
	app.CommandNotFound = cmdNotFound
	app.Usage = "Create and manage Docker environments."
	app.Version = Version
	app.Flags = GlobalFlags

	log.Debug("docker-env version: ", app.Version)

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}

func cmdNotFound(c *cli.Context, command string) {
	log.Fatalf(
		"%s: '%s' is not a %s command. See '%s --help'.",
		c.App.Name,
		command,
		c.App.Name,
		os.Args[0],
	)
}
