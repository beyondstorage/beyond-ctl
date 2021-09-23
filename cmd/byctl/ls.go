package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/beyondstorage/beyond-ctl/operations"
)

const (
	lsFlagLongName = "l"
	lsFlagFormat   = "format"
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

		conn, path, err := cfg.ParseProfileInput(c.Args().Get(0))
		if err != nil {
			logger.Error("parse profile input", zap.Error(err))
			return err
		}

		store, err := services.NewStoragerFromString(conn)
		if err != nil {
			logger.Error("init storager", zap.Error(err))
			return err
		}

		so := operations.NewSingleOperator(store)

		// TODO: we need support more format that gnsls supports.
		format := shortListFormat
		if c.Bool("l") || c.String("format") == "long" {
			format = longListFormat
		}

		ch, err := so.List(path)
		if err != nil {
			logger.Error("list",
				zap.String("path", path),
				zap.Error(err))
			return err
		}

		isFirst := true

		for v := range ch {
			if v.Error != nil {
				logger.Error("read next result", zap.Error(v.Error))
				return v.Error
			}

			oa := parseObject(v.Object)
			fmt.Print(oa.Format(format, isFirst))

			// Update isFirst
			if isFirst {
				isFirst = false
			}
		}
		// End of line
		fmt.Print("\n")
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
	return oa.name + " "
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
