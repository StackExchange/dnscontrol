package misaka

import (
	"strings"
	"testing"

	"github.com/DNSControl/dnscontrol/v5/models"
)

const testOrigin = "example.com"

func TestToRC(t *testing.T) {
	tests := []struct {
		name  string
		rs    apiRecordSet
		label string
		want  []string // GetRDATA().String() of each record
		txt   []string // GetTargetTXTJoined() of each TXT record
	}{
		{
			name:  "apex",
			rs:    apiRecordSet{Name: "@", Type: "A", TTL: 300, Records: []apiRecord{{Value: "192.0.2.1"}}},
			label: "@",
			want:  []string{"192.0.2.1"},
		},
		{
			name:  "relative name",
			rs:    apiRecordSet{Name: "www.sub", Type: "AAAA", TTL: 300, Records: []apiRecord{{Value: "2001:db8::1"}}},
			label: "www.sub",
			want:  []string{"2001:db8::1"},
		},
		{
			name:  "multi-value set",
			rs:    apiRecordSet{Name: "www", Type: "A", TTL: 300, Records: []apiRecord{{Value: "192.0.2.1"}, {Value: "192.0.2.2"}}},
			label: "www",
			want:  []string{"192.0.2.1", "192.0.2.2"},
		},
		{
			name:  "dotless CNAME target is a FQDN",
			rs:    apiRecordSet{Name: "docs", Type: "CNAME", TTL: 300, Records: []apiRecord{{Value: "docs.example.net"}}},
			label: "docs",
			want:  []string{"docs.example.net."},
		},
		{
			name:  "single-label target is a FQDN too",
			rs:    apiRecordSet{Name: "docs", Type: "CNAME", TTL: 300, Records: []apiRecord{{Value: "www"}}},
			label: "docs",
			want:  []string{"www."},
		},
		{
			name:  "dotless MX target",
			rs:    apiRecordSet{Name: "@", Type: "MX", TTL: 300, Records: []apiRecord{{Value: "10 mail.example.net"}}},
			label: "@",
			want:  []string{"10 mail.example.net."},
		},
		{
			name:  "dotless ALIAS target",
			rs:    apiRecordSet{Name: "cdn", Type: "ALIAS", TTL: 300, Records: []apiRecord{{Value: "edge.example.net"}}},
			label: "cdn",
			want:  []string{"edge.example.net."},
		},
		{
			name:  "dotted ALIAS target",
			rs:    apiRecordSet{Name: "cdn", Type: "ALIAS", TTL: 300, Records: []apiRecord{{Value: "edge.example.net."}}},
			label: "cdn",
			want:  []string{"edge.example.net."},
		},
		{
			name:  "quoted TXT",
			rs:    apiRecordSet{Name: "t", Type: "TXT", TTL: 300, Records: []apiRecord{{Value: `"challengetoken"`}}},
			label: "t",
			txt:   []string{"challengetoken"},
		},
		{
			name:  "quoted TXT with escaped quotes",
			rs:    apiRecordSet{Name: "t", Type: "TXT", TTL: 300, Records: []apiRecord{{Value: `"\"test2\""`}}},
			label: "t",
			txt:   []string{`"test2"`},
		},
		{
			name:  "quoted TXT with an escaped backslash",
			rs:    apiRecordSet{Name: "t", Type: "TXT", TTL: 300, Records: []apiRecord{{Value: `"back\\slash"`}}},
			label: "t",
			txt:   []string{`back\slash`},
		},
		{
			name:  "quoted TXT with two segments",
			rs:    apiRecordSet{Name: "t", Type: "TXT", TTL: 300, Records: []apiRecord{{Value: `"seg1" "seg2"`}}},
			label: "t",
			txt:   []string{"seg1seg2"},
		},
		{
			name:  "empty TXT",
			rs:    apiRecordSet{Name: "t", Type: "TXT", TTL: 300, Records: []apiRecord{{Value: `""`}}},
			label: "t",
			txt:   []string{""},
		},
		{
			name:  "bare TXT with spaces",
			rs:    apiRecordSet{Name: "t", Type: "TXT", TTL: 300, Records: []apiRecord{{Value: "hello world"}}},
			label: "t",
			txt:   []string{"hello world"},
		},
		{
			name:  "bare TXT keeps its backslash",
			rs:    apiRecordSet{Name: "t", Type: "TXT", TTL: 300, Records: []apiRecord{{Value: `back\slash`}}},
			label: "t",
			txt:   []string{`back\slash`},
		},
		{
			name:  "partially quoted TXT is a raw string",
			rs:    apiRecordSet{Name: "t", Type: "TXT", TTL: 300, Records: []apiRecord{{Value: `"seg1" bare`}}},
			label: "t",
			txt:   []string{`"seg1" bare`},
		},
		{
			name:  "CAA",
			rs:    apiRecordSet{Name: "@", Type: "CAA", TTL: 300, Records: []apiRecord{{Value: `0 issue "letsencrypt.org"`}}},
			label: "@",
			want:  []string{`0 issue "letsencrypt.org"`},
		},
		{
			name:  "SRV",
			rs:    apiRecordSet{Name: "_sip._tcp", Type: "SRV", TTL: 300, Records: []apiRecord{{Value: "10 60 5060 sip.example.net"}}},
			label: "_sip._tcp",
			want:  []string{"10 60 5060 sip.example.net."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := models.MustNewDomainConfig(testOrigin)
			recs, err := toRC(dc, tt.rs)
			if err != nil {
				t.Fatalf("toRC() error = %v", err)
			}

			n := max(len(tt.want), len(tt.txt))
			if len(recs) != n {
				t.Fatalf("toRC() returned %d records, want %d", len(recs), n)
			}
			for i, rc := range recs {
				if rc.GetLabel() != tt.label {
					t.Errorf("label = %q, want %q", rc.GetLabel(), tt.label)
				}
				if rc.Type != tt.rs.Type {
					t.Errorf("type = %q, want %q", rc.Type, tt.rs.Type)
				}
				if rc.TTL != tt.rs.TTL {
					t.Errorf("ttl = %d, want %d", rc.TTL, tt.rs.TTL)
				}
				if tt.txt != nil {
					if got := rc.GetTargetTXTJoined(); got != tt.txt[i] {
						t.Errorf("txt = %q, want %q", got, tt.txt[i])
					}
					continue
				}
				if got := rc.GetRDATA().String(); got != tt.want[i] {
					t.Errorf("rdata = %q, want %q", got, tt.want[i])
				}
			}
		})
	}
}

func TestToRCInvalidValue(t *testing.T) {
	dc := models.MustNewDomainConfig(testOrigin)
	_, err := toRC(dc, apiRecordSet{Name: "www", Type: "A", TTL: 300, Records: []apiRecord{{Value: "not-an-address"}}})
	if err == nil || !strings.Contains(err.Error(), "www/A") {
		t.Fatalf("toRC() error = %v, want one naming www/A", err)
	}
}

func TestToNative(t *testing.T) {
	dc := models.MustNewDomainConfig(testOrigin)
	parse := func(label, rtype, value string, ttl uint32) *models.RecordConfig {
		t.Helper()
		rc, err := dc.NewRecordConfigParse(label, ttl, rtype, value)
		if err != nil {
			t.Fatal(err)
		}
		return rc
	}

	tests := []struct {
		name    string
		records models.Records
		ttl     uint32
		want    []string
	}{
		{
			name:    "multi-value set",
			records: models.Records{parse("www", "A", "192.0.2.1", 300), parse("www", "A", "192.0.2.2", 300)},
			ttl:     300,
			want:    []string{"192.0.2.1", "192.0.2.2"},
		},
		{
			name:    "apex with the largest TTL",
			records: models.Records{parse("@", "MX", "10 mail.example.net.", 2147483647)},
			ttl:     2147483647,
			want:    []string{"10 mail.example.net."},
		},
		{
			name:    "target inside the zone is written as a FQDN",
			records: models.Records{parse("docs", "CNAME", "www", 600)},
			ttl:     600,
			want:    []string{"www.example.com."},
		},
		{
			name:    "TXT is quoted and escaped",
			records: models.Records{parse("t", "TXT", `"\"test1\""`, 300)},
			ttl:     300,
			want:    []string{`"\"test1\""`},
		},
		{
			name:    "CAA value is quoted",
			records: models.Records{parse("@", "CAA", `0 issue "letsencrypt.org"`, 300)},
			ttl:     300,
			want:    []string{`0 issue "letsencrypt.org"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := toNative(tt.records)
			if req.TTL != tt.ttl {
				t.Errorf("ttl = %d, want %d", req.TTL, tt.ttl)
			}
			if req.Filters == nil || len(req.Filters) != 0 {
				t.Errorf("filters = %#v, want an empty list", req.Filters)
			}
			if len(req.Records) != len(tt.want) {
				t.Fatalf("records = %v, want %v", req.Records, tt.want)
			}
			for i, r := range req.Records {
				if r.Value != tt.want[i] {
					t.Errorf("value = %q, want %q", r.Value, tt.want[i])
				}
			}
		})
	}
}

func TestToNativeLongTXT(t *testing.T) {
	dc := models.MustNewDomainConfig(testOrigin)
	rc, err := dc.NewRecordConfig("long", 300, "TXT", strings.Repeat("a", 300))
	if err != nil {
		t.Fatal(err)
	}

	value := toNative(models.Records{rc}).Records[0].Value
	want := `"` + strings.Repeat("a", 255) + `" "` + strings.Repeat("a", 45) + `"`
	if value != want {
		t.Errorf("value = %q, want two quoted segments of 255 and 45", value)
	}

	// Reading the value back yields the original string.
	recs, err := toRC(dc, apiRecordSet{Name: "long", Type: "TXT", TTL: 300, Records: []apiRecord{{Value: value}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := recs[0].GetTargetTXTJoined(); got != strings.Repeat("a", 300) {
		t.Errorf("round trip = %q", got)
	}
}
