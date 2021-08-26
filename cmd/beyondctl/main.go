package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	mainFlagConfig  = "config"
	mainFlagWorkers = "workers"
)

var app = cli.App{
	Name: "beyondctl",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    mainFlagConfig,
			Usage:   "Load config from `FILE`",
			Aliases: []string{"c"},
			EnvVars: []string{
				"BEYOND_CTL_CONFIG",
			},
			Value: fmt.Sprintf("%s/beyondctl/config.toml", userConfigDir()),
		},
		&cli.IntFlag{
			Name:  mainFlagWorkers,
			Usage: "Specify the workers number",
			EnvVars: []string{
				"BEYOND_CTL_WORKERS",
			},
			Value: 4,
		},
	},
	Commands: []*cli.Command{
		cpCmd,
		lsCmd,
		profileCmd,
	},
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		// FIXME: we need to respect platform style later.
		os.Exit(1)
	}
}

func userConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic("$HOME is not specified")
	}
	return configDir
}
