package commands

import (
	"fmt"
	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"github.com/urfave/cli/v2"
	"io/ioutil"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args FmtArgs
	return &cli.Command{
		Name:  "fmt",
		Usage: "Format and prettify the given file",
		Action: func(c *cli.Context) error {
			return exit(FmtFile(args))
		},
		Flags: args.flags(),
	}
}())

type FmtArgs struct {
	File string
}

func (args *FmtArgs) flags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, &cli.StringFlag{
		Name:        "file",
		Aliases:     []string{"f"},
		Value:       "dnsconfig.js",
		Usage:       "File to format",
		Destination: &args.File,
	})
	return flags
}

func FmtFile(args FmtArgs) error {
	fileBytes, readErr := ioutil.ReadFile(args.File)
	if readErr != nil {
		return readErr
	}

	opts := jsbeautifier.DefaultOptions()

	str := string(fileBytes)
	beautified, beautifyErr := jsbeautifier.Beautify(&str, opts)
	if beautifyErr != nil {
		return beautifyErr
	}
	fmt.Print(beautified)
	return nil
}
