package dnssort

import "testing"

type domainname struct {
	domain string
	name   string
}

func Test_domaintree(t *testing.T) {

	t.Run("Single domain with name",
		executeTeeTest(
			[]domainname{
				{domain: "example.com", name: "www"},
			},
			[]string{"www.example.com"},
			[]string{"com", "x.example.com", "x.www.example.com", "com", "example.com"},
		),
	)

	t.Run("Single FQDN",
		executeTeeTest(
			[]domainname{
				{domain: "example.com", name: "other.domain.com."},
			},
			[]string{"other.domain.com"},
			[]string{"com", "x.example.com", "x.www.example.com", "example.com"},
		),
	)

	t.Run("Single At sign",
		executeTeeTest(
			[]domainname{
				{domain: "example.com", name: "@"},
			},
			[]string{"example.com"},
			[]string{"com", "x.example.com", "x.www.example.com"},
		),
	)

	t.Run("Wildcard",
		executeTeeTest(
			[]domainname{
				{domain: "example.com", name: "*"},
			},
			[]string{"example.com", "other.example.com"},
			[]string{"com", "example.nl"},
		),
	)

	t.Run("Combined domains",
		executeTeeTest(
			[]domainname{
				{domain: "example.com", name: "*.other"},
				{domain: "example.com", name: "specific"},
				{domain: "example.nl", name: "specific"},
			},
			[]string{"any.other.example.com", "specific.example.com", "specific.example.nl"},
			[]string{"com", "nl", "", "example.nl", "other.nl"},
		),
	)
}

func executeTeeTest(inputs []domainname, founds []string, missings []string) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		tree := CreateTree()
		for _, input := range inputs {
			tree.Add(input.domain, input.name)
		}

		for _, found := range founds {
			if tree.Has(found) == false {
				t.Errorf("Expected %s to be found in tree, but is missing", found)
			}
		}

		for _, missing := range missings {
			if tree.Has(missing) == true {
				t.Errorf("Expected %s to be missing in tree, but is found", missing)
			}
		}
	}
}
