package domaintags

import "testing"

func TestPermitList_Permitted(t *testing.T) {
	// MakeDomainFixForms is not exported, so we can't directly use it here
	// to create complex test cases with IDNs easily without duplicating its logic.
	// However, the existing tests cover a wide range of practical scenarios.
	// For the purpose of this test, we'll assume MakeDomainFixForms works as expected
	// and focus on the logic of the Permitted method itself.

	testCases := []struct {
		name       string
		permitList string
		domain     string
		expected   bool
	}{
		// "all" or empty permit list
		{"all permits everything", "all", "example.com", true},
		{"all permits everything with tag", "all", "example.com!tag1", true},
		{"empty string permits everything", "", "example.com", true},
		{"whitespace string permits everything", "  ", "example.com", true},

		// Simple exact matches
		{"exact match", "example.com", "example.com", true},
		{"exact match with tag", "example.com!tag1", "example.com!tag1", true},
		{"exact mismatch domain", "example.com", "google.com", false},
		{"exact mismatch tag", "example.com!tag1", "example.com!tag2", false},
		{"exact mismatch domain with tag", "example.com!tag1", "google.com!tag1", false},
		{"domain with tag not in list without tag", "example.com", "example.com!tag1", false},
		{"domain without tag not in list with tag", "example.com!tag1", "example.com", false},

		// Wildcard domain name
		{"wildcard domain matches", "*!tag1", "example.com!tag1", true},
		{"wildcard domain mismatch tag", "*!tag1", "example.com!tag2", false},
		{"wildcard domain no tag", "*!tag1", "example.com", false},
		{"wildcard domain and tag", "*", "example.com!tag1", true},
		{"wildcard domain and tag no tag", "*", "example.com", true},

		// Wildcard tag
		{"wildcard tag matches", "example.com!*", "example.com!tag1", true},
		{"wildcard tag matches no tag", "example.com!*", "example.com", true},
		{"wildcard tag mismatch domain", "example.com!*", "google.com!tag1", false},

		// Suffix matching
		{"suffix match base domain", "*.example.com", "example.com", true},
		{"suffix match subdomain", "*.example.com", "foo.example.com", true},
		{"suffix match another subdomain", "*.example.com", "foo.bar.example.com", true},
		{"suffix mismatch different domain", "*.example.com", "google.com", false},
		{"suffix mismatch partial", "*.example.com", "badexample.com", false},
		{"suffix match with tag", "*.example.com!tag1", "foo.example.com!tag1", true},
		{"suffix match base domain with tag", "*.example.com!tag1", "example.com!tag1", true},
		{"suffix mismatch tag", "*.example.com!tag1", "foo.example.com!tag2", false},
		{"suffix mismatch domain with tag", "*.example.com!tag1", "google.com!tag1", false},

		// Multiple items in list
		{"multiple items first match", "google.com,example.com", "google.com", true},
		{"multiple items second match", "google.com,example.com", "example.com", true},
		{"multiple items no match", "google.com,example.com", "other.com", false},
		{"multiple items with tags match", "google.com!tag1,example.com!tag2", "example.com!tag2", true},
		{"multiple items with tags mismatch", "google.com!tag1,example.com!tag2", "example.com!tag1", false},
		{"multiple complex items match", "a.com,*.b.com!tag1,c.com!*", "foo.b.com!tag1", true},
		{"multiple complex items match 2", "a.com,*.b.com!tag1,c.com!*", "c.com!anytag", true},
		{"multiple complex items no match", "a.com,*.b.com!tag1,c.com!*", "foo.b.com!tag2", false},

		// IDN/Unicode cases (assuming MakeDomainFixForms works)
		{"IDN exact match punycode", "xn--e1a4c.com", "xn--e1a4c.com", true}, // д.com
		{"IDN exact match unicode", "д.com", "д.com", true},
		{"IDN mixed match", "xn--e1a4c.com", "д.com", true},
		{"IDN mixed match reversed", "д.com", "xn--e1a4c.com", true},
		{"IDN suffix match punycode", "*.xn--e1a4c.com", "sub.xn--e1a4c.com", true},
		{"IDN suffix match unicode", "*.д.com", "sub.д.com", true},
		{"IDN suffix match mixed", "*.xn--e1a4c.com", "sub.д.com", true},
		{"IDN suffix match mixed reversed", "*.д.com", "sub.xn--e1a4c.com", true},
		{"IDN suffix match base", "*.д.com", "д.com", true},

		// Edge cases
		{"empty list", " ", "example.com", true}, // TrimSpace makes it "", which is "all"
		{"list with empty items", "one.com,,two.com", "one.com", true},
		{"list with empty items 2", "one.com,,two.com", "two.com", true},
		{"list with empty items no match", "one.com,,two.com", "three.com", false},
		{"no match on empty list", "nonexistent", "example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pl := CompilePermitList(tc.permitList)
			got := pl.Permitted(tc.domain)
			if got != tc.expected {
				t.Errorf("PermitList(%q).Permitted(%q) = %v; want %v", tc.permitList, tc.domain, got, tc.expected)
			}
		})
	}
}
