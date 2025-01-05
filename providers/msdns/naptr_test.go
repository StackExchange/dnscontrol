package msdns

import (
	"reflect"
	"testing"

	"github.com/StackExchange/dnscontrol/v4/models"
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
		if err := tt.want.SetTarget(tt.want.Name); err != nil {
			t.Fatal(err)
		}
		tt.want.Name = ""
		t.Run(tt.name, func(t *testing.T) {
			if got, err := decodeRecordDataNaptr(tt.args.s); err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeRecordDataNaptr() = %+v (%v), want %+v", got, err, tt.want)
			}
		})
	}
}
