package main

import (
	"errors"
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
			return fmt.Errorf("cat command wants one args, but got %d", args)
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

		var storePathMap = make(map[*operations.SingleOperator][]string)
		for i := 0; i < c.Args().Len(); i++ {
			arg := c.Args().Get(i)
			conn, key, err := cfg.ParseProfileInput(arg)
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

			if hasMeta(key) {
				objects, err := so.Glob(key)
				if err != nil {
					logger.Error("glob", zap.Error(err), zap.String("path", arg))
					continue
				}
				for _, o := range objects {
					if o.Mode.IsDir() {
						// so.StatStorager().Service + ":" + o.Path
						fmt.Printf("cat: '%s': Is a directory\n", o.Path)
						continue
					}
					storePathMap[so] = append(storePathMap[so], o.Path)
				}
			} else {
				o, err := so.Stat(key)
				if err != nil {
					if errors.Is(err, services.ErrObjectNotExist) {
						fmt.Printf("cat: '%s': No such file or directory\n", arg)
					} else {
						logger.Error("stat", zap.Error(err), zap.String("path", arg))
					}
					continue
				} else {
					if o.Mode.IsDir() {
						fmt.Printf("cat: '%s': Is a directory\n", arg)
						continue
					} else if o.Mode.IsPart() {
						fmt.Printf("cat: '%s': Is an in progress multipart upload task\n", arg)
						continue
					}
				}
				err = nil
				storePathMap[so] = append(storePathMap[so], key)
			}
		}

		for so, paths := range storePathMap {
			for _, path := range paths {
				ch, err := so.CatFile(path)
				if err != nil {
					logger.Error("run cat", zap.Error(err))
					continue
				}

				for v := range ch {
					if v.Error != nil {
						logger.Error("cat", zap.Error(err))
						continue
					}
				}

				fmt.Printf("\n")
			}
		}

		return nil
	},
}
