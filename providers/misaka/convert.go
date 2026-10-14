package misaka

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v5/models"
	"github.com/DNSControl/dnscontrol/v5/pkg/nrc"
	"github.com/DNSControl/dnscontrol/v5/pkg/privatetypes"
)

// apexLabel is how both DNSControl and the API name the zone apex.
const apexLabel = "@"

// toRC converts a record set to one RecordConfig per value.
//
// Every value is a presentation-format string, and a hostname without a
// trailing dot is served as a fully qualified name, never relative to the zone.
// Two kinds of value need different handling:
//
//   - ALIAS: the parser behind NewRecordConfigParse ignores TargetIsFqdnNoDot
//     for this private type; NewRecordConfig honours it.
//   - TXT: the API reads a value as presentation format only when it is
//     entirely quoted. Any other value is the raw string, so a bare
//     back\slash is published with its backslash.
//
// Original is not set: the API addresses a record set by name and type, and
// every diff2.ByRecordSet instruction carries both.
func toRC(dc *models.DomainConfig, rs apiRecordSet) (models.Records, error) {
	label := dc.LabelFromShort(rs.Name)

	recs := make(models.Records, 0, len(rs.Records))
	for _, r := range rs.Records {
		var rc *models.RecordConfig
		var err error
		switch {
		case rs.Type == "ALIAS":
			rc, err = dc.NewRecordConfig(label, rs.TTL, privatetypes.TypeALIAS, r.Value,
				nrc.Flags{TargetIsFqdnNoDot: true})
		case rs.Type == "TXT" && !isQuoted(r.Value):
			rc, err = dc.NewRecordConfig(label, rs.TTL, dnsv2.TypeTXT, r.Value)
		default:
			rc, err = dc.NewRecordConfigParse(label, rs.TTL, rs.Type, r.Value,
				nrc.Flags{TargetIsFqdnNoDot: true})
		}
		if err != nil {
			return nil, fmt.Errorf("MISAKA: record set %s/%s: value %q: %w", rs.Name, rs.Type, r.Value, err)
		}
		recs = append(recs, rc)
	}
	return recs, nil
}

// isQuoted reports whether a TXT value starts and ends with a double quote.
func isQuoted(value string) bool {
	return len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"'
}

// toNative converts the records of one record set to the body of a create or
// replace. The presentation format of each value is what the API stores and
// serves: hostnames keep their trailing dot, and TXT and CAA values are quoted.
func toNative(records models.Records) recordSetRequest {
	req := recordSetRequest{
		TTL:     records[0].TTL,
		Records: make([]apiRecord, 0, len(records)),
		Filters: []any{},
	}
	for _, rc := range records {
		req.Records = append(req.Records, apiRecord{Value: rc.GetRDATA().String()})
	}
	return req
}
