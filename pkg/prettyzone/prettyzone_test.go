package prettyzone

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
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
	WriteZoneFileRR(buf2, parsed, "bosun.org")

	// Compare:
	if buf2.String() != expected {
		t.Fatalf("Regenerated zonefile does not match: got=(\n%v\n)\nexpected=(\n%v\n)\n", buf2.String(), expected)
	}
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
	x := models.RRstoRCs(records, "bosun.org")
	g = MostCommonTTL(x)
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

	// Mixture of TTLs with an obvious winner.
	records = nil
	records, e = append(records, r1, r2, r2), 200
	g = MostCommonTTL(models.RRstoRCs(records, "bosun.org"))
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

	// 3-way tie. Largest TTL should be used.
	records = nil
	records, e = append(records, r1, r2, r3), 300
	g = MostCommonTTL(models.RRstoRCs(records, "bosun.org"))
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

	// NS records are ignored.
	records = nil
	records, e = append(records, r1, r4, r5), 100
	g = MostCommonTTL(models.RRstoRCs(records, "bosun.org"))
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
	WriteZoneFileRR(buf, []dns.RR{r1, r2, r3}, "bosun.org")
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
	WriteZoneFileRR(buf, []dns.RR{r1, r2, r3, r4}, "bosun.org")
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
	WriteZoneFileRR(buf, []dns.RR{r1, r2, r3, r4, r5, r6, r7, r8, r9}, "bosun.org")
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
	WriteZoneFileRR(buf, []dns.RR{r1, r2, r3, r4, r5}, "bosun.org")
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
	WriteZoneFileRR(buf, []dns.RR{r1, r2, r3}, "bosun.org")
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
	WriteZoneFileRR(buf, []dns.RR{r1, r2, r3, r4, r5, r6}, "bosun.org")
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
	buf := &bytes.Buffer{}
	WriteZoneFileRR(buf, d, "bosun.org")
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
4.5              IN PTR   y.bosun.org.
_443._tcp        IN TLSA  3 1 1 abcdef0
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

	recs := models.RRstoRCs([]dns.RR{r1, r2, r3}, "bosun.org")
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
;myalias          IN R53_ALIAS  atype= zone_id=
;myalias          IN R53_ALIAS  atype= zone_id=
www              IN CNAME bosun.org.
;zalias           IN R53_ALIAS  atype= zone_id=
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
	WriteZoneFileRR(buf, records, "stackoverflow.com")
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
		WriteZoneFileRR(buf, records, "stackoverflow.com")
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

// func formatLine

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
		actual := formatLine(ts.lengths, ts.fields)
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
		actual := zoneLabelLess(test.e1, test.e2)
		if test.expected != actual {
			t.Errorf("%v: expected (%v) got (%v)\n", test.e1, test.e2, actual)
		}
		actual = zoneLabelLess(test.e2, test.e1)
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
