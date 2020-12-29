package msdns

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// Hex strings were found with:
// $records = Get-DnsServerResourceRecord -rrtype naptr -zonename ds.stackexchange.com -name "*.5.6.enum" ; $r = $records[0].RecordData ; $r.Data ; $records[0].HostName
// NAPTR parameters were found with:
// dig +short '*.5.6.enum.ds.stackexchange.com.' naptr
// NOTE: Change // to /

func Test_naptrToHex(t *testing.T) {
	type args struct {
		rc *models.RecordConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "0.8.0.2.enum",
			args: args{&models.RecordConfig{Type: "NAPTR",
				NaptrOrder: 1, NaptrPreference: 10, NaptrFlags: "U",
				NaptrService: `E2U+sip`,
				NaptrRegexp:  `!^.*$!sip:2080@10.110.2.10!`,
			}},
			want: "01000A000155074532552B7369701B215E2E2A24217369703A323038304031302E3131302E322E31302100",
			//want: "0100 0A00 0155 074532552B736970 1B215E2E2A24217369703A323038304031302E3131302E322E313021 00",
			//       u16  u16  bstr bstr             bstr                                                     0
		},
		{
			name: "*.5.6.enum",
			args: args{&models.RecordConfig{Type: "NAPTR",
				NaptrOrder: 1, NaptrPreference: 10, NaptrFlags: "U",
				NaptrService: `E2U+sip`,
				NaptrRegexp:  `!^(.*)$!sip:\1@10.110.2.10!`,
			}},
			want: "01000A000155074532552B7369701B215E282E2A2924217369703A5C314031302E3131302E322E31302100",
			//want: "0100 0A00 0155 074532552B736970 1B215E282E2A2924217369703A5C314031302E3131302E322E313021 00",
			//       u16  u16  bstr bstr             bstr                                                     0
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naptrToHex(tt.args.rc); got != tt.want {
				t.Errorf("naptrToHex(): got=(\n%v\n), want=(\n%v\n)", got, tt.want)
			}
		})
	}
}

func Test_populateFromHex(t *testing.T) {
	type args struct {
		s string
	}
	type want struct {
		NaptrOrder      uint16
		NaptrPreference uint16
		NaptrFlags      string
		NaptrService    string
		NaptrRegexp     string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "*.5.6.enum",
			args: args{`01000A000155074532552B7369701B215E282E2A2924217369703A5C314031302E3131302E322E31302100`},
			want: want{1, 10, "U", `E2U+sip`, `!^(.*)$!sip:\\1@10.110.2.10!`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := &models.RecordConfig{Type: "NAPTR"}
			populateFromHex(rc, tt.args.s)
			if rc.NaptrOrder != tt.want.NaptrOrder {
				t.Errorf("populateFromHex() NaptrOrder got=%v want=%v", rc.NaptrOrder, tt.want.NaptrOrder)
			}
			if rc.NaptrPreference != tt.want.NaptrPreference {
				t.Errorf("populateFromHex() NaptrPreference got=%v want=%v", rc.NaptrPreference, tt.want.NaptrPreference)
			}
			if rc.NaptrFlags != tt.want.NaptrFlags {
				t.Errorf("populateFromHex() NaptrFlags got=%q want=%q", rc.NaptrFlags, tt.want.NaptrFlags)
			}
			if rc.NaptrService != tt.want.NaptrService {
				t.Errorf("populateFromHex() NaptrService got=%q want=%q", rc.NaptrService, tt.want.NaptrService)
			}
			if rc.NaptrRegexp != tt.want.NaptrRegexp {
				t.Errorf("populateFromHex() NaptrRegexp got=%q want=%q", rc.NaptrRegexp, tt.want.NaptrRegexp)
			}
		})
	}
}
