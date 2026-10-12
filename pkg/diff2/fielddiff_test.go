package diff2

import (
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
	dnsrdatav2 "codeberg.org/miekg/dns/rdata"
)

func TestFieldDiff(t *testing.T) {
	soa := dnsrdatav2.SOA{Ns: "ns1.f.com.", Mbox: "hostmaster.f.com.", Serial: 2026101100, Refresh: 3600, Retry: 600, Expire: 604800, Minttl: 300}
	soa2 := soa
	soa2.Serial = 2026101101

	tests := []struct {
		name   string
		a, b   dnsv2.RDATA
		want   string
		wantOK bool
	}{
		{
			name:   "MX preference only",
			a:      dnsrdatav2.MX{Preference: 10, Mx: "mx.f.com."},
			b:      dnsrdatav2.MX{Preference: 20, Mx: "mx.f.com."},
			want:   "preference=(10->20) mx=mx.f.com.",
			wantOK: true,
		},
		{
			name:   "MX host only",
			a:      dnsrdatav2.MX{Preference: 10, Mx: "a.f.com."},
			b:      dnsrdatav2.MX{Preference: 10, Mx: "b.f.com."},
			want:   "preference=10 mx=(a.f.com.->b.f.com.)",
			wantOK: true,
		},
		{
			name:   "SRV weight only",
			a:      dnsrdatav2.SRV{Priority: 10, Weight: 5, Port: 443, Target: "sip.f.com."},
			b:      dnsrdatav2.SRV{Priority: 10, Weight: 10, Port: 443, Target: "sip.f.com."},
			want:   "priority=10 weight=(5->10) port=443 target=sip.f.com.",
			wantOK: true,
		},
		{
			name:   "CAA value with spaces is quoted",
			a:      dnsrdatav2.CAA{Flag: 0, Tag: "issue", Value: "ca.example.net; account=1"},
			b:      dnsrdatav2.CAA{Flag: 0, Tag: "issue", Value: "ca.example.net; account=2"},
			want:   `flag=0 tag=issue value=("ca.example.net; account=1"->"ca.example.net; account=2")`,
			wantOK: true,
		},
		{
			name:   "SOA serial only",
			a:      soa,
			b:      soa2,
			want:   "ns=ns1.f.com. mbox=hostmaster.f.com. serial=(2026101100->2026101101) refresh=3600 retry=600 expire=604800 minttl=300",
			wantOK: true,
		},
		{
			name: "every field changed",
			a:    dnsrdatav2.MX{Preference: 10, Mx: "a.f.com."},
			b:    dnsrdatav2.MX{Preference: 20, Mx: "b.f.com."},
		},
		{
			name: "single-field type",
			a:    dnsrdatav2.CNAME{Target: "a.f.com."},
			b:    dnsrdatav2.CNAME{Target: "b.f.com."},
		},
		{
			name: "no fields changed",
			a:    dnsrdatav2.MX{Preference: 10, Mx: "mx.f.com."},
			b:    dnsrdatav2.MX{Preference: 10, Mx: "mx.f.com."},
		},
		{
			name: "different types",
			a:    dnsrdatav2.MX{Preference: 10, Mx: "mx.f.com."},
			b:    dnsrdatav2.SRV{Priority: 10, Weight: 5, Port: 443, Target: "mx.f.com."},
		},
		{
			name: "non-scalar field (TXT strings)",
			a:    dnsrdatav2.TXT{Txt: []string{"a"}},
			b:    dnsrdatav2.TXT{Txt: []string{"b"}},
		},
		{
			name: "non-scalar field (CSYNC type bitmap)",
			a:    dnsrdatav2.CSYNC{Serial: 1, Flags: 0, TypeBitMap: []uint16{dnsv2.TypeA}},
			b:    dnsrdatav2.CSYNC{Serial: 2, Flags: 0, TypeBitMap: []uint16{dnsv2.TypeA}},
		},
		{
			name: "nil rdata",
			a:    nil,
			b:    dnsrdatav2.MX{Preference: 10, Mx: "mx.f.com."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := fieldDiff(tt.a, tt.b)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("fieldDiff() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
