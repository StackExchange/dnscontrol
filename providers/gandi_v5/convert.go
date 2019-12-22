package gandi5

// Convert the provider's native record description to models.RecordConfig.

import (
	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/printer"
	"github.com/pkg/errors"
	gandi "github.com/tiramiseb/go-gandi-livedns"
)

// nativeToRecord takes a DNS record from Gandi and returns a native RecordConfig struct.
func nativeToRecords(n gandi.ZoneRecord, origin string) (rcs []*models.RecordConfig) {

	// Gandi returns all the values for a given label/rtype pair in each
	// gandi.ZoneRecord.  In other words, if there are multiple A
	// records for a label, all the IP addresses are listed in
	// n.RrsetValues rather than having many gandi.ZoneRecord's.
	// We must split them out into individual records, one for each value.
	for _, value := range n.RrsetValues {
		rc := &models.RecordConfig{
			TTL:      uint32(n.RrsetTTL),
			Original: n,
		}
		rc.SetLabel(n.RrsetName, origin)
		switch rtype := n.RrsetType; rtype {
		default: //  "A", "AAAA", "CAA", "NS", "CNAME", "MX", "PTR", "SRV", "TXT"
			if err := rc.PopulateFromString(rtype, value, origin); err != nil {
				panic(errors.Wrap(err, "unparsable record received from gandi"))
			}
		}
		rcs = append(rcs, rc)
	}

	return rcs
}

func recordsToNative(rcs []*models.RecordConfig, origin string) []gandi.ZoneRecord {
	// Take a list of RecordConfig and return an equivalent list of ZoneRecords.
	// Gandi requires one ZoneRecord for each label:key tuple, therefore we
	// might collapse many RecordConfig into one ZoneRecord.

	var keys = map[models.RecordKey]*gandi.ZoneRecord{}
	var zrs []gandi.ZoneRecord

	for _, r := range rcs {
		label := r.GetLabel()
		if label == "@" {
			label = origin
		}
		key := r.Key()

		if zr, ok := keys[key]; !ok {
			// Allocate a new ZoneRecord:
			zr := gandi.ZoneRecord{
				RrsetType:   r.Type,
				RrsetTTL:    int(r.TTL),
				RrsetName:   label,
				RrsetValues: []string{r.GetTargetCombined()},
			}
			zrs = append(zrs, zr)
			//keys[key] = &zr   // This didn't work.
			keys[key] = &zrs[len(zrs)-1] // This does work. I don't know why.

		} else {
			zr.RrsetValues = append(zr.RrsetValues, r.GetTargetCombined())

			if r.TTL != uint32(zr.RrsetTTL) {
				printer.Warnf("All TTLs for a rrset (%v) must be the same. Using smaller of %v and %v.\n", key, r.TTL, zr.RrsetTTL)
				if r.TTL < uint32(zr.RrsetTTL) {
					zr.RrsetTTL = int(r.TTL)
				}
			}

		}
	}

	return zrs
}
