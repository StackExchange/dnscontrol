package cmd

import (
	"encoding/json"
	"os"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/urfave/cli"
)

var app = cli.NewApp()

// categories of commands
const (
	catMain     = "main"
	catPlumbing = "plumbing"
	catUtils    = "utility"
)

func init() {
	app.Name = "dnscontrol"
	app.Usage = "dnscontrol is a compiler and dsl for managing cloud dns zones"
	app.Commands = []cli.Command{
		*previewCommand,
		*pushCommand,
		*debugJSCommand,
		*debugPreprocessCommand,
	}
	app.EnableBashCompletion = true
}

// Run will execute the CLI
func Run() error {
	app.Run(os.Args)
	return nil
}

// Shared config types

// GetDNSConfigArgs contains what we need to get a valid dns config.
// Could come from parsing js, or from stored json
type GetDNSConfigArgs struct {
	ExecuteDSLArgs
	JSONFile string
}

func (args *GetDNSConfigArgs) flags() []cli.Flag {
	return append(args.ExecuteDSLArgs.flags(),
		cli.StringFlag{
			Destination: &args.JSONFile,
			Name:        "json",
			Usage:       "file containing intermediate json",
		},
	)
}

func GetDNSConfig(args GetDNSConfigArgs) (*models.DNSConfig, error) {
	if args.JSONFile != "" {
		f, err := os.Open(args.JSONFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		cfg := &models.DNSConfig{}
		if err = dec.Decode(cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	return ExecuteDSL(args.ExecuteDSLArgs)
}

// ExecuteDSLArgs are used anytime we need to read and execute dnscontrol javascript
type ExecuteDSLArgs struct {
	JSFile   string
	JSONFile string
	DevMode  bool
}

func (args *ExecuteDSLArgs) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "js",
			Value:       "dnsconfig.js",
			Destination: &args.JSFile,
			Usage:       "Javascript file containing dns config",
		},
		cli.BoolFlag{
			Name:        "dev",
			Destination: &args.DevMode,
			Usage:       "Use helpers.js from disk instead of embedded copy",
		},
	}
}

// PrintJSONArgs are used anytime a command may print some json
type PrintJSONArgs struct {
	Pretty bool
	Output string
}

func (args *PrintJSONArgs) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "pretty",
			Destination: &args.Pretty,
			Usage:       "Pretty print json",
		},
		cli.StringFlag{
			Name:        "out",
			Destination: &args.Output,
			Usage:       "File to write json to",
		},
	}
}

type GetCredentialsArgs struct {
	CredsFile string
}

func (args *GetCredentialsArgs) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "creds",
			Destination: &args.CredsFile,
			Usage:       "Provider credentials JSON file",
			Value:       "creds.json",
		},
	}
}
