package js

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"unicode"

	"github.com/StackExchange/dnscontrol/models"
)

const (
	testDir  = "js/parse_tests"
	errorDir = "js/error_tests"
)

func init() {
	os.Chdir("..") // go up a directory so we helpers.js is in a consistent place.
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
		t.Log(f.Name(), "------")
		content, err := ioutil.ReadFile(filepath.Join(testDir, f.Name()))
		if err != nil {
			t.Fatal(err)
		}
		conf, err := ExecuteJavascript(string(content), true)
		if err != nil {
			t.Fatal(err)
		}
		actualJson, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		expectedFile := filepath.Join(testDir, f.Name()[:len(f.Name())-3]+".json")
		expectedData, err := ioutil.ReadFile(expectedFile)
		if err != nil {
			t.Fatal(err)
		}
		conf = &models.DNSConfig{}
		err = json.Unmarshal(expectedData, conf)
		if err != nil {
			t.Fatal(err)
		}
		expectedJson, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if string(expectedJson) != string(actualJson) {
			t.Error("Expected and actual json don't match")
			t.Log("Expected:", string(expectedJson))
			t.Log("Actual:", string(actualJson))
			t.FailNow()
		}
	}
}

func TestErrors(t *testing.T) {
	files, err := ioutil.ReadDir(errorDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".js" {
			continue
		}
		t.Log(f.Name(), "------")
		content, err := ioutil.ReadFile(filepath.Join(errorDir, f.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = ExecuteJavascript(string(content), true); err == nil {
			t.Fatal("Expected error but found none")
		}
	}
}
