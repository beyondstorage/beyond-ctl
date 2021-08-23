package main

import (
	"fmt"

	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/beyondstorage/beyond-ctl/config"
	"github.com/beyondstorage/beyond-ctl/operations"
)

var cpCmd = &cli.Command{
	Name:      "cp",
	Usage:     "copy file from source storager to target storager",
	UsageText: "beyondctl cp [command options] [source] [target]",
	Flags:     []cli.Flag{},
	Before: func(ctx *cli.Context) error {
		if args := ctx.Args().Len(); args < 2 {
			return fmt.Errorf("cp command wants two args, but got %d", args)
		}
		return nil
	},
	Action: func(ctx *cli.Context) (err error) {
		logger, _ := zap.NewDevelopment()

		path := ctx.String(flagConfig)

		cfg, err := config.LoadFromFile(path)
		if err != nil {
			return err
		}
		cfg.MergeProfileFromEnv()

		srcConn, srcKey, err := cfg.ParseProfileInput(ctx.Args().Get(0))
		if err != nil {
			return err
		}
		dstConn, dstKey, err := cfg.ParseProfileInput(ctx.Args().Get(1))
		if err != nil {
			return err
		}

		src, err := services.NewStoragerFromString(srcConn)
		if err != nil {
			return err
		}

		dst, err := services.NewStoragerFromString(dstConn)
		if err != nil {
			return err
		}

		bo, err := operations.NewDualOperator(src, dst)
		if err != nil {
			return err
		}

		go func() {
			for v := range bo.Errors() {
				logger.Error("copy", zap.Error(v))
			}
		}()

		bo.Copy(srcKey, dstKey)
		return
	},
}
