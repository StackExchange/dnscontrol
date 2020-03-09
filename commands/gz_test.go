package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	_ "github.com/StackExchange/dnscontrol/v2/providers/_all"
)

func TestFormatTypes(t *testing.T) {
	/*
	  Input:                  Converted to:  Should match contents of:
	  test_data/$DOMAIN.zone  js             test_data/$DOMAIN.zone.js
	  test_data/$DOMAIN.zone  tsv            test_data/$DOMAIN.zone.tsv
	  test_data/$DOMAIN.zone  zone           test_data/$DOMAIN.zone.zone
	*/

	for _, domain := range []string{"simple.com", "example.org"} {
		t.Run(domain+"%s/js", func(t *testing.T) { testFormat(t, domain, "js") })
		t.Run(domain+"%s/tsv", func(t *testing.T) { testFormat(t, domain, "tsv") })
		t.Run(domain+"%s/zone", func(t *testing.T) { testFormat(t, domain, "zone") })
	}
}

func TestFormatLoop(t *testing.T) {
	/*
		  Use the .js file that is generated to create a zonefile.
			The records should be the same as the zonefile.
	*/

	//	for _, domain := range []string{"simple.com", "example.org"} {
	//		// Go from the sample zonefile to .js:
	//		testFormat(t, domain, "js")
	//		// Go from .js to the zonefile.
	//		jsToZone(t, domain)
	//		// Compare results.
	//	}
}

func testFormat(t *testing.T, domain, format string) {
	t.Helper()

	//sourceFilename := fmt.Sprintf("test_data/%s.zone", domain)
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
	gzargs.CredsFile = "test_data/creds.json"

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

	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("TestFormatTypes mismatch (-want +got):\n%s", diff)
	}

}
