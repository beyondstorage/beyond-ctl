package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/beyondstorage/beyond-ctl/config"
)

func loadConfig(c *cli.Context, loadEnv bool) (*config.Config, error) {
	path := c.String(mainFlagConfig)
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("load config %s: %w", path, err)
	}
	if loadEnv {
		cfg.MergeProfileFromEnv()
	}
	return cfg, nil
}
