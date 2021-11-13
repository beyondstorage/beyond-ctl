package main

import (
	"fmt"
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

type statsMonitor struct {
	statsList []bool
}

func (monitor *statsMonitor) newStatMonitor(num int) {
	for i := 0; i < num; i++ {
		monitor.statsList = append(monitor.statsList, false)
	}
}

func (monitor *statsMonitor) setStat(n int) {
	if n >= 0 && n < len(monitor.statsList) {
		monitor.statsList[n] = true
	}
}

func (monitor *statsMonitor) isDone() bool {
	done := true
	for i := 0; i < len(monitor.statsList); i++ {
		if !monitor.statsList[i] {
			done = false
		}
	}
	return done
}
