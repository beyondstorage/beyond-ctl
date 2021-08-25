package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const flagConfig = "config"

var app = cli.App{
	Name: "beyondctl",
	Flags: []cli.Flag{
		&configFlag,
	},
	Commands: []*cli.Command{
		cpCmd,
		lsCmd,
		profileCmd,
	},
}

var configFlag = cli.StringFlag{
	Name:    flagConfig,
	Usage:   "Load config from `FILE`",
	Aliases: []string{"c"},
	EnvVars: []string{
		"BEYOND_CTL_CONFIG",
	},
	Value: fmt.Sprintf("%s/beyondctl/config.toml", userConfigDir()),
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
