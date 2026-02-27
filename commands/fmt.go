package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"github.com/urfave/cli/v3"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args FmtArgs
	return &cli.Command{
		Name:  "fmt",
		Usage: "[BETA] Format and prettify a given file",
		Action: func(ctx context.Context, c *cli.Command) error {
			return exit(FmtFile(args))
		},
		Flags: args.flags(),
	}
}())

// FmtArgs stores arguments related to the fmt subcommand.
type FmtArgs struct {
	InputFile  string
	OutputFile string
	Verbose    bool
}

func (args *FmtArgs) flags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, &cli.StringFlag{
		Name:        "input",
		Aliases:     []string{"i"},
		Value:       "dnsconfig.js",
		Usage:       "Input file",
		Destination: &args.InputFile,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "output",
		Aliases:     []string{"o"},
		Value:       "dnsconfig.js",
		Usage:       "Output file",
		Destination: &args.OutputFile,
	})
	flags = append(flags, &cli.BoolFlag{
		Name:        "verbose",
		Aliases:     []string{"v"},
		Value:       false,
		Usage:       "Enable verbose output",
		Destination: &args.Verbose,
	})
	return flags
}

// FmtFile reads and formats a file.
func FmtFile(args FmtArgs) error {
	var fileBytes []byte
	if args.InputFile == "" {
		var err error
		fileBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		var err error
		fileBytes, err = os.ReadFile(args.InputFile)
		if err != nil {
			return err
		}
	}
	original := string(fileBytes)

	opts := jsbeautifier.DefaultOptions()
	beautified, beautifyErr := jsbeautifier.Beautify(&original, opts)
	if beautifyErr != nil {
		return beautifyErr
	}

	beautified = strings.TrimSpace(beautified)
	if len(beautified) != 0 {
		beautified = beautified + "\n"
	}

	if args.OutputFile == "" {
		fmt.Print(beautified)
	} else {
		changed := original != beautified
		if changed {
			if err := os.WriteFile(args.OutputFile, []byte(beautified), 0o744); err != nil {
				return err
			}
		}
		if args.Verbose || changed {
			if changed {
				fmt.Fprintf(os.Stderr, "%s (formatted)\n", args.OutputFile)
			} else {
				fmt.Fprintf(os.Stderr, "%s (unchanged)\n", args.OutputFile)
			}
		}
	}
	return nil
}
