package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
)

const (
	signFlagExpire = "expire"
)

var signFlags = []cli.Flag{
	&cli.IntFlag{
		Name:  signFlagExpire,
		Usage: "the number of seconds until the signed URL expires",
		Value: 300,
	},
}

var signCmd = &cli.Command{
	Name:      "sign",
	Usage:     "get the signed URL by the source",
	UsageText: "byctl sign [command options] [source]",
	Flags:     mergeFlags(globalFlags, signFlags),
	Before: func(c *cli.Context) error {
		if args := c.Args().Len(); args < 1 {
			return fmt.Errorf("sign command wants at least one args, but got %d", args)
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

		isFirst := true
		args := c.Args().Len()
		for i := 0; i < args; i++ {
			conn, key, err := cfg.ParseProfileInput(c.Args().Get(i))
			if err != nil {
				logger.Error("parse profile input from source", zap.Error(err))
				continue
			}

			store, err := services.NewStoragerFromString(conn)
			if err != nil {
				logger.Error("init source storager", zap.Error(err), zap.String("conn string", conn))
				continue
			}

			so := operations.NewSingleOperator(store)

			// The default is 300 second.
			second := c.Int(signFlagExpire)
			expire := time.Duration(second) * time.Second

			url, err := so.Sign(key, expire)
			if err != nil {
				logger.Error("run sign", zap.Error(err))
				continue
			}

			if args > 1 {
				if isFirst {
					isFirst = false
				} else {
					fmt.Printf("\n")
				}
				fmt.Printf("%s:\n", c.Args().Get(i))
			}
			fmt.Println(url)
		}

		return nil
	},
}
