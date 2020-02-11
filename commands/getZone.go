package commands

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v2/providers"
	"github.com/StackExchange/dnscontrol/v2/providers/bind"
	"github.com/StackExchange/dnscontrol/v2/providers/config"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
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

	// Read it in:

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
	zonename := args.ZoneName

	provider, err := providers.CreateDNSProvider(args.ProviderName, providerConfigs[args.CredName], nil)
	if err != nil {
		return err
	}
	//fmt.Printf("FOO=%+v\n", provider)

	recs, err := provider.GetZoneRecords(zonename)
	//fmt.Printf("RECS=%v\n", recs)
	//fmt.Printf("err = %v\n", err)
	if err != nil {
		return err
	}

	// Write it out:

	switch args.OutputFormat {
	case "pretty":
		bind.WriteZoneFile(os.Stdout, recs, zonename)
	case "dsl":
		fmt.Printf(`var CHANGEME = NewDnsProvider("%s", "%s");`+"\n",
			args.CredName, args.ProviderName)
		fmt.Printf(`D("%s", REG_CHANGEME,`+"\n", zonename)
		fmt.Printf(`  DnsProvider(CHANGEME),` + "\n")
		rrFormat(zonename, args.OutputFile, recs, defTTL, true)
		fmt.Println("\n)")
	case "tsv":
		rrFormat(zonename, args.OutputFile, recs, defTTL, false)
	default:
		return fmt.Errorf("format %q unknown", args.OutputFile)
	}

	return nil
}

// rrFormat outputs the zonefile in either DSL or TSV format.
func rrFormat(zonename string, filename string, recs []dns.RR, defaultTTL uint32, dsl bool) {
	zonenamedot := zonename + "."

	for _, x := range recs {

		// Skip comments. Parse the formatted version.
		line := x.String()
		if line[0] == ';' {
			continue
		}
		items := strings.SplitN(line, "\t", 5)
		if len(items) < 5 {
			log.Fatalf("Too few items in: %v", line)
		}

		target := items[4]

		hdr := x.Header()
		nameFqdn := hdr.Name
		name := dnsutil.TrimDomainName(nameFqdn, zonenamedot)
		ttl := strconv.FormatUint(uint64(hdr.Ttl), 10)
		classStr := dns.ClassToString[hdr.Class]
		typeStr := dns.TypeToString[hdr.Rrtype]

		// MX records should split out the prio vs. target.
		if hdr.Rrtype == dns.TypeMX {
			target = strings.Replace(target, " ", "\t", 1)
		}

		var ttlop string
		if hdr.Ttl == defaultTTL {
			ttlop = ""
		} else {
			ttlop = fmt.Sprintf(", TTL(%d)", hdr.Ttl)
		}

		// NS records at the apex should be NAMESERVER() records.
		if hdr.Rrtype == dns.TypeNS && name == "@" {
			fmt.Printf(",\n\tNAMESERVER('%s'%s)", target, ttlop)
			continue
		}

		if !dsl { // TSV format:
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", name, ttl, classStr, typeStr, target)
		} else { // DSL format:
			switch hdr.Rrtype { // #rtype_variations
			case dns.TypeMX:
				m := strings.SplitN(target, "\t", 2)
				target = m[0] + ", '" + m[1] + "'"
			case dns.TypeSOA:
				continue
			case dns.TypeTXT:
				if len(x.(*dns.TXT).Txt) == 1 {
					target = `'` + x.(*dns.TXT).Txt[0] + `'`
				} else {
					target = `['` + strings.Join(x.(*dns.TXT).Txt, `', '`) + `']`
				}
			default:
				target = "'" + target + "'"
			}
			fmt.Printf(",\n\t%s('%s', %s%s)", typeStr, name, target, ttlop)
		}
	}

}
