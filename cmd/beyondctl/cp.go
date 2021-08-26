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
			logger.Error("load config", zap.Error(err))
			return err
		}
		cfg.MergeProfileFromEnv()

		srcConn, srcKey, err := cfg.ParseProfileInput(ctx.Args().Get(0))
		if err != nil {
			logger.Error("parse profile input from src", zap.Error(err))
			return err
		}
		dstConn, dstKey, err := cfg.ParseProfileInput(ctx.Args().Get(1))
		if err != nil {
			logger.Error("parse profile input from dst", zap.Error(err))
			return err
		}

		src, err := services.NewStoragerFromString(srcConn)
		if err != nil {
			logger.Error("init src storager", zap.Error(err), zap.String("conn string", srcConn))
			return err
		}

		dst, err := services.NewStoragerFromString(dstConn)
		if err != nil {
			logger.Error("init dst storager", zap.Error(err), zap.String("conn string", dstConn))
			return err
		}

		so := operations.NewSingleOperator(src)

		srcObject, err := so.Stat(srcKey)
		if err != nil {
			logger.Error("stat", zap.String("path", srcKey), zap.Error(err))
			return err
		}

		if srcObject.Mode.IsDir() {
			logger.Error("copy dir is not supported for now")
			return fmt.Errorf("copy dir not supported")
		}

		size, ok := srcObject.GetContentLength()
		if !ok {
			logger.Error("can't get object content length", zap.String("path", srcKey))
			return fmt.Errorf("get object content length failed")
		}

		do := operations.NewDualOperator(src, dst)

		var ch chan *operations.EmptyResult
		// FIXME: we hardcoded 1GB here, we will allow user config it.
		if size < 512*1024*1024 {
			ch, err = do.CopyFileViaWrite(srcKey, dstKey, size)
		} else {
			// TODO: we will support other copy method later.
			ch, err = do.CopyFileViaMultipart(srcKey, dstKey, size)
		}
		if err != nil {
			logger.Error("start copy",
				zap.String("src", srcKey),
				zap.String("dst", dstKey),
				zap.Error(err))
			return err
		}

		for v := range ch {
			logger.Error("read next result", zap.Error(v.Error))
			return v.Error
		}
		return
	},
}
