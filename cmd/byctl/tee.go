package main

import (
	"fmt"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
)

const (
	teeFlagExpectSize = "expected-size"
)

var teeFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  teeFlagExpectSize,
		Usage: "expected size of the input file",
		Value: "128MiB",
	},
}

var teeCmd = &cli.Command{
	Name:      "tee",
	Usage:     "used to read data from standard input and output its contents to a file",
	UsageText: "byctl tee [command options] [target]",
	Flags:     mergeFlags(globalFlags, teeFlags),
	Before: func(c *cli.Context) error {
		if args := c.Args().Len(); args < 1 {
			return fmt.Errorf("tee command wants one args, but got %d", args)
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

		conn, key, err := cfg.ParseProfileInput(c.Args().Get(0))
		if err != nil {
			logger.Error("parse profile input from target", zap.Error(err))
			return err
		}

		store, err := services.NewStoragerFromString(conn)
		if err != nil {
			logger.Error("init target storager", zap.Error(err), zap.String("conn string", conn))
			return err
		}

		so := operations.NewSingleOperator(store)

		expectedSize, err := units.RAMInBytes(c.String(teeFlagExpectSize))
		if err != nil {
			logger.Error("expected-size is invalid", zap.String("input", c.String(teeFlagExpectSize)), zap.Error(err))
			return err
		}

		ch, err := so.TeeRun(key, expectedSize, c.App.Reader)
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

		fmt.Printf("Stdin is saved to <%s>\n", key)

		return nil
	},
}
