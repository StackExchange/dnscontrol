package cmd

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/pkg/normalize"
	_ "github.com/StackExchange/dnscontrol/providers/_all"
	"github.com/urfave/cli"
)

var debugPreprocessCommand = &cli.Command{
	Name:  "debug-preprocess",
	Usage: "Run validation and normalization logic, and print resulting json",
	Action: func(c *cli.Context) error {
		return exit(DebugPreprocess(globalDebugPreprocessArgs))
	},
	Category: catPlumbing,
	Flags:    globalDebugPreprocessArgs.flags(),
}

type DebugPreprocessArgs struct {
	GetDNSConfigArgs
	PrintJSONArgs
}

func (args *DebugPreprocessArgs) flags() []cli.Flag {
	return append(args.GetDNSConfigArgs.flags(), args.PrintJSONArgs.flags()...)
}

var globalDebugPreprocessArgs DebugPreprocessArgs

func DebugPreprocess(args DebugPreprocessArgs) error {
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	fmt.Println(len(cfg.Domains))
	errs := normalize.NormalizeAndValidateConfig(cfg)
	if len(errs) > 0 {
		fmt.Printf("%d Validation errors:\n", len(errs))
		fatal := false
		for _, err := range errs {
			if _, ok := err.(normalize.Warning); ok {
				fmt.Printf("WARNING: %s\n", err)
			} else {
				fatal = true
				fmt.Printf("ERROR: %s\n", err)
			}
		}
		if fatal {
			return fmt.Errorf("Exiting due to validation errors")
		}
	}
	return PrintJSON(args.PrintJSONArgs, cfg)
}
