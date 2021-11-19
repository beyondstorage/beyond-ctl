package main

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"path/filepath"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"
)

const (
	cpFlagMultipartThresholdName = "multipart-threshold"
	cpFlagRecursive              = "recursive"
)

var cpFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  cpFlagMultipartThresholdName,
		Usage: "Specify multipart threshold. If source file size is larger than this value, byctl will use multipart method to copy file.",
		EnvVars: []string{
			"BEYOND_CTL_MULTIPART_THRESHOLD",
		},
		Value: "1GiB", // Use 1 GiB as the default value.
	},
	&cli.BoolFlag{
		Name: cpFlagRecursive,
		Aliases: []string{
			"r",
			"R",
		},
		Usage: "copy directories recursively",
	},
}

var cpCmd = &cli.Command{
	Name:      "cp",
	Usage:     "copy file from source storager to target storager",
	UsageText: "byctl cp [command options] [source] [target]",
	Flags:     mergeFlags(globalFlags, ioFlags, cpFlags),
	Before: func(c *cli.Context) error {
		if args := c.Args().Len(); args < 2 {
			return fmt.Errorf("cp command wants at least two args, but got %d", args)
		}
		return nil
	},
	Action: func(c *cli.Context) (err error) {
		logger, _ := zap.NewDevelopment()

		cfg, err := loadConfig(c, true)
		if err != nil {
			logger.Error("load config", zap.Error(err))
			return err
		}

		argsNum := c.Args().Len()

		dstConn, dstKey, err := cfg.ParseProfileInput(c.Args().Get(argsNum - 1))
		if err != nil {
			logger.Error("parse profile input from dst", zap.Error(err))
			return err
		}

		dst, err := services.NewStoragerFromString(dstConn)
		if err != nil {
			logger.Error("init dst storager", zap.Error(err), zap.String("conn string", dstConn))
			return err
		}

		dstSo := operations.NewSingleOperator(dst)

		dstObject, err := dstSo.Stat(dstKey)
		if err != nil {
			if errors.Is(err, services.ErrObjectNotExist) {
				err = nil
			} else {
				logger.Error("stat", zap.Error(err), zap.String("dst path", dstKey))
				return err
			}
		}
		if argsNum > 2 {
			if err == nil && !dstObject.Mode.IsDir() {
				fmt.Printf("cp: target '%s' is not a directory\n", dstKey)
				return fmt.Errorf("cp: target '%s' is not a directory", dstKey)
			}
		}

		// Handle read pairs.
		var readPairs []types.Pair
		if c.IsSet(flagReadSpeedLimitName) {
			limitPair, err := parseLimit(c.String(flagReadSpeedLimitName))
			if err != nil {
				logger.Error("read limit is invalid",
					zap.String("input", c.String(flagReadSpeedLimitName)),
					zap.Error(err))
				return err
			}

			readPairs = append(readPairs, limitPair)
		}

		// Handle write pairs.
		var writePairs []types.Pair
		if c.IsSet(flagWriteSpeedLimitName) {
			limitPair, err := parseLimit(c.String(flagWriteSpeedLimitName))
			if err != nil {
				logger.Error("write limit is invalid",
					zap.String("input", c.String(flagWriteSpeedLimitName)),
					zap.Error(err))
				return err
			}

			writePairs = append(writePairs, limitPair)
		}

		// parse flag multipart-threshold, 1GB is the default value
		multipartThreshold, err := units.FromHumanSize(c.String(cpFlagMultipartThresholdName))
		if err != nil {
			logger.Error("multipart-threshold is invalid",
				zap.String("input", c.String(cpFlagMultipartThresholdName)),
				zap.Error(err))
			return err
		}

		for i := 0; i < argsNum-1; i++ {
			srcConn, srcKey, err := cfg.ParseProfileInput(c.Args().Get(i))
			if err != nil {
				logger.Error("parse profile input from src", zap.Error(err))
				continue
			}

			src, err := services.NewStoragerFromString(srcConn)
			if err != nil {
				logger.Error("init src storager", zap.Error(err), zap.String("conn string", srcConn))
				continue
			}

			so := operations.NewSingleOperator(src)

			srcObject, err := so.Stat(srcKey)
			if err != nil {
				logger.Error("stat", zap.String("path", srcKey), zap.Error(err))
				continue
			}

			if srcObject.Mode.IsDir() && !c.Bool(cpFlagRecursive) {
				fmt.Printf("cp: -r not specified; omitting directory '%s'\n", srcKey)
				continue
			}

			var size int64
			if srcObject.Mode.IsRead() {
				n, ok := srcObject.GetContentLength()
				if !ok {
					logger.Error("can't get object content length", zap.String("path", srcKey))
					continue
				}
				size = n
			}

			do := operations.NewDualOperator(src, dst)
			if c.IsSet(flagWorkersName) {
				do.WithWorkers(c.Int(flagWorkersName))
			}

			// set read pairs
			do.WithReadPairs(readPairs...)
			// set write pairs
			do.WithWritePairs(writePairs...)

			realDstKey := dstKey
			if argsNum > 2 || (dstObject != nil && dstObject.Mode.IsDir()) {
				realDstKey = filepath.Join(dstKey, filepath.Base(srcKey))
			}

			var ch chan *operations.EmptyResult
			if c.Bool(cpFlagRecursive) && srcObject.Mode.IsDir() {
				ch, err = do.CopyRecursively(srcKey, realDstKey, multipartThreshold)
			} else if size < multipartThreshold {
				ch, err = do.CopyFileViaWrite(srcKey, realDstKey, size)
			} else {
				// TODO: we will support other copy method later.
				ch, err = do.CopyFileViaMultipart(srcKey, realDstKey, size)
			}
			if err != nil {
				logger.Error("start copy",
					zap.String("src", srcKey),
					zap.String("dst", realDstKey),
					zap.Error(err))
				continue
			}

			for v := range ch {
				logger.Error("read next result", zap.Error(v.Error))
			}
		}

		return
	},
}
