package bind

// Generate zonefiles.
// This generates a zonefile that prioritizes beauty over efficiency.

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
)

type zoneGenData struct {
	Origin     string
	DefaultTTL uint32
	Records    []dns.RR
}

func (z *zoneGenData) Len() int      { return len(z.Records) }
func (z *zoneGenData) Swap(i, j int) { z.Records[i], z.Records[j] = z.Records[j], z.Records[i] }
func (z *zoneGenData) Less(i, j int) bool {
	a, b := z.Records[i], z.Records[j]
	compA, compB := dnsutil.AddOrigin(a.Header().Name, z.Origin+"."), dnsutil.AddOrigin(b.Header().Name, z.Origin+".")
	if compA != compB {
		if compA == z.Origin+"." {
			compA = "@"
		}
		if compB == z.Origin+"." {
			compB = "@"
		}
		return zoneLabelLess(compA, compB)
	}
	rrtypeA, rrtypeB := a.Header().Rrtype, b.Header().Rrtype
	if rrtypeA != rrtypeB {
		return zoneRrtypeLess(rrtypeA, rrtypeB)
	}
	switch rrtypeA { // #rtype_variations
	case dns.TypeA:
		ta2, tb2 := a.(*dns.A), b.(*dns.A)
		ipa, ipb := ta2.A.To4(), tb2.A.To4()
		if ipa == nil || ipb == nil {
			log.Fatalf("should not happen: IPs are not 4 bytes: %#v %#v", ta2, tb2)
		}
		return bytes.Compare(ipa, ipb) == -1
	case dns.TypeAAAA:
		ta2, tb2 := a.(*dns.AAAA), b.(*dns.AAAA)
		ipa, ipb := ta2.AAAA.To16(), tb2.AAAA.To16()
		return bytes.Compare(ipa, ipb) == -1
	case dns.TypeMX:
		ta2, tb2 := a.(*dns.MX), b.(*dns.MX)
		pa, pb := ta2.Preference, tb2.Preference
		// sort by priority. If they are equal, sort by Mx.
		if pa != pb {
			return pa < pb
		}
		return ta2.Mx < tb2.Mx
	case dns.TypeSRV:
		ta2, tb2 := a.(*dns.SRV), b.(*dns.SRV)
		pa, pb := ta2.Port, tb2.Port
		if pa != pb {
			return pa < pb
		}
		pa, pb = ta2.Priority, tb2.Priority
		if pa != pb {
			return pa < pb
		}
		pa, pb = ta2.Weight, tb2.Weight
		if pa != pb {
			return pa < pb
		}
	case dns.TypePTR:
		ta2, tb2 := a.(*dns.PTR), b.(*dns.PTR)
		pa, pb := ta2.Ptr, tb2.Ptr
		if pa != pb {
			return pa < pb
		}
	case dns.TypeCAA:
		ta2, tb2 := a.(*dns.CAA), b.(*dns.CAA)
		// sort by tag
		pa, pb := ta2.Tag, tb2.Tag
		if pa != pb {
			return pa < pb
		}
		// then flag
		fa, fb := ta2.Flag, tb2.Flag
		if fa != fb {
			// flag set goes before ones without flag set
			return fa > fb
		}
	default:
		// pass through. String comparison is sufficient.
	}
	return a.String() < b.String()
}

// mostCommonTTL returns the most common TTL in a set of records. If there is
// a tie, the highest TTL is selected. This makes the results consistent.
// NS records are not included in the analysis because Tom said so.
func mostCommonTTL(records []dns.RR) uint32 {
	// Index the TTLs in use:
	d := make(map[uint32]int)
	for _, r := range records {
		if r.Header().Rrtype != dns.TypeNS {
			d[r.Header().Ttl]++
		}
	}
	// Find the largest count:
	var mc int
	for _, value := range d {
		if value > mc {
			mc = value
		}
	}
	// Find the largest key with that count:
	var mk uint32
	for key, value := range d {
		if value == mc {
			if key > mk {
				mk = key
			}
		}
	}
	return mk
}

// WriteZoneFile writes a beautifully formatted zone file.
func WriteZoneFile(w io.Writer, records []dns.RR, origin string) error {
	// This function prioritizes beauty over efficiency.
	// * The zone records are sorted by label, grouped by subzones to
	//   be easy to read and pleasant to the eye.
	// * Within a label, SOA and NS records are listed first.
	// * MX records are sorted numericly by preference value.
	// * SRV records are sorted numericly by port, then priority, then weight.
	// * A records are sorted by IP address, not lexicographically.
	// * Repeated labels are removed.
	// * $TTL is used to eliminate clutter. The most common TTL value is used.
	// * "@" is used instead of the apex domain name.

	defaultTTL := mostCommonTTL(records)

	z := &zoneGenData{
		Origin:     dnsutil.AddOrigin(origin, "."),
		DefaultTTL: defaultTTL,
	}
	z.Records = nil
	for _, r := range records {
		z.Records = append(z.Records, r)
	}
	return z.generateZoneFileHelper(w)
}

// generateZoneFileHelper creates a pretty zonefile.
func (z *zoneGenData) generateZoneFileHelper(w io.Writer) error {

	nameShortPrevious := ""

	sort.Sort(z)
	fmt.Fprintln(w, "$TTL", z.DefaultTTL)
	for i, rr := range z.Records {
		line := rr.String()
		if line[0] == ';' {
			continue
		}
		hdr := rr.Header()

		items := strings.SplitN(line, "\t", 5)
		if len(items) < 5 {
			log.Fatalf("Too few items in: %v", line)
		}

		// items[0]: name
		nameFqdn := hdr.Name
		nameShort := dnsutil.TrimDomainName(nameFqdn, z.Origin)
		name := nameShort
		if i > 0 && nameShort == nameShortPrevious {
			name = ""
		} else {
			name = nameShort
		}
		nameShortPrevious = nameShort

		// items[1]: ttl
		ttl := ""
		if hdr.Ttl != z.DefaultTTL && hdr.Ttl != 0 {
			ttl = items[1]
		}

		// items[2]: class
		if hdr.Class != dns.ClassINET {
			log.Fatalf("generateZoneFileHelper: Unimplemented class=%v", items[2])
		}

		// items[3]: type
		typeStr := dns.TypeToString[hdr.Rrtype]

		// items[4]: the remaining line
		target := items[4]

		fmt.Fprintln(w, formatLine([]int{10, 5, 2, 5, 0}, []string{name, ttl, "IN", typeStr, target}))
	}
	return nil
}

func formatLine(lengths []int, fields []string) string {
	c := 0
	result := ""
	for i, length := range lengths {
		item := fields[i]
		for len(result) < c {
			result += " "
		}
		if item != "" {
			result += item + " "
		}
		c += length + 1
	}
	return strings.TrimRight(result, " ")
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func zoneLabelLess(a, b string) bool {
	// Compare two zone labels for the purpose of sorting the RRs in a Zone.

	// If they are equal, we are done. All other code is simplified
	// because we can assume a!=b.
	if a == b {
		return false
	}

	// Sort @ at the top, then *, then everything else lexigraphically.
	// i.e. @ always is less. * is is less than everything but @.
	if a == "@" {
		return true
	}
	if b == "@" {
		return false
	}
	if a == "*" {
		return true
	}
	if b == "*" {
		return false
	}

	// Split into elements and match up last elements to first. Compare the
	// first non-equal elements.

	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	ia := len(as) - 1
	ib := len(bs) - 1

	var min int
	if ia < ib {
		min = len(as) - 1
	} else {
		min = len(bs) - 1
	}

	// Skip the matching highest elements, then compare the next item.
	for i, j := ia, ib; min >= 0; i, j, min = i-1, j-1, min-1 {
		// Compare as[i] < bs[j]
		// Sort @ at the top, then *, then everything else.
		// i.e. @ always is less. * is is less than everything but @.
		// If both are numeric, compare as integers, otherwise as strings.

		if as[i] != bs[j] {

			// If the first element is *, it is always less.
			if i == 0 && as[i] == "*" {
				return true
			}
			if j == 0 && bs[j] == "*" {
				return false
			}

			// If the elements are both numeric, compare as integers:
			au, aerr := strconv.ParseUint(as[i], 10, 64)
			bu, berr := strconv.ParseUint(bs[j], 10, 64)
			if aerr == nil && berr == nil {
				return au < bu
			}
			// otherwise, compare as strings:
			return as[i] < bs[j]
		}
	}
	// The min top elements were equal, so the shorter name is less.
	return ia < ib
}

func zoneRrtypeLess(a, b uint16) bool {
	// Compare two RR types for the purpose of sorting the RRs in a Zone.

	// If they are equal, we are done. All other code is simplified
	// because we can assume a!=b.
	if a == b {
		return false
	}

	// List SOAs, then NSs, then all others.
	// i.e. SOA is always less. NS is less than everything but SOA.
	if a == dns.TypeSOA {
		return true
	}
	if b == dns.TypeSOA {
		return false
	}
	if a == dns.TypeNS {
		return true
	}
	if b == dns.TypeNS {
		return false
	}
	return a < b
}
