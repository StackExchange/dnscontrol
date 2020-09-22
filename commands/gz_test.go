package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/andreyvit/diff"

	_ "github.com/StackExchange/dnscontrol/v3/providers/_all"
)

func TestFormatTypes(t *testing.T) {
	/*
	  Input:                   Converted to:   Should match contents of:
	  test_data/$DOMAIN.zone   js              test_data/$DOMAIN.zone.js
	  test_data/$DOMAIN.zone   tsv             test_data/$DOMAIN.zone.tsv
	  test_data/$DOMAIN.zone   zone            test_data/$DOMAIN.zone.zone
	*/

	for _, domain := range []string{"simple.com", "example.org"} {
		t.Run(domain+"/js", func(t *testing.T) { testFormat(t, domain, "js") })
		t.Run(domain+"/tsv", func(t *testing.T) { testFormat(t, domain, "tsv") })
		t.Run(domain+"/zone", func(t *testing.T) { testFormat(t, domain, "zone") })
	}
}

func testFormat(t *testing.T, domain, format string) {
	t.Helper()

	expectedFilename := fmt.Sprintf("test_data/%s.zone.%s", domain, format)
	outputFiletmpl := fmt.Sprintf("%s.zone.%s.*.txt", domain, format)

	outfile, err := ioutil.TempFile("", outputFiletmpl)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(outfile.Name())

	// Convert test data to the experiment output.
	gzargs := GetZoneArgs{
		ZoneNames:    []string{domain},
		OutputFormat: format,
		OutputFile:   outfile.Name(),
		CredName:     "bind",
		ProviderName: "BIND",
	}
	gzargs.CredsFile = "test_data/bind-creds.json"

	// Read the zonefile and convert
	err = GetZone(gzargs)
	if err != nil {
		log.Fatal(err)
	}

	// Read the actual result:
	got, err := ioutil.ReadFile(outfile.Name())
	if err != nil {
		log.Fatal(err)
	}

	// Read the expected result
	want, err := ioutil.ReadFile(expectedFilename)
	if err != nil {
		log.Fatal(err)
	}

	//	// Update got -> want
	//	err = ioutil.WriteFile(expectedFilename, got, 0644)
	//	if err != nil {
	//		log.Fatal(err)
	//	}

	if w, g := string(want), string(got); w != g {
		t.Errorf("testFormat mismatch (-got +want):\n%s", diff.LineDiff(g, w))
	}

}
