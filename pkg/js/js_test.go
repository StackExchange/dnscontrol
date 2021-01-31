package js

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"unicode"

	testifyrequire "github.com/stretchr/testify/require"

	"github.com/StackExchange/dnscontrol/v3/pkg/normalize"
	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
	"github.com/StackExchange/dnscontrol/v3/providers"
	_ "github.com/StackExchange/dnscontrol/v3/providers/_all"
)

const (
	testDir  = "pkg/js/parse_tests"
	errorDir = "pkg/js/error_tests"
)

func init() {
	os.Chdir("../..") // go up a directory so we helpers.js is in a consistent place.
}

func TestParsedFiles(t *testing.T) {
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		name := f.Name()

		// run all js files that start with a number. Skip others.
		if filepath.Ext(name) != ".js" || !unicode.IsNumber(rune(name[0])) {
			continue
		}
		t.Run(name, func(t *testing.T) {
			var err error

			// Compile the .js file:
			conf, err := ExecuteJavascript(string(filepath.Join(testDir, name)), true, nil)
			if err != nil {
				t.Fatal(err)
			}
			// for _, dc := range conf.Domains {
			// 	normalize.UpdateNameSplitHorizon(dc)
			// }

			// Initialize any DNS providers mentioned.
			for _, dProv := range conf.DNSProviders {
				var pcfg = map[string]string{}
				// Fake out any provider's validation tests.
				switch dProv.Type {
				case "CLOUDFLAREAPI":
					pcfg["apitoken"] = "fake"
				default:
				}
				_, err := providers.CreateDNSProvider(dProv.Type, pcfg, nil)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Test the JS compiled as expected (compare to the .json file)
			actualJSON, err := json.MarshalIndent(conf, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			testName := name[:len(name)-3]
			expectedFile := filepath.Join(testDir, testName+".json")
			expectedJSON, err := ioutil.ReadFile(expectedFile)
			if err != nil {
				t.Fatal(err)
			}
			es := string(expectedJSON)
			as := string(actualJSON)
			_, _ = es, as
			// When debugging, leave behind the actual result:
			//ioutil.WriteFile(expectedFile+".ACTUAL", []byte(es), 0644)
			testifyrequire.JSONEqf(t, es, as, "EXPECTING %q = \n```\n%s\n```", expectedFile, as)

			// For each domain, if there is a zone file, test against it:

			errs := normalize.ValidateAndNormalizeConfig(conf)
			if len(errs) != 0 {
				t.Fatal(errs[0])
			}

			var dCount int
			for _, dc := range conf.Domains {
				zoneFile := filepath.Join(testDir, testName, dc.Name+".zone")
				expectedZone, err := ioutil.ReadFile(zoneFile)
				if err != nil {
					continue
				}
				dCount++

				// Generate the zonefile
				var buf bytes.Buffer
				err = prettyzone.WriteZoneFileRC(&buf, dc.Records, dc.Name, 300, nil)
				if err != nil {
					t.Fatal(err)
				}
				actualZone := buf.String()

				es := string(expectedZone)
				as := actualZone
				if es != as {
					// On failure, leave behind the .ACTUAL file.
					ioutil.WriteFile(zoneFile+".ACTUAL", []byte(actualZone), 0644)
				}
				testifyrequire.Equal(t, es, as, "EXPECTING %q =\n```\n%s```", zoneFile, as)
			}
			if dCount > 0 && (len(conf.Domains) != dCount) {
				t.Fatal(fmt.Errorf("only %d of %d domains in %q have zonefiles", dCount, len(conf.Domains), name))
			}

		})

	}
}

func TestErrors(t *testing.T) {
	tests := []struct{ desc, text string }{
		{"old dsp style", `D("foo.com","reg","dsp")`},
		{"MX no priority", `D("foo.com","reg",MX("@","test."))`},
		{"MX reversed", `D("foo.com","reg",MX("@","test.", 5))`},
		{"CF_REDIRECT With comma", `D("foo.com","reg",CF_REDIRECT("foo.com,","baaa"))`},
		{"CF_TEMP_REDIRECT With comma", `D("foo.com","reg",CF_TEMP_REDIRECT("foo.com","baa,a"))`},
		{"Bad cidr", `D(reverse("foo.com"), "reg")`},
		{"Dup domains", `D("example.org", "reg"); D("example.org", "reg")`},
		{"Bad NAMESERVER", `D("example.com","reg", NAMESERVER("@","ns1.foo.com."))`},
	}
	for _, tst := range tests {
		t.Run(tst.desc, func(t *testing.T) {
			if _, err := ExecuteJavascript(tst.text, true, nil); err == nil {
				t.Fatal("Expected error but found none")
			}
		})

	}
}
