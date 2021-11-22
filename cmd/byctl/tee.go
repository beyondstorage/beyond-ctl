package main

import (
	"bytes"
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
			return fmt.Errorf("tee command wants at least one args, but got %d", args)
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

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(c.App.Reader)
		if err != nil {
			logger.Error("read data", zap.Error(err))
			return err
		}

		for i := 0; i < c.Args().Len(); i++ {
			conn, key, err := cfg.ParseProfileInput(c.Args().Get(i))
			if err != nil {
				logger.Error("parse profile input from target", zap.Error(err))
				continue
			}

			store, err := services.NewStoragerFromString(conn)
			if err != nil {
				logger.Error("init target storager", zap.Error(err), zap.String("conn string", conn))
				continue
			}

			so := operations.NewSingleOperator(store)

			expectedSize, err := units.RAMInBytes(c.String(teeFlagExpectSize))
			if err != nil {
				logger.Error("expected-size is invalid", zap.String("input", c.String(teeFlagExpectSize)), zap.Error(err))
				continue
			}

			ch, err := so.TeeRun(key, expectedSize, bytes.NewReader(buf.Bytes()))
			if err != nil {
				logger.Error("run tee", zap.Error(err))
				continue
			}

			for v := range ch {
				if v.Error != nil {
					logger.Error("tee", zap.Error(err))
					continue
				}
			}

			fmt.Printf("Stdin is saved to <%s>\n", key)
		}

		return nil
	},
}
