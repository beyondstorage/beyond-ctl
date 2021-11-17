package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/Xuanwo/go-bufferpool"
	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
	"golang.org/x/time/rate"

	"go.beyondstorage.io/beyond-ctl/config"
	"go.beyondstorage.io/v5/pairs"
	"go.beyondstorage.io/v5/types"
)

var pool = bufferpool.New(128)

func loadConfig(c *cli.Context, loadEnv bool) (*config.Config, error) {
	path := c.String(flagConfigName)
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("load config %s: %w", path, err)
	}
	if loadEnv {
		cfg.MergeProfileFromEnv()
	}
	return cfg, nil
}

func parseLimit(text string) (types.Pair, error) {
	limit, err := units.FromHumanSize(text)
	if err != nil {
		return types.Pair{}, err
	}

	limter := rate.NewLimiter(rate.Limit(limit), int(limit))

	return pairs.WithIoCallback(func(bs []byte) {
		l := len(bs)

		for l > 0 {
			n := limter.Burst()
			if n > l {
				n = l
			}
			r := limter.ReserveN(time.Now(), n)
			time.Sleep(r.Delay())
			l -= n
		}
	}), nil
}

func hasMeta(path string) bool {
	magicChars := `*?[{`
	if runtime.GOOS != "windows" {
		magicChars = `*?[{\`
	}
	return strings.ContainsAny(path, magicChars)
}
