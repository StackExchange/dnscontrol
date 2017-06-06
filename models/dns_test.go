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
		CaaFlags: 1,
	}
	expected = "example.com.\t300\tIN\tCAA\t1 iodef \"mailto:test@example.com\""
	found = experiment.ToRR().String()
	if found != expected {
		t.Errorf("RR expected (%#v) got (%#v)\n", expected, found)
	}
}
