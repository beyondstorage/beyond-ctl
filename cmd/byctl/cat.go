package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
)

var catCmd = &cli.Command{
	Name:      "cat",
	Usage:     "pipe data from storage services into stdout",
	UsageText: "byctl cat [command options] [source]",
	Flags:     mergeFlags(globalFlags),
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

		ch, err := so.CatFile(key)
		if err != nil {
			logger.Error("run cat", zap.Error(err))
			return err
		}

		for v := range ch {
			if v.Error != nil {
				logger.Error("cat", zap.Error(err))
				return v.Error
			}
		}

		return nil
	},
}
