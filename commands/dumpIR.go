package commands

import (
	"fmt"
	"os"

	"github.com/StackExchange/dnscontrol/pkg/normalize"
	"github.com/urfave/cli"
)

var _ = cmd(catDebug, func() *cli.Command {
	var args DumpIRArgs
	return &cli.Command{
		Name:  "dump-ir",
		Usage: "Output intermediate representation (IR) after running validation and normalization logic.",
		Action: func(c *cli.Context) error {
			return exit(DumpIR(args))
		},
		Flags: args.flags(),
	}
}())

var _ = cmd(catDebug, func() *cli.Command {
	var args DumpIRArgs
	// This is the same as dump-ir but output defaults to /dev/null.
	return &cli.Command{
		Name:  "check",
		Usage: "Check and validate dnsconfig.js. Do not access providers.",
		Action: func(c *cli.Context) error {
			if args.Output == "" {
				args.Output = os.DevNull
			}
			return exit(DumpIR(args))
		},
		Flags: args.flags(),
	}
}())

type DumpIRArgs struct {
	GetDNSConfigArgs
	PrintJSONArgs
}

func (args *DumpIRArgs) flags() []cli.Flag {
	return append(args.GetDNSConfigArgs.flags(), args.PrintJSONArgs.flags()...)
}

func DumpIR(args DumpIRArgs) error {
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
