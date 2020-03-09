package commands

import (
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
  example.org.zone -> dsl      $NAME.dsl          *.expected
  example.org.zone -> zonefile $NAME.zon          *.expected

// Loop test:
  zonefile -> dsl              $NAME.zon.dsl      *.expected
  dsl -> zonefile              $NAME.zon.dsl.zon  *.expected

*/

func TestZoneToZone(t *testing.T) {

	// Compute: test_data/example.org.zone + creds.json ==> tempfile
	// Compare: tempfile to test_data/expected_zonefile.txt

	// Create temp file for output.
	outfile, err := ioutil.TempFile("", "example.*.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(outfile.Name())

	// Convert test data to the experiment output.
	gzargs := GetZoneArgs{
		CredName:     "bind",
		ProviderName: "BIND",
		ZoneNames:    []string{"example.org"},
		OutputFormat: "pretty",
		OutputFile:   outfile.Name(),
		//DefaultTTL:   0,
	}
	gzargs.CredsFile = "test_data/creds.json"
	err = GetZone(gzargs)
	if err != nil {
		log.Fatal(err)
	}

	want, err := ioutil.ReadFile("test_data/expected_zonefile.txt")
	if err != nil {
		log.Fatal(err)
	}

	got, err := ioutil.ReadFile(outfile.Name())
	if err != nil {
		log.Fatal(err)
	}

	//err = ioutil.WriteFile("test_data/expected_zonefile.txt", got, 0644)
	//if err != nil {
	//	log.Fatal(err)
	//}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestZoneToZone() mismatch (-want +got):\n%s", diff)
	}

}
