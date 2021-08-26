package main

import (
	"fmt"

	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/beyondstorage/beyond-ctl/operations"
)

const (
	cpFlagMultipartThreshold = "multipart-threshold"
)

var cpCmd = &cli.Command{
	Name:      "cp",
	Usage:     "copy file from source storager to target storager",
	UsageText: "beyondctl cp [command options] [source] [target]",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  cpFlagMultipartThreshold,
			Usage: "Specify multipart threshold. If source file size is larger than this value, beyondctl will use multipart method to copy file.",
			EnvVars: []string{
				"BEYOND_CTL_MULTIPART_THRESHOLD",
			},
			Value:       1024 * 1024 * 1024, // Use 1 GB as the default value.
			DefaultText: "1GB",
		},
	},
	Before: func(c *cli.Context) error {
		if args := c.Args().Len(); args < 2 {
			return fmt.Errorf("cp command wants two args, but got %d", args)
		}
		return nil
	},
	Action: func(c *cli.Context) (err error) {
		logger, _ := zap.NewDevelopment()

		cfg, err := loadConfig(c, true)
		if err != nil {
			return err
		}

		srcConn, srcKey, err := cfg.ParseProfileInput(c.Args().Get(0))
		if err != nil {
			logger.Error("parse profile input from src", zap.Error(err))
			return err
		}
		dstConn, dstKey, err := cfg.ParseProfileInput(c.Args().Get(1))
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
		if c.IsSet(mainFlagWorkers) {
			do.WithWorkers(c.Int(mainFlagWorkers))
		}

		var ch chan *operations.EmptyResult
		if size < c.Int64(cpFlagMultipartThreshold) {
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
