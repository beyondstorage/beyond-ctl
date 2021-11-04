package main

import (
	"fmt"
	"strconv"
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

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
	PETABYTE
	EXABYTE
)

// ByteSize returns a human-readable byte string of the form 10M, 12.5K, and so forth. The following units are available:
//  EiB: Exabyte
//  PiB: Petabyte
//  TiB: Terabyte
//  GiB: Gigabyte
//  MiB: Megabyte
//  KiB: Kilobyte
//  B: Byte
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
func ByteSize(bytes uint64) string {
	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= EXABYTE:
		unit = "EiB"
		value = value / EXABYTE
	case bytes >= PETABYTE:
		unit = "PiB"
		value = value / PETABYTE
	case bytes >= TERABYTE:
		unit = "TiB"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "GiB"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "MiB"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "KiB"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	result := strconv.FormatFloat(value, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + " " + unit
}
