package activedir

import (
	"fmt"
	"testing"

	"github.com/StackExchange/dnscontrol/models"
)

func TestGetExistingRecords(t *testing.T) {

	cf := &adProvider{}

	*flagFakePowerShell = true
	actual, err := cf.getExistingRecords("test2")
	if err != nil {
		t.Fatal(err)
	}
	expected := []*models.RecordConfig{
		{Name: "@", NameFQDN: "test2", Type: "A", TTL: 600, Target: "10.166.2.11"},
		//{Name: "_msdcs", NameFQDN: "_msdcs.test2", Type: "NS", TTL: 300, Target: "other_record"}, // Will be filtered.
		{Name: "co-devsearch02", NameFQDN: "co-devsearch02.test2", Type: "A", TTL: 3600, Target: "10.8.2.64"},
		{Name: "co-devservice01", NameFQDN: "co-devservice01.test2", Type: "A", TTL: 1200, Target: "10.8.2.48"}, // Downcased.
		{Name: "yum", NameFQDN: "yum.test2", Type: "A", TTL: 3600, Target: "10.8.0.59"},
	}

	actualS := ""
	for i, x := range actual {
		actualS += fmt.Sprintf("%d %v\n", i, x)
	}

	expectedS := ""
	for i, x := range expected {
		expectedS += fmt.Sprintf("%d %v\n", i, x)
	}

	if actualS != expectedS {
		t.Fatalf("got\n(%s)\nbut expected\n(%s)", actualS, expectedS)
	}
}
