package cmd

import (
	"fmt"

	"github.com/urfave/cli"
)

var pushCommand = &cli.Command{
	Name:  "push",
	Usage: "identify changes to be made, and perform them",
	Action: func(ctx *cli.Context) error {
		return exit(Push(globalPushArgs))
	},
	Category: catMain,
	Flags:    globalPushArgs.flags(),
}

type PushArgs struct {
	GetDNSConfigArgs
}

func (args *PushArgs) flags() []cli.Flag {
	flags := globalPreviewArgs.GetDNSConfigArgs.flags()
	return flags
}

var globalPushArgs PushArgs

func Push(args PushArgs) error {
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	fmt.Println(len(cfg.Domains))
	return nil
}
