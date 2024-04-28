package prettyzone

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
)

func parseAndRegen(t *testing.T, buf *bytes.Buffer, expected string) {
	// Take a zonefile, parse it, then generate a zone. We should
	// get back the same string.
	// This is used after any WriteZoneFile test as an extra verification step.

	zp := dns.NewZoneParser(buf, "bosun.org", "bozun.org.zone")
	var parsed []dns.RR
	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		parsed = append(parsed, rr)
	}
	if err := zp.Err(); err != nil {
		log.Fatalf("Error in zonefile: %v", err)
	}

	// Generate it back:
	buf2 := &bytes.Buffer{}
	writeZoneFileRR(buf2, parsed, "bosun.org")

	// Compare:
	if buf2.String() != expected {
		t.Fatalf("Regenerated zonefile does not match: got=(\n%v\n)\nexpected=(\n%v\n)\n", buf2.String(), expected)
	}
}

// rrstoRCs converts []dns.RR to []RecordConfigs.
func rrstoRCs(rrs []dns.RR, origin string) (models.Records, error) {
	rcs := make(models.Records, 0, len(rrs))
	for _, r := range rrs {
		rc, err := models.RRtoRC(r, origin)
		if err != nil {
			return nil, err
		}

		rcs = append(rcs, &rc)
	}
	return rcs, nil
}

// writeZoneFileRR is a helper for when you have []dns.RR instead of models.Records
func writeZoneFileRR(w io.Writer, records []dns.RR, origin string) error {
	rcs, err := rrstoRCs(records, origin)
	if err != nil {
		return err
	}

	return WriteZoneFileRC(w, rcs, origin, 0, nil)
}
func TestMostCommonTtl(t *testing.T) {
	var records []dns.RR
	var g, e uint32
	r1, _ := dns.NewRR("bosun.org. 100 IN A 1.1.1.1")
	r2, _ := dns.NewRR("bosun.org. 200 IN A 1.1.1.1")
	r3, _ := dns.NewRR("bosun.org. 300 IN A 1.1.1.1")
	r4, _ := dns.NewRR("bosun.org. 400 IN NS foo.bosun.org.")
	r5, _ := dns.NewRR("bosun.org. 400 IN NS bar.bosun.org.")

	// All records are TTL=100
	records = nil
	records, e = append(records, r1, r1, r1), 100
	x, err := rrstoRCs(records, "bosun.org")
	if err != nil {
		panic(err)
	}
	g = MostCommonTTL(x)
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

	// Mixture of TTLs with an obvious winner.
	records = nil
	records, e = append(records, r1, r2, r2), 200
	rcs, err := rrstoRCs(records, "bosun.org")
	if err != nil {
		panic(err)
	}
	g = MostCommonTTL(rcs)
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

	// 3-way tie. Largest TTL should be used.
	records = nil
	records, e = append(records, r1, r2, r3), 300
	rcs, err = rrstoRCs(records, "bosun.org")
	if err != nil {
		panic(err)
	}
	g = MostCommonTTL(rcs)
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

	// NS records are ignored.
	records = nil
	records, e = append(records, r1, r4, r5), 100
	rcs, err = rrstoRCs(records, "bosun.org")
	if err != nil {
		panic(err)
	}
	g = MostCommonTTL(rcs)
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

}

// func WriteZoneFile

func TestWriteZoneFileSimple(t *testing.T) {
	r1, _ := dns.NewRR("bosun.org. 300 IN A 192.30.252.153")
	r2, _ := dns.NewRR("bosun.org. 300 IN A 192.30.252.154")
	r3, _ := dns.NewRR("www.bosun.org. 300 IN CNAME bosun.org.")
	buf := &bytes.Buffer{}
	writeZoneFileRR(buf, []dns.RR{r1, r2, r3}, "bosun.org")
	expected := `$TTL 300
@                IN A     192.30.252.153
                 IN A     192.30.252.154
www              IN CNAME bosun.org.
`
	if buf.String() != expected {
		t.Log(buf.String())
		t.Log(expected)
		t.Fatalf("Zone file does not match.")
	}

	parseAndRegen(t, buf, expected)
}

func TestWriteZoneFileSimpleTtl(t *testing.T) {
	r1, _ := dns.NewRR("bosun.org. 100 IN A 192.30.252.153")
	r2, _ := dns.NewRR("bosun.org. 100 IN A 192.30.252.154")
	r3, _ := dns.NewRR("bosun.org. 100 IN A 192.30.252.155")
	r4, _ := dns.NewRR("www.bosun.org. 300 IN CNAME bosun.org.")
	buf := &bytes.Buffer{}
	writeZoneFileRR(buf, []dns.RR{r1, r2, r3, r4}, "bosun.org")
	expected := `$TTL 100
@                IN A     192.30.252.153
                 IN A     192.30.252.154
                 IN A     192.30.252.155
www        300   IN CNAME bosun.org.
`
	if buf.String() != expected {
		t.Log(buf.String())
		t.Log(expected)
		t.Fatalf("Zone file does not match")
	}

	parseAndRegen(t, buf, expected)
}

func TestWriteZoneFileMx(t *testing.T) {
	// sort by priority
	r1, _ := dns.NewRR("aaa.bosun.org. IN MX 1 aaa.example.com.")
	r2, _ := dns.NewRR("aaa.bosun.org. IN MX 5 aaa.example.com.")
	r3, _ := dns.NewRR("aaa.bosun.org. IN MX 10 aaa.example.com.")
	// same priority? sort by name
	r4, _ := dns.NewRR("bbb.bosun.org. IN MX 10 ccc.example.com.")
	r5, _ := dns.NewRR("bbb.bosun.org. IN MX 10 bbb.example.com.")
	r6, _ := dns.NewRR("bbb.bosun.org. IN MX 10 aaa.example.com.")
	// a mix
	r7, _ := dns.NewRR("ccc.bosun.org. IN MX 40 zzz.example.com.")
	r8, _ := dns.NewRR("ccc.bosun.org. IN MX 40 aaa.example.com.")
	r9, _ := dns.NewRR("ccc.bosun.org. IN MX 1 ttt.example.com.")
	buf := &bytes.Buffer{}
	writeZoneFileRR(buf, []dns.RR{r1, r2, r3, r4, r5, r6, r7, r8, r9}, "bosun.org")
	if buf.String() != testdataZFMX {
		t.Log(buf.String())
		t.Log(testdataZFMX)
		t.Fatalf("Zone file does not match.")
	}
	parseAndRegen(t, buf, testdataZFMX)
}

var testdataZFMX = `$TTL 3600
aaa              IN MX    1 aaa.example.com.
                 IN MX    5 aaa.example.com.
                 IN MX    10 aaa.example.com.
bbb              IN MX    10 aaa.example.com.
                 IN MX    10 bbb.example.com.
                 IN MX    10 ccc.example.com.
ccc              IN MX    1 ttt.example.com.
                 IN MX    40 aaa.example.com.
                 IN MX    40 zzz.example.com.
`

func TestWriteZoneFileSrv(t *testing.T) {
	// exhibits explicit ttls and long name
	r1, _ := dns.NewRR(`bosun.org. 300 IN SRV 10 10 9999 foo.com.`)
	r2, _ := dns.NewRR(`bosun.org. 300 IN SRV 10 20 5050 foo.com.`)
	r3, _ := dns.NewRR(`bosun.org. 300 IN SRV 10 10 5050 foo.com.`)
	r4, _ := dns.NewRR(`bosun.org. 300 IN SRV 20 10 5050 foo.com.`)
	r5, _ := dns.NewRR(`bosun.org. 300 IN SRV 10 10 5050 foo.com.`)
	buf := &bytes.Buffer{}
	writeZoneFileRR(buf, []dns.RR{r1, r2, r3, r4, r5}, "bosun.org")
	if buf.String() != testdataZFSRV {
		t.Log(buf.String())
		t.Log(testdataZFSRV)
		t.Fatalf("Zone file does not match.")
	}
	parseAndRegen(t, buf, testdataZFSRV)
}

var testdataZFSRV = `$TTL 300
@                IN SRV   10 10 5050 foo.com.
                 IN SRV   10 10 5050 foo.com.
                 IN SRV   10 20 5050 foo.com.
                 IN SRV   20 10 5050 foo.com.
                 IN SRV   10 10 9999 foo.com.
`

func TestWriteZoneFilePtr(t *testing.T) {
	// exhibits explicit ttls and long name
	r1, _ := dns.NewRR(`bosun.org. 300 IN PTR chell.bosun.org`)
	r2, _ := dns.NewRR(`bosun.org. 300 IN PTR barney.bosun.org.`)
	r3, _ := dns.NewRR(`bosun.org. 300 IN PTR alex.bosun.org.`)
	buf := &bytes.Buffer{}
	writeZoneFileRR(buf, []dns.RR{r1, r2, r3}, "bosun.org")
	if buf.String() != testdataZFPTR {
		t.Log(buf.String())
		t.Log(testdataZFPTR)
		t.Fatalf("Zone file does not match.")
	}
	parseAndRegen(t, buf, testdataZFPTR)
}

var testdataZFPTR = `$TTL 300
@                IN PTR   alex.bosun.org.
                 IN PTR   barney.bosun.org.
                 IN PTR   chell.bosun.org.
`

func TestWriteZoneFileCaa(t *testing.T) {
	// exhibits explicit ttls and long name
	r1, _ := dns.NewRR(`bosun.org. 300 IN CAA 0 issuewild ";"`)
	r2, _ := dns.NewRR(`bosun.org. 300 IN CAA 0 issue "letsencrypt.org"`)
	r3, _ := dns.NewRR(`bosun.org. 300 IN CAA 1 iodef "http://example.com"`)
	r4, _ := dns.NewRR(`bosun.org. 300 IN CAA 0 iodef "https://example.com"`)
	r5, _ := dns.NewRR(`bosun.org. 300 IN CAA 0 iodef "https://example.net"`)
	r6, _ := dns.NewRR(`bosun.org. 300 IN CAA 1 iodef "mailto:example.com"`)
	buf := &bytes.Buffer{}
	writeZoneFileRR(buf, []dns.RR{r1, r2, r3, r4, r5, r6}, "bosun.org")
	if buf.String() != testdataZFCAA {
		t.Log(buf.String())
		t.Log(testdataZFCAA)
		t.Fatalf("Zone file does not match.")
	}
	parseAndRegen(t, buf, testdataZFCAA)
}

var testdataZFCAA = `$TTL 300
@                IN CAA   1 iodef "http://example.com"
                 IN CAA   1 iodef "mailto:example.com"
                 IN CAA   0 iodef "https://example.com"
                 IN CAA   0 iodef "https://example.net"
                 IN CAA   0 issue "letsencrypt.org"
                 IN CAA   0 issuewild ";"
`

// r is shorthand for strings.Repeat()
func r(s string, c int) string { return strings.Repeat(s, c) }

func TestWriteZoneFileTxt(t *testing.T) {
	// Do round-trip tests on various length TXT records.
	t10 := `t10              IN TXT   "ten4567890"`
	t254 := `t254             IN TXT   "` + r("a", 254) + `"`
	t255 := `t255             IN TXT   "` + r("b", 255) + `"`
	t256 := `t256             IN TXT   "` + r("c", 255) + `" "` + r("D", 1) + `"`
	t509 := `t509             IN TXT   "` + r("e", 255) + `" "` + r("F", 254) + `"`
	t510 := `t510             IN TXT   "` + r("g", 255) + `" "` + r("H", 255) + `"`
	t511 := `t511             IN TXT   "` + r("i", 255) + `" "` + r("J", 255) + `" "` + r("k", 1) + `"`
	t512 := `t511             IN TXT   "` + r("L", 255) + `" "` + r("M", 255) + `" "` + r("n", 2) + `"`
	t513 := `t511             IN TXT   "` + r("o", 255) + `" "` + r("P", 255) + `" "` + r("q", 3) + `"`
	for i, d := range []string{t10, t254, t255, t256, t509, t510, t511, t512, t513} {
		// Make the rr:
		rr, err := dns.NewRR(d)
		if err != nil {
			t.Fatal(err)
		}

		// Make the expected zonefile:
		ez := "$TTL 3600\n" + d + "\n"

		// Generate the zonefile:
		buf := &bytes.Buffer{}
		writeZoneFileRR(buf, []dns.RR{rr}, "bosun.org")
		gz := buf.String()
		if gz != ez {
			t.Log("got: " + gz)
			t.Log("wnt: " + ez)
			t.Fatalf("Zone file %d does not match.", i)
		}

		// Reverse the process. Turn the zonefile into a list of records
		parseAndRegen(t, buf, ez)
	}

}

// Test 1 of each record type

func mustNewRR(s string) dns.RR {
	r, err := dns.NewRR(s)
	if err != nil {
		panic(err)
	}
	return r
}

func TestWriteZoneFileEach(t *testing.T) {
	// Each rtype should be listed in this test exactly once.
	// If an rtype has more than one variations, add a test like TestWriteZoneFileCaa to test each.
	var d []dns.RR
	// #rtype_variations
	d = append(d, mustNewRR(`4.5                  300 IN PTR   y.bosun.org.`)) // Wouldn't actually be in this domain.
	d = append(d, mustNewRR(`bosun.org.           300 IN A     1.2.3.4`))
	d = append(d, mustNewRR(`bosun.org.           300 IN MX    1 bosun.org.`))
	d = append(d, mustNewRR(`bosun.org.           300 IN TXT   "my text"`))
	d = append(d, mustNewRR(`bosun.org.           300 IN AAAA  4500:fe::1`))
	d = append(d, mustNewRR(`bosun.org.           300 IN SRV   10 10 9999 foo.com.`))
	d = append(d, mustNewRR(`bosun.org.           300 IN CAA   0 issue "letsencrypt.org"`))
	d = append(d, mustNewRR(`_443._tcp.bosun.org. 300 IN TLSA  3 1 1 abcdef0`)) // Label must be _port._proto
	d = append(d, mustNewRR(`sub.bosun.org.       300 IN NS    bosun.org.`))    // Must be a label with no other records.
	d = append(d, mustNewRR(`x.bosun.org.         300 IN CNAME bosun.org.`))    // Must be a label with no other records.
	d = append(d, mustNewRR(`bosun.org.           300 IN DHCID   AAIBY2/AuCccgoJbsaxcQc9TUapptP69lOjxfNuVAA2kjEA=`))
	d = append(d, mustNewRR(`dname.bosun.org.     300 IN DNAME   example.com.`))
	d = append(d, mustNewRR(`dnssec.bosun.org.    300 IN DS      31334 13 2 94cc505ebc36b1f4e051268b820efb230f1572d445e833bb5bf7380d6c2cbc0a`))
	d = append(d, mustNewRR(`dnssec.bosun.org.    300 IN DNSKEY  257 3 13 rNR701yiOPHfqDP53GnsHZdlsRqI7O1ksk60rnFILZVk7Z4eTBd1U49oSkTNVNox9tb7N15N2hboXoMEyFFzcw==`))
	d = append(d, mustNewRR(`bosun.org.           300 IN HTTPS 1 . alpn="h3,h2"`))
	d = append(d, mustNewRR(`bosun.org.           300 IN SVCB 1 . alpn="h3,h2"`))
	buf := &bytes.Buffer{}
	writeZoneFileRR(buf, d, "bosun.org")
	if buf.String() != testdataZFEach {
		t.Log(buf.String())
		t.Log(testdataZFEach)
		t.Fatalf("Zone file does not match.")
	}
	parseAndRegen(t, buf, testdataZFEach)
}

var testdataZFEach = `$TTL 300
@                IN A     1.2.3.4
                 IN AAAA  4500:fe::1
                 IN MX    1 bosun.org.
                 IN SRV   10 10 9999 foo.com.
                 IN TXT   "my text"
                 IN CAA   0 issue "letsencrypt.org"
                 IN DHCID AAIBY2/AuCccgoJbsaxcQc9TUapptP69lOjxfNuVAA2kjEA=
                 IN HTTPS 1 . alpn="h3,h2"
                 IN SVCB  1 . alpn="h3,h2"
4.5              IN PTR   y.bosun.org.
_443._tcp        IN TLSA  3 1 1 abcdef0
dname            IN DNAME example.com.
dnssec           IN DNSKEY 257 3 13 rNR701yiOPHfqDP53GnsHZdlsRqI7O1ksk60rnFILZVk7Z4eTBd1U49oSkTNVNox9tb7N15N2hboXoMEyFFzcw==
                 IN DS    31334 13 2 94CC505EBC36B1F4E051268B820EFB230F1572D445E833BB5BF7380D6C2CBC0A
sub              IN NS    bosun.org.
x                IN CNAME bosun.org.
`

func TestWriteZoneFileSynth(t *testing.T) {
	r1, _ := dns.NewRR("bosun.org. 300 IN A 192.30.252.153")
	r2, _ := dns.NewRR("bosun.org. 300 IN A 192.30.252.154")
	r3, _ := dns.NewRR("www.bosun.org. 300 IN CNAME bosun.org.")
	rsynm := &models.RecordConfig{Type: "R53_ALIAS", TTL: 300}
	rsynm.SetLabel("myalias", "bosun.org")
	rsynz := &models.RecordConfig{Type: "R53_ALIAS", TTL: 300}
	rsynz.SetLabel("zalias", "bosun.org")

	recs, err := rrstoRCs([]dns.RR{r1, r2, r3}, "bosun.org")
	if err != nil {
		panic(err)
	}
	recs = append(recs, rsynm)
	recs = append(recs, rsynm)
	recs = append(recs, rsynz)

	buf := &bytes.Buffer{}
	WriteZoneFileRC(buf, recs, "bosun.org", 0, []string{"c1", "c2", "c3\nc4"})
	expected := `$TTL 300
; c1
; c2
; c3
; c4
@                IN A     192.30.252.153
                 IN A     192.30.252.154
;myalias          IN R53_ALIAS  atype= zone_id= evaluate_target_health=
;myalias          IN R53_ALIAS  atype= zone_id= evaluate_target_health=
www              IN CNAME bosun.org.
;zalias           IN R53_ALIAS  atype= zone_id= evaluate_target_health=
`
	if buf.String() != expected {
		t.Log(buf.String())
		t.Log(expected)
		t.Fatalf("Zone file does not match.")
	}
}

// Test sorting

func TestWriteZoneFileOrder(t *testing.T) {
	var records []dns.RR
	for i, td := range []string{
		"@",
		"@",
		"@",
		"stackoverflow.com.",
		"*",
		"foo",
		"bar.foo",
		"hip.foo",
		"mup",
		"a.mup",
		"bzt.mup",
		"aaa.bzt.mup",
		"zzz.bzt.mup",
		"nnn.mup",
		"zt.mup",
		"zap",
	} {
		name := dnsutil.AddOrigin(td, "stackoverflow.com.")
		r, _ := dns.NewRR(fmt.Sprintf("%s 300 IN A 1.2.3.%d", name, i))
		records = append(records, r)
	}

	buf := &bytes.Buffer{}
	writeZoneFileRR(buf, records, "stackoverflow.com")
	// Compare
	if buf.String() != testdataOrder {
		t.Log("Found:")
		t.Log(buf.String())
		t.Log("Expected:")
		t.Log(testdataOrder)
		t.Fatalf("Zone file does not match.")
	}
	parseAndRegen(t, buf, testdataOrder)

	// Now shuffle the list many times and make sure it still works:
	for iteration := 5; iteration > 0; iteration-- {
		// Randomize the list:
		perm := rand.Perm(len(records))
		for i, v := range perm {
			records[i], records[v] = records[v], records[i]
		}
		// Generate
		buf := &bytes.Buffer{}
		writeZoneFileRR(buf, records, "stackoverflow.com")
		// Compare
		if buf.String() != testdataOrder {
			t.Log(buf.String())
			t.Log(testdataOrder)
			t.Fatalf("Zone file does not match.")
		}
		parseAndRegen(t, buf, testdataOrder)
	}
}

var testdataOrder = `$TTL 300
@                IN A     1.2.3.0
                 IN A     1.2.3.1
                 IN A     1.2.3.2
                 IN A     1.2.3.3
*                IN A     1.2.3.4
foo              IN A     1.2.3.5
bar.foo          IN A     1.2.3.6
hip.foo          IN A     1.2.3.7
mup              IN A     1.2.3.8
a.mup            IN A     1.2.3.9
bzt.mup          IN A     1.2.3.10
aaa.bzt.mup      IN A     1.2.3.11
zzz.bzt.mup      IN A     1.2.3.12
nnn.mup          IN A     1.2.3.13
zt.mup           IN A     1.2.3.14
zap              IN A     1.2.3.15
`

// func FormatLine

func TestFormatLine(t *testing.T) {
	tests := []struct {
		lengths  []int
		fields   []string
		expected string
	}{
		{[]int{2, 2, 0}, []string{"a", "b", "c"}, "a  b  c"},
		{[]int{2, 2, 0}, []string{"aaaaa", "b", "c"}, "aaaaa b c"},
	}
	for _, ts := range tests {
		actual := FormatLine(ts.lengths, ts.fields)
		if actual != ts.expected {
			t.Errorf("\"%s\" != \"%s\"", actual, ts.expected)
		}
	}
}

// func zoneLabelLess

func TestZoneLabelLess(t *testing.T) {
	/*
			The zone should sort in prefix traversal order:

		  @
		  *
		  foo
		  bar.foo
		  hip.foo
		  mup
		  a.mup
		  bzt.mup
		  *.bzt.mup
		  1.bzt.mup
		  2.bzt.mup
		  10.bzt.mup
		  aaa.bzt.mup
		  zzz.bzt.mup
		  nnn.mup
		  zt.mup
		  zap
	*/

	var tests = []struct {
		e1, e2   string
		expected bool
	}{
		{"@", "@", false},
		{"@", "*", true},
		{"@", "b", true},
		{"*", "@", false},
		{"*", "*", false},
		{"*", "b", true},
		{"foo", "foo", false},
		{"foo", "bar", false},
		{"bar", "foo", true},
		{"a.mup", "mup", false},
		{"mup", "a.mup", true},
		{"a.mup", "a.mup", false},
		{"a.mup", "bzt.mup", true},
		{"a.mup", "aa.mup", true},
		{"zt.mup", "aaa.bzt.mup", false},
		{"aaa.bzt.mup", "mup", false},
		{"*.bzt.mup", "aaa.bzt.mup", true},
		{"1.bzt.mup", "aaa.bzt.mup", true},
		{"1.bzt.mup", "2.bzt.mup", true},
		{"10.bzt.mup", "2.bzt.mup", false},
		{"nnn.mup", "aaa.bzt.mup", false},
		{`www\.miek.nl`, `www.miek.nl`, false},
	}

	for _, test := range tests {
		actual := LabelLess(test.e1, test.e2)
		if test.expected != actual {
			t.Errorf("%v: expected (%v) got (%v)\n", test.e1, test.e2, actual)
		}
		actual = LabelLess(test.e2, test.e1)
		// The reverse should work too:
		var expected bool
		if test.e1 == test.e2 {
			expected = false
		} else {
			expected = !test.expected
		}
		if expected != actual {
			t.Errorf("%v: expected (%v) got (%v)\n", test.e1, test.e2, actual)
		}
	}
}

func TestZoneRrtypeLess(t *testing.T) {
	/*
		In zonefiles we want to list SOAs, then NSs, then all others.
	*/

	var tests = []struct {
		e1, e2   string
		expected bool
	}{
		{"SOA", "SOA", false},
		{"SOA", "A", true},
		{"SOA", "TXT", true},
		{"SOA", "NS", true},
		{"NS", "SOA", false},
		{"NS", "A", true},
		{"NS", "TXT", true},
		{"NS", "NS", false},
		{"A", "SOA", false},
		{"A", "A", false},
		{"A", "TXT", true},
		{"A", "NS", false},
		{"MX", "SOA", false},
		{"MX", "A", false},
		{"MX", "TXT", true},
		{"MX", "NS", false},
	}

	for _, test := range tests {
		actual := zoneRrtypeLess(test.e1, test.e2)
		if test.expected != actual {
			t.Errorf("%v: expected (%v) got (%v)\n", test.e1, test.e2, actual)
		}
		actual = zoneRrtypeLess(test.e2, test.e1)
		// The reverse should work too:
		var expected bool
		if test.e1 == test.e2 {
			expected = false
		} else {
			expected = !test.expected
		}
		if expected != actual {
			t.Errorf("%v: expected (%v) got (%v)\n", test.e1, test.e2, actual)
		}
	}
}
