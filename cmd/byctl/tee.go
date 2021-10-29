package main

import (
	"os"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
)

const (
	teeFlagMultipartThresholdName = "multipart-threshold"
)

var teeFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  teeFlagMultipartThresholdName,
		Usage: "Specify multipart threshold. If source file size is larger than this value, byctl will use multipart method to copy file.",
		EnvVars: []string{
			"BEYOND_CTL_MULTIPART_THRESHOLD",
		},
		Value: "1GiB", // Use 1 GiB as the default value.
	},
}

var teeCmd = &cli.Command{
	Name:      "tee",
	Usage:     "used to read data from standard input and output its contents to a file",
	UsageText: "byctl tee [command options] [target]",
	Flags:     mergeFlags(globalFlags, teeFlags),
	Before: func(c *cli.Context) error {
		logger, _ := zap.NewDevelopment()

		cfg, err := loadConfig(c, true)
		if err != nil {
			logger.Error("load config", zap.Error(err))
			return err
		}

		conn, key, err := cfg.ParseProfileInput(c.Args().Get(0))
		if err != nil {
			logger.Error("parse profile input from src", zap.Error(err))
			return err
		}

		store, err := services.NewStoragerFromString(conn)
		if err != nil {
			logger.Error("init src storager", zap.Error(err), zap.String("conn string", conn))
			return err
		}

		so := operations.NewSingleOperator(store)

		// parse flag multipart-threshold, 1GB is the default value
		multipartThreshold, err := units.FromHumanSize(c.String(cpFlagMultipartThresholdName))
		if err != nil {
			logger.Error("multipart-threshold is invalid",
				zap.String("input", c.String(cpFlagMultipartThresholdName)),
				zap.Error(err))
			return err
		}

		status, err := os.Stdin.Stat()
		if err != nil {
			logger.Error("stat stdin", zap.Error(err))
			return err
		}

		var ch chan *operations.EmptyResult
		if (status.Mode() & os.ModeNamedPipe) != os.ModeNamedPipe {
			ch, err = so.TeeRun(key)
		} else {
			ch, err = so.TeeRunViaPipe(key, multipartThreshold)
		}
		if err != nil {
			logger.Error("run tee", zap.Error(err))
			return err
		}

		for v := range ch {
			if v.Error != nil {
				logger.Error("tee", zap.Error(err))
				return v.Error
			}
		}

		return nil
	},
}
