package activedir

import (
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func Test_generatePSZoneDump(t *testing.T) {
	type args struct {
		domainname string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{domainname: "example.com"},
			want: `Get-DnsServerResourceRecord -ZoneName "example.com" | ConvertTo-Json -depth 4`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generatePSZoneDump(tt.args.domainname); got != tt.want {
				t.Errorf("generatePSZoneDump() = %v, want %v", got, tt.want)
			}
		})
	}
}

//func Test_generatePSDelete(t *testing.T) {
//	type args struct {
//		domain string
//		rec    *models.RecordConfig
//	}
//	tests := []struct {
//		name string
//		args args
//		want string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := generatePSDelete(tt.args.domain, tt.args.rec); got != tt.want {
//				t.Errorf("generatePSDelete() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

// func Test_generatePSCreate(t *testing.T) {
// 	type args struct {
// 		domain string
// 		rec    *models.RecordConfig
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := generatePSCreate(tt.args.domain, tt.args.rec); got != tt.want {
// 				t.Errorf("generatePSCreate() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_generatePSModify(t *testing.T) {

	recA1 := &models.RecordConfig{
		Type:   "A",
		Name:   "@",
		Target: "1.2.3.4",
	}
	recA2 := &models.RecordConfig{
		Type:   "A",
		Name:   "@",
		Target: "10.20.30.40",
	}

	recMX1 := &models.RecordConfig{
		Type:         "MX",
		Name:         "@",
		Target:       "foo.com.",
		MxPreference: 5,
	}
	recMX2 := &models.RecordConfig{
		Type:         "MX",
		Name:         "@",
		Target:       "foo2.com.",
		MxPreference: 50,
	}

	type args struct {
		domain string
		old    *models.RecordConfig
		rec    *models.RecordConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "A", args: args{old: recA1, rec: recA2},
			want: `echo DELETE "A" "@" "1.2.3.4" ; Remove-DnsServerResourceRecord -Force -ZoneName "" -Name "@" -RRType "A" -RecordData "1.2.3.4" ; echo CREATE "A" "@" "10.20.30.40" ; Add-DnsServerResourceRecord -ZoneName "" -Name "@" -TimeToLive $(New-TimeSpan -Seconds 0) -A -IPv4Address "10.20.30.40"`,
		},
		{name: "MX-1", args: args{old: recMX1, rec: recMX2},
			want: `echo DELETE "MX" "@" "5 foo.com." ; Remove-DnsServerResourceRecord -Force -ZoneName "" -Name "@" -RRType "MX" -RecordData 5,"foo.com." ; echo CREATE "MX" "@" "50 foo2.com." ; Add-DnsServerResourceRecord -ZoneName "" -Name "@" -TimeToLive $(New-TimeSpan -Seconds 0) -MX -MailExchange "foo2.com." -Preference 50`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generatePSModify(tt.args.domain, tt.args.old, tt.args.rec); strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("generatePSModify() = got=(\n%s\n) want=(\n%s\n)", got, tt.want)
			}
		})
	}
}
