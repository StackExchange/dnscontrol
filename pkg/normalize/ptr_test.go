package normalize

import (
	"fmt"
	"testing"
)

// ptrmagic(name, domain string, al int) (string, error)

func TestPtrMagic(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		version int
		output  string
		fail    bool
	}{
		{"1", "2.3.4.in-addr.arpa", 4, "1", false},
	}
	for _, tst := range tests {
		t.Run(fmt.Sprintf("%s %s", tst.name, tst.domain), func(t *testing.T) {
			o, errs := ptrmagic(tst.name, tst.domain, tst.version)
			if errs != nil && !tst.fail {
				t.Error("Got error but expected none")
			}
			if errs == nil && tst.fail {
				t.Error("Expected error but got none")
			}
		})
	}
}
