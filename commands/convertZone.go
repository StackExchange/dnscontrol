package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/pkg/printer"
	"github.com/StackExchange/dnscontrol/providers/bind"
	"github.com/StackExchange/dnscontrol/providers/octodns/octoyaml"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var _ = cmd(catUtils, func() *cli.Command {
	var opts ConvertZoneOptions
	return &cli.Command{
		Name:      "convert-zone",
		Usage:     "Read and write DNS zone files.",
		ArgsUsage: "zonename (filename|stdin)",
		Action: func(ctx *cli.Context) error {
			return exit(ConvertZoneAction(ctx, opts))
		},
		Flags: opts.flags(),
	}
}())

type ConvertZoneOptions struct {
	InFormat  string
	OutFormat string
	TTL       uint
	Registrar string
	Provider  string
}

func (opts *ConvertZoneOptions) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "in",
			Value:       "bind",
			Destination: &opts.InFormat,
			Usage:       "input format to use (bind|octodns)",
		},
		cli.StringFlag{
			Name:        "out",
			Value:       "dsl",
			Destination: &opts.OutFormat,
			Usage:       "output format to use (dsl|tsv|pretty)",
		},
		cli.UintFlag{
			Name:        "ttl",
			Value:       300,
			Destination: &opts.TTL,
			Usage:       "TTL value",
		},
		cli.StringFlag{
			Name:        "registrar",
			Value:       "REG_FILL_IN",
			Destination: &opts.Registrar,
			Usage:       "registrar name",
		},
		cli.StringFlag{
			Name:        "provider",
			Value:       "DNS_FILL_IN",
			Destination: &opts.Provider,
			Usage:       "provider name",
		},
	}
}

func ConvertZoneAction(ctx *cli.Context, opts ConvertZoneOptions) error {
	// 1 args: first arg is the zonename. Read stdin.
	// 2 args: first arg is the zonename. 2nd is the filename.
	// Anything else returns an error.
	var zonename string
	var filename string
	var reader io.Reader
	var err error

	if len(ctx.Args()) < 1 {
		return errors.Errorf("no command line parameters. Zone name required")
	}

	zonename = ctx.Args().Get(0)

	if len(ctx.Args()) == 1 {
		filename = "stdin"
		reader = bufio.NewReader(os.Stdin)
	} else if len(ctx.Args()) == 2 {
		filename = ctx.Args().Get(1)
		reader, err = os.Open(filename)
		if err != nil {
			return errors.Wrapf(err, "Could not open file: %s", filename)
		}
	} else {
		return errors.Errorf("too many command line parameters")
	}

	return ConvertZone(filename, reader, zonename, opts, printer.ConsolePrinter{})
}

func ConvertZone(filename string, reader io.Reader, zonename string, opts ConvertZoneOptions, out printer.CLI) error {
	var recs []dns.RR
	var err error

	// Read it in:

	switch opts.InFormat {
	case "bind", "zone":
		recs, err = readZone(zonename, reader, filename)
	case "oct", "octo", "octodns":
		recs, err = readOctodns(zonename, reader, filename)
	default:
		out.Debugf("unrecognized input format: %s", opts.InFormat)
		return nil
	}
	if err != nil {
		return err
	}

	// Write it out:

	switch opts.OutFormat {
	case "pretty":
		bind.WriteZoneFile(os.Stdout, recs, zonename)
	case "dsl":
		out.Debugf(`D("%s", %s, DnsProvider(%s)`, zonename, opts.Registrar, opts.Provider)
		rrFormat(zonename, filename, recs, opts.TTL, true)
		out.Debugf("\n)\n")
	case "tsv":
		rrFormat(zonename, filename, recs, opts.TTL, false)
	default:
		out.Debugf("convertzone [-flags] ZONENAME FILENAME")
		// flag.Usage()
	}
	return nil
}

func readZone(zonename string, r io.Reader, filename string) ([]dns.RR, error) {
	var l []dns.RR
	for x := range dns.ParseZone(r, zonename, filename) {
		if x.Error != nil {
			return nil, x.Error
		} else {
			l = append(l, x.RR)
		}
	}
	return l, nil
}

func readOctodns(zonename string, r io.Reader, filename string) ([]dns.RR, error) {
	var l []dns.RR

	foundRecords, err := octoyaml.ReadYaml(r, zonename)
	if err != nil {
		return nil, errors.Wrapf(err, "can not get corrections")
	}

	for _, x := range foundRecords {
		l = append(l, x.ToRR())
	}
	return l, nil
}

// pretty outputs the zonefile using the prettyprinter.
func writePretty(zonename string, recs []dns.RR, defaultTTL uint) {
	bind.WriteZoneFile(os.Stdout, recs, zonename)
}

// rrFormat outputs the zonefile in either DSL or TSV format.
func rrFormat(zonename string, filename string, recs []dns.RR, defaultTTL uint, dsl bool) error {
	zonenamedot := zonename + "."

	for _, x := range recs {

		// Skip comments. Parse the formatted version.
		line := x.String()
		if line[0] == ';' {
			continue
		}
		items := strings.SplitN(line, "\t", 5)
		if len(items) < 5 {
			return errors.Errorf("Too few items in: %v", line)
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
		if hdr.Ttl == uint32(defaultTTL) {
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
	return nil
}
