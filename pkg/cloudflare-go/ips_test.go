package cloudflare

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockTransport struct {
	http.Transport
	Server *httptest.Server
	Path   string
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	url, err := url.Parse(m.Server.URL + m.Path)
	if err != nil {
		return nil, err
	}

	req.URL = url

	return m.Transport.RoundTrip(req)
}

func Test_IPs(t *testing.T) {
	setup()
	defer teardown()

	mux := http.NewServeMux()
	server = httptest.NewServer(mux)
	defer server.Close()

	defaultTransport := http.DefaultTransport
	http.DefaultTransport = &MockTransport{
		Server: server,
		Path:   "/ips",
	}
	defer func() { http.DefaultTransport = defaultTransport }()

	mux.HandleFunc("/ips", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "ipv4_cidrs": ["198.51.100.0/24"],
    "ipv6_cidrs": ["ffff:ffff::/32"],
    "china_colos": ["42.81.6.0/25", "2408:871a:1801:7::/72"]
  }
}`)
	})

	ipRanges, err := IPs()

	assert.NoError(t, err)

	assert.Len(t, ipRanges.IPv4CIDRs, 1)
	assert.Equal(t, "198.51.100.0/24", ipRanges.IPv4CIDRs[0])
	assert.Len(t, ipRanges.IPv6CIDRs, 1)
	assert.Equal(t, "ffff:ffff::/32", ipRanges.IPv6CIDRs[0])
	assert.Len(t, ipRanges.ChinaIPv4CIDRs, 1)
	assert.Equal(t, "42.81.6.0/25", ipRanges.ChinaIPv4CIDRs[0])
	assert.Len(t, ipRanges.ChinaIPv6CIDRs, 1)
	assert.Equal(t, "2408:871a:1801:7::/72", ipRanges.ChinaIPv6CIDRs[0])
}
