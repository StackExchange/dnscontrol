package rtypeloc

import (
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
	"github.com/miekg/dns"
)

// FromRawArgsBuilder creates a Rdata...
// update a RecordConfig using the args (from a
// RawRecord.Args). In other words, use the data from dnsconfig.js's
// rawrecordBuilder to create (actually... update) a models.RecordConfig.
func FromRawArgsBuilder(items []any) (*LOC, error) {
	n := &LOC{}

	// The first item dictates what to expect:
	if err := rtypecontrol.PaveArgs([]any{items[0]}, "s"); err != nil {
		return nil, err
	}
	switch items[0].(string) {
	case "RFC1876":
		if err := rtypecontrol.PaveArgs(items[1:], "s"); err != nil {
			return nil, err
		}
		loc, err := dns.NewRR(". IN LOC " + items[1].(string))
		if err != nil {
			return nil, err
		}
		n.Version = loc.(*dns.LOC).Version
		n.Size = loc.(*dns.LOC).Size
		n.HorizPre = loc.(*dns.LOC).HorizPre
		n.VertPre = loc.(*dns.LOC).VertPre
		n.Latitude = loc.(*dns.LOC).Latitude
		n.Longitude = loc.(*dns.LOC).Longitude
		n.Altitude = loc.(*dns.LOC).Altitude

		return n, err
	}

	return n, nil
}
