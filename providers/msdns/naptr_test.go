package msdns

import (
	"reflect"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func Test_decodeRecordDataNaptr(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want models.RecordConfig
	}{
		{"01", args{"C8AFB0B30153075349502B4432540474657374165F7369702E5F7463702E6578616D706C652E6F72672E"}, models.RecordConfig{NaptrOrder: 45000, NaptrPreference: 46000, NaptrFlags: "S", NaptrService: "SIP+D2T", NaptrRegexp: "test", Name: "_sip._tcp.example.org."}},
	}
	for _, tt := range tests {
		tt.want.SetTarget(tt.want.Name)
		tt.want.Name = ""
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeRecordDataNaptr(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeRecordDataNaptr() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
