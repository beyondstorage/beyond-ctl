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
	syncFlagExisting           = "existing"
	syncFlagIgnoreExisting     = "ignore-existing"
	syncFlagRecursive          = "recursive"
	syncFlagUpdate             = "update"
	syncFlagRemove             = "remove"
	syncFlagExclude            = "exclude"
	syncFlagInclude            = "include"
	syncFlagMultipartThreshold = "multipart-threshold"
)

var syncFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:  syncFlagExisting,
		Usage: "skip creating new files in target dirs",
	},
	&cli.BoolFlag{
		Name:  syncFlagIgnoreExisting,
		Usage: "skip updating files in target dirs, only copy those not exist",
	},
	&cli.BoolFlag{
		Name: syncFlagRecursive,
		Aliases: []string{
			"r",
			"R",
		},
		Usage: "recurse into sub directories",
	},
	&cli.BoolFlag{
		Name:  syncFlagUpdate,
		Usage: "skip files that are newer in target dirs",
	},
	&cli.BoolFlag{
		Name:  syncFlagRemove,
		Usage: "remove extraneous object(s) on target",
	},
	&cli.StringFlag{
		Name:  syncFlagExclude,
		Usage: "regular expression for files to exclude",
	},
	&cli.StringFlag{
		Name:  syncFlagInclude,
		Usage: "regular expression for files to include (not work if exclude not set)",
	},
	&cli.StringFlag{
		Name:  syncFlagMultipartThreshold,
		Usage: "Specify multipart threshold. If source file size is larger than this value, byctl will use multipart method to sync file.",
		EnvVars: []string{
			"BEYOND_CTL_MULTIPART_THRESHOLD",
		},
		Value: "1GiB",
	},
}

var syncCmd = &cli.Command{
	Name:      "sync",
	Usage:     "sync file from source storager to target storager",
	UsageText: "byctl sync [command options] [source] [target]",
	Flags:     mergeFlags(globalFlags, ioFlags, syncFlags),
	Before: func(c *cli.Context) error {
		if args := c.Args().Len(); args < 2 {
			return fmt.Errorf("sync command wants at least two args, but got %d", args)
		}
		return nil
	},
	Action: func(c *cli.Context) error {
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

		if !strings.HasSuffix(dstKey, "/") {
			logger.Error("target is not a directory", zap.String("target", dstKey))
			return fmt.Errorf("target is not a directory")
		}

		dst, err := services.NewStoragerFromString(dstConn)
		if err != nil {
			logger.Error("init dst storager", zap.Error(err), zap.String("conn string", dstConn))
			return err
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
		multipartThreshold, err := units.RAMInBytes(c.String(syncFlagMultipartThreshold))
		if err != nil {
			logger.Error("multipart-threshold is invalid",
				zap.String("input", c.String(syncFlagMultipartThreshold)),
				zap.Error(err))
		}

		// Initialization of `sync` options.
		opts := operations.SyncOptions{
			MultipartThreshold: multipartThreshold,
			Recursive:          c.Bool(syncFlagRecursive),
			Existing:           c.Bool(syncFlagExisting),
			IgnoreExisting:     c.Bool(syncFlagIgnoreExisting),
			Update:             c.Bool(syncFlagUpdate),
			Remove:             c.Bool(syncFlagRemove),
			IsExclude:          c.IsSet(syncFlagExclude),
			Exclude:            c.String(syncFlagExclude),
			IsInclude:          c.IsSet(syncFlagInclude),
			Include:            c.String(syncFlagInclude),
			IsArgs:             c.Args().Len() > 2,
		}

		for i := 0; i < argsNum-1; i++ {
			srcConn, srcKey, err := cfg.ParseProfileInput(c.Args().Get(i))
			if err != nil {
				logger.Error("parse profile input from src", zap.Error(err))
				continue
			}

			if !strings.HasSuffix(srcKey, "/") {
				logger.Error("source is not a directory", zap.String("source", dstKey))
				return fmt.Errorf("source is not a directory")
			}

			src, err := services.NewStoragerFromString(srcConn)
			if err != nil {
				logger.Error("init src storager", zap.Error(err), zap.String("conn string", srcConn))
				continue
			}

			so := operations.NewSingleOperator(src)

			_, err = so.Stat(srcKey)
			if err != nil {
				logger.Error("stat", zap.String("path", srcKey), zap.Error(err))
				return err
			}

			do := operations.NewDualOperator(src, dst)
			if c.IsSet(flagWorkersName) {
				do.WithWorkers(c.Int(flagWorkersName))
			}

			do.WithReadPairs(readPairs...)
			do.WithWritePairs(writePairs...)

			ch, err := do.SyncDir(srcKey, dstKey, opts)
			if err != nil {
				logger.Error("sync", zap.Error(err))
				return err
			}

			for v := range ch {
				if v.Error != nil {
					logger.Error("run sync", zap.Error(v.Error))
					return v.Error
				}
			}
		}

		return nil
	},
}
