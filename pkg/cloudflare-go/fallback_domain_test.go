package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListFallbackDomain(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
    {
      "success": true,
      "errors": [],
      "messages": [],
      "result": [
        {
          "suffix": "example.com",
          "description": "Domain bypass for local development"
       }
      ]
    }
    `)
	}

	want := []FallbackDomain{{
		Suffix:      "example.com",
		Description: "Domain bypass for local development",
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/fallback_domains", handler)

	actual, err := client.ListFallbackDomains(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListFallbackDomainsDeviceSettingsPolicy(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
    {
      "success": true,
      "errors": [],
      "messages": [],
      "result": [
        {
          "suffix": "example.com",
          "description": "Domain bypass for local development"
       }
      ]
    }
    `)
	}

	want := []FallbackDomain{{
		Suffix:      "example.com",
		Description: "Domain bypass for local development",
	}}

	policyID := "a842fa8a-a583-482e-9cd9-eb43362949fd"

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/"+policyID+"/fallback_domains", handler)

	actual, err := client.ListFallbackDomainsDeviceSettingsPolicy(context.Background(), testAccountID, policyID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestFallbackDomainDNSServer(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
    {
      "success": true,
      "errors": [],
      "messages": [],
      "result": [
        {
          "suffix": "example.com",
          "description": "Domain bypass for local development",
					"dns_server": ["192.168.0.1", "10.1.1.1"]
       }
      ]
    }
    `)
	}

	want := []FallbackDomain{{
		Suffix:      "example.com",
		Description: "Domain bypass for local development",
		DNSServer:   []string{"192.168.0.1", "10.1.1.1"},
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/fallback_domains", handler)

	actual, err := client.ListFallbackDomains(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateFallbackDomain(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
    {
      "success": true,
      "errors": [],
      "messages": [],
      "result": [
        {
          "suffix": "example_one.com",
          "description": "example one",
					"dns_server": ["192.168.0.1", "10.1.1.1"]
       },
       {
          "suffix": "example_two.com",
          "description": "example two"
       },
       {
          "suffix": "example_three.com",
          "description": "example three"
       }
      ]
    }
    `)
	}

	domains := []FallbackDomain{
		{
			Suffix:      "example_one.com",
			Description: "example one",
			DNSServer:   []string{"192.168.0.1", "10.1.1.1"},
		},
		{
			Suffix:      "example_two.com",
			Description: "example two",
		},
		{
			Suffix:      "example_three.com",
			Description: "example three",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/fallback_domains", handler)

	actual, err := client.UpdateFallbackDomain(context.Background(), testAccountID, domains)

	if assert.NoError(t, err) {
		assert.Equal(t, domains, actual)
	}
}

func TestUpdateFallbackDomainDeviceSettingsPolicy(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
    {
      "success": true,
      "errors": [],
      "messages": [],
      "result": [
        {
          "suffix": "example_one.com",
          "description": "example one",
					"dns_server": ["192.168.0.1", "10.1.1.1"]
       },
       {
          "suffix": "example_two.com",
          "description": "example two"
       },
       {
          "suffix": "example_three.com",
          "description": "example three"
       }
      ]
    }
    `)
	}

	domains := []FallbackDomain{
		{
			Suffix:      "example_one.com",
			Description: "example one",
			DNSServer:   []string{"192.168.0.1", "10.1.1.1"},
		},
		{
			Suffix:      "example_two.com",
			Description: "example two",
		},
		{
			Suffix:      "example_three.com",
			Description: "example three",
		},
	}

	policyID := "a842fa8a-a583-482e-9cd9-eb43362949fd"

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/"+policyID+"/fallback_domains", handler)

	actual, err := client.UpdateFallbackDomainDeviceSettingsPolicy(context.Background(), testAccountID, policyID, domains)

	if assert.NoError(t, err) {
		assert.Equal(t, domains, actual)
	}
}
