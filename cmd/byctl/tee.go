package main

import (
	"fmt"
	"os"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
)

const (
	teeFlagExpectSize = "expect-size"
)

var teeFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  teeFlagExpectSize,
		Usage: "",
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
			return fmt.Errorf("rm command wants one args, but got %d", args)
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
			logger.Error("parse profile input from src", zap.Error(err))
			return err
		}

		store, err := services.NewStoragerFromString(conn)
		if err != nil {
			logger.Error("init src storager", zap.Error(err), zap.String("conn string", conn))
			return err
		}

		so := operations.NewSingleOperator(store)

		expectSize, err := units.RAMInBytes(c.String(teeFlagExpectSize))
		if err != nil {
			logger.Error("expect-size is invalid", zap.String("input", c.String(teeFlagExpectSize)), zap.Error(err))
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
			err = so.TeeRunViaPipe(key, expectSize)
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
