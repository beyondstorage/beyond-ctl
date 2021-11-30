package main

import (
	"errors"
	"fmt"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"path/filepath"

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
		Usage: "Specify multipart threshold. If source file size is larger than this value, byctl will use multipart method to move file.",
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
			return fmt.Errorf("mv command wants at least two args, but got %d", args)
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
		multipartThreshold, err := units.RAMInBytes(c.String(mvFlagMultipartThresholdName))
		if err != nil {
			logger.Error("multipart-threshold is invalid",
				zap.String("input", c.String(mvFlagMultipartThresholdName)),
				zap.Error(err))
			return err
		}

		// parse src args
		srcNum := 0
		var storeObjectMap = make(map[types.Storager][]*types.Object)
		for i := 0; i < c.Args().Len()-1; i++ {
			arg := c.Args().Get(i)
			conn, key, err := cfg.ParseProfileInput(arg)
			if err != nil {
				logger.Error("parse profile input from src", zap.Error(err))
				continue
			}

			store, err := services.NewStoragerFromString(conn)
			if err != nil {
				logger.Error("init src storager", zap.Error(err), zap.String("conn string", conn))
				continue
			}

			so := operations.NewSingleOperator(store)

			if hasMeta(key) {
				objects, err := so.Glob(key)
				if err != nil {
					logger.Error("glob", zap.Error(err), zap.String("path", arg))
					continue
				}
				for _, o := range objects {
					if o.Mode.IsDir() && !c.Bool(mvFlagRecursive) {
						// so.StatStorager().Service + ":" + o.Path
						fmt.Printf("mv: -r not specified; omitting directory '%s'\n", o.Path)
						continue
					}
					storeObjectMap[store] = append(storeObjectMap[store], o)
					srcNum++
				}
			} else {
				o, err := so.Stat(key)
				if err != nil && !errors.Is(err, services.ErrObjectNotExist) {
					if errors.Is(err, services.ErrObjectNotExist) {
						fmt.Printf("mv: cannot stat '%s': No such file or directory\n", arg)
					} else {
						logger.Error("stat", zap.Error(err), zap.String("path", arg))
					}
					continue
				}
				if err == nil {
					if o.Mode.IsDir() && !c.Bool(mvFlagRecursive) {
						fmt.Printf("mv: -r not specified; omitting directory '%s'\n", arg)
						continue
					} else if o.Mode.IsPart() {
						fmt.Printf("mv: cannot move '%s': Is an in progress multipart upload task\n", arg)
						continue
					}
				}

				err = nil
				storeObjectMap[store] = append(storeObjectMap[store], o)
				srcNum++
			}
		}

		// check dst
		dstConn, dstKey, err := cfg.ParseProfileInput(c.Args().Get(c.Args().Len() - 1))
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
		if dstObject != nil {
			if dstObject.Mode.IsPart() {
				fmt.Printf("mv: target '%s' is an in progress multipart upload task\n", dstKey)
				return fmt.Errorf("mv: target '%s' is an in progress multipart upload task", dstKey)
			}
			if srcNum > 1 && !dstObject.Mode.IsDir() {
				fmt.Printf("mv: target '%s' is not a directory\n", dstKey)
				return fmt.Errorf("mv: target '%s' is not a directory", dstKey)
			}
		}

		for store, objects := range storeObjectMap {
			for _, o := range objects {
				var size int64
				if o.Mode.IsRead() {
					n, ok := o.GetContentLength()
					if !ok {
						logger.Error("can't get object content length", zap.String("path", o.Path))
						continue
					}
					size = n
				}

				do := operations.NewDualOperator(store, dst)
				if c.IsSet(flagWorkersName) {
					do.WithWorkers(c.Int(flagWorkersName))
				}

				// set read pairs
				do.WithReadPairs(readPairs...)
				// set write pairs
				do.WithWritePairs(writePairs...)

				realDstKey := dstKey
				if srcNum > 1 || (dstObject != nil && dstObject.Mode.IsDir()) {
					realDstKey = filepath.Join(dstKey, filepath.Base(o.Path))
				}

				if c.Bool(mvFlagRecursive) && o.Mode.IsDir() {
					err = do.MoveRecursively(o.Path, realDstKey, multipartThreshold)
				} else if size < multipartThreshold {
					err = do.MoveFileViaWrite(o.Path, realDstKey, size)
				} else {
					err = do.MoveFileViaMultipart(o.Path, realDstKey, size)
				}
				if err != nil {
					logger.Error("start move",
						zap.String("src", o.Path),
						zap.String("dst", realDstKey),
						zap.Error(err))
					continue
				}
			}
		}
		return
	},
}
