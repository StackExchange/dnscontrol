package models

import (
	"testing"
)

func TestHasRecordTypeName(t *testing.T) {
	x := &RecordConfig{
		Type: "A",
		Name: "@",
	}
	dc := DomainConfig{}
	if dc.HasRecordTypeName("A", "@") {
		t.Errorf("%v: expected (%v) got (%v)\n", dc.Records, false, true)
	}
	dc.Records = append(dc.Records, x)
	if !dc.HasRecordTypeName("A", "@") {
		t.Errorf("%v: expected (%v) got (%v)\n", dc.Records, true, false)
	}
	if dc.HasRecordTypeName("AAAA", "@") {
		t.Errorf("%v: expected (%v) got (%v)\n", dc.Records, false, true)
	}
}

func TestRR(t *testing.T) {
	experiment := RecordConfig{
		Type:         "A",
		Name:         "foo",
		Target:       "1.2.3.4",
		TTL:          0,
		NameFQDN:     "foo.example.com",
		MxPreference: 0,
	}
	expected := "foo.example.com.\t300\tIN\tA\t1.2.3.4"
	found := experiment.ToRR().String()
	if found != expected {
		t.Errorf("RR expected (%#v) got (%#v)\n", expected, found)
	}

	experiment = RecordConfig{
		Type:     "CAA",
		Name:     "@",
		Target:   "mailto:test@example.com",
		TTL:      300,
		NameFQDN: "example.com",
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
		Target:           "abcdef0123456789",
		TTL:              300,
		NameFQDN:         "_443._tcp.example.com",
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

// func TestSetTxtParse(t *testing.T) {
// 	tests := []struct {
// 		d1 string
// 		e1 string
// 		e2 []string
// 	}{
// 		{``, ``, []string{``}},
// 		{`foo`, `foo`, []string{`foo`}},
// 	}
// 	for i, test := range tests {
// 		x := &RecordConfig{Type: "TXT"}
// 		x.SetTxtParse(test.d1)
// 		if x.Target != test.e1 {
// 			t.Errorf("%v: expected Target=(%v) got (%v)", i, x.Target, test.e1)
// 		}
// 		if len()
// 	}
// }

func TestDowncase(t *testing.T) {
	dc := DomainConfig{Records: Records{
		&RecordConfig{Type: "MX", Name: "lower", Target: "targetmx"},
		&RecordConfig{Type: "MX", Name: "UPPER", Target: "TARGETMX"},
	}}
	Downcase(dc.Records)
	if !dc.HasRecordTypeName("MX", "lower") {
		t.Errorf("%v: expected (%v) got (%v)\n", dc.Records, false, true)
	}
	if !dc.HasRecordTypeName("MX", "upper") {
		t.Errorf("%v: expected (%v) got (%v)\n", dc.Records, false, true)
	}
	if dc.Records[0].Target != "targetmx" {
		t.Errorf("%v: target0 expected (%v) got (%v)\n", dc.Records, "targetmx", dc.Records[0].Target)
	}
	if dc.Records[1].Target != "targetmx" {
		t.Errorf("%v: target1 expected (%v) got (%v)\n", dc.Records, "targetmx", dc.Records[1].Target)
	}
}
