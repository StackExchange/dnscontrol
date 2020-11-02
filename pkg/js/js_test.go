package js

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"unicode"

	testifyrequire "github.com/stretchr/testify/require"
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
		// run all js files that start with a number. Skip others.
		if filepath.Ext(f.Name()) != ".js" || !unicode.IsNumber(rune(f.Name()[0])) {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			conf, err := ExecuteJavascript(string(filepath.Join(testDir, f.Name())), true, nil)
			if err != nil {
				t.Fatal(err)
			}
			actualJSON, err := json.MarshalIndent(conf, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			expectedFile := filepath.Join(testDir, f.Name()[:len(f.Name())-3]+".json")
			expectedJSON, err := ioutil.ReadFile(expectedFile)
			if err != nil {
				t.Fatal(err)
			}
			es := string(expectedJSON)
			as := string(actualJSON)
			testifyrequire.JSONEqf(t, es, as, "EXPECTING %s", as)
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
