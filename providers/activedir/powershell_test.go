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
			want: `Get-DnsServerResourceRecord -ZoneName "example.com" | ConvertTo-Json -depth 10`,
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

func Test_generatePSCreate(t *testing.T) {
	type args struct {
		domain string
		rec    *models.RecordConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generatePSCreate(tt.args.domain, tt.args.rec); got != tt.want {
				t.Errorf("generatePSCreate() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			want: "echo \"MODIFY @  A old=(1.2.3.4) new=(10.20.30.40):\" ; $OldObj = Get-DnsServerResourceRecord -ZoneName \"\" -Name \"@\" -RRType \"A\"| Where-Object {$_.HostName eq \"@\" -and -RRType -eq \"A\" -and $_.RecordData.IPv4Address -eq \"1.2.3.4\"} ; if($OldObj.Length -ne 1){ throw \"Error, multiple results for Get-DnsServerResourceRecord\" } ; $NewObj = $OldObj.Clone() ;$NewObj.RecordData.IPv4Address = \"10.20.30.40\" ; Set-DnsServerResourceRecord -ZoneName \"\" -NewInputObject $NewObj -OldInputObject $OldObj",
		},
		{name: "MX-1", args: args{old: recMX1, rec: recMX2},
			want: "echo \"MODIFY @  MX old=foo.com. new=foo2.com.\" ; $OldObj = Get-DnsServerResourceRecord -ZoneName \"\" -Name \"@\" -RRType \"MX\" | Where-Object {$_.RecordData. -eq \"foo.com.\" -and $_.HostName -eq \"@\"} ; if($OldObj.Length -ne $null){ throw \"Error, multiple results for Get-DnsServerResourceRecord\" } ; $NewObj = $OldObj.Clone() ; $NewObj.RecordData.Preference = 50 ; $NewObj.RecordData.MailExchange = \"foo2.com.\" ; Set-DnsServerResourceRecord -ZoneName \"\" -NewInputObject $NewObj -OldInputObject $OldObj",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generatePSModify(tt.args.domain, tt.args.old, tt.args.rec); strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("generatePSModify() = %q, want %q", got, tt.want)
			}
		})
	}
}
