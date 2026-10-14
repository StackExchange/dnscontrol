package misaka

import (
	"fmt"
	"strings"

	"github.com/DNSControl/dnscontrol/v5/models"
	"github.com/DNSControl/dnscontrol/v5/pkg/diff2"
	"github.com/DNSControl/dnscontrol/v5/pkg/printer"
	"github.com/DNSControl/dnscontrol/v5/pkg/providers"
)

// hostnameTypes are the types whose value ends in a hostname.
var hostnameTypes = map[string]bool{
	"ALIAS": true,
	"CNAME": true,
	"DNAME": true,
	"MX":    true,
	"NS":    true,
	"PTR":   true,
	"SRV":   true,
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
//
// The SOA record set and the NS record set at the zone apex are left out: both
// are managed by Misaka.io. A record set that DNSControl could not write back
// unchanged makes the read fail; see checkRecordSet.
func (p *misakaProvider) GetZoneRecords(dc *models.DomainConfig) (models.Records, error) {
	sets, err := p.client.listRecordSets(dc.Name)
	if err != nil {
		return nil, err
	}

	recs := make(models.Records, 0, len(sets))
	for _, rs := range sets {
		if rs.Type == "SOA" || (rs.Type == "NS" && rs.Name == apexLabel) {
			continue
		}
		if err := checkRecordSet(rs); err != nil {
			return nil, err
		}

		before := providers.BeginToRC(p.observer, "toRC", rs)
		rcs, err := toRC(dc, rs)
		providers.EndToRC(p.observer, "toRC", before, rs, rcs, err)
		if err != nil {
			return nil, err
		}
		recs = append(recs, rcs...)
	}

	return recs, nil
}

// checkRecordSet refuses a record set that uses something DNSControl does not
// model. Reading it as plain records would let the next change to it erase
// the feature, so the user has to resolve it in the Misaka.io console first.
func checkRecordSet(rs apiRecordSet) error {
	fail := func(format string, args ...any) error {
		return fmt.Errorf("MISAKA: record set %s/%s %s", rs.Name, rs.Type, fmt.Sprintf(format, args...))
	}

	// Names are case-sensitive, verbatim identifiers on the API side while
	// DNSControl lowercases every label: a write to the lowercased name would
	// miss this set or create a second one next to it.
	if rs.Name != strings.ToLower(rs.Name) || strings.HasSuffix(rs.Name, ".") {
		return fail("has a name that is not lowercase or ends with a dot; rename it in the Misaka.io console")
	}
	if len(rs.Filters) > 0 {
		return fail("uses filters (GeoDNS), which DNSControl does not manage")
	}
	if rs.Alias != nil {
		return fail("uses alias %q, which DNSControl does not manage", *rs.Alias)
	}
	if !rs.Enabled {
		return fail("is disabled (enabled: false), which DNSControl does not manage")
	}
	if rs.Protected {
		return fail("is protected, which DNSControl does not manage")
	}
	if len(rs.Records) == 0 {
		return fail("has no values; delete it or give it values in the Misaka.io console")
	}

	// The API serves a hostname of "@" as the literal name "\@.", while
	// DNSControl would read it as the root, so it can be neither kept nor
	// rewritten truthfully.
	if hostnameTypes[rs.Type] {
		for _, r := range rs.Records {
			fields := strings.Fields(r.Value)
			if len(fields) > 0 && fields[len(fields)-1] == apexLabel {
				return fail("has the hostname value %q; replace \"@\" with the zone's full name in the Misaka.io console", r.Value)
			}
		}
	}

	return nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (p *misakaProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, int, error) {
	filterApexNS(dc)

	instructions, actualChangeCount, err := diff2.ByRecordSet(existing, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	var corrections []*models.Correction
	for _, inst := range instructions {
		switch inst.Type {
		case diff2.REPORT:
			corrections = append(corrections, inst.CreateMessage())

		case diff2.CREATE, diff2.CHANGE:
			name, rtype := inst.New[0].GetLabel(), inst.Key.Type
			before := providers.BeginToNative(p.observer, "toNative", inst.New)
			req := toNative(inst.New)
			providers.EndToNative(p.observer, "toNative", before, inst.New, req, nil)

			write := p.client.createRecordSet
			if inst.Type == diff2.CHANGE {
				write = p.client.updateRecordSet
			}
			corrections = append(corrections, inst.CreateCorrection(func() error {
				return write(dc.Name, name, rtype, req)
			}))

		case diff2.DELETE:
			name, rtype := inst.Old[0].GetLabel(), inst.Key.Type
			corrections = append(corrections, inst.CreateCorrection(func() error {
				return p.client.deleteRecordSet(dc.Name, name, rtype)
			}))

		default:
			panic(fmt.Sprintf("unhandled inst.Type %s", inst.Type))
		}
	}

	return corrections, actualChangeCount, nil
}

// filterApexNS removes the NS records at the zone apex from dc.Records.
// Misaka.io manages the apex NS set, so DNSControl leaves it alone. A record
// that repeats one of the zone's nameservers is dropped silently; any other
// apex NS record is dropped with a warning, because the user asked for
// something that will not happen.
func filterApexNS(dc *models.DomainConfig) {
	declared := make(map[string]bool, len(dc.Nameservers))
	for _, ns := range dc.Nameservers {
		declared[strings.TrimSuffix(ns.Name, ".")] = true
	}

	kept := make(models.Records, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" && rec.GetLabel() == apexLabel {
			target := strings.TrimSuffix(rec.AsNS().Ns, ".")
			if !declared[target] {
				printer.Warnf("MISAKA: the NS records at the zone apex are managed by Misaka.io. %s will not be added.\n", rec.AsNS().Ns)
			}
			continue
		}
		kept = append(kept, rec)
	}
	dc.Records = kept
}
