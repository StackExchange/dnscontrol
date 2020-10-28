package main

/*
convertzone: Read and write DNS zone files.

     convertzone [-in=INPUT] [-out=OUTPUT] zonename [filename]

	Input format:
	-in=bind      BIND-style zonefiles (DEFAULT)
    -in=octodns   OctoDNS YAML "config" files.

    Output format:

    -out=dsl      DNSControl DSL language (dnsconfig.js) (DEFAULT)
    -out=tsv      TAB-separated values
    -out=pretty   pretty-printed (BIND-style zonefiles)

    zonename    The FQDN of the zone name.
    filename    File to read (optional. Defaults to stdin)

	The DSL output format is useful for creating the first
	draft of your dnsconfig.js when importing zones from
	other services.

	The TSV format makes it easy to process a zonefile with
	shell tools.  `awk -F"\t" $2 = "A" { print $3 }`

	The PRETTY format is just a nice way to clean up a zonefile.

	If no filename is specified, stdin is assumed.
	Output is sent to stdout.

	The zonename is required as it can not be guessed automatically from the input.
*/

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
	"github.com/StackExchange/dnscontrol/v3/providers/octodns/octoyaml"
)

var flagInfmt = flag.String("in", "zone", "zone|octodns")
var flagOutfmt = flag.String("out", "dsl", "dsl|tsv|pretty")
var flagDefaultTTL = flag.Uint("ttl", 300, "Default TTL")
var flagRegText = flag.String("registrar", "REG_FILL_IN", "registrar text")
var flagProviderText = flag.String("provider", "DNS_FILL_IN", "provider text")

func main() {
	flag.Parse()
	zonename, filename, reader, err := parseargs(flag.Args())
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
		fmt.Println("convertzone [-flags] ZONENAME FILENAME")
		flag.Usage()
		os.Exit(1)
	}

	defTTL := uint32(*flagDefaultTTL)

	var recs []dns.RR

	// Read it in:

	switch *flagInfmt {
	case "zone":
		recs = readZone(zonename, reader, filename)
	case "oct", "octo", "octodns":
		recs = readOctodns(zonename, reader, filename)
	}

	// Write it out:

	switch *flagOutfmt {
	case "pretty":
		prettyzone.WriteZoneFileRR(os.Stdout, recs, zonename)
	case "dsl":
		fmt.Printf(`D("%s", %s, DnsProvider(%s)`, zonename, *flagRegText, *flagProviderText)
		rrFormat(zonename, filename, recs, defTTL, true)
		fmt.Println("\n)")
	case "tsv":
		rrFormat(zonename, filename, recs, defTTL, false)
	default:
		fmt.Println("convertzone [-flags] ZONENAME FILENAME")
		flag.Usage()
	}

}

// parseargs parses the non-flag arguments.
func parseargs(args []string) (zonename string, filename string, r io.Reader, err error) {
	// 1 args: first arg is the zonename. Read stdin.
	// 2 args: first arg is the zonename. 2nd is the filename.
	// Anything else returns an error.

	if len(args) < 2 {
		return "", "", nil, fmt.Errorf("no command line parameters. Zone name required")
	}

	zonename = args[0]

	if len(args) == 1 {
		filename = "stdin"
		r = bufio.NewReader(os.Stdin)
	} else if len(args) == 2 {
		filename = flag.Arg(1)
		r, err = os.Open(filename)
		if err != nil {
			return "", "", nil, fmt.Errorf("could not open file: %s: %w", filename, err)
		}
	} else {
		return "", "", nil, fmt.Errorf("too many command line parameters")
	}

	return zonename, filename, r, nil
}

func readZone(zonename string, r io.Reader, filename string) []dns.RR {

	zp := dns.NewZoneParser(r, zonename, filename)

	var parsed []dns.RR
	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		parsed = append(parsed, rr)
	}
	if err := zp.Err(); err != nil {
		log.Fatalf("Error in zonefile: %v", err)
	}
	return parsed
}

func readOctodns(zonename string, r io.Reader, filename string) []dns.RR {
	var l []dns.RR

	foundRecords, err := octoyaml.ReadYaml(r, zonename)
	if err != nil {
		log.Println(fmt.Errorf("can not get corrections: %w", err))
	}

	for _, x := range foundRecords {
		l = append(l, x.ToRR())
	}
	return l
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
