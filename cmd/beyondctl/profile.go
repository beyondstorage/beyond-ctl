package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v2"

	"github.com/beyondstorage/beyond-ctl/config"
)

var profileCmd = &cli.Command{
	Name:  "profile",
	Usage: "manage profile in config file",
	Subcommands: []*cli.Command{
		profileAddCmd,
		profileListCmd,
		profileRemoveCmd,
	},
}

var profileAddCmd = &cli.Command{
	Name:  "add",
	Usage: "add profile [name] [connection_string]",
	Before: func(ctx *cli.Context) error {
		if args := ctx.Args().Len(); args < 2 {
			return fmt.Errorf("add command wants two args, but got %d", args)
		}
		return nil
	},
	Action: func(c *cli.Context) error {
		cfg, err := loadConfig(c, false)
		if err != nil {
			return err
		}

		name, connStr := c.Args().Get(0), c.Args().Get(1)
		err = cfg.AddProfile(name, config.Profile{
			Connection: connStr,
		})
		if err != nil {
			return err
		}

		if err := cfg.WriteToFile(c.String(globalFlagConfig)); err != nil {
			return err
		}
		return nil
	},
}

var profileRemoveCmd = &cli.Command{
	Name:  "remove",
	Usage: "remove profile [name]",
	Before: func(ctx *cli.Context) error {
		if args := ctx.Args().Len(); args < 1 {
			return fmt.Errorf("remove command wants one arg at least, but got %d", args)
		}
		return nil
	},
	Action: func(c *cli.Context) error {
		cfg, err := loadConfig(c, false)
		if err != nil {
			return err
		}

		cfg.RemoveProfile(c.Args().First())

		if err := cfg.WriteToFile(c.String(globalFlagConfig)); err != nil {
			return err
		}
		return nil
	},
}

var profileListCmd = &cli.Command{
	Name:  "list",
	Usage: "list profiles",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "json",
			Usage: "list profile from config",
			Value: false,
		},
	},
	Action: func(c *cli.Context) error {
		cfg, err := loadConfig(c, false)
		if err != nil {
			return err
		}

		if c.Bool("json") {
			err = json.NewEncoder(os.Stdout).Encode(cfg.Profiles)
		} else {
			err = toml.NewEncoder(os.Stdout).Encode(cfg.Profiles)
		}

		return err
	},
}
