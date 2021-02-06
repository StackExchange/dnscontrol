package activedir

import (
	"fmt"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func makeRC(label, domain, target string, rc models.RecordConfig) *models.RecordConfig {
	rc.SetLabel(label, domain)
	rc.SetTarget(target)
	return &rc
}

func TestGetExistingRecords(t *testing.T) {

	cf := &activedirProvider{}

	cf.fake = true
	actual, err := cf.getExistingRecords("test2")
	if err != nil {
		t.Fatal(err)
	}
	expected := []*models.RecordConfig{
		makeRC("@", "test2", "10.166.2.11", models.RecordConfig{Type: "A", TTL: 600}),
		makeRC("_msdcs", "test2", "other_record", models.RecordConfig{Type: "NS", TTL: 300}),
		makeRC("co-devsearch02", "test2", "10.8.2.64", models.RecordConfig{Type: "A", TTL: 3600}),
		makeRC("co-devservice01", "test2", "10.8.2.48", models.RecordConfig{Type: "A", TTL: 1200}), // Downcased.
		makeRC("yum", "test2", "10.8.0.59", models.RecordConfig{Type: "A", TTL: 3600}),
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
