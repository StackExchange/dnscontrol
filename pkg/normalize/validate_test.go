package normalize

import (
	"testing"

	"fmt"

	"github.com/StackExchange/dnscontrol/models"
)

func TestCheckLabel(t *testing.T) {
	var tests = []struct {
		experiment string
		isError    bool
	}{
		{"@", false},
		{"foo", false},
		{"foo.bar", false},
		{"foo.", true},
		{"foo.bar.", true},
		{"foo_bar", true},
		{"_domainkey", false},
	}

	for _, test := range tests {
		err := checkLabel(test.experiment, "A", "foo.com")
		checkError(t, err, test.isError, test.experiment)
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
	//do not allow ns records for @
	rec := &models.RecordConfig{Name: "test", Type: "NS", Target: "ns1.name.com."}
	errs := checkTargets(rec, "foo.com")
	if len(errs) > 0 {
		t.Error("Expect no error with ns record on subdomain")
	}
	rec.Name = "@"
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
				{Type: "A", Target: test.givenIP, Metadata: map[string]string{"transform": transform}},
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
			if rec.Target != test.expectedRecords[r] {
				t.Errorf("test %d at index %d: records don't match. Expect %s but found %s.", i, r, test.expectedRecords[r], rec.Target)
				continue
			}
		}
	}
}

func TestCNAMEMutex(t *testing.T) {
	var recA = &models.RecordConfig{Type: "CNAME", Name: "foo", NameFQDN: "foo.example.com", Target: "example.com."}
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
			var recB = &models.RecordConfig{Type: tst.rType, Name: tst.name, NameFQDN: tst.name + ".example.com", Target: "example2.com."}
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
				Name:      "example.com",
				Registrar: "BIND",
				Records: []*models.RecordConfig{
					{Name: "@", Type: "CAA", CaaTag: "invalid", Target: "example.com"},
				},
			},
		},
	}
	errs := NormalizeAndValidateConfig(config)
	if len(errs) != 1 {
		t.Error("Expect error on invalid CAA but got none")
	}
}
