package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"
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

		var storeObjectMap = make(map[*operations.SingleOperator][]*types.Object)
		for i := 0; i < c.Args().Len(); i++ {
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
					if o.Mode.IsDir() && !c.Bool(rmFlagRecursive) {
						// so.StatStorager().Service + ":" + o.Path
						fmt.Printf("rm: cannot remove '%s': Is a directory\n", o.Path)
						continue
					}
					storeObjectMap[so] = append(storeObjectMap[so], o)
				}
			} else {
				o, err := so.Stat(key)
				if err != nil && !errors.Is(err, services.ErrObjectNotExist) {
					if errors.Is(err, services.ErrObjectNotExist) {
						fmt.Printf("rm: cannot remove '%s': No such file or directory\n", arg)
					} else {
						logger.Error("stat", zap.Error(err), zap.String("path", arg))
					}
					continue
				}
				if err == nil {
					if o.Mode.IsDir() && !c.Bool(rmFlagRecursive) {
						fmt.Printf("rm: cannot remove '%s': Is a directory\n", arg)
						continue
					} else if o.Mode.IsPart() && !c.Bool(rmFlagMultipart) {
						fmt.Printf("rm: cannot remove '%s': Is an in progress multipart upload task\n", arg)
						continue
					}
				}

				err = nil
				storeObjectMap[so] = append(storeObjectMap[so], o)
			}
		}

		for so, objects := range storeObjectMap {
			for _, o := range objects {
				if o.Mode.IsDir() {
					// recursive remove a dir.
					ch, err := so.DeleteRecursively(o.Path)
					if err != nil {
						logger.Error("delete recursively",
							zap.String("path", o.Path),
							zap.Error(err))
						continue
					}

					if ch != nil {
						for v := range ch {
							if v.Error != nil {
								logger.Error("delete", zap.Error(err))
								continue
							}
						}
					}
				} else if o.Mode.IsPart() {
					if !c.Bool(rmFlagRecursive) {
						// Remove all multipart objects whose path is `key`
						ch, err := so.DeleteMultipart(o.Path)
						if err != nil {
							logger.Error("delete multipart",
								zap.String("path", o.Path),
								zap.Error(err))
							continue
						}

						if ch != nil {
							for v := range ch {
								if v.Error != nil {
									logger.Error("delete", zap.Error(err))
									continue
								}
							}
						}
					} else {
						// Remove all multipart objects prefixed with `key`.
						ch, err := so.DeleteMultipartViaRecursively(o.Path)
						if err != nil {
							logger.Error("delete multipart recursively",
								zap.String("path", o.Path),
								zap.Error(err))
							continue
						}

						if ch != nil {
							for v := range ch {
								if v.Error != nil {
									logger.Error("delete", zap.Error(err))
									continue
								}
							}
						}
					}
				} else {
					// remove single file
					err = so.Delete(o.Path)
					if err != nil {
						logger.Error("delete", zap.String("path", o.Path), zap.Error(err))
						continue
					}
				}
			}
		}
		return nil
	},
}
