package cloudflare

import (
	"net"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/transform"
)

func newDomainConfig() *models.DomainConfig {
	return &models.DomainConfig{
		Name:     "test.com",
		Records:  []*models.RecordConfig{},
		Metadata: map[string]string{},
	}
}

func makeRCmeta(meta map[string]string) *models.RecordConfig {
	rc := models.RecordConfig{
		Type:     "A",
		Metadata: meta,
	}
	rc.SetLabel("foo", "example.tld")
	rc.SetTarget("1.2.3.4")
	return &rc
}

func TestPreprocess_BoolValidation(t *testing.T) {
	cf := &cloudflareProvider{}

	domain := newDomainConfig()
	domain.Records = append(domain.Records, makeRCmeta(map[string]string{metaProxy: "on"}))
	domain.Records = append(domain.Records, makeRCmeta(map[string]string{metaProxy: "fUll"}))
	domain.Records = append(domain.Records, makeRCmeta(map[string]string{}))
	domain.Records = append(domain.Records, makeRCmeta(map[string]string{metaProxy: "Off"}))
	domain.Records = append(domain.Records, makeRCmeta(map[string]string{metaProxy: "off"}))
	err := cf.preprocessConfig(domain)
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"on", "full", "off", "off", "off"}
	// make sure only "on" or "off", and "full" are actually set
	for i, rec := range domain.Records {
		if rec.Metadata[metaProxy] != expected[i] {
			t.Fatalf("At index %d: expect '%s' but found '%s'", i, expected[i], rec.Metadata[metaProxy])
		}
	}
}

func TestPreprocess_BoolValidation_Fails(t *testing.T) {
	cf := &cloudflareProvider{}
	domain := newDomainConfig()
	domain.Records = append(domain.Records, &models.RecordConfig{Metadata: map[string]string{metaProxy: "true"}})
	err := cf.preprocessConfig(domain)
	if err == nil {
		t.Fatal("Expected validation error, but got none")
	}
}

func TestPreprocess_DefaultProxy(t *testing.T) {
	cf := &cloudflareProvider{}
	domain := newDomainConfig()
	domain.Metadata[metaProxyDefault] = "full"
	domain.Records = append(domain.Records, makeRCmeta(map[string]string{metaProxy: "on"}))
	domain.Records = append(domain.Records, makeRCmeta(map[string]string{metaProxy: "off"}))
	domain.Records = append(domain.Records, makeRCmeta(map[string]string{}))
	err := cf.preprocessConfig(domain)
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"on", "off", "full"}
	for i, rec := range domain.Records {
		if rec.Metadata[metaProxy] != expected[i] {
			t.Fatalf("At index %d: expect '%s' but found '%s'", i, expected[i], rec.Metadata[metaProxy])
		}
	}
}

func TestPreprocess_DefaultProxy_Validation(t *testing.T) {
	cf := &cloudflareProvider{}
	domain := newDomainConfig()
	domain.Metadata[metaProxyDefault] = "true"
	err := cf.preprocessConfig(domain)
	if err == nil {
		t.Fatal("Expected validation error, but got none")
	}
}

func TestIpRewriting(t *testing.T) {
	var tests = []struct {
		Given, Expected string
		Proxy           string
	}{
		// outside of range
		{"5.5.5.5", "5.5.5.5", "full"},
		{"5.5.5.5", "5.5.5.5", "on"},
		// inside range, but not proxied
		{"1.2.3.4", "1.2.3.4", "on"},
		// inside range and proxied
		{"1.2.3.4", "255.255.255.4", "full"},
	}
	cf := &cloudflareProvider{}
	domain := newDomainConfig()
	cf.ipConversions = []transform.IPConversion{{
		Low:      net.ParseIP("1.2.3.0"),
		High:     net.ParseIP("1.2.3.40"),
		NewBases: []net.IP{net.ParseIP("255.255.255.0")},
		NewIPs:   nil}}
	for _, tst := range tests {
		rec := &models.RecordConfig{Type: "A", Metadata: map[string]string{metaProxy: tst.Proxy}}
		rec.SetTarget(tst.Given)
		domain.Records = append(domain.Records, rec)
	}
	err := cf.preprocessConfig(domain)
	if err != nil {
		t.Fatal(err)
	}
	for i, tst := range tests {
		rec := domain.Records[i]
		if rec.GetTargetField() != tst.Expected {
			t.Fatalf("At index %d, expected target of %s, but found %s.", i, tst.Expected, rec.GetTargetField())
		}
		if tst.Proxy == "full" && tst.Given != tst.Expected && rec.Metadata[metaOriginalIP] != tst.Given {
			t.Fatalf("At index %d, expected original_ip to be set", i)
		}
	}
}
