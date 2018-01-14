package utfutil_test

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"github.com/TomOnTime/utfutil"
)

func TestReadFile(t *testing.T) {

	expected, err := ioutil.ReadFile(filepath.Join("testdata", "calblur8.htm"))
	if err != nil {
		log.Fatal(err)
	}

	// The test files were generated with:
	//    for i in $(iconv  -l|grep UTF)  ; do
	//        iconv -f UTF-8 -t $i calblur8.htm > calblur8.htm.$i
	//    done
	for _, tst := range []struct {
		works bool // is combination is expected to work?
		ume   utfutil.EncodingHint
		name  string
	}{
		// Assume missing BOM means UTF8
		{true, utfutil.UTF8, "calblur8.htm.UTF-8"},     // No BOM
		{true, utfutil.UTF8, "calblur8.htm.UTF-16"},    // BOM=fffe
		{false, utfutil.UTF8, "calblur8.htm.UTF-16LE"}, // no BOM
		{false, utfutil.UTF8, "calblur8.htm.UTF-16BE"}, // no BOM
		// Assume missing BOM means UFT16LE
		{false, utfutil.UTF16LE, "calblur8.htm.UTF-8"},    // No BOM
		{true, utfutil.UTF16LE, "calblur8.htm.UTF-16"},    // BOM=fffe
		{true, utfutil.UTF16LE, "calblur8.htm.UTF-16LE"},  // no BOM
		{false, utfutil.UTF16LE, "calblur8.htm.UTF-16BE"}, // no BOM
		// Assume missing BOM means UFT16BE
		{false, utfutil.UTF16BE, "calblur8.htm.UTF-8"},    // No BOM
		{true, utfutil.UTF16BE, "calblur8.htm.UTF-16"},    // BOM=fffe
		{false, utfutil.UTF16BE, "calblur8.htm.UTF-16LE"}, // no BOM
		{true, utfutil.UTF16BE, "calblur8.htm.UTF-16BE"},  // no BOM
	} {

		actual, err := utfutil.ReadFile(filepath.Join("testdata", tst.name), tst.ume)
		if err != nil {
			log.Fatal(err)
		}

		if tst.works {
			if string(expected) == string(actual) {
				t.Log("SUCCESS:", tst.ume, tst.name)
			} else {
				t.Errorf("FAIL: %v/%v: expected %#v got %#v\n", tst.ume, tst.name, string(expected)[:4], actual[:4])
			}
		} else {
			if string(expected) != string(actual) {
				t.Logf("SUCCESS: %v/%v: failed as expected.", tst.ume, tst.name)
			} else {
				t.Errorf("FAILUREish: %v/%v: unexpected success!", tst.ume, tst.name)
			}
		}
	}

}
