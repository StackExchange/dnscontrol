package gandi5

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func TestRecordsToNative_1(t *testing.T) {
	var rcs = []*models.RecordConfig{{}}
	rcs[0].SetLabelFromFQDN("foo.example.com", "example.com")
	rcs[0].Type = "A"
	rcs[0].SetTarget("1.2.3.4")

	ns := recordsToNative(rcs, "example.com")

	if len(ns) != 1 {
		t.Errorf("len(ns) != 1; got=%v", len(ns))
	}
	if len(ns[0].RrsetValues) != 1 {
		t.Errorf("len(ns[0].RrsetValues) != 1; got=%v", ns[0].RrsetValues)
	}

}

func TestRecordsToNative_2(t *testing.T) {
	var rcs = []*models.RecordConfig{{}, {}}
	rcs[0].SetLabelFromFQDN("foo.example.com", "example.com")
	rcs[0].Type = "A"
	rcs[0].SetTarget("1.2.3.4")
	rcs[1].SetLabelFromFQDN("foo.example.com", "example.com")
	rcs[1].Type = "A"
	rcs[1].SetTarget("5.6.7.8")

	ns := recordsToNative(rcs, "example.com")

	if len(ns) != 1 {
		t.Errorf("len(ns) != 1; got=%v", len(ns))
	}
	if len(ns[0].RrsetValues) != 2 {
		t.Errorf("len(ns[0].RrsetValues) != 2; got=%v", ns[0].RrsetValues)
	}

}
