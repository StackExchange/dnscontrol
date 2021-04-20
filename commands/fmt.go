package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"github.com/urfave/cli/v2"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args FmtArgs
	return &cli.Command{
		Name:  "fmt",
		Usage: "[BETA] Format and prettify a given file",
		Action: func(c *cli.Context) error {
			return exit(FmtFile(args))
		},
		Flags: args.flags(),
	}
}())

// FmtArgs stores arguments related to the fmt subcommand.
type FmtArgs struct {
	InputFile  string
	OutputFile string
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
		Usage:       "Output file",
		Destination: &args.OutputFile,
	})
	return flags
}

// FmtFile reads and formats a file.
func FmtFile(args FmtArgs) error {
	fileBytes, readErr := ioutil.ReadFile(args.InputFile)
	if readErr != nil {
		return readErr
	}

	opts := jsbeautifier.DefaultOptions()

	str := string(fileBytes)
	beautified, beautifyErr := jsbeautifier.Beautify(&str, opts)
	if beautifyErr != nil {
		return beautifyErr
	}

	if len(beautified) != 0 && beautified[len(beautified)-1] != '\n' {
		beautified = beautified + "\n"
	}

	if args.OutputFile == "" {
		fmt.Print(beautified)
	} else {
		if err := ioutil.WriteFile(args.OutputFile, []byte(beautified), 0744); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "File %s successfully written\n", args.OutputFile)
	}
	return nil
}
