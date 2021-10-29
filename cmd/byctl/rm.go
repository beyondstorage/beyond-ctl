package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
)

const (
	rmFlagRecursive = "recursive"
	rmFlagMultipart = "multipart"
)

var rmFlags = []cli.Flag{
	&cli.BoolFlag{
		Name: rmFlagRecursive,
		Aliases: []string{
			"r",
			"R",
		},
		Usage: "remove directories and their contents recursively\n",
	},
	&cli.BoolFlag{
		Name:  rmFlagMultipart,
		Usage: "remove multipart object",
	},
}

var rmCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove file from storager",
	UsageText: "byctl rm [command options] [source]",
	Flags:     mergeFlags(globalFlags, rmFlags),
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

		var ch chan *operations.EmptyResult
		if c.Bool(rmFlagMultipart) && !c.Bool(rmFlagRecursive) {
			// Remove all multipart objects whose path is `key`
			ch, err = so.DeleteMultipart(key)
			if err != nil {
				logger.Error("delete multipart",
					zap.String("path", key),
					zap.Error(err))
				return err
			}
		} else if c.Bool(rmFlagMultipart) && c.Bool(rmFlagRecursive) {
			// Remove all multipart objects prefixed with `key`.
			ch, err = so.DeleteMultipartViaRecursively(key)
			if err != nil {
				logger.Error("delete multipart recursively",
					zap.String("path", key),
					zap.Error(err))
				return err
			}
		} else if !c.Bool(rmFlagMultipart) && c.Bool(rmFlagRecursive) {
			// recursive remove a dir.
			ch, err = so.DeleteRecursively(key)
			if err != nil {
				logger.Error("delete recursively",
					zap.String("path", key),
					zap.Error(err))
				return err
			}
		} else {
			// remove single file
			_, err = so.Stat(key)
			if err != nil && errors.Is(err, services.ErrObjectNotExist) {
				fmt.Printf("rm: cannot remove '%s': No such file or directory\n", key)
				return nil
			}
			if err != nil {
				logger.Error("stat", zap.String("path", key), zap.Error(err))
				return err
			}

			err = so.Delete(key)
			if err != nil {
				logger.Error("delete", zap.String("path", key), zap.Error(err))
				return err
			}
		}

		for v := range ch {
			if v.Error != nil {
				logger.Error("delete", zap.Error(err))
				return v.Error
			}
		}
		return nil
	},
}
