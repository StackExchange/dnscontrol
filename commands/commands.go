package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/urfave/cli"
)

// categories of commands
const (
	catMain  = "\b main" // screwed up to alphebatize first
	catDebug = "debug"
	catUtils = "utility"
)

var commands = []cli.Command{}
var version string

func cmd(cat string, c *cli.Command) bool {
	c.Category = cat
	commands = append(commands, *c)
	return true
}

var _ = cmd(catDebug, &cli.Command{
	Name:  "version",
	Usage: "Print version information",
	Action: func(c *cli.Context) {
		fmt.Println(version)
	},
})

// Run will execute the CLI
func Run(v string) int {
	version = v
	app := cli.NewApp()
	app.Version = version
	app.Name = "dnscontrol"
	app.HideVersion = true
	app.Usage = "dnscontrol is a compiler and DSL for managing dns zones"
	sort.Sort(cli.CommandsByName(commands))
	app.Commands = commands
	app.EnableBashCompletion = true
	if err := app.Run(os.Args); err != nil {
		return 1
	}
	return 0
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
			Name:        "ir",
			Usage:       "Read IR (json) directly from this file. Do not process DSL at all",
		},
		cli.StringFlag{
			Destination: &args.JSONFile,
			Name:        "json",
			Hidden:      true,
			Usage:       "same as -ir. only here for backwards compatibility, hence hidden",
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

// ExecuteDSLArgs are used anytime we need to read and execute dnscontrol DSL
type ExecuteDSLArgs struct {
	JSFile   string
	JSONFile string
	DevMode  bool
}

func (args *ExecuteDSLArgs) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Value:       "dnsconfig.js",
			Destination: &args.JSFile,
			Usage:       "File containing dns config in javascript DSL",
		},
		cli.StringFlag{
			Name:        "js",
			Value:       "dnsconfig.js",
			Hidden:      true,
			Destination: &args.JSFile,
			Usage:       "same as config. for back compatibility",
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
			Usage:       "Pretty print IR JSON",
		},
		cli.StringFlag{
			Name:        "out",
			Destination: &args.Output,
			Usage:       "File to write IR JSON to (default stdout)",
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

type FilterArgs struct {
	Providers string
	Domains   string
}

func (args *FilterArgs) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "providers",
			Destination: &args.Providers,
			Usage:       `Providers to enable (comma separated list); default is all. Can exclude individual providers from default by adding '"_exclude_from_defaults": "true"' to the credentials file for a provider`,
			Value:       "",
		},
		cli.StringFlag{
			Name:        "domains",
			Destination: &args.Domains,
			Usage:       `Comma separated list of domain names to include`,
			Value:       "",
		},
	}
}

func (args *FilterArgs) shouldRunProvider(p string, dc *models.DomainConfig, nonDefaultProviders []string) bool {
	if args.Providers == "all" {
		return true
	}
	if args.Providers == "" {
		for _, pr := range nonDefaultProviders {
			if pr == p {
				return false
			}
		}
		return true
	}
	for _, prov := range strings.Split(args.Providers, ",") {
		if prov == p {
			return true
		}
	}
	return false
}

func (args *FilterArgs) shouldRunDomain(d string) bool {
	if args.Domains == "" {
		return true
	}
	for _, dom := range strings.Split(args.Domains, ",") {
		if dom == d {
			return true
		}
	}
	return false
}
