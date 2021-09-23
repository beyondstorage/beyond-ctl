package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

var (
	// global flags will be applied to all operations.
	globalFlags = []cli.Flag{
		flagConfig,
		flagWorkers,
	}
	// IO flags will be applied to all operations that will have read or write IO
	// operations
	ioFlags = []cli.Flag{
		flagReadSpeedLimit,
		flagWriteSpeedLimit,
	}
)

const (
	flagConfigName          = "config"
	flagWorkersName         = "workers"
	flagReadSpeedLimitName  = "read-speed-limit"
	flagWriteSpeedLimitName = "write-speed-limit"
)

var (
	flagConfig = &cli.StringFlag{
		Name:    flagConfigName,
		Usage:   "Load config from `FILE`",
		Aliases: []string{"c"},
		EnvVars: []string{
			"BEYOND_CTL_CONFIG",
		},
		Value: fmt.Sprintf("%s/byctl/config.toml", userConfigDir()),
	}
	flagWorkers = &cli.IntFlag{
		Name:  flagWorkersName,
		Usage: "Specify the workers number",
		EnvVars: []string{
			"BEYOND_CTL_WORKERS",
		},
		Value: 4,
	}
	flagReadSpeedLimit = &cli.StringFlag{
		Name:  flagReadSpeedLimitName,
		Usage: "Specify speed limit for read I/O operations, for example, 1MB, 10mb, 3GiB.",
		EnvVars: []string{
			"BEYOND_CTL_READ_SPEED_LIMIT",
		},
	}
	flagWriteSpeedLimit = &cli.StringFlag{
		Name:  flagWriteSpeedLimitName,
		Usage: "Specify speed limit for write I/O operations, for example, 1MB, 10mb, 3GiB.",
		EnvVars: []string{
			"BEYOND_CTL_WRITE_SPEED_LIMIT",
		},
	}
)

func mergeFlags(fs ...[]cli.Flag) []cli.Flag {
	var x []cli.Flag
	for _, v := range fs {
		x = append(x, v...)
	}
	return x
}
