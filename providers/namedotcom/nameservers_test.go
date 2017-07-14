package namedotcom

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/models"
)

var (
	mux    *http.ServeMux
	client *nameDotCom
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = &nameDotCom{
		APIUser: "bob",
		APIKey:  "123",
		APIUrl:  server.URL,
	}
}

func teardown() {
	server.Close()
}

func TestGetNameserversRaw(t *testing.T) {
	for i, test := range []struct {
		givenNs, expected string
	}{
		{"", ""},
		{`"foo.ns.tld","bar.ns.tld"`, "bar.ns.tld,foo.ns.tld"},
		{"ERR", "ERR"},
		{"MSGERR", "ERR"},
	} {
		setup()
		defer teardown()

		mux.HandleFunc("/domain/get/example.tld", func(w http.ResponseWriter, r *http.Request) {
			if test.givenNs == "ERR" {
				http.Error(w, "UH OH", 500)
				return
			}
			if test.givenNs == "MSGERR" {
				w.Write(nameComError)
				return
			}
			w.Write(domainResponse(test.givenNs))
		})

		found, err := client.getNameserversRaw("example.tld")
		if err != nil {
			if test.expected == "ERR" {
				continue
			}
			t.Errorf("Error on test %d: %s", i, err)
			continue
		}
		if test.expected == "ERR" {
			t.Errorf("Expected error on test %d, but was none", i)
			continue
		}
		if strings.Join(found, ",") != test.expected {
			t.Errorf("Test %d: Expected '%s', but found '%s'", i, test.expected, found)
		}
	}
}

func TestGetCorrections(t *testing.T) {
	for i, test := range []struct {
		givenNs  string
		expected int
	}{
		{"", 1},
		{`"foo.ns.tld","bar.ns.tld"`, 0},
		{`"bar.ns.tld","foo.ns.tld"`, 0},
		{`"foo.ns.tld"`, 1},
		{`"1.ns.aaa","2.ns.www"`, 1},
		{"ERR", -1}, //-1 means we expect an error
		{"MSGERR", -1},
	} {
		setup()
		defer teardown()

		mux.HandleFunc("/domain/get/example.tld", func(w http.ResponseWriter, r *http.Request) {
			if test.givenNs == "ERR" {
				http.Error(w, "UH OH", 500)
				return
			}
			if test.givenNs == "MSGERR" {
				w.Write(nameComError)
				return
			}
			w.Write(domainResponse(test.givenNs))
		})
		dc := &models.DomainConfig{
			Name: "example.tld",
			Nameservers: []*models.Nameserver{
				{Name: "foo.ns.tld"},
				{Name: "bar.ns.tld"},
			},
		}
		corrections, err := client.GetRegistrarCorrections(dc)
		if err != nil {
			if test.expected == -1 {
				continue
			}
			t.Errorf("Error on test %d: %s", i, err)
			continue
		}
		if test.expected == -1 {
			t.Errorf("Expected error on test %d, but was none", i)
			continue
		}
		if len(corrections) != test.expected {
			t.Errorf("Test %d: Expected '%d', but found '%d'", i, test.expected, len(corrections))
		}
	}
}

func domainResponse(ns string) []byte {
	return []byte(fmt.Sprintf(`{
  "result": {
    "code": 100,
    "message": "Command Successful"
  },
  "domain_name": "example.tld",
  "create_date": "2015-12-28 18:08:05",
  "expire_date": "2016-12-28 23:59:59",
  "locked": true,
  "nameservers": [%s],
  "contacts": [],
  "addons": {
    "whois_privacy": {
      "price": "3.99"
    },
    "domain\/renew": {
      "price": "10.99"
    }
  }
}`, ns))
}

var nameComError = []byte(`{"result":{"code":251,"message":"Authentication Error - Invalid Username Or Api Token"}}`)

func TestGetNameservers(t *testing.T) {
	const d = "ns1.name.com,ns2.name.com,ns3.name.com,ns4.name.com"
	for i, test := range []struct {
		givenNs, expected string
	}{
		//empty or external dsp, use ns1-4.name.com
		{"", d},
		{`"foo.ns.tld","bar.ns.tld"`, d},
		//if already on name.com, use the existing nameservers
		{`"ns1aaa.name.com","ns2bbb.name.com","ns3ccc.name.com","ns4ddd.name.com"`, "ns1aaa.name.com,ns2bbb.name.com,ns3ccc.name.com,ns4ddd.name.com"},
		//also handle half and half
		{`"ns1aaa.name.com","ns2bbb.name.com","ns3ccc.aws.net","ns4ddd.awsdns.org"`, "ns1aaa.name.com,ns2bbb.name.com,ns3.name.com,ns4.name.com"},
		{`"nsa.azuredns.com","ns2b.gandhi.net","ns3ccc.name.com","ns4ddd.name.com"`, "ns1.name.com,ns2.name.com,ns3ccc.name.com,ns4ddd.name.com"},
	} {
		setup()
		defer teardown()

		mux.HandleFunc("/domain/get/example.tld", func(w http.ResponseWriter, r *http.Request) {
			w.Write(domainResponse(test.givenNs))
		})
		found, err := client.GetNameservers("example.tld")
		if err != nil {
			t.Errorf("Test %d: %s", i, err)
			continue
		}
		fStrs := []string{}
		for _, n := range found {
			fStrs = append(fStrs, n.Name)
		}
		if strings.Join(fStrs, ",") != test.expected {
			t.Errorf("Test %d: Expected '%s', but found '%s'", i, test.expected, found)
		}
	}
}
