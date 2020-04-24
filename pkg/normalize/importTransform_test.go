package normalize

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func makeRC(label, domain, target string, rc models.RecordConfig) *models.RecordConfig {
	rc.SetLabel(label, domain)
	rc.SetTarget(target)
	return &rc
}

func TestImportTransform(t *testing.T) {

	const transformDouble = "0.0.0.0~1.1.1.1~~9.0.0.0,10.0.0.0"
	const transformSingle = "0.0.0.0~1.1.1.1~~8.0.0.0"
	src := &models.DomainConfig{
		Name: "stackexchange.com",
		Records: []*models.RecordConfig{
			makeRC("*", "stackexchange.com", "0.0.2.2", models.RecordConfig{Type: "A"}),
			makeRC("www", "stackexchange.com", "0.0.1.1", models.RecordConfig{Type: "A"}),
		},
	}
	dst := &models.DomainConfig{
		Name: "internal",
		Records: []*models.RecordConfig{
			makeRC("*.stackexchange.com", "*.stackexchange.com.internal", "0.0.3.3", models.RecordConfig{Type: "A", Metadata: map[string]string{"transform_table": transformSingle}}),
			makeRC("@", "internal", "stackexchange.com", models.RecordConfig{Type: "IMPORT_TRANSFORM", Metadata: map[string]string{"transform_table": transformDouble}}),
		},
	}
	cfg := &models.DNSConfig{
		Domains: []*models.DomainConfig{src, dst},
	}
	if errs := ValidateAndNormalizeConfig(cfg); len(errs) != 0 {
		for _, err := range errs {
			t.Error(err)
		}
		t.FailNow()
	}
	d := cfg.FindDomain("internal")
	if len(d.Records) != 3 {
		for _, r := range d.Records {
			t.Error(r)
		}
		t.Fatalf("Expected 3 records in internal, but got %d", len(d.Records))
	}
}
