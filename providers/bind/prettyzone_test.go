package bind

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
)

func parseAndRegen(t *testing.T, buf *bytes.Buffer, expected string) {
	// Take a zonefile, parse it, then generate a zone. We should
	// get back the same string.
	// This is used after any WriteZoneFile test as an extra verification step.

	// Parse the output:
	var parsed []dns.RR
	for x := range dns.ParseZone(buf, "bosun.org", "bosun.org.zone") {
		if x.Error != nil {
			log.Fatalf("Error in zonefile: %v", x.Error)
		} else {
			parsed = append(parsed, x.RR)
		}
	}
	// Generate it back:
	buf2 := &bytes.Buffer{}
	WriteZoneFile(buf2, parsed, "bosun.org.")

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
	g = mostCommonTtl(records)
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

	// Mixture of TTLs with an obvious winner.
	records = nil
	records, e = append(records, r1, r2, r2), 200
	g = mostCommonTtl(records)
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

	// 3-way tie. Largest TTL should be used.
	records = nil
	records, e = append(records, r1, r2, r3), 300
	g = mostCommonTtl(records)
	if e != g {
		t.Fatalf("expected %d; got %d\n", e, g)
	}

	// NS records are ignored.
	records = nil
	records, e = append(records, r1, r4, r5), 100
	g = mostCommonTtl(records)
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
	WriteZoneFile(buf, []dns.RR{r1, r2, r3}, "bosun.org.")
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
	WriteZoneFile(buf, []dns.RR{r1, r2, r3, r4}, "bosun.org.")
	expected := `$TTL 100
@                IN A     192.30.252.153
                 IN A     192.30.252.154
                 IN A     192.30.252.155
www        300   IN CNAME bosun.org.
`
	if buf.String() != expected {
		t.Log(buf.String())
		t.Log(expected)
		t.Fatalf("Zone file does not match.")
	}

	parseAndRegen(t, buf, expected)
}

func TestWriteZoneFileMx(t *testing.T) {
	//exhibits explicit ttls and long name
	r1, _ := dns.NewRR(`bosun.org. 300 IN TXT "aaa"`)
	r2, _ := dns.NewRR(`bosun.org. 300 IN TXT "bbb"`)
	r2.(*dns.TXT).Txt[0] = `b"bb`
	r3, _ := dns.NewRR("bosun.org. 300 IN MX 1 ASPMX.L.GOOGLE.COM.")
	r4, _ := dns.NewRR("bosun.org. 300 IN MX 5 ALT1.ASPMX.L.GOOGLE.COM.")
	r5, _ := dns.NewRR("bosun.org. 300 IN MX 10 ASPMX3.GOOGLEMAIL.COM.")
	r6, _ := dns.NewRR("bosun.org. 300 IN A 198.252.206.16")
	r7, _ := dns.NewRR("*.bosun.org. 600 IN A 198.252.206.16")
	r8, _ := dns.NewRR(`_domainkey.bosun.org. 300 IN TXT "vvvv"`)
	r9, _ := dns.NewRR(`google._domainkey.bosun.org. 300 IN TXT "\"foo\""`)
	buf := &bytes.Buffer{}
	WriteZoneFile(buf, []dns.RR{r1, r2, r3, r4, r5, r6, r7, r8, r9}, "bosun.org")
	if buf.String() != testdataZFMX {
		t.Log(buf.String())
		t.Log(testdataZFMX)
		t.Fatalf("Zone file does not match.")
	}
	parseAndRegen(t, buf, testdataZFMX)
}

var testdataZFMX = `$TTL 300
@                IN A     198.252.206.16
                 IN MX    1 ASPMX.L.GOOGLE.COM.
                 IN MX    5 ALT1.ASPMX.L.GOOGLE.COM.
                 IN MX    10 ASPMX3.GOOGLEMAIL.COM.
                 IN TXT   "aaa"
                 IN TXT   "b\"bb"
*          600   IN A     198.252.206.16
_domainkey       IN TXT   "vvvv"
google._domainkey IN TXT  "\"foo\""
`

func TestWriteZoneFileSrv(t *testing.T) {
	//exhibits explicit ttls and long name
	r1, _ := dns.NewRR(`bosun.org. 300 IN SRV 10 10 9999 foo.com.`)
	r2, _ := dns.NewRR(`bosun.org. 300 IN SRV 10 20 5050 foo.com.`)
	r3, _ := dns.NewRR(`bosun.org. 300 IN SRV 10 10 5050 foo.com.`)
	r4, _ := dns.NewRR(`bosun.org. 300 IN SRV 20 10 5050 foo.com.`)
	r5, _ := dns.NewRR(`bosun.org. 300 IN SRV 10 10 5050 foo.com.`)
	buf := &bytes.Buffer{}
	WriteZoneFile(buf, []dns.RR{r1, r2, r3, r4, r5}, "bosun.org")
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

func TestWriteZoneFileCaa(t *testing.T) {
	//exhibits explicit ttls and long name
	r1, _ := dns.NewRR(`bosun.org. 300 IN CAA 0 issuewild ";"`)
	r2, _ := dns.NewRR(`bosun.org. 300 IN CAA 0 issue "letsencrypt.org"`)
	r3, _ := dns.NewRR(`bosun.org. 300 IN CAA 1 iodef "http://example.com"`)
	r4, _ := dns.NewRR(`bosun.org. 300 IN CAA 0 iodef "https://example.com"`)
	r5, _ := dns.NewRR(`bosun.org. 300 IN CAA 0 iodef "https://example.net"`)
	r6, _ := dns.NewRR(`bosun.org. 300 IN CAA 1 iodef "mailto:example.com"`)
	buf := &bytes.Buffer{}
	WriteZoneFile(buf, []dns.RR{r1, r2, r3, r4, r5, r6}, "bosun.org")
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
	WriteZoneFile(buf, records, "stackoverflow.com.")
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
			//fmt.Println(i, v)
		}
		// Generate
		buf := &bytes.Buffer{}
		WriteZoneFile(buf, records, "stackoverflow.com.")
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
		e1, e2   uint16
		expected bool
	}{
		{dns.TypeSOA, dns.TypeSOA, false},
		{dns.TypeSOA, dns.TypeA, true},
		{dns.TypeSOA, dns.TypeTXT, true},
		{dns.TypeSOA, dns.TypeNS, true},
		{dns.TypeNS, dns.TypeSOA, false},
		{dns.TypeNS, dns.TypeA, true},
		{dns.TypeNS, dns.TypeTXT, true},
		{dns.TypeNS, dns.TypeNS, false},
		{dns.TypeA, dns.TypeSOA, false},
		{dns.TypeA, dns.TypeA, false},
		{dns.TypeA, dns.TypeTXT, true},
		{dns.TypeA, dns.TypeNS, false},
		{dns.TypeMX, dns.TypeSOA, false},
		{dns.TypeMX, dns.TypeA, false},
		{dns.TypeMX, dns.TypeTXT, true},
		{dns.TypeMX, dns.TypeNS, false},
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
