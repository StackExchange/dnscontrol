package bind

import (
	"fmt"
	"testing"
	"time"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func Test_makeSoa(t *testing.T) {
	origin := "example.com"
	var tests = []struct {
		def            *SoaInfo
		existing       *models.RecordConfig
		desired        *models.RecordConfig
		expectedSoa    *models.RecordConfig
		expectedSerial uint32
	}{
		{
			// If everything is blank, the hard-coded defaults should kick in.
			&SoaInfo{"", "", 0, 0, 0, 0, 0},
			&models.RecordConfig{Target: "", SoaMbox: "", SoaSerial: 0, SoaRefresh: 0, SoaRetry: 0, SoaExpire: 0, SoaMinttl: 0},
			&models.RecordConfig{Target: "", SoaMbox: "", SoaSerial: 0, SoaRefresh: 0, SoaRetry: 0, SoaExpire: 0, SoaMinttl: 0},
			&models.RecordConfig{Target: "DEFAULT_NOT_SET.", SoaMbox: "DEFAULT_NOT_SET.", SoaSerial: 1, SoaRefresh: 3600, SoaRetry: 600, SoaExpire: 604800, SoaMinttl: 1440},
			2019022300,
		},
		{
			// If everything is filled, leave the desired values in place.
			&SoaInfo{"ns.example.com", "root.example.com", 1, 2, 3, 4, 5},
			&models.RecordConfig{Target: "a", SoaMbox: "aa", SoaSerial: 10, SoaRefresh: 11, SoaRetry: 12, SoaExpire: 13, SoaMinttl: 14},
			&models.RecordConfig{Target: "b", SoaMbox: "bb", SoaSerial: 15, SoaRefresh: 16, SoaRetry: 17, SoaExpire: 18, SoaMinttl: 19},
			&models.RecordConfig{Target: "b", SoaMbox: "bb", SoaSerial: 15, SoaRefresh: 16, SoaRetry: 17, SoaExpire: 18, SoaMinttl: 19},
			2019022300,
		},
		{
			// Test incrementing serial.
			&SoaInfo{"ns.example.com", "root.example.com", 1, 2, 3, 4, 5},
			&models.RecordConfig{Target: "a", SoaMbox: "aa", SoaSerial: 2019022301, SoaRefresh: 11, SoaRetry: 12, SoaExpire: 13, SoaMinttl: 14},
			&models.RecordConfig{Target: "b", SoaMbox: "bb", SoaSerial: 0, SoaRefresh: 16, SoaRetry: 17, SoaExpire: 18, SoaMinttl: 19},
			&models.RecordConfig{Target: "b", SoaMbox: "bb", SoaSerial: 2019022301, SoaRefresh: 16, SoaRetry: 17, SoaExpire: 18, SoaMinttl: 19},
			2019022302,
		},
		{
			// Test incrementing serial_2.
			&SoaInfo{"ns.example.com", "root.example.com", 1, 2, 3, 4, 5},
			&models.RecordConfig{Target: "a", SoaMbox: "aa", SoaSerial: 0, SoaRefresh: 11, SoaRetry: 12, SoaExpire: 13, SoaMinttl: 14},
			&models.RecordConfig{Target: "b", SoaMbox: "bb", SoaSerial: 2019022304, SoaRefresh: 16, SoaRetry: 17, SoaExpire: 18, SoaMinttl: 19},
			&models.RecordConfig{Target: "b", SoaMbox: "bb", SoaSerial: 2019022304, SoaRefresh: 16, SoaRetry: 17, SoaExpire: 18, SoaMinttl: 19},
			2019022305,
		},
		{
			// If there are gaps in existing or desired, fill in as appropriate.
			&SoaInfo{"ns.example.com", "root.example.com", 1, 2, 3, 4, 5},
			&models.RecordConfig{Target: "", SoaMbox: "aa", SoaSerial: 0, SoaRefresh: 11, SoaRetry: 0, SoaExpire: 13, SoaMinttl: 0},
			&models.RecordConfig{Target: "b", SoaMbox: "", SoaSerial: 15, SoaRefresh: 0, SoaRetry: 17, SoaExpire: 0, SoaMinttl: 19},
			&models.RecordConfig{Target: "b", SoaMbox: "aa", SoaSerial: 15, SoaRefresh: 11, SoaRetry: 17, SoaExpire: 13, SoaMinttl: 19},
			2019022300,
		},
		{
			// Gaps + existing==nil
			&SoaInfo{"ns.example.com", "root.example.com", 1, 2, 3, 4, 5},
			nil,
			&models.RecordConfig{Target: "b", SoaMbox: "", SoaSerial: 15, SoaRefresh: 0, SoaRetry: 17, SoaExpire: 0, SoaMinttl: 19},
			&models.RecordConfig{Target: "b", SoaMbox: "root.example.com", SoaSerial: 15, SoaRefresh: 2, SoaRetry: 17, SoaExpire: 4, SoaMinttl: 19},
			2019022300,
		},
		{
			// Gaps + desired==nil
			// NB(tom): In the code as of 2020-02-23, desired will never be nil.
			&SoaInfo{"ns.example.com", "root.example.com", 1, 2, 3, 4, 5},
			&models.RecordConfig{Target: "", SoaMbox: "aa", SoaSerial: 0, SoaRefresh: 11, SoaRetry: 0, SoaExpire: 13, SoaMinttl: 0},
			nil,
			&models.RecordConfig{Target: "ns.example.com", SoaMbox: "aa", SoaSerial: 1, SoaRefresh: 11, SoaRetry: 3, SoaExpire: 13, SoaMinttl: 5},
			2019022300,
		},
	}

	// Fake out the tests so they think today is 2019-02-23
	nowFunc = func() time.Time {
		fakeToday, _ := time.Parse("20060102", "20190223")
		return fakeToday
	}

	for i, tst := range tests {

		if tst.existing != nil {
			tst.existing.SetLabel("@", origin)
			tst.existing.Type = "SOA"
		}
		if tst.desired != nil {
			tst.desired.SetLabel("@", origin)
			tst.desired.Type = "SOA"
		}

		tst.expectedSoa.SetLabel("@", origin)
		tst.expectedSoa.Type = "SOA"

		r1, r2 := makeSoa(origin, tst.def, tst.existing, tst.desired)
		if !areEqualSoa(r1, tst.expectedSoa) {
			t.Fatalf("Test %d soa:\nExpected (%v)\n     got (%v)\n", i, tst.expectedSoa.String(), r1.String())
		}
		if r2 != tst.expectedSerial {
			t.Fatalf("Test:%d soa: Expected (%v) got (%v)\n", i, tst.expectedSerial, r2)
		}
	}
}

func areEqualSoa(r1, r2 *models.RecordConfig) bool {
	if r1.NameFQDN != r2.NameFQDN {
		fmt.Printf("ERROR: fqdn %q != %q\n", r1.NameFQDN, r2.NameFQDN)
		return false
	}
	if r1.Name != r2.Name {
		fmt.Printf("ERROR: name %q != %q\n", r1.Name, r2.Name)
		return false
	}
	if r1.Target != r2.Target {
		fmt.Printf("ERROR: target %q != %q\n", r1.Target, r2.Target)
		return false
	}
	if r1.SoaMbox != r2.SoaMbox {
		fmt.Printf("ERROR: mbox %v != %v\n", r1.SoaMbox, r2.SoaMbox)
		return false
	}
	if r1.SoaSerial != r2.SoaSerial {
		fmt.Printf("ERROR: serial %v != %v\n", r1.SoaSerial, r2.SoaSerial)
		return false
	}
	if r1.SoaRefresh != r2.SoaRefresh {
		fmt.Printf("ERROR: refresh %v != %v\n", r1.SoaRefresh, r2.SoaRefresh)
		return false
	}
	if r1.SoaRetry != r2.SoaRetry {
		fmt.Printf("ERROR: retry %v != %v\n", r1.SoaRetry, r2.SoaRetry)
		return false
	}
	if r1.SoaExpire != r2.SoaExpire {
		fmt.Printf("ERROR: expire %v != %v\n", r1.SoaExpire, r2.SoaExpire)
		return false
	}
	if r1.SoaMinttl != r2.SoaMinttl {
		fmt.Printf("ERROR: minttl %v != %v\n", r1.SoaMinttl, r2.SoaMinttl)
		return false
	}
	return true
}
