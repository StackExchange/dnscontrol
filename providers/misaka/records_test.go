package misaka

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DNSControl/dnscontrol/v5/models"
)

// recordSetsServer answers the record-set listing of example.com with sets.
func recordSetsServer(t *testing.T, sets string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/zones/example.com/recordsets" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail": "Zone not found"}`))
			return
		}
		var list []json.RawMessage
		if err := json.Unmarshal([]byte(sets), &list); err != nil {
			t.Fatal(err)
		}
		resp, _ := json.Marshal(map[string]any{"results": list, "count": len(list)})
		_, _ = w.Write(resp)
	}))
}

func testProvider(baseURL string) *misakaProvider {
	return &misakaProvider{client: newAPIClient(baseURL, "test-token")}
}

func TestGetZoneRecords(t *testing.T) {
	server := recordSetsServer(t, `[
		{"name": "@", "type": "SOA", "ttl": 600, "enabled": true, "protected": false, "alias": null, "records": [{"value": "a.m1ns.net. hostmaster.example.com. 1 2 3 4 5"}], "filters": []},
		{"name": "@", "type": "NS", "ttl": 600, "enabled": true, "protected": false, "alias": null, "records": [{"value": "a.m1ns.net"}, {"value": "b.m1ns.com"}], "filters": []},
		{"name": "sub", "type": "NS", "ttl": 600, "enabled": true, "protected": false, "alias": null, "records": [{"value": "ns1.example.org."}], "filters": []},
		{"name": "www", "type": "A", "ttl": 300, "enabled": true, "protected": false, "alias": null, "records": [{"value": "192.0.2.1"}, {"value": "192.0.2.2"}], "filters": []}
	]`)
	defer server.Close()

	recs, err := testProvider(server.URL).GetZoneRecords(models.MustNewDomainConfig(testOrigin))
	if err != nil {
		t.Fatalf("GetZoneRecords() error = %v", err)
	}

	// The SOA and the apex NS are managed by Misaka.io and must not be
	// reported; the NS record of a delegated child must be.
	var got []string
	for _, rc := range recs {
		got = append(got, rc.GetLabel()+" "+rc.Type+" "+rc.GetRDATA().String())
	}
	want := []string{"sub NS ns1.example.org.", "www A 192.0.2.1", "www A 192.0.2.2"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("GetZoneRecords() = %q, want %q", got, want)
	}
}

func TestGetZoneRecordsMissingZone(t *testing.T) {
	server := recordSetsServer(t, `[]`)
	defer server.Close()

	_, err := testProvider(server.URL).GetZoneRecords(models.MustNewDomainConfig("missing.example"))
	if err == nil || !strings.Contains(err.Error(), "missing.example") || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("GetZoneRecords() error = %v, want one naming the zone and HTTP 404", err)
	}
}

func TestGetZoneRecordsRefusesUnmodelledSets(t *testing.T) {
	tests := []struct {
		name string
		set  string
		want []string
	}{
		{
			name: "filters",
			set:  `{"name": "geo", "type": "A", "ttl": 300, "enabled": true, "protected": false, "alias": null, "records": [{"value": "192.0.2.1"}], "filters": [{"type": "geo"}]}`,
			want: []string{"geo/A", "filters"},
		},
		{
			name: "alias",
			set:  `{"name": "cdn", "type": "A", "ttl": 300, "enabled": true, "protected": false, "alias": "edge.example.net.", "records": [], "filters": []}`,
			want: []string{"cdn/A", "alias"},
		},
		{
			name: "disabled",
			set:  `{"name": "off", "type": "A", "ttl": 300, "enabled": false, "protected": false, "alias": null, "records": [{"value": "192.0.2.1"}], "filters": []}`,
			want: []string{"off/A", "enabled"},
		},
		{
			name: "protected",
			set:  `{"name": "keep", "type": "A", "ttl": 300, "enabled": true, "protected": true, "alias": null, "records": [{"value": "192.0.2.1"}], "filters": []}`,
			want: []string{"keep/A", "protected"},
		},
		{
			name: "no values",
			set:  `{"name": "www", "type": "A", "ttl": 300, "enabled": true, "protected": false, "alias": null, "records": [], "filters": []}`,
			want: []string{"www/A", "no values", "console"},
		},
		{
			name: "CNAME hostname @",
			set:  `{"name": "alias", "type": "CNAME", "ttl": 300, "enabled": true, "protected": false, "alias": null, "records": [{"value": "@"}], "filters": []}`,
			want: []string{"alias/CNAME", `"@"`, "full name"},
		},
		{
			name: "MX hostname @",
			set:  `{"name": "@", "type": "MX", "ttl": 300, "enabled": true, "protected": false, "alias": null, "records": [{"value": "10 @"}], "filters": []}`,
			want: []string{"@/MX", `"10 @"`},
		},
		{
			name: "mixed-case name",
			set:  `{"name": "Www", "type": "A", "ttl": 300, "enabled": true, "protected": false, "alias": null, "records": [{"value": "192.0.2.1"}], "filters": []}`,
			want: []string{"Www/A", "rename"},
		},
		{
			name: "name with a trailing dot",
			set:  `{"name": "www.example.com.", "type": "A", "ttl": 300, "enabled": true, "protected": false, "alias": null, "records": [{"value": "192.0.2.1"}], "filters": []}`,
			want: []string{"www.example.com./A", "rename"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := recordSetsServer(t, "["+tt.set+"]")
			defer server.Close()

			_, err := testProvider(server.URL).GetZoneRecords(models.MustNewDomainConfig(testOrigin))
			if err == nil {
				t.Fatal("GetZoneRecords() error = nil")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
		})
	}
}

func TestGetZoneRecordsAcceptsTXTAt(t *testing.T) {
	// "@" is only refused as a hostname; as TXT text it is just a character.
	server := recordSetsServer(t, `[{"name": "t", "type": "TXT", "ttl": 300, "enabled": true, "protected": false, "alias": null, "records": [{"value": "@"}], "filters": []}]`)
	defer server.Close()

	recs, err := testProvider(server.URL).GetZoneRecords(models.MustNewDomainConfig(testOrigin))
	if err != nil {
		t.Fatalf("GetZoneRecords() error = %v", err)
	}
	if len(recs) != 1 || recs[0].GetTargetTXTJoined() != "@" {
		t.Errorf("GetZoneRecords() = %v", recs)
	}
}

// request is one API call seen by the fake server.
type request struct {
	Method string
	Path   string
	Body   string
}

func TestGetZoneRecordsCorrections(t *testing.T) {
	var seen []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		seen = append(seen, request{Method: r.Method, Path: r.URL.EscapedPath(), Body: string(body)})
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	p := testProvider(server.URL)

	dc := models.MustNewDomainConfig(testOrigin)
	rec := func(label, rtype, value string, ttl uint32) *models.RecordConfig {
		t.Helper()
		rc, err := dc.NewRecordConfigParse(label, ttl, rtype, value)
		if err != nil {
			t.Fatal(err)
		}
		return rc
	}

	existing := models.Records{
		rec("www", "A", "192.0.2.1", 300),
		rec("ttl", "A", "192.0.2.5", 300),
		rec("@", "MX", "10 mail.example.net.", 300),
	}
	dc.Records = models.Records{
		rec("www", "A", "192.0.2.1", 300),
		rec("www", "A", "192.0.2.2", 300),
		rec("ttl", "A", "192.0.2.5", 600),
		rec("@", "TXT", "hello", 300),
	}

	corrections, count, err := p.GetZoneRecordsCorrections(dc, existing)
	if err != nil {
		t.Fatalf("GetZoneRecordsCorrections() error = %v", err)
	}
	if count != 4 {
		t.Errorf("actual change count = %d, want 4", count)
	}
	if len(seen) != 0 {
		t.Fatalf("computing corrections sent requests: %v", seen)
	}
	for _, c := range corrections {
		if c.F == nil {
			t.Fatalf("correction %q has no function", c.Msg)
		}
		if err := c.F(); err != nil {
			t.Fatal(err)
		}
	}

	want := map[string]string{
		"PUT /dns/zones/example.com/recordsets/www/A":   `{"ttl":300,"records":[{"value":"192.0.2.1"},{"value":"192.0.2.2"}],"filters":[]}`,
		"PUT /dns/zones/example.com/recordsets/ttl/A":   `{"ttl":600,"records":[{"value":"192.0.2.5"}],"filters":[]}`,
		"POST /dns/zones/example.com/recordsets/@/TXT":  `{"ttl":300,"records":[{"value":"\"hello\""}],"filters":[]}`,
		"DELETE /dns/zones/example.com/recordsets/@/MX": ``,
	}
	if len(seen) != len(want) {
		t.Errorf("requests = %v, want %d", seen, len(want))
	}
	for _, r := range seen {
		body, ok := want[r.Method+" "+r.Path]
		if !ok {
			t.Errorf("unexpected request %s %s", r.Method, r.Path)
			continue
		}
		if r.Body != body {
			t.Errorf("%s %s body = %s, want %s", r.Method, r.Path, r.Body, body)
		}
	}
}

func TestGetZoneRecordsCorrectionsReport(t *testing.T) {
	p := testProvider("http://127.0.0.1:0")
	dc := models.MustNewDomainConfig(testOrigin)
	dc.KeepUnknown = true // NO_PURGE

	existing, err := dc.NewRecordConfigParse("old", 300, "A", "192.0.2.9")
	if err != nil {
		t.Fatal(err)
	}

	corrections, count, err := p.GetZoneRecordsCorrections(dc, models.Records{existing})
	if err != nil {
		t.Fatalf("GetZoneRecordsCorrections() error = %v", err)
	}
	if count != 0 {
		t.Errorf("actual change count = %d, want 0", count)
	}
	if len(corrections) != 1 || corrections[0].F != nil || !strings.Contains(corrections[0].Msg, "NO_PURGE") {
		t.Fatalf("corrections = %+v, want one report without a function", corrections)
	}
}

func TestFilterApexNS(t *testing.T) {
	dc := models.MustNewDomainConfig(testOrigin)
	nss, err := models.ToNameservers([]string{"a.m1ns.net", "b.m1ns.com"})
	if err != nil {
		t.Fatal(err)
	}
	dc.Nameservers = nss
	for _, v := range []string{"a.m1ns.net.", "ns1.other.example."} {
		rc, err := dc.NewRecordConfigParse("@", 300, "NS", v)
		if err != nil {
			t.Fatal(err)
		}
		dc.Records = append(dc.Records, rc)
	}
	sub, err := dc.NewRecordConfigParse("sub", 300, "NS", "ns1.example.org.")
	if err != nil {
		t.Fatal(err)
	}
	dc.Records = append(dc.Records, sub)

	filterApexNS(dc)

	if len(dc.Records) != 1 || dc.Records[0].GetLabel() != "sub" {
		t.Errorf("records after filterApexNS = %v, want only the sub delegation", dc.Records)
	}
}

func TestGetZoneRecordsCorrectionsLowersTTL(t *testing.T) {
	var seen []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		seen = append(seen, request{Method: r.Method, Path: r.URL.EscapedPath(), Body: string(body)})
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	p := testProvider(server.URL)

	for _, tt := range []struct {
		name        string
		existingTTL uint32
		wantBody    string
	}{
		{
			name:        "too large a TTL is written as the maximum",
			existingTTL: 300,
			wantBody:    `{"ttl":2147483647,"records":[{"value":"192.0.2.1"}],"filters":[]}`,
		},
		{
			name:        "the maximum already stored is no change",
			existingTTL: maxTTL,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			seen = nil
			dc := models.MustNewDomainConfig(testOrigin)
			existing, err := dc.NewRecordConfigParse("big", tt.existingTTL, "A", "192.0.2.1")
			if err != nil {
				t.Fatal(err)
			}
			desired, err := dc.NewRecordConfigParse("big", 4294967295, "A", "192.0.2.1")
			if err != nil {
				t.Fatal(err)
			}
			dc.Records = models.Records{desired}

			corrections, _, err := p.GetZoneRecordsCorrections(dc, models.Records{existing})
			if err != nil {
				t.Fatalf("GetZoneRecordsCorrections() error = %v", err)
			}
			if tt.wantBody == "" {
				if len(corrections) != 0 {
					t.Fatalf("corrections = %v, want none", corrections)
				}
				return
			}
			if len(corrections) != 1 {
				t.Fatalf("corrections = %v, want one", corrections)
			}
			if err := corrections[0].F(); err != nil {
				t.Fatal(err)
			}
			if len(seen) != 1 || seen[0].Method != http.MethodPut || seen[0].Body != tt.wantBody {
				t.Errorf("requests = %+v, want one PUT with %s", seen, tt.wantBody)
			}
		})
	}
}
