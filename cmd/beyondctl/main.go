package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const flagConcurrent = "concurrent"

var app = cli.App{
	Name: "beyondctl",
	Commands: []*cli.Command{
		cpCmd,
		lsCmd,
	},
	Flags: []cli.Flag{
		&concurrentFlag,
	},
}

var concurrentFlag = cli.IntFlag{
	Name:  flagConcurrent,
	Usage: "adjust count for concurrency",
	EnvVars: []string{
		"BEYCTL_CONCURRENT",
	},
	Value: 10,
}

func main() {
	logger, _ := zap.NewDevelopment()

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal("beyondctl execute", zap.Error(err))
	}
}
