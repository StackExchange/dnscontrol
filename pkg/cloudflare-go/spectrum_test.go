package cloudflare

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpectrumApplication(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "f68579455bd947efb65ffa1bcf33b52c",
				"protocol": "tcp/22",
				"ipv4": true,
				"dns": {
					"type": "CNAME",
					"name": "spectrum.example.com"
				},
				"origin_direct": [
					"tcp://192.0.2.1:22"
				],
				"ip_firewall": true,
				"proxy_protocol": "off",
				"tls": "off",
				"created_on": "2018-03-28T21:25:55.643771Z",
				"modified_on": "2018-03-28T21:25:55.643771Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/spectrum/apps/f68579455bd947efb65ffa1bcf33b52c", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	want := SpectrumApplication{
		ID:         "f68579455bd947efb65ffa1bcf33b52c",
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
		Protocol:   "tcp/22",
		IPv4:       true,
		DNS: SpectrumApplicationDNS{
			Name: "spectrum.example.com",
			Type: "CNAME",
		},
		OriginDirect:  []string{"tcp://192.0.2.1:22"},
		IPFirewall:    true,
		ProxyProtocol: "off",
		TLS:           "off",
	}

	actual, err := client.SpectrumApplication(context.Background(), "01a7362d577a6c3019a474fd6f485823", "f68579455bd947efb65ffa1bcf33b52c")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestSpectrumApplications(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": [
					{
					"id": "f68579455bd947efb65ffa1bcf33b52c",
					"protocol": "tcp/22",
					"ipv4": true,
					"dns": {
						"type": "CNAME",
						"name": "spectrum.example.com"
					},
					"origin_direct": [
						"tcp://192.0.2.1:22"
					],
					"ip_firewall": true,
					"proxy_protocol": "off",
					"tls": "off",
					"created_on": "2018-03-28T21:25:55.643771Z",
					"modified_on": "2018-03-28T21:25:55.643771Z"
				}
			],
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/spectrum/apps", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	want := []SpectrumApplication{
		{
			ID:         "f68579455bd947efb65ffa1bcf33b52c",
			CreatedOn:  &createdOn,
			ModifiedOn: &modifiedOn,
			Protocol:   "tcp/22",
			IPv4:       true,
			DNS: SpectrumApplicationDNS{
				Name: "spectrum.example.com",
				Type: "CNAME",
			},
			OriginDirect:  []string{"tcp://192.0.2.1:22"},
			IPFirewall:    true,
			ProxyProtocol: "off",
			TLS:           "off",
		},
	}

	actual, err := client.SpectrumApplications(context.Background(), "01a7362d577a6c3019a474fd6f485823")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateSpectrumApplication(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "f68579455bd947efb65ffa1bcf33b52c",
				"protocol": "tcp/23",
				"ipv4": true,
				"dns": {
					"type": "CNAME",
					"name": "spectrum1.example.com"
				},
				"origin_direct": [
					"tcp://192.0.2.1:23"
				],
				"ip_firewall": true,
				"proxy_protocol": "off",
				"tls": "full",
				"created_on": "2018-03-28T21:25:55.643771Z",
				"modified_on": "2018-03-28T21:25:55.643771Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/spectrum/apps/f68579455bd947efb65ffa1bcf33b52c", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	want := SpectrumApplication{
		ID:       "f68579455bd947efb65ffa1bcf33b52c",
		Protocol: "tcp/23",
		IPv4:     true,
		DNS: SpectrumApplicationDNS{
			Type: "CNAME",
			Name: "spectrum1.example.com",
		},
		OriginDirect:  []string{"tcp://192.0.2.1:23"},
		IPFirewall:    true,
		ProxyProtocol: "off",
		TLS:           "full",
		CreatedOn:     &createdOn,
		ModifiedOn:    &modifiedOn,
	}

	actual, err := client.UpdateSpectrumApplication(context.Background(), "01a7362d577a6c3019a474fd6f485823", "f68579455bd947efb65ffa1bcf33b52c", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateSpectrumApplication(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "f68579455bd947efb65ffa1bcf33b52c",
				"protocol": "tcp/22",
				"ipv4": true,
				"dns": {
					"type": "CNAME",
					"name": "spectrum.example.com"
				},
				"origin_direct": [
					"tcp://192.0.2.1:22"
				],
				"ip_firewall": true,
				"proxy_protocol": "off",
				"tls": "full",
				"created_on": "2018-03-28T21:25:55.643771Z",
				"modified_on": "2018-03-28T21:25:55.643771Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/spectrum/apps", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	want := SpectrumApplication{
		ID:       "f68579455bd947efb65ffa1bcf33b52c",
		Protocol: "tcp/22",
		IPv4:     true,
		DNS: SpectrumApplicationDNS{
			Type: "CNAME",
			Name: "spectrum.example.com",
		},
		OriginDirect:  []string{"tcp://192.0.2.1:22"},
		IPFirewall:    true,
		ProxyProtocol: "off",
		TLS:           "full",
		CreatedOn:     &createdOn,
		ModifiedOn:    &modifiedOn,
	}

	actual, err := client.CreateSpectrumApplication(context.Background(), "01a7362d577a6c3019a474fd6f485823", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateSpectrumApplication_OriginDNS(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "5683dc9a12ba4dc6bceaca011bcafcf5",
				"protocol": "tcp/22",
				"ipv4": true,
				"dns": {
					"type": "CNAME",
					"name": "spectrum.example.com"
				},
				"origin_dns": {
					"name" : "spectrum.origin.example.com"
				},
				"origin_port": 2022,
				"ip_firewall": true,
				"proxy_protocol": "off",
				"tls": "full",
				"created_on": "2018-03-28T21:25:55.643771Z",
				"modified_on": "2018-03-28T21:25:55.643771Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/spectrum/apps", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	want := SpectrumApplication{
		ID:       "5683dc9a12ba4dc6bceaca011bcafcf5",
		Protocol: "tcp/22",
		IPv4:     true,
		DNS: SpectrumApplicationDNS{
			Type: "CNAME",
			Name: "spectrum.example.com",
		},
		OriginDNS: &SpectrumApplicationOriginDNS{
			Name: "spectrum.origin.example.com",
		},
		OriginPort: &SpectrumApplicationOriginPort{
			Port: 2022,
		},
		IPFirewall:    true,
		ProxyProtocol: "off",
		TLS:           "full",
		CreatedOn:     &createdOn,
		ModifiedOn:    &modifiedOn,
	}

	actual, err := client.CreateSpectrumApplication(context.Background(), "01a7362d577a6c3019a474fd6f485823", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteSpectrumApplication(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "40d67c87c6cd4b889a4fd57805225e85"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/spectrum/apps/f68579455bd947efb65ffa1bcf33b52c", handler)

	err := client.DeleteSpectrumApplication(context.Background(), "01a7362d577a6c3019a474fd6f485823", "f68579455bd947efb65ffa1bcf33b52c")
	assert.NoError(t, err)
}

func TestSpectrumApplicationProxyProtocolDeprecations(t *testing.T) {
	for _, testCase := range []struct {
		actualProxyProtocol   bool
		actualSPP             bool
		expectedProxyProtocol ProxyProtocol
	}{
		{
			actualProxyProtocol:   false,
			actualSPP:             false,
			expectedProxyProtocol: "off",
		},
		{
			actualProxyProtocol:   true,
			actualSPP:             false,
			expectedProxyProtocol: "v1",
		},
		{
			actualProxyProtocol:   false,
			actualSPP:             true,
			expectedProxyProtocol: "simple",
		},
	} {
		setup()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
			w.Header().Set("content-type", "application/json")
			fmt.Fprintf(w, `{
			"result": {
				"id": "f68579455bd947efb65ffa1bcf33b52c",
				"protocol": "tcp/22",
				"ipv4": true,
				"dns": {
					"type": "CNAME",
					"name": "spectrum.example.com"
				},
				"origin_direct": [
					"tcp://192.0.2.1:22"
				],
				"ip_firewall": true,
				"proxy_protocol": %v,
				"spp": %v,
				"tls": "off",
				"created_on": "2018-03-28T21:25:55.643771Z",
				"modified_on": "2018-03-28T21:25:55.643771Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`, testCase.actualProxyProtocol, testCase.actualSPP)
		}

		mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/spectrum/apps/f68579455bd947efb65ffa1bcf33b52c", handler)
		createdOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
		modifiedOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
		want := SpectrumApplication{
			ID:         "f68579455bd947efb65ffa1bcf33b52c",
			CreatedOn:  &createdOn,
			ModifiedOn: &modifiedOn,
			Protocol:   "tcp/22",
			IPv4:       true,
			DNS: SpectrumApplicationDNS{
				Name: "spectrum.example.com",
				Type: "CNAME",
			},
			OriginDirect:  []string{"tcp://192.0.2.1:22"},
			IPFirewall:    true,
			ProxyProtocol: testCase.expectedProxyProtocol,
			TLS:           "off",
		}

		actual, err := client.SpectrumApplication(context.Background(), "01a7362d577a6c3019a474fd6f485823", "f68579455bd947efb65ffa1bcf33b52c")
		if assert.NoError(t, err) {
			assert.Equal(t, want, actual)
		}

		teardown()
	}
}

func TestSpectrumApplicationEdgeIPs(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "f68579455bd947efb65ffa1bcf33b52c",
				"protocol": "tcp/22",
				"ipv4": true,
				"dns": {
					"type": "CNAME",
					"name": "spectrum.example.com"
				},
				"origin_direct": [
					"tcp://192.0.2.1:22"
				],
				"ip_firewall": true,
				"proxy_protocol": "off",
				"tls": "off",
				"edge_ips": {
					"type": "static",
					"ips": [
						"192.0.2.1",
						"2001:db8::1"
					]
				},
				"created_on": "2018-03-28T21:25:55.643771Z",
				"modified_on": "2018-03-28T21:25:55.643771Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/spectrum/apps/f68579455bd947efb65ffa1bcf33b52c", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	want := SpectrumApplication{
		ID:         "f68579455bd947efb65ffa1bcf33b52c",
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
		Protocol:   "tcp/22",
		IPv4:       true,
		DNS: SpectrumApplicationDNS{
			Name: "spectrum.example.com",
			Type: "CNAME",
		},
		OriginDirect:  []string{"tcp://192.0.2.1:22"},
		IPFirewall:    true,
		ProxyProtocol: "off",
		TLS:           "off",
		EdgeIPs: &SpectrumApplicationEdgeIPs{
			Type: SpectrumEdgeTypeStatic,
			IPs:  []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("2001:db8::1")},
		},
	}

	actual, err := client.SpectrumApplication(context.Background(), "01a7362d577a6c3019a474fd6f485823", "f68579455bd947efb65ffa1bcf33b52c")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestSpectrumApplicationPortRange(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "f68579455bd947efb65ffa1bcf33b52c",
				"protocol": "tcp/22-23",
				"ipv4": true,
				"dns": {
					"type": "CNAME",
					"name": "spectrum.example.com"
				},
				"origin_dns": {
				  "name": "cloudflare.com"
				},
				"origin_port": "2022-2023",
				"ip_firewall": true,
				"proxy_protocol": "off",
				"tls": "off",
				"edge_ips": {
					"type": "static",
					"ips": [
						"192.0.2.1",
						"2001:db8::1"
					]
				},
				"created_on": "2018-03-28T21:25:55.643771Z",
				"modified_on": "2018-03-28T21:25:55.643771Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/spectrum/apps/f68579455bd947efb65ffa1bcf33b52c", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-03-28T21:25:55.643771Z")
	want := SpectrumApplication{
		ID:         "f68579455bd947efb65ffa1bcf33b52c",
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
		Protocol:   "tcp/22-23",
		IPv4:       true,
		DNS: SpectrumApplicationDNS{
			Name: "spectrum.example.com",
			Type: "CNAME",
		},
		OriginDNS: &SpectrumApplicationOriginDNS{
			Name: "cloudflare.com",
		},
		OriginPort: &SpectrumApplicationOriginPort{
			Start: 2022,
			End:   2023,
		},
		IPFirewall:    true,
		ProxyProtocol: "off",
		TLS:           "off",
		EdgeIPs: &SpectrumApplicationEdgeIPs{
			Type: SpectrumEdgeTypeStatic,
			IPs:  []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("2001:db8::1")},
		},
	}

	actual, err := client.SpectrumApplication(context.Background(), "01a7362d577a6c3019a474fd6f485823", "f68579455bd947efb65ffa1bcf33b52c")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
