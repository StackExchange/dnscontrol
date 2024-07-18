package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitTunnelIncludeHost(t *testing.T) {
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
          "host": "*.example.com",
          "description": "default"
       }
      ]
    }
    `)
	}

	want := []SplitTunnel{{
		Host:        "*.example.com",
		Description: "default",
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/include", handler)

	actual, err := client.ListSplitTunnels(context.Background(), testAccountID, "include")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestSplitTunnelIncludeAddress(t *testing.T) {
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
          "address": "192.0.2.0/24",
          "description": "TEST-NET-1"
       }
      ]
    }
    `)
	}

	want := []SplitTunnel{{
		Address:     "192.0.2.0/24",
		Description: "TEST-NET-1",
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/include", handler)

	actual, err := client.ListSplitTunnels(context.Background(), testAccountID, "include")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateSplitTunnelInclude(t *testing.T) {
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
          "address": "192.0.2.0/24",
          "description": "TEST-NET-1"
       },
       {
          "address": "198.51.100.0/24",
          "description": "TEST-NET-2"
       },
       {
          "host": "*.example.com",
          "description": "example host name"
       }
      ]
    }
    `)
	}

	tunnels := []SplitTunnel{
		{
			Address:     "192.0.2.0/24",
			Description: "TEST-NET-1",
		},
		{
			Address:     "198.51.100.0/24",
			Description: "TEST-NET-2",
		},
		{
			Host:        "*.example.com",
			Description: "example host name",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/include", handler)

	actual, err := client.UpdateSplitTunnel(context.Background(), testAccountID, "include", tunnels)

	if assert.NoError(t, err) {
		assert.Equal(t, tunnels, actual)
	}
}

func TestSplitTunnelExcludeHost(t *testing.T) {
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
          "host": "*.example.com",
          "description": "default"
       }
      ]
    }
    `)
	}

	want := []SplitTunnel{{
		Host:        "*.example.com",
		Description: "default",
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/exclude", handler)

	actual, err := client.ListSplitTunnels(context.Background(), testAccountID, "exclude")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestSplitTunnelExcludeAddress(t *testing.T) {
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
          "address": "192.0.2.0/24",
          "description": "TEST-NET-1"
       }
      ]
    }
    `)
	}

	want := []SplitTunnel{{
		Address:     "192.0.2.0/24",
		Description: "TEST-NET-1",
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/exclude", handler)

	actual, err := client.ListSplitTunnels(context.Background(), testAccountID, "exclude")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateSplitTunnelExclude(t *testing.T) {
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
          "address": "192.0.2.0/24",
          "description": "TEST-NET-1"
       },
       {
          "address": "198.51.100.0/24",
          "description": "TEST-NET-2"
       },
       {
          "host": "*.example.com",
          "description": "example host name"
       }
      ]
    }
    `)
	}

	tunnels := []SplitTunnel{
		{
			Address:     "192.0.2.0/24",
			Description: "TEST-NET-1",
		},
		{
			Address:     "198.51.100.0/24",
			Description: "TEST-NET-2",
		},
		{
			Host:        "*.example.com",
			Description: "example host name",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/exclude", handler)

	actual, err := client.UpdateSplitTunnel(context.Background(), testAccountID, "exclude", tunnels)

	if assert.NoError(t, err) {
		assert.Equal(t, tunnels, actual)
	}
}

func TestSplitTunnelsDeviceSettingsPolicy(t *testing.T) {
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
          "host": "*.example.com",
          "description": "default"
       }
      ]
    }
    `)
	}

	want := []SplitTunnel{{
		Host:        "*.example.com",
		Description: "default",
	}}

	policyID := "a842fa8a-a583-482e-9cd9-eb43362949fd"

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/"+policyID+"/include", handler)

	actual, err := client.ListSplitTunnelsDeviceSettingsPolicy(context.Background(), testAccountID, policyID, "include")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
