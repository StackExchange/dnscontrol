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
	"strings"

	"github.com/StackExchange/dnscontrol/providers/bind"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
)

var output_mode = flag.String("mode", "tsv", "tsv|dsl")
var flag_defaultTtl = flag.Uint("ttl", 300, "help message")
var flag_registrar = flag.String("registrar", "REG_FILL_IN", "registrar text")
var flag_provider = flag.String("provider", "DNS_FILL_IN", "provider text")

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	zonename := flag.Arg(0)
	zonenamedot := zonename + "."

	var err error
	var reader io.Reader
	var filename string
	if flag.NArg() < 2 {
		filename = "stdin"
		reader = bufio.NewReader(os.Stdin)
	} else {
		filename = flag.Arg(1)
		reader, err = os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
	}

	var pretty_list []dns.RR

	switch *output_mode {
	case "dsl":
		fmt.Printf("D(\"%s\", %s, DnsProvider(%s)", zonename, *flag_registrar, *flag_provider)
	default:
	}

	for x := range dns.ParseZone(reader, zonename, filename) {
		if x.Error == nil {

			line := x.String()
			if line[0] == ';' {
				continue
			}

			items := strings.SplitN(line, "\t", 5)
			if len(items) < 5 {
				log.Fatalf("Too few items in: %v", line)
			}

			hdr := x.Header()
			nameFqdn := hdr.Name
			name := dnsutil.TrimDomainName(nameFqdn, zonenamedot)
			ttl := fmt.Sprintf("%d", hdr.Ttl)
			classStr := items[2]
			typeStr := dns.TypeToString[hdr.Rrtype]
			target := items[4]

			// MX records should split out the prio vs. target.
			if hdr.Rrtype == dns.TypeMX {
				target = strings.Replace(target, " ", "\t", 1)
			}

			switch *output_mode {

			case "tsv":
				fmt.Printf("%s\t%s\t%s\t%s\t%s\n", name, ttl, classStr, typeStr, target)

			case "dsl":
				switch hdr.Rrtype {
				case dns.TypeSOA:
					continue
				case dns.TypeTXT:
				case dns.TypeMX:
					mxi := strings.SplitN(target, "\t", 2)
					target = mxi[0] + ", '" + mxi[1] + "'"
				default:
					target = "'" + target + "'"
				}
				if int64(hdr.Ttl) == int64(*flag_defaultTtl) {
					ttl = ""
				} else {
					ttl = fmt.Sprintf(", TTL(%d)", hdr.Ttl)
				}
				fmt.Printf(",\n\t%s('%s', %s%s)", typeStr, name, target, ttl)

			case "pretty":
				pretty_list = append(pretty_list, x.RR)
			}

		}
	}

	switch *output_mode {
	case "dsl":
		fmt.Println("\n)")
	case "pretty":

		bind.WriteZoneFile(bufio.NewWriter(os.Stdout), pretty_list, zonename, uint32(*flag_defaultTtl))
	default:
	}
}
