package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/beyond-ctl/operations"
	"go.beyondstorage.io/v5/pairs"
	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"
)

const (
	lsFlagLongName  = "l"
	lsFlagFormat    = "format"
	lsFlagSummarize = "summarize"
)

var lsFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:  lsFlagLongName,
		Usage: "use a long listing format",
	},
	&cli.StringFlag{
		Name:  lsFlagFormat,
		Usage: "across long -l",
	},
	&cli.BoolFlag{
		Name:  lsFlagSummarize,
		Usage: "display summary information",
	},
}

var lsCmd = &cli.Command{
	Name:  "ls",
	Flags: mergeFlags(globalFlags, lsFlags),
	Action: func(c *cli.Context) (err error) {
		logger, _ := zap.NewDevelopment()

		cfg, err := loadConfig(c, true)
		if err != nil {
			logger.Error("load config", zap.Error(err))
			return err
		}

		// parse args
		var args []string
		if c.Args().Len() == 0 {
			args = append(args, "")
		} else {
			args = c.Args().Slice()
		}

		var storeObjectMap = make(map[*operations.SingleOperator][]*types.Object)
		for _, arg := range args {
			conn, path, err := cfg.ParseProfileInput(arg)
			if err != nil {
				logger.Error("parse profile input", zap.Error(err))
				continue
			}

			store, err := services.NewStoragerFromString(conn)
			if err != nil {
				logger.Error("init storager", zap.Error(err))
				continue
			}

			so := operations.NewSingleOperator(store)

			var objects []*types.Object
			if hasMeta(path) {
				objects, err = so.Glob(path)
				if err != nil {
					logger.Error("glob", zap.Error(err))
					fmt.Printf("ls: cannot access '%s': No such file or directory\n", path)
					continue
				}
				storeObjectMap[so] = objects
			} else {
				var o *types.Object
				if path == "" {
					o = store.Create(path, pairs.WithObjectMode(types.ModeDir))
				} else {
					o, err = so.Stat(path)
					if err != nil {
						logger.Error("stat", zap.Error(err))
						fmt.Printf("stat: cannot access '%s': No such file or directory\n", path)
						continue
					}
				}

				storeObjectMap[so] = append(storeObjectMap[so], o)
			}
		}

		format := shortListFormat
		if c.Bool("l") || c.String("format") == "long" {
			format = longListFormat
		}

		isFirstSrc := true
		for so, objects := range storeObjectMap {
			for _, o := range objects {
				// print src path if more than 1 arg
				if len(storeObjectMap) > 1 || len(objects) > 1 {
					if isFirstSrc {
						isFirstSrc = false
					} else {
						fmt.Printf("\n")
					}
					//so.StatStorager().Service + ":" + path
					fmt.Printf("%s:\n", o.Path)
				}

				if o.Mode.IsDir() {
					ch, err := so.List(o.Path)
					if err != nil {
						logger.Error("list",
							zap.String("path", o.Path),
							zap.Error(err))
						continue
					}

					isFirst := true
					var totalNum int
					var totalSize int64

					for v := range ch {
						if v.Error != nil {
							logger.Error("read next result", zap.Error(v.Error))
							break
						}

						oa := parseObject(v.Object)
						fmt.Print(oa.Format(format, isFirst))

						// Update isFirst
						if isFirst {
							isFirst = false
						}

						totalNum += 1
						totalSize += oa.size
					}
					// End of line
					fmt.Print("\n")

					// display summary information
					if c.Bool(lsFlagSummarize) {
						fmt.Printf("\n%14s %d\n", "Total Objects:", totalNum)
						fmt.Printf("%14s %s\n", "Total Size:", units.BytesSize(float64(totalSize)))
					}
				} else {
					oa := parseObject(o)
					fmt.Print(oa.Format(format, true))
					// End of line
					fmt.Print("\n")
				}
			}
		}
		return
	},
}

const (
	shortListFormat = iota
	longListFormat
)

type objectAttr struct {
	mode      types.ObjectMode
	name      string
	size      int64
	updatedAt time.Time
}

func (oa objectAttr) Format(layout int, isFirst bool) string {
	switch layout {
	case shortListFormat:
		return oa.shortFormat(isFirst)
	case longListFormat:
		return oa.longFormat(isFirst)
	default:
		panic("not supported format")
	}
}

func (oa objectAttr) shortFormat(isFirst bool) string {
	if isFirst {
		return oa.name
	}
	return " " + oa.name
}

func (oa objectAttr) longFormat(isFirst bool) string {
	buf := pool.Get()
	defer buf.Free()

	// If not the first entry, we need to print a new line.
	if !isFirst {
		buf.AppendString("\n")
	}

	if oa.mode.IsRead() {
		buf.AppendString("read")
	} else if oa.mode.IsDir() {
		// Keep align with read.
		buf.AppendString("dir ")
	} else if oa.mode.IsPart() {
		buf.AppendString("part")
	}
	// FIXME: it's hard to calculate the padding, so we hardcoded the padding here.
	buf.AppendString(fmt.Sprintf("%12d", oa.size))
	buf.AppendString(" ")
	// gnuls will print year instead if not the same year.
	if time.Now().Year() != oa.updatedAt.Year() {
		buf.AppendString(oa.updatedAt.Format("Jan 02  2006"))
	} else {
		buf.AppendString(oa.updatedAt.Format("Jan 02 15:04"))
	}
	buf.AppendString(" ")
	buf.AppendString(oa.name)

	return buf.String()
}

func parseObject(o *types.Object) objectAttr {
	oa := objectAttr{
		name: filepath.Base(o.Path),
	}

	if v, ok := o.GetContentLength(); ok {
		oa.size = v
	}

	if v, ok := o.GetLastModified(); ok {
		oa.updatedAt = v
	}

	// Mode could be updated after lazy stat.
	oa.mode = o.Mode
	return oa
}
