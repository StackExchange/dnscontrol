package normalize

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/spflib"
)

// hasSpfRecords returns true if this record requests SPF unrolling.
func flattenSPFs(cfg *models.DNSConfig) []error {
	cache, err := spflib.NewCache("spfcache.json")
	if err != nil {
		return []error{err}
	}
	var errs []error
	for _, domain := range cfg.Domains {
		apexTXTs := domain.Records.Grouped()[models.RecordKey{Type: "TXT", Name: "@"}]
		// flatten all spf records that have the "flatten" metadata
		for _, txt := range apexTXTs {
			if flatten, ok := txt.Metadata["flatten"]; ok && strings.HasPrefix(txt.Target, "v=spf1") {
				rec, err := spflib.Parse(txt.Target, cache)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				rec = rec.Flatten(flatten)
				txt.Target = rec.TXT()
			}
		}
	}
	// check if cache is stale
	for _, e := range cache.ResolveErrors() {
		errs = append(errs, Warning{fmt.Errorf("problem resolving SPF record: %s", e)})
	}
	changed := cache.ChangedRecords()
	if len(changed) > 0 {
		if err := cache.Save("spfcache.updated.json"); err != nil {
			errs = append(errs, err)
		} else {
			errs = append(errs, Warning{fmt.Errorf("%d spf record lookups are out of date with cache (%s). Wrote changes to dnscache.updated.json. Please rename and commit", len(changed), strings.Join(changed, ","))})
		}
	}
	return errs
}

// func flattenSpf(domain *models.DomainConfig) error {
// 	// Assume there is enough room.

// 	fmt.Println("flattenSpf")

// 	// Compute the total payload of all the TXT records at the apex.

// 	// Find the SPF record for the apex. Extract unroll_patterns, pattern_spec.
// 	// dnsres := dnsresolver.NewResolverPreloaded( DNS cache filename )
// 	// Parse it.
// 	// Flatten each segment of the unroll list.
// 	// rec.TXTSplit( pattern_spec + "." + domain.Name)
// 	// Generate 1 TXT record for each split.

// 	// Generate the new SPF records.
// 	// Replace the original SPF record with the new list.

// 	// 	res, err := dnsresolver.NewResolverPreloaded("testdata-dns1.json")
// 	// 	if err != nil {
// 	// 		t.Fatal(err)
// 	// 	}
// 	// 	rec, err := Parse(strings.Join([]string{"v=spf1",
// 	// 		"ip4:198.252.206.0/24",
// 	// 		"ip4:192.111.0.0/24",
// 	// 		"include:_spf.google.com",
// 	// 		"include:mailgun.org",
// 	// 		"include:spf-basic.fogcreek.com",
// 	// 		"include:mail.zendesk.com",
// 	// 		"include:servers.mcsv.net",
// 	// 		"include:sendgrid.net",
// 	// 		"include:spf.mtasv.net",
// 	// 		"~all"}, " "), res)
// 	// 	if err != nil {
// 	// 		t.Fatal(err)
// 	// 	}
// 	// 	t.Log(rec.Print())
// 	// 	rec = rec.Flatten("mailgun.org")
// 	// 	//fmt.Println(rec.TXT())
// 	// 	//fmt.Println(rec.TXTSplit("_spf%d.stackoverflow.com"))
// 	// 	t.Log(rec.Print())

// 	return nil
// }
