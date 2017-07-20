package js

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"unicode"

	"github.com/tdewolff/minify"
	minjson "github.com/tdewolff/minify/json"
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
		//run all js files that start with a number. Skip others.
		if filepath.Ext(f.Name()) != ".js" || !unicode.IsNumber(rune(f.Name()[0])) {
			continue
		}
		m := minify.New()
		m.AddFunc("json", minjson.Minify)
		t.Run(f.Name(), func(t *testing.T) {
			content, err := ioutil.ReadFile(filepath.Join(testDir, f.Name()))
			if err != nil {
				t.Fatal(err)
			}
			conf, err := ExecuteJavascript(string(content), true)
			if err != nil {
				t.Fatal(err)
			}
			actualJSON, err := json.MarshalIndent(conf, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			actualJSON, err = m.Bytes("json", actualJSON)
			if err != nil {
				t.Fatal(err)
			}
			expectedFile := filepath.Join(testDir, f.Name()[:len(f.Name())-3]+".json")
			expectedData, err := ioutil.ReadFile(expectedFile)
			if err != nil {
				t.Fatal(err)
			}

			expectedJSON, err := m.String("json", string(expectedData))
			if err != nil {
				t.Fatal(err)
			}
			if string(expectedJSON) != string(actualJSON) {
				t.Error("Expected and actual json don't match")
				t.Log("Expected:", string(expectedJSON))
				t.Log("Actual  :", string(actualJSON))
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
	}
	for _, tst := range tests {
		t.Run(tst.desc, func(t *testing.T) {
			if _, err := ExecuteJavascript(tst.text, true); err == nil {
				t.Fatal("Expected error but found none")
			}
		})

	}
}
