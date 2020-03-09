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

/*

This file does a number of tests:

// Basic conversions:        OUTPUTFILE:        EXPECTED FILE:
  example.org.zone -> tsv      $NAME.tsv          *.expected
  example.org.zone -> jsd      $NAME.jsd          *.expected
  example.org.zone -> zonefile $NAME.zon          *.expected

// Loop test:
  zonefile -> dsl              $NAME.zon.dsl      *.expected
  dsl -> zonefile              $NAME.zon.dsl.zon  *.expected

*/

func TestFormatTypes(t *testing.T) {
	for i, domainname := range []string{"simple.com", "example.org"} {
		t.Run(fmt.Sprintf("%s/tsv", domain), func(t *testing.T) { testFormat(t, domainname, "tsv") })
		t.Run(fmt.Sprintf("%s/zon", domain), func(t *testing.T) { testFormat(t, domainname, "zon") })
		t.Run(fmt.Sprintf("%s/js", domain), func(t *testing.T) { testFormat(t, domainname, "js") })
	}
}

func testFormat(t *testing.T, domainname, format string) {
	t.Helper()

	sourceFilename := fmt.Sprintf("test_data/%s.zone", domainname)
	expectedFilename := fmt.Sprintf("test_data/%s.zone.%s", domainname, format)
	outputFiletmpl := fmt.Sprintf("%s.zone.%s.*.txt", domainname, format)

	outfile, err := ioutil.TempFile("", outputFiletmpl)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(outfile.Name())

	// Convert test data to the experiment output.
	gzargs := GetZoneArgs{
		ZoneNames:    []string{domainname}
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

	// Copy got -> want
	err = ioutil.WriteFile(expectedFilename, got, 0644)
	if err != nil {
		log.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestFormatTypes mismatch (-want +got):\n%s", diff)
	}

}
