package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const flagConfig = "config"

var app = cli.App{
	Name: "beyondctl",
	Flags: []cli.Flag{
		&configFlag,
	},
	Commands: []*cli.Command{
		lsCmd,
		profileCmd,
	},
}

var configFlag = cli.StringFlag{
	Name:    flagConfig,
	Usage:   "Load config from `FILE`",
	Aliases: []string{"c"},
	EnvVars: []string{
		"BEYCTL_CONFIG",
	},
	Value: "~/.beyondctl/config.toml",
}

func main() {
	logger, _ := zap.NewDevelopment()

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal("beyondctl execute", zap.Error(err))
	}
}
