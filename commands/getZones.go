package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/StackExchange/dnscontrol/v3/providers/config"
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
   --format=js        dnsconfig.js format (not perfect, just a decent first draft)
   --format=djs       js with disco commas (leading commas)
   --format=zone      BIND zonefile format
   --format=tsv       TAB separated value (useful for AWK)
   --format=nameonly  Just print the zone names

The columns in --format=tsv are:
   FQDN (the label with the domain)
   ShortName (just the label, "@" if it is the naked domain)
   TTL
   Record Type (A, AAAA, CNAME, etc.)
   Target and arguments (quoted like in a zonefile)
   Either empty or a comma-separated list of properties like "cloudflare_proxy=true"

The --ttl flag only applies to zone/js/djs formats.

EXAMPLES:
   dnscontrol get-zones myr53 ROUTE53 example.com
   dnscontrol get-zones gmain GANDI_V5 example.com other.com
   dnscontrol get-zones cfmain CLOUDFLAREAPI all
   dnscontrol get-zones --format=tsv bind BIND example.com
   dnscontrol get-zones --format=djs --out=draft.js glcoud GCLOUD example.com`,
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
		Value:       "zone",
		Usage:       `Output format: js djs zone tsv nameonly`,
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

	// fetch all of the records
	zoneRecs := make([]models.Records, len(zones))
	for i, zone := range zones {
		recs, err := provider.GetZoneRecords(zone)
		if err != nil {
			return err
		}
		zoneRecs[i] = recs
	}

	// Write the heading:

	if args.OutputFormat == "js" || args.OutputFormat == "djs" {
		fmt.Fprintf(w, `var %s = NewDnsProvider("%s", "%s");`+"\n",
			args.CredName, args.CredName, args.ProviderName)
		fmt.Fprintf(w, `var REG_CHANGEME = NewRegistrar("ThirdParty", "NONE");`+"\n")
	}

	// print each zone
	for i, recs := range zoneRecs {
		zoneName := zones[i]

		z := prettyzone.PrettySort(recs, zoneName, 0, nil)
		switch args.OutputFormat {

		case "zone":
			fmt.Fprintf(w, "$ORIGIN %s.\n", zoneName)
			prettyzone.WriteZoneFileRC(w, z.Records, zoneName, uint32(args.DefaultTTL), nil)
			fmt.Fprintln(w)

		case "js", "djs":
			sep := ",\n\t" // Commas at EOL
			if args.OutputFormat == "djs" {
				sep = "\n\t, " // Funky comma mode
			}
			fmt.Fprintf(w, `D("%s", REG_CHANGEME%s`, zoneName, sep)
			var o []string
			o = append(o, fmt.Sprintf("DnsProvider(%s)", args.CredName))
			defaultTTL := uint32(args.DefaultTTL)
			if defaultTTL == 0 {
				defaultTTL = prettyzone.MostCommonTTL(recs)
			}
			if defaultTTL != models.DefaultTTL && defaultTTL != 0 {
				o = append(o, fmt.Sprintf("DefaultTTL(%d)", defaultTTL))
			}
			for _, rec := range recs {
				o = append(o, formatDsl(zoneName, rec, defaultTTL))
			}
			out := strings.Join(o, sep)
			fmt.Fprint(w, strings.ReplaceAll(out, "\n\t, //", "\n\t//, "))
			fmt.Fprint(w, "\n)\n")

		case "tsv":
			for _, rec := range recs {

				cfproxy := ""
				if cp, ok := rec.Metadata["cloudflare_proxy"]; ok {
					if cp == "true" {
						cfproxy = "\tcloudflare_proxy=true"
					}
				}

				fmt.Fprintf(w, "%s\t%s\t%d\tIN\t%s\t%s%s\n",
					rec.NameFQDN, rec.Name, rec.TTL, rec.Type, rec.GetTargetCombined(), cfproxy)
			}

		default:
			return fmt.Errorf("format %q unknown", args.OutputFormat)
		}
	}
	return nil
}

func formatDsl(zonename string, rec *models.RecordConfig, defaultTTL uint32) string {

	target := rec.GetTargetCombined()

	ttl := uint32(0)
	ttlop := ""
	if rec.TTL != defaultTTL && rec.TTL != 0 {
		ttl = rec.TTL
		ttlop = fmt.Sprintf(", TTL(%d)", ttl)
	}

	cfproxy := ""
	if cp, ok := rec.Metadata["cloudflare_proxy"]; ok {
		if cp == "true" {
			cfproxy = ", CF_PROXY_ON"
		}
	}

	switch rec.Type { // #rtype_variations
	case "CAA":
		return makeCaa(rec, ttlop)
	case "MX":
		target = fmt.Sprintf("%d, '%s'", rec.MxPreference, rec.GetTargetField())
	case "SSHFP":
		target = fmt.Sprintf("%d, %d, '%s'", rec.SshfpAlgorithm, rec.SshfpFingerprint, rec.GetTargetField())
	case "SOA":
		rec.Type = "//SOA"
		target = fmt.Sprintf("'%s', '%s', %d, %d, %d, %d, %d", rec.GetTargetField(), rec.SoaMbox, rec.SoaSerial, rec.SoaRefresh, rec.SoaRetry, rec.SoaExpire, rec.SoaMinttl)
	case "SRV":
		target = fmt.Sprintf("%d, %d, %d, '%s'", rec.SrvPriority, rec.SrvWeight, rec.SrvPort, rec.GetTargetField())
	case "TLSA":
		target = fmt.Sprintf("%d, %d, %d, '%s'", rec.TlsaUsage, rec.TlsaSelector, rec.TlsaMatchingType, rec.GetTargetField())
	case "TXT":
		if len(rec.TxtStrings) == 1 {
			target = `'` + rec.TxtStrings[0] + `'`
		} else {
			target = `['` + strings.Join(rec.TxtStrings, `', '`) + `']`
		}
		// TODO(tlim): If this is an SPF record, generate a SPF_BUILDER().
	case "NS":
		// NS records at the apex should be NAMESERVER() records.
		// DnsControl uses the API to get this info. NAMESERVER() is just
		// to override that when needed.
		if rec.Name == "@" {
			return fmt.Sprintf("//NAMESERVER('%s')", target)
		}
		target = "'" + target + "'"
	case "R53_ALIAS":
		return makeR53alias(rec, ttl)
	default:
		target = "'" + target + "'"
	}

	return fmt.Sprintf("%s('%s', %s%s%s)", rec.Type, rec.Name, target, cfproxy, ttlop)
}

func makeCaa(rec *models.RecordConfig, ttlop string) string {
	var target string
	if rec.CaaFlag == 128 {
		target = fmt.Sprintf("'%s', '%s', CAA_CRITICAL", rec.CaaTag, rec.GetTargetField())
	} else {
		target = fmt.Sprintf("'%s', '%s'", rec.CaaTag, rec.GetTargetField())
	}
	return fmt.Sprintf("%s('%s', %s%s)", rec.Type, rec.Name, target, ttlop)

	// TODO(tlim): Generate a CAA_BUILDER() instead?
}

func makeR53alias(rec *models.RecordConfig, ttl uint32) string {
	items := []string{
		"'" + rec.Name + "'",
		"'" + rec.R53Alias["type"] + "'",
		"'" + rec.GetTargetField() + "'",
	}
	if z, ok := rec.R53Alias["zone_id"]; ok {
		items = append(items, "R53_ZONE('"+z+"')")
	}
	if ttl != 0 {
		items = append(items, fmt.Sprintf("TTL(%d)", ttl))
	}
	return rec.Type + "(" + strings.Join(items, ", ") + ")"
}
