package spflib

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	dnsres, err := NewCache("testdata-dns1.json")
	if err != nil {
		t.Fatal(err)
	}
	rec, err := Parse(strings.Join([]string{"v=spf1",
		"ip4:198.252.206.0/24",
		"ip4:192.111.0.0/24",
		"include:_spf.google.com",
		"include:mailgun.org",
		//"include:spf-basic.fogcreek.com",
		"include:mail.zendesk.com",
		"include:servers.mcsv.net",
		"include:sendgrid.net",
		"include:spf.mtasv.net",
		"exists:%{i}._spf.sparkpostmail.com",
		"ptr:sparkpostmail.com",
		"~all"}, " "), dnsres)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rec.Print())
}

func TestParseWithDoubleSpaces(t *testing.T) {
	dnsres, err := NewCache("testdata-dns1.json")
	if err != nil {
		t.Fatal(err)
	}
	rec, err := Parse("v=spf1 ip4:192.111.0.0/24  ip4:192.111.1.0/24 -all", dnsres)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rec.Print())
}

func TestParseRedirectNotLast(t *testing.T) {
	// Make sure redirect=foo fails if it isn't the last item.
	dnsres, err := NewCache("testdata-dns1.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Parse(strings.Join([]string{"v=spf1",
		"redirect=servers.mcsv.net",
		"~all"}, " "), dnsres)
	if err == nil {
		t.Fatal("should fail")
	}
}

func TestParseRedirectColon(t *testing.T) {
	// Make sure redirect:foo fails.
	dnsres, err := NewCache("testdata-dns1.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Parse(strings.Join([]string{"v=spf1",
		"redirect:servers.mcsv.net",
	}, " "), dnsres)
	if err == nil {
		t.Fatal("should fail")
	}
}

func TestParseRedirectOnly(t *testing.T) {
	dnsres, err := NewCache("testdata-dns1.json")
	if err != nil {
		t.Fatal(err)
	}
	rec, err := Parse(strings.Join([]string{"v=spf1",
		"redirect=servers.mcsv.net"}, " "), dnsres)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rec.Print())
}

func TestParseRedirectLast(t *testing.T) {
	dnsres, err := NewCache("testdata-dns1.json")
	if err != nil {
		t.Fatal(err)
	}
	rec, err := Parse(strings.Join([]string{"v=spf1",
		"ip4:198.252.206.0/24",
		"redirect=servers.mcsv.net"}, " "), dnsres)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rec.Print())
}
