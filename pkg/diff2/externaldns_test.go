package diff2

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func makeTestRecord(name, rtype, target, domain string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: rtype,
	}
	rc.SetLabel(name, domain)
	if rtype == "TXT" {
		rc.SetTargetTXT(target)
	} else {
		rc.SetTarget(target)
	}
	return rc
}

func TestIsExternalDNSTxtRecord(t *testing.T) {
	domain := "example.com"

	tests := []struct {
		name           string
		record         *models.RecordConfig
		wantIsExtDNS   bool
		wantLabel      string
		wantRecordType string
	}{
		{
			name:           "external-dns A record TXT",
			record:         makeTestRecord("a-myapp", "TXT", "heritage=external-dns,external-dns/owner=my-cluster,external-dns/resource=ingress/default/myapp", domain),
			wantIsExtDNS:   true,
			wantLabel:      "myapp",
			wantRecordType: "A",
		},
		{
			name:           "external-dns AAAA record TXT",
			record:         makeTestRecord("aaaa-myapp", "TXT", "heritage=external-dns,external-dns/owner=my-cluster", domain),
			wantIsExtDNS:   true,
			wantLabel:      "myapp",
			wantRecordType: "AAAA",
		},
		{
			name:           "external-dns CNAME record TXT",
			record:         makeTestRecord("cname-www", "TXT", "heritage=external-dns,external-dns/owner=default", domain),
			wantIsExtDNS:   true,
			wantLabel:      "www",
			wantRecordType: "CNAME",
		},
		{
			name:           "external-dns apex A record TXT",
			record:         makeTestRecord("a-", "TXT", "heritage=external-dns,external-dns/owner=k8s", domain),
			wantIsExtDNS:   true,
			wantLabel:      "@",
			wantRecordType: "A",
		},
		{
			name:         "non-external-dns TXT record",
			record:       makeTestRecord("myapp", "TXT", "some random txt content", domain),
			wantIsExtDNS: false,
		},
		{
			name:         "A record (not TXT)",
			record:       makeTestRecord("myapp", "A", "1.2.3.4", domain),
			wantIsExtDNS: false,
		},
		{
			name:         "SPF record",
			record:       makeTestRecord("@", "TXT", "v=spf1 include:_spf.google.com ~all", domain),
			wantIsExtDNS: false,
		},
		{
			name:         "DKIM record",
			record:       makeTestRecord("selector._domainkey", "TXT", "v=DKIM1; k=rsa; p=MIGfMA0G...", domain),
			wantIsExtDNS: false,
		},
		{
			name:           "external-dns with quoted heritage",
			record:         makeTestRecord("a-test", "TXT", "\"heritage=external-dns,external-dns/owner=test\"", domain),
			wantIsExtDNS:   true,
			wantLabel:      "test",
			wantRecordType: "A",
		},
		{
			name:           "external-dns legacy format (no prefix)",
			record:         makeTestRecord("legacy-app", "TXT", "heritage=external-dns,external-dns/owner=old-cluster", domain),
			wantIsExtDNS:   true,
			wantLabel:      "legacy-app",
			wantRecordType: "", // Empty means match any type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsExtDNS, gotInfo := isExternalDNSTxtRecord(tt.record, domain, "")

			if gotIsExtDNS != tt.wantIsExtDNS {
				t.Errorf("isExternalDNSTxtRecord() isExtDNS = %v, want %v", gotIsExtDNS, tt.wantIsExtDNS)
			}

			if tt.wantIsExtDNS {
				if gotInfo == nil {
					t.Errorf("isExternalDNSTxtRecord() returned nil info for external-dns record")
					return
				}
				if gotInfo.Label != tt.wantLabel {
					t.Errorf("isExternalDNSTxtRecord() label = %q, want %q", gotInfo.Label, tt.wantLabel)
				}
				if gotInfo.RecordType != tt.wantRecordType {
					t.Errorf("isExternalDNSTxtRecord() recordType = %q, want %q", gotInfo.RecordType, tt.wantRecordType)
				}
			}
		})
	}
}

func TestParseExternalDNSTxtLabel(t *testing.T) {
	tests := []struct {
		name           string
		label          string
		wantLabel      string
		wantRecordType string
	}{
		{
			name:           "A record prefix",
			label:          "a-myapp",
			wantLabel:      "myapp",
			wantRecordType: "A",
		},
		{
			name:           "AAAA record prefix",
			label:          "aaaa-myapp",
			wantLabel:      "myapp",
			wantRecordType: "AAAA",
		},
		{
			name:           "CNAME record prefix",
			label:          "cname-www",
			wantLabel:      "www",
			wantRecordType: "CNAME",
		},
		{
			name:           "NS record prefix",
			label:          "ns-subdomain",
			wantLabel:      "subdomain",
			wantRecordType: "NS",
		},
		{
			name:           "MX record prefix",
			label:          "mx-mail",
			wantLabel:      "mail",
			wantRecordType: "MX",
		},
		{
			name:           "apex domain (A)",
			label:          "a-",
			wantLabel:      "@",
			wantRecordType: "A",
		},
		{
			name:           "uppercase prefix (should handle case)",
			label:          "A-myapp",
			wantLabel:      "myapp",
			wantRecordType: "A",
		},
		{
			name:           "no recognized prefix (legacy)",
			label:          "myapp",
			wantLabel:      "myapp",
			wantRecordType: "",
		},
		{
			name:           "subdomain with dots in name",
			label:          "a-sub.domain",
			wantLabel:      "sub.domain",
			wantRecordType: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseExternalDNSTxtLabel(tt.label, "")

			if got.Label != tt.wantLabel {
				t.Errorf("parseExternalDNSTxtLabel() Label = %q, want %q", got.Label, tt.wantLabel)
			}
			if got.RecordType != tt.wantRecordType {
				t.Errorf("parseExternalDNSTxtLabel() RecordType = %q, want %q", got.RecordType, tt.wantRecordType)
			}
		})
	}
}

func TestFindExternalDNSManagedRecords(t *testing.T) {
	domain := "example.com"

	existing := models.Records{
		// External-dns managed records
		makeTestRecord("a-myapp", "TXT", "heritage=external-dns,external-dns/owner=cluster1", domain),
		makeTestRecord("myapp", "A", "10.0.0.1", domain),
		makeTestRecord("cname-www", "TXT", "heritage=external-dns,external-dns/owner=cluster1", domain),
		makeTestRecord("www", "CNAME", "myapp.example.com.", domain),
		// Non-external-dns records
		makeTestRecord("static", "A", "1.2.3.4", domain),
		makeTestRecord("@", "TXT", "v=spf1 -all", domain),
		makeTestRecord("@", "MX", "mail.example.com.", domain),
	}

	managed := findExternalDNSManagedRecords(existing, domain, "")

	// Check that expected keys are present
	expectedKeys := []string{
		"a-myapp:TXT",   // The TXT record itself
		"myapp:A",       // The A record it manages
		"cname-www:TXT", // The TXT record itself
		"www:CNAME",     // The CNAME record it manages
	}

	for _, key := range expectedKeys {
		if !managed[key] {
			t.Errorf("Expected key %q to be marked as managed", key)
		}
	}

	// Check that non-external-dns records are not marked
	notExpected := []string{
		"static:A",
		"@:TXT",
		"@:MX",
	}

	for _, key := range notExpected {
		if managed[key] {
			t.Errorf("Key %q should not be marked as managed", key)
		}
	}
}

func TestGetExternalDNSIgnoredRecords(t *testing.T) {
	domain := "example.com"

	existing := models.Records{
		// External-dns managed records
		makeTestRecord("a-myapp", "TXT", "heritage=external-dns,external-dns/owner=cluster1", domain),
		makeTestRecord("myapp", "A", "10.0.0.1", domain),
		// Non-external-dns records
		makeTestRecord("static", "A", "1.2.3.4", domain),
		makeTestRecord("@", "TXT", "v=spf1 -all", domain),
	}

	ignored := GetExternalDNSIgnoredRecords(existing, domain, "")

	if len(ignored) != 2 {
		t.Errorf("Expected 2 ignored records, got %d", len(ignored))
	}

	// Verify the ignored records
	foundTXT := false
	foundA := false
	for _, rec := range ignored {
		if rec.GetLabel() == "a-myapp" && rec.Type == "TXT" {
			foundTXT = true
		}
		if rec.GetLabel() == "myapp" && rec.Type == "A" {
			foundA = true
		}
	}

	if !foundTXT {
		t.Error("Expected TXT record a-myapp to be ignored")
	}
	if !foundA {
		t.Error("Expected A record myapp to be ignored")
	}
}

func TestGetExternalDNSIgnoredRecords_NoExternalDNS(t *testing.T) {
	domain := "example.com"

	existing := models.Records{
		makeTestRecord("static", "A", "1.2.3.4", domain),
		makeTestRecord("@", "TXT", "v=spf1 -all", domain),
		makeTestRecord("www", "CNAME", "static.example.com.", domain),
	}

	ignored := GetExternalDNSIgnoredRecords(existing, domain, "")

	if len(ignored) != 0 {
		t.Errorf("Expected 0 ignored records when no external-dns records exist, got %d", len(ignored))
	}
}

func TestGetExternalDNSIgnoredRecords_LegacyFormat(t *testing.T) {
	domain := "example.com"

	// Legacy format: TXT record has same name as the record it manages (no type prefix)
	existing := models.Records{
		makeTestRecord("legacyapp", "TXT", "heritage=external-dns,external-dns/owner=old-cluster", domain),
		makeTestRecord("legacyapp", "A", "10.0.0.1", domain),
		makeTestRecord("legacyapp", "AAAA", "::1", domain),
	}

	ignored := GetExternalDNSIgnoredRecords(existing, domain, "")

	// Legacy format should match the TXT and common record types
	if len(ignored) < 3 {
		t.Errorf("Expected at least 3 ignored records for legacy format, got %d", len(ignored))
	}
}

func TestGetExternalDNSIgnoredRecords_CustomPrefix(t *testing.T) {
	domain := "example.com"

	// Custom prefix format: e.g., "extdns-www" for "www" record
	existing := models.Records{
		makeTestRecord("extdns-www", "TXT", "heritage=external-dns,external-dns/owner=k3s-cluster", domain),
		makeTestRecord("www", "A", "10.0.0.1", domain),
		makeTestRecord("extdns-api", "TXT", "heritage=external-dns,external-dns/owner=k3s-cluster", domain),
		makeTestRecord("api", "CNAME", "app.example.com.", domain),
		// Non-external-dns records
		makeTestRecord("static", "A", "1.2.3.4", domain),
	}

	// Without custom prefix, only the TXT records themselves should be detected
	// (but the A/CNAME won't be linked because "extdns-" isn't a known type prefix)
	ignoredDefault := GetExternalDNSIgnoredRecords(existing, domain, "")

	// With custom prefix, both TXT and their managed records should be detected
	ignoredCustom := GetExternalDNSIgnoredRecords(existing, domain, "extdns-")

	// Custom prefix should find more records
	if len(ignoredCustom) <= len(ignoredDefault) {
		t.Errorf("Expected custom prefix to find more records: default=%d, custom=%d",
			len(ignoredDefault), len(ignoredCustom))
	}

	// Should find: extdns-www:TXT, www:A/AAAA/CNAME/etc, extdns-api:TXT, api:A/AAAA/CNAME/etc
	if len(ignoredCustom) < 4 {
		t.Errorf("Expected at least 4 ignored records with custom prefix, got %d", len(ignoredCustom))
	}
}

func TestParseExternalDNSTxtLabel_CustomPrefix(t *testing.T) {
	tests := []struct {
		name           string
		label          string
		customPrefix   string
		wantLabel      string
		wantRecordType string
	}{
		{
			name:           "custom prefix with record type",
			label:          "extdns-a-myapp",
			customPrefix:   "extdns-",
			wantLabel:      "myapp",
			wantRecordType: "A",
		},
		{
			name:           "custom prefix without record type",
			label:          "extdns-www",
			customPrefix:   "extdns-",
			wantLabel:      "www",
			wantRecordType: "", // No record type in this format
		},
		{
			name:           "custom prefix apex domain",
			label:          "extdns-",
			customPrefix:   "extdns-",
			wantLabel:      "@",
			wantRecordType: "",
		},
		{
			name:           "prefix not found - fallback to legacy",
			label:          "other-www",
			customPrefix:   "extdns-",
			wantLabel:      "other-www",
			wantRecordType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseExternalDNSTxtLabel(tt.label, tt.customPrefix)

			if got.Label != tt.wantLabel {
				t.Errorf("parseExternalDNSTxtLabel() Label = %q, want %q", got.Label, tt.wantLabel)
			}
			if got.RecordType != tt.wantRecordType {
				t.Errorf("parseExternalDNSTxtLabel() RecordType = %q, want %q", got.RecordType, tt.wantRecordType)
			}
		})
	}
}
