package commands

import (
	"fmt"
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
		Name:    "get-zones",
		Aliases: []string{"get-zone"},
		Usage:   "gets a zone from a provider (stand-alone)",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 3 {
				return cli.NewExitError("Arguments should be: credskey providername zone(s) (Ex: r53 ROUTE53 example.com)", 1)

			}
			args.CredName = ctx.Args().Get(0)
			args.ProviderName = ctx.Args().Get(1)
			args.ZoneNames = ctx.Args().Slice()[2:]
			return exit(GetZone(args))
		},
		Flags:     args.flags(),
		UsageText: "dnscontrol get-zones [command options] credkey provider zone [...]",
		Description: `Download a zone from a provider.  This is a stand-alone utility.

ARGUMENTS:
   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js)
   provider: The name of the provider (second parameter to NewDnsProvider() in dnsconfig.js)
   zone:     One or more zones (domains) to download; or "all".

FORMATS:
   --format=dsl      dnsconfig.js format (not perfect, but a decent first draft)
   --format=nameonly Just print the zone names
   --format=pretty   BIND Zonefile format
   --format=tsv      TAB separated value (useful for AWK)

EXAMPLES:
   dnscontrol get-zones myr53 ROUTE53 example.com
   dnscontrol get-zones gmain GANDI_V5 example.comn other.com
   dnscontrol get-zones cfmain CLOUDFLAREAPI all
   dnscontrol get-zones -format=tsv bind BIND example.com
   dnscontrol get-zones -format=dsl -out=draft.js glcoud GCLOUD example.com`,
	}
}())

// check-creds foo bar
// is the same as
// get-zones --format=nameonly foo bar all
var _ = cmd(catUtils, func() *cli.Command {
	var args GetZoneArgs
	return &cli.Command{
		Name:  "check-creds",
		Usage: "Do a small operation to verify credentials (stand-alone)",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() != 2 {
				return cli.NewExitError("Arguments should be: credskey providername (Ex: r53 ROUTE53)", 1)

			}
			args.CredName = ctx.Args().Get(0)
			args.ProviderName = ctx.Args().Get(1)
			args.ZoneNames = []string{"all"}
			args.OutputFormat = "nameonly"
			return exit(GetZone(args))
		},
		Flags:     args.flags(),
		UsageText: "dnscontrol check-creds [command options] credkey provider",
		Description: `Do a trivia operation to verify credentials.  This is a stand-alone utility.

If successful, a list of zones will be output. If not, hopefully you
see verbose error messages.

ARGUMENTS:
   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js)
   provider: The name of the provider (second parameter to NewDnsProvider() in dnsconfig.js)

EXAMPLES:
   dnscontrol get-zones myr53 ROUTE53 
   dnscontrol get-zones --out=/dev/null myr53 ROUTE53`,
	}
}())

// GetZoneArgs args required for the create-domain subcommand.
type GetZoneArgs struct {
	GetCredentialsArgs          // Args related to creds.json
	CredName           string   // key in creds.json
	ProviderName       string   // provider name: BIND, GANDI_V5, etc or "-"
	ZoneNames          []string // The zones to get
	OutputFormat       string   // Output format
	OutputFile         string   // Filename to send output ("" means stdout)
	DefaultTTL         int      // default TTL for providers where it is unknown
}

func (args *GetZoneArgs) flags() []cli.Flag {
	flags := args.GetCredentialsArgs.flags()
	flags = append(flags, &cli.StringFlag{
		Name:        "format",
		Destination: &args.OutputFormat,
		Value:       "pretty",
		Usage:       `Output format: dsl pretty tsv nameonly`,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "out",
		Destination: &args.OutputFile,
		Usage:       `Instead of stdout, write to this file`,
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "ttl",
		Destination: &args.DefaultTTL,
		Usage:       `Default TTL (0 picks the zone's most common TTL)`,
	})
	return flags
}

// GetZone contains all data/flags needed to run get-zones, independently of CLI.
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

	// decide which zones we need to convert
	zones := args.ZoneNames
	if len(args.ZoneNames) == 1 && args.ZoneNames[0] == "all" {
		lister, ok := provider.(providers.ZoneLister)
		if !ok {
			return fmt.Errorf("provider type %s cannot list zones to use the 'all' feature", args.ProviderName)
		}
		zones, err = lister.ListZones()
		if err != nil {
			return err
		}
	}

	// first open output stream and print initial header (if applicable)
	w := os.Stdout
	if args.OutputFile != "" {
		w, err = os.Create(args.OutputFile)
	}
	if err != nil {
		return err
	}
	defer w.Close()

	if args.OutputFormat == "nameonly" {
		for _, zone := range zones {
			fmt.Fprintln(w, zone)
		}
		return nil
	}

	// actually fetch all of the records
	zoneRecs := make([]models.Records, len(zones))
	for i, zone := range zones {
		recs, err := provider.GetZoneRecords(zone)
		if err != nil {
			return err
		}
		zoneRecs[i] = recs
	}

	// Write it out:

	if args.OutputFormat == "dsl" {
		fmt.Fprintf(w, `var %s = NewDnsProvider("%s", "%s");`+"\n",
			args.CredName, args.CredName, args.ProviderName)
	}

	// now print all zones
	for i, recs := range zoneRecs {
		zoneName := zones[i]

		z := prettyzone.PrettySort(recs, zoneName, 0, nil)
		switch args.OutputFormat {

		case "pretty":
			fmt.Fprintf(w, "$ORIGIN %s.\n", zoneName)
			prettyzone.WriteZoneFileRC(w, z.Records, zoneName, uint32(args.DefaultTTL), nil)
			fmt.Fprintln(w)

		case "dsl":
			fmt.Fprintf(w, `D("%s", REG_CHANGEME,`, zoneName)
			fmt.Fprintf(w, "\n\tDnsProvider(%s)", args.CredName)
			defaultTTL := uint32(args.DefaultTTL)
			if defaultTTL == 0 {
				defaultTTL = prettyzone.MostCommonTTL(recs)
			}
			if defaultTTL != models.DefaultTTL && defaultTTL != 0 {
				fmt.Fprintf(w, "\n\tDefaultTTL(%d)", defaultTTL)
			}
			for _, rec := range recs {
				fmt.Fprint(w, formatDsl(zoneName, rec, defaultTTL))
			}
			fmt.Fprint(w, "\n)\n")

		case "tsv":
			for _, rec := range recs {
				fmt.Fprintf(w,
					fmt.Sprintf("%s\t%s\t%d\tIN\t%s\t%s\n",
						rec.NameFQDN, rec.Name, rec.TTL, rec.Type, rec.GetTargetCombined()))
			}

		default:
			return fmt.Errorf("format %q unknown", args.OutputFile)
		}
	}
	return nil
}

func formatDsl(zonename string, rec *models.RecordConfig, defaultTTL uint32) string {

	target := rec.GetTargetCombined()

	ttlop := ""
	if rec.TTL != defaultTTL && rec.TTL != 0 {
		ttlop = fmt.Sprintf(", TTL(%d)", rec.TTL)
	}

	switch rec.Type { // #rtype_variations
	case "MX":
		target = fmt.Sprintf("%d, '%s'", rec.MxPreference, rec.GetTargetField())
	case "SOA":
	case "TXT":
		if len(rec.TxtStrings) == 1 {
			target = `'` + rec.TxtStrings[0] + `'`
		} else {
			target = `['` + strings.Join(rec.TxtStrings, `', '`) + `']`
		}
	case "NS":
		// NS records at the apex should be NAMESERVER() records.
		if rec.Name == "@" {
			return fmt.Sprintf(",\n\tNAMESERVER('%s')", target)
		}
	default:
		target = "'" + target + "'"
	}

	return fmt.Sprintf(",\n\t%s('%s', %s%s)", rec.Type, rec.Name, target, ttlop)
}
