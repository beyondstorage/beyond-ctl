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

		ch, err := so.List(path)
		if err != nil {
			logger.Error("list",
				zap.String("path", path),
				zap.Error(err))
			return err
		}

		for v := range ch {
			if v.Error != nil {
				logger.Error("read next object", zap.Error(v.Error))
				break
			}
			fmt.Print(parseToShell(v.Object))
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
