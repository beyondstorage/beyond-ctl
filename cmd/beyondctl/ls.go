package main

import (
	"fmt"

	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/beyondstorage/beyond-ctl/config"
	"github.com/beyondstorage/beyond-ctl/operations"
)

var lsCmd = &cli.Command{
	Name: "ls",
	Action: func(ctx *cli.Context) (err error) {
		logger, _ := zap.NewDevelopment()

		cfg, err := config.LoadFromFile(ctx.String(flagConfig))
		if err != nil {
			return err
		}
		cfg.MergeProfileFromEnv()

		conn, path, err := cfg.ParseProfileInput(ctx.Args().Get(0))
		if err != nil {
			return err
		}

		store, err := services.NewStoragerFromString(conn)
		if err != nil {
			return err
		}

		so, err := operations.NewSingleOperator(store)
		if err != nil {
			return err
		}

		go func() {
			for v := range so.Errors() {
				logger.Error("", zap.Error(v))
			}
		}()

		for v := range so.List(path) {
			fmt.Print(parseToShell(v))
		}
		// End of line
		fmt.Print("\n")
		return
	},
}

func parseToShell(o *types.Object) (content string) {
	buf := pool.Get()
	defer buf.Free()

	buf.AppendString(o.Path)
	buf.AppendString(" ")
	return string(buf.Bytes())
}
