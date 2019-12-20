package gandi

// Convert the provider's native record description to models.RecordConfig.

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/printer"
	"github.com/pkg/errors"
	gandi "github.com/tiramiseb/go-gandi-livedns"
)

// nativeToRecord takes a DNS record from Gandi and returns our native RecordConfig format.
func nativeToRecords(n gandi.ZoneRecord, origin string) (rcs []*models.RecordConfig) {

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
	fmt.Printf("R2N IN:\n")
	for i, j := range rcs {
		fmt.Printf("  %v: %+v\n", i, j)
	}

	// Take a list of RecordConfig and return an equivalent list of
	// ZoneRecords.  Gandi requires one ZoneRecord for each label:key tuple.

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
			zr := &gandi.ZoneRecord{
				RrsetType:   r.Type,
				RrsetTTL:    int(r.TTL),
				RrsetName:   label,
				RrsetValues: []string{r.GetTargetCombined()},
			}
			keys[key] = zr
			zrs = append(zrs, *zr)

		} else {
			// Update an existing ZoneRecord:
			fmt.Printf("APPENDING: %v\n", r.GetTargetCombined())
			fmt.Printf("        A: %v\n", zr.RrsetValues)
			zr.RrsetValues = append(zr.RrsetValues, r.GetTargetCombined())
			fmt.Printf("        B: %v\n", zr.RrsetValues)
			fmt.Printf("XXXXXXX: %v\n", zrs[len(zrs)-1])
			fmt.Printf("YYYYYYY: %v || %v\n", *zr, zrs)

			if r.TTL != uint32(zr.RrsetTTL) {
				printer.Warnf("All TTLs for a rrset (%v) must be the same. Using smaller of %v and %v.\n", key, r.TTL, zr.RrsetTTL)
				if r.TTL < uint32(zr.RrsetTTL) {
					zr.RrsetTTL = int(r.TTL)
				}
			}

		}
	}

	fmt.Printf("R2N OUT:\n")
	for i, j := range zrs {
		fmt.Printf("  %v: %+v\n", i, j)
	}
	fmt.Println()

	return zrs
}
