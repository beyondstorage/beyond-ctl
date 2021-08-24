package main

import (
	"fmt"

	"github.com/beyondstorage/go-storage/v4/services"
	"github.com/beyondstorage/go-storage/v4/types"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/beyondstorage/beyond-ctl/operations"
)

var lsCmd = &cli.Command{
	Name: "ls",
	Action: func(context *cli.Context) (err error) {
		logger, _ := zap.NewDevelopment()

		store, err := services.NewStoragerFromString(context.Args().First())
		if err != nil {
			return err
		}

		oo, err := operations.NewSingleOperator(store)
		if err != nil {
			return err
		}

		go func() {
			for v := range oo.Errors() {
				logger.Error("", zap.Error(v))
			}
		}()

		for v := range oo.List("") {
			fmt.Print(parseToShell(v))
		}
		// End of line
		fmt.Print("\n")
		return
	},
}

func parseToShell(o *types.Object) (content string) {
	buf := pool.Get()
	defer buf.Free()

	buf.AppendString(o.Path)
	buf.AppendString(" ")
	return string(buf.Bytes())
}
