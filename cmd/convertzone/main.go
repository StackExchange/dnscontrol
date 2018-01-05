package main

/*
convertzone: Read BIND-style zonefile and output.

     convertzone [-mode=MODE] zonename [filename]

     -mode=tsv   TAB-separated values (default)
     -mode=dsl   DNSControl DSL
     -mode=pretty   Sort and pretty-print records.

     zonename    The FQDN of the zone name.
     filename    File to read (default: stdin)
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

	"github.com/StackExchange/dnscontrol/providers/bind"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
	"github.com/pkg/errors"
)

var flagMode = flag.String("mode", "tsv", "tsv|dsl|pretty")
var flagDefaultTTL = flag.Uint("ttl", 300, "Default TTL")
var flagRegText = flag.String("registrar", "REG_FILL_IN", "registrar text")
var flagProviderText = flag.String("provider", "DNS_FILL_IN", "provider text")

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
			return "", "", nil, errors.Wrapf(err, "Could not open file: %s", filename)
		}
	} else {
		return "", "", nil, fmt.Errorf("too many command line parameters")
	}

	return zonename, filename, r, nil
}

// pretty outputs the zonefile using the prettyprinter.
func pretty(zonename string, filename string, r io.Reader, defaultTTL uint32) {
	var l []dns.RR
	for x := range dns.ParseZone(r, zonename, filename) {
		if x.Error == nil {
			l = append(l, x.RR)
		}
	}
	bind.WriteZoneFile(os.Stdout, l, zonename)
}

// rrFormat outputs the zonefile in either DSL or TSV format.
func rrFormat(zonename string, filename string, r io.Reader, defaultTTL uint32, dsl bool) {
	zonenamedot := zonename + "."

	for x := range dns.ParseZone(r, zonename, filename) {
		if x.Error != nil {
			continue
		}

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

		// NS records at the apex should be NAMESERVER() records.
		if hdr.Rrtype == dns.TypeNS && name == "@" {
			typeStr = "NAMESERVER"
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
				// Leave target as-is.
				//				if len(
				//				target =
			default:
				target = "'" + target + "'"
			}
			if hdr.Ttl == defaultTTL {
				ttl = ""
			} else {
				ttl = fmt.Sprintf(", TTL(%d)", hdr.Ttl)
			}
			fmt.Printf(",\n\t%s('%s', %s%s)", typeStr, name, target, ttl)
		}
	}

}

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

	switch *flagMode {
	case "pretty":
		pretty(zonename, filename, reader, defTTL)
	case "dsl":
		fmt.Printf(`D("%s", %s, DnsProvider(%s)`, zonename, *flagRegText, *flagProviderText)
		rrFormat(zonename, filename, reader, defTTL, true)
		fmt.Println("\n)")
	case "tsv":
		rrFormat(zonename, filename, reader, defTTL, false)
	default:
		fmt.Println("convertzone [-flags] ZONENAME FILENAME")
		flag.Usage()
	}

}
