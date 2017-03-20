package main

/*
convertzone: Read BIND-style zonefile and output.

     convertzone [-mode=MODE] zonename [filename]

     -mode=tsv   TAB-separated values (default)
     -mode=dsl   DNSControl DSL
     -mode=pretty   Sort and pretty-print records.

     zonename    The FQDN of the zone name.
     filename    File to read (default: stdin)

mode=tsv:

This is useful for AWK and other systems that deal best with a
uniform set of input.

Example: Print all CNAMEs:

    convertzone -mode=tsv foo.com <zone.foo.com | awk '$4 == "CNAME" { print $1 " -> " $5 }'


mode=pretty:

This is useful for cleaning up a zonefile. It sorts the records,
moving SOA and NS records to the top of the zone; all other records
are alphabetically sorted; if a label has mutiple records, they are
listed in a logical (not numeric) order, multiple A records are
listed sorted by IP address, MX records are sorted by priority,
etc.  Use -ttl to set a default TTL.

Example: Clean up a zone file:

    convertzone -mode=pretty foo.com <old/zone.foo.com >new/zone.foo.com


mode=dsl:

This is useful for generating your draft dnsconfig.js configuration.
Pass the old zone through this program with -mode=dsl and append
it to your dnsconfig.js file. You'll probably need to clean it up
a bit: remove NS records (DnsProvider() inserts NS records for you,
change the order to be more logical and readable, manually check
over the results.

When converting a zonefile to DSL, we recommend first doing a straightforward
conversion, do not change any records at this time.  Now you can run `dnscontrol preview`
to verify that dnsconfig.js file is correct, and you will see that it has found zero
changes are needed. That means you have done the conversion property.  After that step,
do any cleanups you'd like to do (remove obsolete records, etc.).  If you do such cleanups
earlier in the process you can't be entirely sure you've done the conversion correctly.

Example: Convert a zone filem and add it to your configuration:

    convertzone -mode=dsl foo.com <old/zone.foo.com >>dnsconfig.js
		# Do these next two steps until you've fixed all errors.
		dnscontrol preview
		vim dnsconfig.js
		# When everything is as you wish, push the changes live:
		dnscontrol push

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

	if len(args) < 1 {
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
	bind.WriteZoneFile(os.Stdout, l, zonename, defaultTTL)
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

		if !dsl { // TSV format:
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", name, ttl, classStr, typeStr, target)
		} else { // DSL format:
			switch hdr.Rrtype {
			case dns.TypeMX:
				m := strings.SplitN(target, "\t", 2)
				target = m[0] + ", '" + m[1] + "'"
			case dns.TypeSOA:
				continue
			case dns.TypeTXT:
				// Leave target as-is.
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
		flag.Usage()
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
		flag.Usage()
	}

}
