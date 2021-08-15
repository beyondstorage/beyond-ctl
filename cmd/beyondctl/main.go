package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var app = cli.App{
	Name: "beyondctl",
	Commands: []*cli.Command{
		lsCmd,
	},
}

func main() {
	logger, _ := zap.NewDevelopment()

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal("beyondctl execute", zap.Error(err))
	}
}
