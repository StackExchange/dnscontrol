package spf

import "testing"
import "strings"
import "fmt"

func TestParse(t *testing.T) {
	rec, err := Parse(strings.Join([]string{"v=spf1",
		"ip4:198.252.206.0/24",
		"ip4:192.111.0.0/24",
		"include:_spf.google.com",
		"include:mailgun.org",
		"include:spf-basic.fogcreek.com",
		"include:mail.zendesk.com",
		"include:servers.mcsv.net",
		"include:sendgrid.net",
		"include:spf.mtasv.net",
		"~all"}, " "))
	if err != nil {
		t.Fatal(err)
	}
	printSPF(rec, "")
}

func printSPF(rec *SPFRecord, indent string) {
	fmt.Printf("%sTotal Lookups: %d\n", indent, rec.Lookups)
	fmt.Print(indent + "v=spf1")
	for _, p := range rec.Parts {
		fmt.Print(" " + p.Text)
	}
	fmt.Println()
	indent += "\t"
	for _, p := range rec.Parts {
		if p.IncludeRecord != nil {
			fmt.Println(indent + p.Text)
			printSPF(p.IncludeRecord, indent+"\t")
		}
	}
}
