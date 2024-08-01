package models

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v4/rtypes/rtypemx"
)

func TestRR(t *testing.T) {
	experiment := RecordConfig{
		Type:     "A",
		Name:     "foo",
		NameFQDN: "foo.example.com",
		target:   "1.2.3.4",
		TTL:      0,
	}
	expected := "foo.example.com.\t300\tIN\tA\t1.2.3.4"
	found := experiment.ToRR().String()
	if found != expected {
		t.Errorf("RR expected (%#v) got (%#v)\n", expected, found)
	}

	experiment = RecordConfig{
		Type:     "CAA",
		Name:     "@",
		NameFQDN: "example.com",
		target:   "mailto:test@example.com",
		TTL:      300,
		CaaTag:   "iodef",
		CaaFlag:  1,
	}
	expected = "example.com.\t300\tIN\tCAA\t1 iodef \"mailto:test@example.com\""
	found = experiment.ToRR().String()
	if found != expected {
		t.Errorf("RR expected (%#v) got (%#v)\n", expected, found)
	}

	experiment = RecordConfig{
		Type:             "TLSA",
		Name:             "@",
		NameFQDN:         "_443._tcp.example.com",
		target:           "abcdef0123456789",
		TTL:              300,
		TlsaUsage:        0,
		TlsaSelector:     0,
		TlsaMatchingType: 1,
	}
	expected = "_443._tcp.example.com.\t300\tIN\tTLSA\t0 0 1 abcdef0123456789"
	found = experiment.ToRR().String()
	if found != expected {
		t.Errorf("RR expected (%#v) got (%#v)\n", expected, found)
	}
}

func TestDowncase(t *testing.T) {
	r1 := &RecordConfig{Type: "MX", Name: "lower"}
	r1.Rdata = &rtypemx.MX{}
	r1.AsMX().SetTargetMX(10, "targetmx")
	r1.ReSeal()

	r2 := &RecordConfig{Type: "MX", Name: "UPPER"}
	r2.Rdata = &rtypemx.MX{}
	r2.AsMX().SetTargetMX(10, "TARGETMX")
	r2.ReSeal()

	dc := DomainConfig{Records: Records{
		r1, r2,
	}}
	Downcase(dc.Records)
	if !dc.Records.HasRecordTypeName("MX", "lower") {
		t.Errorf("%v: expected (%v) got (%v)\n", dc.Records, false, true)
	}
	if !dc.Records.HasRecordTypeName("MX", "upper") {
		t.Errorf("%v: expected (%v) got (%v)\n", dc.Records, false, true)
	}
	if dc.Records[0].GetTargetField() != "targetmx" {
		t.Errorf("%v: target0 expected (%v) got (%v)\n", dc.Records, "targetmx", dc.Records[0].GetTargetField())
	}
	if dc.Records[1].GetTargetField() != "targetmx" {
		t.Errorf("%v: target1 expected (%v) got (%v)\n", dc.Records, "targetmx", dc.Records[1].GetTargetField())
	}
}
