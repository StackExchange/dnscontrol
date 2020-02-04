package commands

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v2/providers"
	"github.com/StackExchange/dnscontrol/v2/providers/config"
	"github.com/urfave/cli/v2"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args GetZoneArgs
	return &cli.Command{
		Name:  "get-zone",
		Usage: "gets a zone from a provider (stand-alone)",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() != 3 {
				return cli.NewExitError("Arguments should be: zonename credskey providername (Ex: example.com r53 ROUTE53)", 1)

			}
			args.CredName = ctx.Args().Get(0)
			args.ProviderName = ctx.Args().Get(1)
			args.ZoneName = ctx.Args().Get(2)
			return exit(GetZone(args))
		},
		Flags:     args.flags(),
		UsageText: "main get-zone [command options] credkey provider zone",
		Description: `Download a zone from a provider.  This is a stand-alone utility.

ARGUMENTS:
   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js) 
   provider: The name of the provider (second parameter to NewDnsProvider() in dnsconfig.js)	
   zone:     The name of the zone (domain) to download

EXAMPLES:
   dnscontrol get-zone myr53 ROUTE53 example.com
   dnscontrol get-zone -format=tsv bind BIND example.com
   dnscontrol get-zone -format=dsl -out=draft.js glcoud GCLOUD example.com`,
	}
}())

// GetZoneArgs args required for the create-domain subcommand.
type GetZoneArgs struct {
	GetCredentialsArgs        // Args related to creds.json
	ZoneName           string // The zone to get
	CredName           string // key in creds.json
	ProviderName       string // provider name: BIND, GANDI_V5, etc or "-"
	OutputFormat       string // Output format
	OutputFile         string // File to send output (default stdout)
}

func (args *GetZoneArgs) flags() []cli.Flag {
	flags := args.GetCredentialsArgs.flags()
	flags = append(flags, &cli.StringFlag{
		Name:        "format",
		Destination: &args.OutputFormat,
		Value:       "pretty",
		Usage:       `Output format: dsl tsv pretty`,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "out",
		Destination: &args.OutputFile,
		Usage:       `Instead of stdout, write to this file`,
	})
	return flags
}

// GetZone contains all data/flags needed to run get-zone, independently of CLI.
func GetZone(args GetZoneArgs) error {
	var providerConfigs map[string]map[string]string
	var err error

	providerConfigs, err = config.LoadProviderConfigs(args.CredsFile)
	if err != nil {
		return err
	}

	fmt.Printf("CONFIGS: %d loaded\n", len(providerConfigs))
	//fmt.Printf("OUTPUTFORMAT = %q\n", args.OutputFormat)
	//fmt.Printf("ARGS = %+v\n", args)
	fmt.Printf("cred = %v\n", args.CredName)
	fmt.Printf("prov = %v\n", args.ProviderName)
	fmt.Printf("zone = %v\n", args.ZoneName)

	provider, err := providers.CreateDNSProvider(args.ProviderName, providerConfigs[args.CredName], nil)
	if err != nil {
		return err
	}
	fmt.Printf("FOO=%+v\n", provider)

	recs, err := provider.GetZoneRecords(args.ZoneName)
	fmt.Printf("RECS=%v\n", recs)
	fmt.Printf("err = %v\n", err)

	return err
}
