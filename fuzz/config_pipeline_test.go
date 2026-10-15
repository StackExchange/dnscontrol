package fuzz

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/DNSControl/dnscontrol/v5/models"
	"github.com/DNSControl/dnscontrol/v5/pkg/normalize"
	"github.com/DNSControl/dnscontrol/v5/pkg/prettyzone"
)

// FuzzConfigJSON fuzzes the JSON document that dnsconfig.js produces
// (via JSON.stringify(conf)) and the Go pipeline that consumes it.  The
// input is in the same shape that pkg/js.ExecuteJavascriptString
// unmarshals, and goes through the same steps: record import, then
// validation and normalization, then zonefile generation.
//
// Any error from the pipeline is acceptable: dnscontrol reports it to
// the user.  A panic is not: it crashes dnscontrol with a stack trace
// on a crafted dnsconfig.js.
func FuzzConfigJSON(f *testing.F) {
	f.Add([]byte(seedBasic))
	f.Add([]byte(seedRecordTypes))
	f.Add([]byte(seedSubdomainAndMetas))
	f.Add([]byte(seedSplitHorizon))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"domains":[{"name":"example.com","registrar":"r","dnsProviders":{"d":-1},"rawrecords":[{"type":"A","args":["@","1.2.3.4"],"ttl":300}]}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		conf := &models.DNSConfig{}
		if err := json.Unmarshal(data, conf); err != nil {
			return
		}

		// The same processing that pkg/js.ExecuteJavascriptString does:
		for _, dc := range conf.Domains {
			dc.PopulateNamesFromRaw(dc.Name)
		}
		if err := conf.PostProcess(); err != nil {
			return
		}
		if err := conf.ImportRawRecords(); err != nil {
			return
		}

		// Errors are fine; we only care that this never panics.
		_ = normalize.ValidateAndNormalizeConfig(conf)

		for _, dc := range conf.Domains {
			//nolint:errcheck // Zonefile generation errors are fine; only panics matter here.
			_ = prettyzone.WriteZoneFileRC(io.Discard, dc.Records, dc.Name, 300, nil)
		}
	})
}

// TestFuzzConfigJSONSeedsAreValid asserts that the interesting seed
// corpora survive the pipeline and actually produce records.  This
// protects the fuzzer's coverage from silently rotting (for example
// when an unrelated change makes seeds fail early in the pipeline).
func TestFuzzConfigJSONSeedsAreValid(t *testing.T) {
	for _, seed := range []string{seedBasic, seedRecordTypes, seedSubdomainAndMetas} {
		conf := &models.DNSConfig{}
		if err := json.Unmarshal([]byte(seed), conf); err != nil {
			t.Fatalf("seed %q: %v", seed[:20], err)
		}
		for _, dc := range conf.Domains {
			dc.PopulateNamesFromRaw(dc.Name)
		}
		if err := conf.PostProcess(); err != nil {
			t.Fatalf("seed %q: %v", seed[:20], err)
		}
		if err := conf.ImportRawRecords(); err != nil {
			t.Fatalf("seed %q: %v", seed[:20], err)
		}
		total := 0
		for _, dc := range conf.Domains {
			total += len(dc.Records)
		}
		if total == 0 {
			t.Fatalf("seed %q produced no records", seed[:20])
		}
	}
}
