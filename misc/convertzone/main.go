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
	"errors"
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
)

var output_mode = flag.String("mode", "tsv", "tsv|dsl|pretty")
var flag_defaultTtl = flag.Uint("ttl", 300, "Default TTL")
var flag_registrar = flag.String("registrar", "REG_FILL_IN", "registrar text")
var flag_provider = flag.String("provider", "DNS_FILL_IN", "provider text")

// parseargs parses the non-flag arguments.
func parseargs(args []string) (zonename string, filename string, r io.Reader, err error) {
	// 1 args: first arg is the zonename. Read stdin.
	// 2 args: first arg is the zonename. 2nd is the filename.

	if len(args) < 1 {
		return "", "", nil, errors.New("Not enough parameters")
	}

	zonename = args[0]

	if len(args) < 2 {
		filename = "stdin"
		r = bufio.NewReader(os.Stdin)
	} else {
		filename = flag.Arg(1)
		r, err = os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
	}

	return zonename, filename, r, nil
}

// pretty outputs the zonefile using the prettyprinter.
func pretty(zonename string, filename string, r io.Reader, defaultTtl uint32) {
	var pretty_list []dns.RR
	for x := range dns.ParseZone(r, zonename, filename) {
		if x.Error == nil {
			pretty_list = append(pretty_list, x.RR)
		}
	}
	bind.WriteZoneFile(bufio.NewWriter(os.Stdout), pretty_list, zonename, defaultTtl)
}

// rr_format outputs the zonefile in either DSL or TSV format.
func rr_format(zonename string, filename string, r io.Reader, defaultTtl uint32, dsl bool) {
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

		if !dsl { // TSV format:
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", name, ttl, classStr, typeStr, target)
		} else { // DSL format:
			switch hdr.Rrtype {
			case dns.TypeMX:
				mxi := strings.SplitN(target, "\t", 2)
				target = mxi[0] + ", '" + mxi[1] + "'"
			case dns.TypeSOA:
				continue
			case dns.TypeTXT:
				// Leave target as-is.
			default:
				target = "'" + target + "'"
			}
			if hdr.Ttl == defaultTtl {
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
		flag.Usage()
	}

	defttl := uint32(*flag_defaultTtl)

	switch *output_mode {
	case "pretty":
		pretty(zonename, filename, reader, defttl)
	case "dsl":
		fmt.Printf(`D("%s", %s, DnsProvider(%s)`, zonename, *flag_registrar, *flag_provider)
		rr_format(zonename, filename, reader, defttl, true)
		fmt.Println("\n)")
	case "tsv":
		rr_format(zonename, filename, reader, defttl, false)
	default:
		flag.Usage()
	}

}
