package normalize

import (
	"github.com/StackExchange/dnscontrol/models"
	"testing"
)

func TestImportTransform(t *testing.T) {

	const transformDouble = "0.0.0.0~1.1.1.1~~9.0.0.0,10.0.0.0"
	const transformSingle = "0.0.0.0~1.1.1.1~~8.0.0.0"
	src := &models.DomainConfig{
		Name: "stackexchange.com",
		Records: []*models.RecordConfig{
			{Type: "A", Name: "*", NameFQDN: "*.stackexchange.com", Target: "0.0.2.2"},
			{Type: "A", Name: "www", NameFQDN: "", Target: "0.0.1.1"},
		},
	}
	dst := &models.DomainConfig{
		Name: "internal",
		Records: []*models.RecordConfig{
			{Type: "A", Name: "*.stackexchange.com", NameFQDN: "*.stackexchange.com.internal", Target: "0.0.3.3", Metadata: map[string]string{"transform_table": transformSingle}},
			{Type: "IMPORT_TRANSFORM", Name: "@", Target: "stackexchange.com", Metadata: map[string]string{"transform_table": transformDouble}},
		},
	}
	cfg := &models.DNSConfig{
		Domains: []*models.DomainConfig{src, dst},
	}
	if errs := NormalizeAndValidateConfig(cfg); len(errs) != 0 {
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
