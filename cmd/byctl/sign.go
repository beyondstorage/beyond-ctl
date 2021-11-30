package main

import (
	"errors"
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
						fmt.Printf("sign: '%s': Is a directory\n", o.Path)
						continue
					}
					storePathMap[so] = append(storePathMap[so], o.Path)
				}
			} else {
				o, err := so.Stat(key)
				if err != nil {
					if errors.Is(err, services.ErrObjectNotExist) {
						fmt.Printf("sign: '%s': No such file or directory\n", arg)
					} else {
						logger.Error("stat", zap.Error(err), zap.String("path", arg))
					}
					continue
				} else {
					if o.Mode.IsDir() {
						fmt.Printf("sign: '%s': Is a directory\n", arg)
						continue
					} else if o.Mode.IsPart() {
						fmt.Printf("sign: '%s': Is an in progress multipart upload task\n", arg)
						continue
					}
				}
				err = nil
				storePathMap[so] = append(storePathMap[so], key)
			}
		}

		// The default is 300 second.
		second := c.Int(signFlagExpire)
		expire := time.Duration(second) * time.Second

		isFirst := true
		for so, paths := range storePathMap {
			for _, path := range paths {
				url, err := so.Sign(path, expire)
				if err != nil {
					logger.Error("run sign", zap.Error(err))
					continue
				}

				if len(paths) > 1 {
					if isFirst {
						isFirst = false
					} else {
						fmt.Printf("\n")
					}
					// so.StatStorager().Service + ":" + o.Path
					fmt.Printf("%s:\n", path)
				}
				fmt.Println(url)
			}
		}
		return nil
	},
}
