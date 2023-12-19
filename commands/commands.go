package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/pkg/js"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/urfave/cli/v2"

	"github.com/fatih/color"
)

// categories of commands
const (
	catMain  = "\b main" // screwed up to alphebatize first
	catDebug = "debug"
	catUtils = "utility"
)

var commands = []*cli.Command{}

func cmd(cat string, c *cli.Command) bool {
	c.Category = cat
	commands = append(commands, c)
	return true
}

var _ = cmd(catDebug, &cli.Command{
	Name:  "version",
	Usage: "Print version information",
	Action: func(c *cli.Context) error {
		_, err := fmt.Println(version)
		return err
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
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "v",
			Usage:       "Enable detailed logging",
			Destination: &printer.DefaultPrinter.Verbose,
		},
		&cli.BoolFlag{
			Name:        "allow-fetch",
			Usage:       "Enable JS fetch(), dangerous on untrusted code!",
			Destination: &js.EnableFetch,
		},
		&cli.BoolFlag{
			Name:   "diff2",
			Usage:  "Obsolete flag. Will be removed in v5 or later",
			Hidden: true,
			Action: func(ctx *cli.Context, v bool) error {
				obsoleteDiff2FlagUsed = true
				return nil
			},
		},
		&cli.BoolFlag{
			Name:        "disableordering",
			Usage:       "Disables update reordering",
			Destination: &diff2.DisableOrdering,
		},
		&cli.BoolFlag{
			Name:        "no-colors",
			Usage:       "Disable colors",
			Destination: &color.NoColor,
			Value:       false,
		},
	}
	sort.Sort(cli.CommandsByName(commands))
	app.Commands = commands
	app.EnableBashCompletion = true
	app.BashComplete = func(cCtx *cli.Context) {
		// ripped from cli.DefaultCompleteWithFlags
		var lastArg string

		if len(os.Args) > 2 {
			lastArg = os.Args[len(os.Args)-2]
		}

		if lastArg != "" {
			if strings.HasPrefix(lastArg, "-") {
				if !islastFlagComplete(lastArg, app.Flags) {
					dnscontrolPrintFlagSuggestions(lastArg, app.Flags, cCtx.App.Writer)
					return
				}
			}
		}
		dnscontrolPrintCommandSuggestions(app.Commands, cCtx.App.Writer)
	}
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
		&cli.StringFlag{
			Destination: &args.JSONFile,
			Name:        "ir",
			Usage:       "Read IR (json) directly from this file. Do not process DSL at all",
		},
		&cli.StringFlag{
			Destination: &args.JSONFile,
			Name:        "json",
			Hidden:      true,
			Usage:       "same as -ir. only here for backwards compatibility, hence hidden",
		},
	)
}

// GetDNSConfig reads the json-formatted IR file. Or executes javascript. All depending on flags provided.
func GetDNSConfig(args GetDNSConfigArgs) (*models.DNSConfig, error) {
	var err error
	cfg := &models.DNSConfig{}

	if args.JSONFile == "" {
		// No IR file specified. Generate the IR by running dnsconfig.json
		// as normal.
		cfg, err = ExecuteDSL(args.ExecuteDSLArgs)
		if err != nil {
			return nil, err
		}
	} else {
		// Read an IR file.
		f, err := os.Open(args.JSONFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		if err = dec.Decode(cfg); err != nil {
			return nil, err
		}
	}

	return preloadProviders(cfg)
}

// the json only contains provider names inside domains. This denormalizes the data for more
// convenient access patterns. Does everything we need to prepare for the validation phase, but
// cannot do anything that requires the credentials file yet.
func preloadProviders(cfg *models.DNSConfig) (*models.DNSConfig, error) {
	//build name to type maps
	cfg.RegistrarsByName = map[string]*models.RegistrarConfig{}
	cfg.DNSProvidersByName = map[string]*models.DNSProviderConfig{}
	for _, reg := range cfg.Registrars {
		cfg.RegistrarsByName[reg.Name] = reg
	}
	for _, p := range cfg.DNSProviders {
		cfg.DNSProvidersByName[p.Name] = p
	}
	// make registrar and dns provider shims. Include name, type, and other metadata, but can't instantiate
	// driver until we load creds in later
	for _, d := range cfg.Domains {
		reg, ok := cfg.RegistrarsByName[d.RegistrarName]
		if !ok {
			return nil, fmt.Errorf("registrar named %s expected for %s, but never registered", d.RegistrarName, d.Name)
		}
		d.RegistrarInstance = &models.RegistrarInstance{
			ProviderBase: models.ProviderBase{
				Name:         reg.Name,
				ProviderType: reg.Type,
			},
		}
		for pName, n := range d.DNSProviderNames {
			prov, ok := cfg.DNSProvidersByName[pName]
			if !ok {
				return nil, fmt.Errorf("DNS Provider named %s expected for %s, but never registered", pName, d.Name)
			}
			d.DNSProviderInstances = append(d.DNSProviderInstances, &models.DNSProviderInstance{
				ProviderBase: models.ProviderBase{
					Name:         pName,
					ProviderType: prov.Type,
				},
				NumberOfNameservers: n,
			})
		}
		// sort so everything is deterministic
		sort.Slice(d.DNSProviderInstances, func(i, j int) bool {
			return d.DNSProviderInstances[i].Name < d.DNSProviderInstances[j].Name
		})
	}
	return cfg, nil
}

// ExecuteDSLArgs are used anytime we need to read and execute dnscontrol DSL
type ExecuteDSLArgs struct {
	JSFile   string
	JSONFile string
	DevMode  bool
	Variable cli.StringSlice
}

func (args *ExecuteDSLArgs) flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Value:       "dnsconfig.js",
			Destination: &args.JSFile,
			Usage:       "File containing dns config in javascript DSL",
		},
		&cli.StringFlag{
			Name:        "js",
			Value:       "dnsconfig.js",
			Hidden:      true,
			Destination: &args.JSFile,
			Usage:       "same as config. for back compatibility",
		},
		&cli.BoolFlag{
			Name:        "dev",
			Destination: &args.DevMode,
			Usage:       "Use helpers.js from disk instead of embedded copy",
		},
		&cli.StringSliceFlag{
			Name:        "variable",
			Aliases:     []string{"v"},
			Destination: &args.Variable,
			Usage:       "Add variable that is passed to JS",
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
		&cli.BoolFlag{
			Name:        "pretty",
			Destination: &args.Pretty,
			Usage:       "Pretty print IR JSON",
		},
		&cli.StringFlag{
			Name:        "out",
			Destination: &args.Output,
			Usage:       "File to write IR JSON to (default stdout)",
		},
	}
}

// GetCredentialsArgs encapsulates the flags/args for sub-commands that use the creds.json file.
type GetCredentialsArgs struct {
	CredsFile string
}

func (args *GetCredentialsArgs) flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "creds",
			Destination: &args.CredsFile,
			Usage:       "Provider credentials JSON file (or !program to execute program that outputs json)",
			Value:       "creds.json",
		},
	}
}

// FilterArgs encapsulates the flags/args for sub-commands that can filter by provider or domain.
type FilterArgs struct {
	Providers string
	Domains   string
}

func (args *FilterArgs) flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "providers",
			Destination: &args.Providers,
			Usage:       `Providers to enable (comma separated list); default is all. Can exclude individual providers from default by adding '"_exclude_from_defaults": "true"' to the credentials file for a provider`,
			Value:       "",
		},
		&cli.StringFlag{
			Name:        "domains",
			Destination: &args.Domains,
			Usage:       `Comma separated list of domain names to include`,
			Value:       "",
		},
	}
}

func (args *FilterArgs) shouldRunProvider(name string, dc *models.DomainConfig) bool {
	if args.Providers == "all" {
		return true
	}
	if args.Providers == "" {
		for _, pri := range dc.DNSProviderInstances {
			if pri.Name == name {
				return pri.IsDefault
			}
		}
		return true
	}
	for _, prov := range strings.Split(args.Providers, ",") {
		if prov == name {
			return true
		}
	}
	return false
}

func (args *FilterArgs) shouldRunDomain(d string) bool {
	if args.Domains == "" {
		return true
	}
	return domainInList(d, strings.Split(args.Domains, ","))
}

func domainInList(domain string, list []string) bool {
	for _, item := range list {
		if strings.HasPrefix(item, "*") && strings.HasSuffix(domain, item[1:]) {
			return true
		}
		if item == domain {
			return true
		}
	}
	return false
}
