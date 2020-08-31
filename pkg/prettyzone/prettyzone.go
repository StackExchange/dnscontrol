package prettyzone

// Generate zonefiles.
// This generates a zonefile that prioritizes beauty over efficiency.

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/miekg/dns"
)

// MostCommonTTL returns the most common TTL in a set of records. If there is
// a tie, the highest TTL is selected. This makes the results consistent.
// NS records are not included in the analysis because Tom said so.
func MostCommonTTL(records models.Records) uint32 {
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
func WriteZoneFileRR(w io.Writer, records []dns.RR, origin string) error {
	return WriteZoneFileRC(w, models.RRstoRCs(records, origin), origin, 0, nil)
}

// WriteZoneFileRC writes a beautifully formatted zone file.
func WriteZoneFileRC(w io.Writer, records models.Records, origin string, defaultTTL uint32, comments []string) error {
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

	if defaultTTL == 0 {
		defaultTTL = MostCommonTTL(records)
	}

	z := PrettySort(records, origin, defaultTTL, comments)

	return z.generateZoneFileHelper(w)
}

// PrettySort sorts the records in a pretty order.
func PrettySort(records models.Records, origin string, defaultTTL uint32, comments []string) *ZoneGenData {
	if defaultTTL == 0 {
		defaultTTL = MostCommonTTL(records)
	}
	z := &ZoneGenData{
		Origin:     origin + ".",
		DefaultTTL: defaultTTL,
		Comments:   comments,
	}
	if z.DefaultTTL == 0 {
		z.DefaultTTL = 300
	}
	z.Records = nil
	z.Records = append(z.Records, records...)
	return z
}

// generateZoneFileHelper creates a pretty zonefile.
func (z *ZoneGenData) generateZoneFileHelper(w io.Writer) error {

	nameShortPrevious := ""

	sort.Sort(z)
	if z.DefaultTTL == 0 {
		z.DefaultTTL = 300
	}
	fmt.Fprintln(w, "$TTL", z.DefaultTTL)
	for _, comment := range z.Comments {
		for _, line := range strings.Split(comment, "\n") {
			if line != "" {
				fmt.Fprintln(w, ";", line)
			}
		}
	}
	for i, rr := range z.Records {

		// Fake types are commented out.
		prefix := ""
		_, ok := dns.StringToType[rr.Type]
		if !ok {
			prefix = ";"
		}

		// name
		nameShort := rr.Name
		name := nameShort
		if (prefix == "") && (i > 0 && nameShort == nameShortPrevious) {
			name = ""
		}
		nameShortPrevious = nameShort

		// ttl
		ttl := ""
		if rr.TTL != z.DefaultTTL && rr.TTL != 0 {
			ttl = fmt.Sprint(rr.TTL)
		}

		// type
		typeStr := rr.Type

		// the remaining line
		target := rr.GetTargetCombined()

		fmt.Fprintf(w, "%s%s\n",
			prefix, formatLine([]int{10, 5, 2, 5, 0}, []string{name, ttl, "IN", typeStr, target}))
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
