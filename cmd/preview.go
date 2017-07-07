package cmd

import (
	"github.com/urfave/cli"
	"fmt"
)

var previewCommand = &cli.Command{
	Name:  "preview",
	Usage: "read live configuration and identify changes to be made, without applying them",
	Action: func(ctx *cli.Context) error {
		return exit(Preview(globalPreviewArgs))
	},
	Category: catMain,
	Flags:    globalPreviewArgs.flags(),
}

// PreviewArgs contains all data/flags needed to run preview, independently of CLI
type PreviewArgs struct {
	GetDNSConfigArgs
}

func (args *PreviewArgs) flags() []cli.Flag {
	flags := globalPreviewArgs.GetDNSConfigArgs.flags()
	return flags
}

var globalPreviewArgs PreviewArgs

func Preview(args PreviewArgs) error {
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	fmt.Println(len(cfg.Domains))
	return nil
}
