package prettyzone

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

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/miekg/dns"
)

type zoneGenData struct {
	Origin     string
	DefaultTTL uint32
	Records    models.Records
}

func (z *zoneGenData) Len() int      { return len(z.Records) }
func (z *zoneGenData) Swap(i, j int) { z.Records[i], z.Records[j] = z.Records[j], z.Records[i] }
func (z *zoneGenData) Less(i, j int) bool {
	a, b := z.Records[i], z.Records[j]

	// Sort by name.
	compA, compB := a.NameFQDN, b.NameFQDN
	if compA != compB {
		if a.Name == "@" {
			compA = "@"
		}
		if b.Name == "@" {
			compB = "@"
		}
		return zoneLabelLess(compA, compB)
	}

	// sub-sort by type
	if a.Type != b.Type {
		return zoneRrtypeLess(a.Type, b.Type)
	}

	// sub-sort within type:
	switch a.Type { // #rtype_variations
	case "A":
		ta2, tb2 := a.GetTargetIP(), b.GetTargetIP()
		ipa, ipb := ta2.To4(), tb2.To4()
		if ipa == nil || ipb == nil {
			log.Fatalf("should not happen: IPs are not 4 bytes: %#v %#v", ta2, tb2)
		}
		return bytes.Compare(ipa, ipb) == -1
	case "AAAA":
		ta2, tb2 := a.GetTargetIP(), b.GetTargetIP()
		ipa, ipb := ta2.To16(), tb2.To16()
		if ipa == nil || ipb == nil {
			log.Fatalf("should not happen: IPs are not 16 bytes: %#v %#v", ta2, tb2)
		}
		return bytes.Compare(ipa, ipb) == -1
	case "MX":
		// sort by priority. If they are equal, sort by Mx.
		if a.MxPreference != b.MxPreference {
			return a.MxPreference < b.MxPreference
		}
		return a.GetTargetField() < b.GetTargetField()
	case "SRV":
		//ta2, tb2 := a.(*dns.SRV), b.(*dns.SRV)
		pa, pb := a.SrvPort, b.SrvPort
		if pa != pb {
			return pa < pb
		}
		pa, pb = a.SrvPriority, b.SrvPriority
		if pa != pb {
			return pa < pb
		}
		pa, pb = a.SrvWeight, b.SrvWeight
		if pa != pb {
			return pa < pb
		}
	case "PTR":
		//ta2, tb2 := a.(*dns.PTR), b.(*dns.PTR)
		pa, pb := a.GetTargetField(), b.GetTargetField()
		if pa != pb {
			return pa < pb
		}
	case "CAA":
		//ta2, tb2 := a.(*dns.CAA), b.(*dns.CAA)
		// sort by tag
		pa, pb := a.CaaTag, b.CaaTag
		if pa != pb {
			return pa < pb
		}
		// then flag
		fa, fb := a.CaaFlag, b.CaaFlag
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
func mostCommonTTL(records models.Records) uint32 {
	// Index the TTLs in use:
	d := make(map[uint32]int)
	for _, r := range records {
		if r.Type != "NS" {
			d[r.TTL]++
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

// WriteZoneFileRR is a helper for when you have []dns.RR instead of models.Records
func WriteZoneFileRR(w io.Writer, records []dns.RR, origin string, serial uint32) error {
	return WriteZoneFileRC(w, models.RRstoRCs(records, origin, serial), origin)
}

// WriteZoneFileRC writes a beautifully formatted zone file.
func WriteZoneFileRC(w io.Writer, records models.Records, origin string) error {
	// This function prioritizes beauty over output size.
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
		Origin:     origin + ".",
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

		// name
		nameShort := rr.Name
		name := nameShort
		if i > 0 && nameShort == nameShortPrevious {
			name = ""
		} else {
			name = nameShort
		}
		nameShortPrevious = nameShort

		// ttl
		ttl := ""
		if rr.TTL != z.DefaultTTL && rr.TTL != 0 {
			ttl = fmt.Sprint(rr.TTL)
		}

		// type
		typeStr := rr.Type

		// items[4]: the remaining line
		target := rr.GetTargetCombined()

		fmt.Fprintln(w, formatLine([]int{10, 5, 2, 5, 0}, []string{name, ttl, "IN", typeStr, target}))
		//f := formatLine([]int{10, 5, 2, 5, 0}, []string{name, ttl, "IN", typeStr, target})
		//fmt.Printf("LINE: %v", f)
		//fmt.Fprintln(w, f)
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

func zoneRrtypeLess(a, b string) bool {
	// Compare two RR types for the purpose of sorting the RRs in a Zone.

	if a == b {
		return false
	}

	// List SOAs, NSs, etc. then all others alphabetically.

	for _, t := range []string{"SOA", "NS", "CNAME",
		"A", "AAAA", "MX", "SRV", "TXT",
	} {
		if a == t {
			return true
		}
		if b == t {
			return false
		}
	}
	return a < b
}
