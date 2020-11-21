package normalize

import (
	"strings"
	"testing"

	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

func TestCheckLabel(t *testing.T) {
	var tests = []struct {
		label       string
		rType       string
		target      string
		isError     bool
		hasSkipMeta bool
	}{
		{"@", "A", "zap", false, false},
		{"foo.bar", "A", "zap", false, false},
		{"_foo", "A", "zap", false, false},
		{"_foo", "SRV", "zap", false, false},
		{"_foo", "TLSA", "zap", false, false},
		{"_foo", "TXT", "zap", false, false},
		{"_y2", "CNAME", "foo", false, false},
		{"s1._domainkey", "CNAME", "foo", false, false},
		{"_y3", "CNAME", "asfljds.acm-validations.aws.", false, false},
		{"test.foo.tld", "A", "zap", true, false},
		{"test.foo.tld", "A", "zap", false, true},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%s %s", test.label, test.rType), func(t *testing.T) {
			meta := map[string]string{}
			if test.hasSkipMeta {
				meta["skip_fqdn_check"] = "true"
			}
			err := checkLabel(test.label, test.rType, test.target, "foo.tld", meta)
			if err != nil && !test.isError {
				t.Errorf("%02d: Expected no error but got %s", i, err)
			}
			if err == nil && test.isError {
				t.Errorf("%02d: Expected error but got none", i)
			}
		})

	}
}

func checkError(t *testing.T, err error, shouldError bool, experiment string) {
	if err != nil && !shouldError {
		t.Errorf("%v: Error (%v)\n", experiment, err)
	}
	if err == nil && shouldError {
		t.Errorf("%v: Expected error but got none \n", experiment)
	}
}

func Test_assert_valid_ipv4(t *testing.T) {
	var tests = []struct {
		experiment string
		isError    bool
	}{
		{"1.2.3.4", false},
		{"1.2.3.4/10", true},
		{"1.2.3", true},
		{"foo", true},
	}

	for _, test := range tests {
		err := checkIPv4(test.experiment)
		checkError(t, err, test.isError, test.experiment)
	}
}

func Test_assert_valid_target(t *testing.T) {
	var tests = []struct {
		experiment string
		isError    bool
	}{
		{"@", false},
		{"foo", false},
		{"foo.bar.", false},
		{"foo.", false},
		{"foo.bar", true},
		{"foo&bar", true},
		{"foo bar", true},
		{"elb21.freshdesk.com/", true},
		{"elb21.freshdesk.com/.", true},
	}

	for _, test := range tests {
		err := checkTarget(test.experiment)
		checkError(t, err, test.isError, test.experiment)
	}
}

func Test_transform_cname(t *testing.T) {
	var tests = []struct {
		experiment string
		expected   string
	}{
		{"@", "old.com.new.com."},
		{"foo", "foo.old.com.new.com."},
		{"foo.bar", "foo.bar.old.com.new.com."},
		{"foo.bar.", "foo.bar.new.com."},
		{"chat.stackexchange.com.", "chat.stackexchange.com.new.com."},
	}

	for _, test := range tests {
		actual := transformCNAME(test.experiment, "old.com", "new.com")
		if test.expected != actual {
			t.Errorf("%v: expected (%v) got (%v)\n", test.experiment, test.expected, actual)
		}
	}
}

func TestNSAtRoot(t *testing.T) {
	// do not allow ns records for @
	rec := &models.RecordConfig{Type: "NS"}
	rec.SetLabel("test", "foo.com")
	rec.SetTarget("ns1.name.com.")
	errs := checkTargets(rec, "foo.com")
	if len(errs) > 0 {
		t.Error("Expect no error with ns record on subdomain")
	}
	rec.SetLabel("@", "foo.com")
	errs = checkTargets(rec, "foo.com")
	if len(errs) != 1 {
		t.Error("Expect error with ns record on @")
	}
}

func TestTransforms(t *testing.T) {
	var tests = []struct {
		givenIP         string
		expectedRecords []string
	}{
		{"0.0.5.5", []string{"2.0.5.5"}},
		{"3.0.5.5", []string{"5.5.5.5"}},
		{"7.0.5.5", []string{"9.9.9.9", "10.10.10.10"}},
	}
	const transform = "0.0.0.0~1.0.0.0~2.0.0.0~;   3.0.0.0~4.0.0.0~~5.5.5.5; 7.0.0.0~8.0.0.0~~9.9.9.9,10.10.10.10"
	for i, test := range tests {
		dc := &models.DomainConfig{
			Records: []*models.RecordConfig{
				makeRC("f", "example.tld", test.givenIP, models.RecordConfig{Type: "A", Metadata: map[string]string{"transform": transform}}),
			},
		}
		err := applyRecordTransforms(dc)
		if err != nil {
			t.Errorf("error on test %d: %s", i, err)
			continue
		}
		if len(dc.Records) != len(test.expectedRecords) {
			t.Errorf("test %d: expect %d records but found %d", i, len(test.expectedRecords), len(dc.Records))
			continue
		}
		for r, rec := range dc.Records {
			if rec.GetTargetField() != test.expectedRecords[r] {
				t.Errorf("test %d at index %d: records don't match. Expect %s but found %s.", i, r, test.expectedRecords[r], rec.GetTargetField())
				continue
			}
		}
	}
}

func TestCNAMEMutex(t *testing.T) {
	var recA = &models.RecordConfig{Type: "CNAME"}
	recA.SetLabel("foo", "foo.example.com")
	recA.SetTarget("example.com.")
	tests := []struct {
		rType string
		name  string
		fail  bool
	}{
		{"A", "foo", true},
		{"A", "foo2", false},
		{"CNAME", "foo", true},
		{"CNAME", "foo2", false},
	}
	for _, tst := range tests {
		t.Run(fmt.Sprintf("%s %s", tst.rType, tst.name), func(t *testing.T) {
			var recB = &models.RecordConfig{Type: tst.rType}
			recB.SetLabel(tst.name, "example.com")
			recB.SetTarget("example2.com.")
			dc := &models.DomainConfig{
				Name:    "example.com",
				Records: []*models.RecordConfig{recA, recB},
			}
			errs := checkCNAMEs(dc)
			if errs != nil && !tst.fail {
				t.Error("Got error but expected none")
			}
			if errs == nil && tst.fail {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestCAAValidation(t *testing.T) {
	config := &models.DNSConfig{
		Domains: []*models.DomainConfig{
			{
				Name:          "example.com",
				RegistrarName: "BIND",
				Records: []*models.RecordConfig{
					makeRC("@", "example.com", "example.com", models.RecordConfig{Type: "CAA", CaaTag: "invalid"}),
				},
			},
		},
	}
	errs := ValidateAndNormalizeConfig(config)
	if len(errs) != 1 {
		t.Error("Expect error on invalid CAA but got none")
	}
}

func TestTXTValidation(t *testing.T) {
	tests := []struct {
		name   string
		record string
		fail   bool
	}{
		{"emoji", "👍🏼", true},
		{"latin1", "\u00ff", false},                    // anything <= u00FF should be supported
		{"long", strings.Repeat("\u00ff", 255), false}, // ensure 255 characters for <= u00FF
	}
	for _, tst := range tests {
		t.Run(fmt.Sprintf("%s", tst.name), func(t *testing.T) {
			config := &models.DNSConfig{
				Domains: []*models.DomainConfig{
					{
						Name:          "example.com",
						RegistrarName: "BIND",
						Records: []*models.RecordConfig{
							makeRC(tst.name, "example.com", "example.com", models.RecordConfig{Type: "TXT", TxtStrings: []string{tst.record}}),
						},
					},
				},
			}
			errs := ValidateAndNormalizeConfig(config)
			if errs != nil && !tst.fail {
				t.Error(errs)
			}
			if errs == nil && tst.fail {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestCheckDuplicates(t *testing.T) {
	records := []*models.RecordConfig{
		// The only difference is the target:
		makeRC("www", "example.com", "4.4.4.4", models.RecordConfig{Type: "A"}),
		makeRC("www", "example.com", "5.5.5.5", models.RecordConfig{Type: "A"}),
		// The only difference is the rType:
		makeRC("aaa", "example.com", "uniquestring.com.", models.RecordConfig{Type: "NS"}),
		makeRC("aaa", "example.com", "uniquestring.com.", models.RecordConfig{Type: "PTR"}),
		// The only difference is the TTL.
		makeRC("zzz", "example.com", "4.4.4.4", models.RecordConfig{Type: "A", TTL: 111}),
		makeRC("zzz", "example.com", "4.4.4.4", models.RecordConfig{Type: "A", TTL: 222}),
		// Three records each with a different target.
		makeRC("@", "example.com", "ns1.foo.com.", models.RecordConfig{Type: "NS"}),
		makeRC("@", "example.com", "ns2.foo.com.", models.RecordConfig{Type: "NS"}),
		makeRC("@", "example.com", "ns3.foo.com.", models.RecordConfig{Type: "NS"}),
	}
	errs := checkDuplicates(records)
	if len(errs) != 0 {
		t.Errorf("Expect duplicate NOT found but found %q", errs)
	}
}

func TestCheckDuplicates_dup_a(t *testing.T) {
	records := []*models.RecordConfig{
		// A records that are exact dupliates.
		makeRC("@", "example.com", "1.1.1.1", models.RecordConfig{Type: "A"}),
		makeRC("@", "example.com", "1.1.1.1", models.RecordConfig{Type: "A"}),
	}
	errs := checkDuplicates(records)
	if len(errs) == 0 {
		t.Error("Expect duplicate found but found none")
	}
}

func TestCheckDuplicates_dup_ns(t *testing.T) {
	records := []*models.RecordConfig{
		// Three records, the last 2 are duplicates.
		// NB: This is a common issue.
		makeRC("@", "example.com", "ns1.foo.com.", models.RecordConfig{Type: "NS"}),
		makeRC("@", "example.com", "ns2.foo.com.", models.RecordConfig{Type: "NS"}),
		makeRC("@", "example.com", "ns2.foo.com.", models.RecordConfig{Type: "NS"}),
	}
	errs := checkDuplicates(records)
	if len(errs) == 0 {
		t.Error("Expect duplicate found but found none")
	}
}

func TestTLSAValidation(t *testing.T) {
	config := &models.DNSConfig{
		Domains: []*models.DomainConfig{
			{
				Name:          "_443._tcp.example.com",
				RegistrarName: "BIND",
				Records: []*models.RecordConfig{
					makeRC("_443._tcp", "_443._tcp.example.com", "abcdef0", models.RecordConfig{
						Type: "TLSA", TlsaUsage: 4, TlsaSelector: 1, TlsaMatchingType: 1}),
				},
			},
		},
	}
	errs := ValidateAndNormalizeConfig(config)
	if len(errs) != 1 {
		t.Error("Expect error on invalid TLSA but got none")
	}
}

const (
	ProviderNoDS        = "NO_DS_SUPPORT"
	ProviderFullDS      = "FULL_DS_SUPPORT"
	ProviderChildDSOnly = "CHILD_DS_SUPPORT"
	ProviderBothDSCaps  = "BOTH_DS_CAPABILITIES"
)

func init() {
	providers.RegisterDomainServiceProviderType(ProviderNoDS, nil, providers.DocumentationNotes{})
	providers.RegisterDomainServiceProviderType(ProviderFullDS, nil, providers.DocumentationNotes{
		providers.CanUseDS: providers.Can(),
	})
	providers.RegisterDomainServiceProviderType(ProviderChildDSOnly, nil, providers.DocumentationNotes{
		providers.CanUseDSForChildren: providers.Can(),
	})
	providers.RegisterDomainServiceProviderType(ProviderBothDSCaps, nil, providers.DocumentationNotes{
		providers.CanUseDS:            providers.Can(),
		providers.CanUseDSForChildren: providers.Can(),
	})
}

func Test_DSChecks(t *testing.T) {
	t.Run("no DS support", func(t *testing.T) {
		err := checkProviderDS(ProviderNoDS, nil)
		if err == nil {
			t.Errorf("Provider %s implements no DS capabilities, so should have failed the check", ProviderNoDS)
		}
	})

	t.Run("full DS support", func(t *testing.T) {
		apexDS := models.RecordConfig{Type: "DS"}
		apexDS.SetLabel("@", "example.com")

		childDS := models.RecordConfig{Type: "DS"}
		childDS.SetLabel("child", "example.com")

		records := models.Records{&apexDS, &childDS}

		// check permutations of ProviderCanDS and having both DS caps
		for _, pType := range []string{ProviderFullDS, ProviderBothDSCaps} {
			err := checkProviderDS(pType, records)
			if err != nil {
				t.Errorf("Provider %s implements full DS capabilities and should process the provided records", ProviderFullDS)
			}
		}
	})

	t.Run("child DS support only", func(t *testing.T) {
		apexDS := models.RecordConfig{Type: "DS"}
		apexDS.SetLabel("@", "example.com")

		childDS := models.RecordConfig{Type: "DS"}
		childDS.SetLabel("child", "example.com")

		// this record is included at the apex to check the Type of the
		// recordset is verified to only inspect records with type == DS
		apexA := models.RecordConfig{Type: "A"}
		apexA.SetLabel("@", "example.com")

		t.Run("accepts when child DS records only", func(t *testing.T) {
			records := models.Records{&childDS, &apexA}
			err := checkProviderDS(ProviderChildDSOnly, records)
			if err != nil {
				t.Errorf("Provider %s implements child DS support so the provided records should be accepted",
					ProviderChildDSOnly,
				)
			}
		})

		t.Run("fails with apex and child DS records", func(t *testing.T) {
			records := models.Records{&apexDS, &childDS, &apexA}
			err := checkProviderDS(ProviderChildDSOnly, records)
			if err == nil {
				t.Errorf("Provider %s does not implement DS support at the zone apex, so should reject provided records",
					ProviderChildDSOnly,
				)
			}
		})
	})
}
