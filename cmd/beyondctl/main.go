package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	commonFlagConfig          = "config"
	commonFlagWorkers         = "workers"
	commonFlagReadSpeedLimit  = "read-speed-limit"
	commonFlagWriteSpeedLimit = "write-speed-limit"
)

var commonFlags = []cli.Flag{
	&cli.IntFlag{
		Name:  commonFlagWorkers,
		Usage: "Specify the workers number",
		EnvVars: []string{
			"BEYOND_CTL_WORKERS",
		},
		Value: 4,
	},
	&cli.StringFlag{
		Name:  commonFlagReadSpeedLimit,
		Usage: "Specify speed limit for read I/O operations, for example, 1MB, 10mb, 3GiB.",
		EnvVars: []string{
			"BEYOND_CTL_READ_SPEED_LIMIT",
		},
	},
	&cli.StringFlag{
		Name:  commonFlagWriteSpeedLimit,
		Usage: "Specify speed limit for write I/O operations, for example, 1MB, 10mb, 3GiB.",
		EnvVars: []string{
			"BEYOND_CTL_WRITE_SPEED_LIMIT",
		},
	},
}

var app = cli.App{
	Name: "beyondctl",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    commonFlagConfig,
			Usage:   "Load config from `FILE`",
			Aliases: []string{"c"},
			EnvVars: []string{
				"BEYOND_CTL_CONFIG",
			},
			Value: fmt.Sprintf("%s/beyondctl/config.toml", userConfigDir()),
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
