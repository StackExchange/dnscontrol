package normalize

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
)

// hasSpfRecords returns true if this record requests SPF unrolling.
func hasSpfRecords(recs []*models.RecordConfig) bool {
	for _, rec := range recs {
		if rec.Type == "TXT" && strings.HasPrefix(rec.Target, "v=spf1 ") {
			if m, ok := rec.Metadata["unroll"]; ok {
				//fmt.Println("spf_unroll", m, rec)
				return true
			}
		}
	}
	return false
}

func flattenSpf(domain *models.DomainConfig) error {
	// Assume there is enough room.

	fmt.Println("flattenSpf")

	// Compute the total payload of all the TXT records at the apex.

	// Find the SPF record for the apex. Extract unroll_patterns, pattern_spec.
	// dnsres := dnsresolver.NewResolverPreloaded( DNS cache filename )
	// Parse it.
	// Flatten each segment of the unroll list.
	// rec.TXTSplit( pattern_spec + "." + domain.Name)
	// Generate 1 TXT record for each split.

	// Generate the new SPF records.
	// Replace the original SPF record with the new list.

	// 	res, err := dnsresolver.NewResolverPreloaded("testdata-dns1.json")
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	rec, err := Parse(strings.Join([]string{"v=spf1",
	// 		"ip4:198.252.206.0/24",
	// 		"ip4:192.111.0.0/24",
	// 		"include:_spf.google.com",
	// 		"include:mailgun.org",
	// 		"include:spf-basic.fogcreek.com",
	// 		"include:mail.zendesk.com",
	// 		"include:servers.mcsv.net",
	// 		"include:sendgrid.net",
	// 		"include:spf.mtasv.net",
	// 		"~all"}, " "), res)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	t.Log(rec.Print())
	// 	rec = rec.Flatten("mailgun.org")
	// 	//fmt.Println(rec.TXT())
	// 	//fmt.Println(rec.TXTSplit("_spf%d.stackoverflow.com"))
	// 	t.Log(rec.Print())

	return nil
}
