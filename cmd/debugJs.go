package cmd

import (
	"fmt"
	"io/ioutil"

	"encoding/json"

	"os"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/js"
	"github.com/urfave/cli"
)

func exit(err error) error {
	if err == nil {
		return nil
	}
	return cli.NewExitError(err, 1)
}

var _ = cmd(catDebug, &cli.Command{
	Name:  "debug-js",
	Usage: "Output intermediate representation (IR) after JavaScript is executed but before any validation/normalization",
	Action: func(c *cli.Context) error {
		return exit(DebugJS(globalDebugJSArgs))
	},
	Flags: globalDebugJSArgs.flags(),
})

type DebugJSArgs struct {
	PrintJSONArgs
	ExecuteDSLArgs
}

func (args *DebugJSArgs) flags() []cli.Flag {
	return append(args.ExecuteDSLArgs.flags(), args.PrintJSONArgs.flags()...)
}

var globalDebugJSArgs DebugJSArgs

func DebugJS(args DebugJSArgs) error {
	config, err := ExecuteDSL(args.ExecuteDSLArgs)
	if err != nil {
		return err
	}
	return PrintJSON(args.PrintJSONArgs, config)
}

func ExecuteDSL(args ExecuteDSLArgs) (*models.DNSConfig, error) {
	if args.JSFile == "" {
		return nil, fmt.Errorf("No config specified")
	}
	text, err := ioutil.ReadFile(args.JSFile)
	if err != nil {
		return nil, fmt.Errorf("Reading js file %s: %s", args.JSFile, err)
	}
	dnsConfig, err := js.ExecuteJavascript(string(text), args.DevMode)
	if err != nil {
		return nil, fmt.Errorf("Executing javascript in %s: %s", args.JSFile, err)
	}
	return dnsConfig, nil
}

func PrintJSON(args PrintJSONArgs, config *models.DNSConfig) (err error) {
	var dat []byte
	if args.Pretty {
		dat, err = json.MarshalIndent(config, "", "  ")
	} else {
		dat, err = json.Marshal(config)
	}
	if err != nil {
		return err
	}
	if args.Output != "" {
		f, err := os.Create(args.Output)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.Write(dat)
		return err
	}
	fmt.Println(string(dat))
	return nil
}
