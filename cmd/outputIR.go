package cmd

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/pkg/normalize"
	"github.com/urfave/cli"
)

var _ = cmd(catDebug, &cli.Command{
	Name:  "output-ir",
	Usage: "Output intermediate representation (IR) after running validation and normalization logic.",
	Action: func(c *cli.Context) error {
		return exit(DebugPreprocess(globalDebugPreprocessArgs))
	},
	Flags: globalDebugPreprocessArgs.flags(),
})

type OutputIRArgs struct {
	GetDNSConfigArgs
	PrintJSONArgs
}

func (args *OutputIRArgs) flags() []cli.Flag {
	return append(args.GetDNSConfigArgs.flags(), args.PrintJSONArgs.flags()...)
}

var globalDebugPreprocessArgs OutputIRArgs

func DebugPreprocess(args OutputIRArgs) error {
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	fmt.Println(len(cfg.Domains))
	errs := normalize.NormalizeAndValidateConfig(cfg)
	if PrintValidationErrors(errs) {
		return fmt.Errorf("Exiting due to validation errors")
	}

	return PrintJSON(args.PrintJSONArgs, cfg)
}

func PrintValidationErrors(errs []error) (fatal bool) {
	if len(errs) == 0 {
		return false
	}
	fmt.Printf("%d Validation errors:\n", len(errs))
	for _, err := range errs {
		if _, ok := err.(normalize.Warning); ok {
			fmt.Printf("WARNING: %s\n", err)
		} else {
			fatal = true
			fmt.Printf("ERROR: %s\n", err)
		}
	}
	return
}
