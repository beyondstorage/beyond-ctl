package main

import (
	"fmt"
	"strings"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"
)

const (
	mvFlagMultipartThresholdName = "multipart-threshold"
	mvFlagRecursive              = "recursive"
)

var mvFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  mvFlagMultipartThresholdName,
		Usage: "Specify multipart threshold. If source file size is larger than this value, byctl will use multipart method to copy file.",
		EnvVars: []string{
			"BEYOND_CTL_MULTIPART_THRESHOLD",
		},
		Value: "1GiB", // Use 1 GiB as the default value.
	},
	&cli.BoolFlag{
		Name: mvFlagRecursive,
		Aliases: []string{
			"r",
			"R",
		},
		Usage: "move directories recursively",
	},
}

var mvCmd = &cli.Command{
	Name:      "mv",
	Usage:     "move file from source storager to target storager",
	UsageText: "byctl mv [command options] [source] [target]",
	Flags:     mergeFlags(globalFlags, ioFlags, mvFlags),
	Before: func(c *cli.Context) error {
		if args := c.Args().Len(); args < 2 {
			return fmt.Errorf("mv command wants two args, but got %d", args)
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

		if c.Bool(mvFlagRecursive) && !strings.HasSuffix(srcKey, "/") {
			srcKey += "/"
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

		size, ok := srcObject.GetContentLength()
		if !ok {
			logger.Error("can't get object content length", zap.String("path", srcKey))
			return fmt.Errorf("get object content length failed")
		}

		do := operations.NewDualOperator(src, dst)
		if c.IsSet(flagWorkersName) {
			do.WithWorkers(c.Int(flagWorkersName))
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
		do.WithReadPairs(readPairs...)

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
		do.WithWritePairs(writePairs...)

		// parse flag multipart-threshold, 1GB is the default value
		multipartThreshold, err := units.FromHumanSize(c.String(cpFlagMultipartThresholdName))
		if err != nil {
			logger.Error("multipart-threshold is invalid",
				zap.String("input", c.String(cpFlagMultipartThresholdName)),
				zap.Error(err))
			return err
		}

		var ch chan *operations.EmptyResult
		if c.Bool(mvFlagRecursive) {
			ch, err = do.MoveRecursively(srcKey, dstKey, multipartThreshold)
		} else if size < multipartThreshold {
			ch, err = do.MoveFileViaWrite(srcKey, dstKey, size)
		} else {
			ch, err = do.MoveFileViaMultipart(srcKey, dstKey, size)
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
