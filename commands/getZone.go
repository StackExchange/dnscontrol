package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/pkg/prettyzone"
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
	OutputFile         string // Filename to send output ("" means stdout)
	DefaultTTL         int    // default TTL for providers where it is unknown
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
	flags = append(flags, &cli.IntFlag{
		Name:        "ttl",
		Destination: &args.DefaultTTL,
		Usage:       `Default TTL`,
		Value:       300,
	})
	return flags
}

// GetZone contains all data/flags needed to run get-zone, independently of CLI.
func GetZone(args GetZoneArgs) error {
	var providerConfigs map[string]map[string]string
	var err error

	// Read it in:
	providerConfigs, err = config.LoadProviderConfigs(args.CredsFile)
	if err != nil {
		return err
	}

	provider, err := providers.CreateDNSProvider(args.ProviderName, providerConfigs[args.CredName], nil)
	if err != nil {
		return err
	}

	recs, err := provider.GetZoneRecords(args.ZoneName)
	if err != nil {
		return err
	}

	z := prettyzone.PrettySort(recs, args.ZoneName, 0)

	// Write it out:

	switch args.OutputFormat {
	case "pretty":
		prettyzone.WriteZoneFileRC(os.Stdout, z.Records, args.ZoneName)
	case "dsl":
		writeDsl(os.Stdout, z.Records, args)
	case "tsv":
		rrFormat(args.ZoneName, args.OutputFile, z.Records, uint32(args.DefaultTTL), false)
	default:
		return fmt.Errorf("format %q unknown", args.OutputFile)
	}

	return nil
}

func writeDsl(w io.Writer, recs models.Records, args GetZoneArgs) error {
	fmt.Fprintf(w, `var CHANGEME = NewDnsProvider("%s", "%s");`+"\n",
		args.CredName, args.ProviderName)
	fmt.Fprintf(w, `D("%s", REG_CHANGEME,`+"\n", args.ZoneName)
	fmt.Fprintf(w, `        DnsProvider(CHANGEME)`)
	rrFormat(args.ZoneName, args.OutputFile, recs, uint32(args.DefaultTTL), true)
	fmt.Fprintln(w, "\n)")
	return nil
}

// rrFormat outputs the zonefile in either DSL or TSV format.
func rrFormat(zonename string, filename string, recs models.Records, defaultTTL uint32, dsl bool) {

	for _, x := range recs {

		target := x.GetTargetCombined()

		var ttlop string
		if x.TTL == defaultTTL {
			ttlop = ""
		} else {
			ttlop = fmt.Sprintf(", TTL(%d)", x.TTL)
		}

		// NS records at the apex should be NAMESERVER() records.
		if x.Type == "NS" && x.Name == "@" {
			fmt.Printf(",\n\tNAMESERVER('%s'%s)", target, ttlop)
			continue
		}

		if !dsl { // TSV format:
			fmt.Printf("%s\t%d\tIN\t%s\t%s\n", x.Name, x.TTL, x.Type, target)
		} else { // DSL format:
			switch x.Type { // #rtype_variations
			case "MX":
				target = fmt.Sprintf("%d, '%s'", x.MxPreference, x.GetTargetField())
			case "SOA":
				continue
			case "TXT":
				if len(x.TxtStrings) == 1 {
					target = `'` + x.TxtStrings[0] + `'`
				} else {
					target = `['` + strings.Join(x.TxtStrings, `', '`) + `']`
				}
			default:
				target = "'" + target + "'"
			}
			fmt.Printf(",\n\t%s('%s', %s%s)", x.Type, x.Name, target, ttlop)
		}
	}

}
