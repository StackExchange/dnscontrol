package octoyaml

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sort"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/natsort"
	"github.com/miekg/dns/dnsutil"
)

type genYamlData struct {
	Origin     string
	DefaultTTL uint32
	Records    models.Records
}

func sortRecs(recs models.Records, origin string) {
	z := &genYamlData{
		Origin:  dnsutil.AddOrigin(origin, "."),
		Records: recs,
	}
	sort.Sort(z)
}

func (z *genYamlData) Len() int      { return len(z.Records) }
func (z *genYamlData) Swap(i, j int) { z.Records[i], z.Records[j] = z.Records[j], z.Records[i] }
func (z *genYamlData) Less(i, j int) bool {
	a, b := z.Records[i], z.Records[j]
	compA, compB := a.GetLabel(), b.GetLabel()
	if compA != compB {
		if compA == z.Origin+"." {
			compA = "@"
		}
		if compB == z.Origin+"." {
			compB = "@"
		}
		return zoneLabelLess(compA, compB)
	}
	rrtypeA, rrtypeB := a.Type, b.Type
	if rrtypeA != rrtypeB {
		return zoneRrtypeLess(rrtypeA, rrtypeB)
	}
	switch rrtypeA { // #rtype_variations
	case "NS", "TXT", "TLSA":
		// pass through.
	case "A":
		ta2, tb2 := net.ParseIP(a.GetTargetField()), net.ParseIP(b.GetTargetField())
		ipa, ipb := ta2.To4(), tb2.To4()
		if ipa == nil || ipb == nil {
			log.Fatalf("should not happen: IPs are not 4 bytes: %#v %#v", ta2, tb2)
		}
		return bytes.Compare(ipa, ipb) == -1
	case "AAAA":
		ta2, tb2 := net.ParseIP(a.GetTargetField()), net.ParseIP(b.GetTargetField())
		ipa, ipb := ta2.To16(), tb2.To16()
		return bytes.Compare(ipa, ipb) == -1
	case "MX":
		pa, pb := a.MxPreference, b.MxPreference
		return pa < pb
	case "SRV":
		pa, pb := a.SrvPort, b.SrvPort
		if pa != pb {
			return pa < pb
		}
		pa, pb = a.SrvPriority, b.SrvPriority
		if pa != pb {
			return pa < pb
		}
		pa, pb = a.SrvWeight, a.SrvWeight
		if pa != pb {
			return pa < pb
		}
	case "PTR":
		pa, pb := a.GetTargetField(), b.GetTargetField()
		if pa != pb {
			return pa < pb
		}
	case "CAA":
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
		panic(fmt.Sprintf("genYamlData Less: unimplemented rtype %v", a.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	return a.GetTargetSortable() < b.GetTargetSortable()
}

func zoneLabelLess(a, b string) bool {
	return natsort.Less(a, b)
	// octodns-validate wants a "natural sort" (i.e. foo10 comes after foo3).
	// We emulate this with the natsort package.
	// If you need to disable that validatation:
	//    Edit env/lib/python2.7/site-packages/octodns/yaml.py
	//    Change line 27: OLD: if key != expected
	//                    NEW: if False and key != expected
}

func zoneRrtypeLess(a, b string) bool {
	// Compare two RR types for the purpose of sorting the RRs in a Zone.

	// If they are equal, we are done. All other code is simplified
	// because we can assume a!=b.
	if a == b {
		return false
	}

	// List SOAs, then NSs, then all others.
	// i.e. SOA is always less. NS is less than everything but SOA.
	if a == "SOA" {
		return true
	}
	if b == "SOA" {
		return false
	}
	if a == "NS" {
		return true
	}
	if b == "NS" {
		return false
	}
	return a < b
}
