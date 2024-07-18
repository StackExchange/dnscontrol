package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListMagicTransitIPsecTunnels(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "ipsec_tunnels": [
          {
            "id": "c4a7362d577a6c3019a474fd6f485821",
            "created_on": "2017-06-14T00:00:00Z",
            "modified_on": "2017-06-14T05:20:00Z",
            "name": "IPsec_1",
            "customer_endpoint": "203.0.113.1",
            "cloudflare_endpoint": "203.0.113.2",
            "interface_address": "192.0.2.0/31",
            "description": "Tunnel for ISP X",
			"replay_protection": true
          }
        ]
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/ipsec_tunnels", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := []MagicTransitIPsecTunnel{
		{
			ID:                 "c4a7362d577a6c3019a474fd6f485821",
			CreatedOn:          &createdOn,
			ModifiedOn:         &modifiedOn,
			Name:               "IPsec_1",
			CustomerEndpoint:   "203.0.113.1",
			CloudflareEndpoint: "203.0.113.2",
			InterfaceAddress:   "192.0.2.0/31",
			Description:        "Tunnel for ISP X",
			ReplayProtection:   BoolPtr(true),
		},
	}

	actual, err := client.ListMagicTransitIPsecTunnels(context.Background(), testAccountID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetMagicTransitIPsecTunnel(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "ipsec_tunnel": {
          "id": "c4a7362d577a6c3019a474fd6f485821",
          "created_on": "2017-06-14T00:00:00Z",
          "modified_on": "2017-06-14T05:20:00Z",
          "name": "IPsec_1",
          "customer_endpoint": "203.0.113.1",
          "cloudflare_endpoint": "203.0.113.2",
          "interface_address": "192.0.2.0/31",
          "description": "Tunnel for ISP X",
          "allow_null_cipher": true,
		  "replay_protection": true
        }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/ipsec_tunnels/c4a7362d577a6c3019a474fd6f485821", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitIPsecTunnel{
		ID:                 "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:          &createdOn,
		ModifiedOn:         &modifiedOn,
		Name:               "IPsec_1",
		CustomerEndpoint:   "203.0.113.1",
		CloudflareEndpoint: "203.0.113.2",
		InterfaceAddress:   "192.0.2.0/31",
		Description:        "Tunnel for ISP X",
		AllowNullCipher:    true,
		ReplayProtection:   BoolPtr(true),
	}

	actual, err := client.GetMagicTransitIPsecTunnel(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateMagicTransitIPsecTunnels(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "ipsec_tunnels": [
          {
            "id": "c4a7362d577a6c3019a474fd6f485821",
            "created_on": "2017-06-14T00:00:00Z",
            "modified_on": "2017-06-14T05:20:00Z",
            "name": "IPsec_1",
            "customer_endpoint": "203.0.113.1",
            "cloudflare_endpoint": "203.0.113.2",
            "interface_address": "192.0.2.0/31",
            "description": "Tunnel for ISP X"
          }
        ]
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/ipsec_tunnels", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := []MagicTransitIPsecTunnel{{
		ID:                 "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:          &createdOn,
		ModifiedOn:         &modifiedOn,
		Name:               "IPsec_1",
		CustomerEndpoint:   "203.0.113.1",
		CloudflareEndpoint: "203.0.113.2",
		InterfaceAddress:   "192.0.2.0/31",
		Description:        "Tunnel for ISP X",
	}}

	actual, err := client.CreateMagicTransitIPsecTunnels(context.Background(), testAccountID, want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateMagicTransitIPsecTunnelsWithHealthcheck(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "ipsec_tunnels": [
          {
            "id": "c4a7362d577a6c3019a474fd6f485821",
            "created_on": "2017-06-14T00:00:00Z",
            "modified_on": "2017-06-14T05:20:00Z",
            "name": "IPsec_1",
            "customer_endpoint": "203.0.113.1",
            "cloudflare_endpoint": "203.0.113.2",
            "interface_address": "192.0.2.0/31",
            "description": "Tunnel for ISP X",
            "health_check": {
              "enabled": true,
              "type": "reply",
              "rate": "mid",
              "direction": "bidirectional"
            }
          }
        ]
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/ipsec_tunnels", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := []MagicTransitIPsecTunnel{{
		ID:                 "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:          &createdOn,
		ModifiedOn:         &modifiedOn,
		Name:               "IPsec_1",
		CustomerEndpoint:   "203.0.113.1",
		CloudflareEndpoint: "203.0.113.2",
		InterfaceAddress:   "192.0.2.0/31",
		Description:        "Tunnel for ISP X",
		HealthCheck: &MagicTransitTunnelHealthcheck{
			Enabled:   true,
			Type:      "reply",
			Rate:      "mid",
			Direction: "bidirectional",
		},
	}}

	actual, err := client.CreateMagicTransitIPsecTunnels(context.Background(), testAccountID, want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateMagicTransitIPsecTunnelsWithReplayProtection(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "ipsec_tunnels": [
          {
            "id": "c4a7362d577a6c3019a474fd6f485821",
            "created_on": "2017-06-14T00:00:00Z",
            "modified_on": "2017-06-14T05:20:00Z",
            "name": "IPsec_1",
            "customer_endpoint": "203.0.113.1",
            "cloudflare_endpoint": "203.0.113.2",
            "interface_address": "192.0.2.0/31",
            "description": "Tunnel for ISP X",
			"replay_protection": true
          }
        ]
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/ipsec_tunnels", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := []MagicTransitIPsecTunnel{{
		ID:                 "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:          &createdOn,
		ModifiedOn:         &modifiedOn,
		Name:               "IPsec_1",
		CustomerEndpoint:   "203.0.113.1",
		CloudflareEndpoint: "203.0.113.2",
		InterfaceAddress:   "192.0.2.0/31",
		Description:        "Tunnel for ISP X",
		ReplayProtection:   BoolPtr(true),
	}}

	actual, err := client.CreateMagicTransitIPsecTunnels(context.Background(), testAccountID, want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateMagicTransitIPsecTunnel(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "modified": true,
        "modified_ipsec_tunnel": {
          "id": "c4a7362d577a6c3019a474fd6f485821",
          "created_on": "2017-06-14T00:00:00Z",
          "modified_on": "2017-06-14T05:20:00Z",
          "name": "IPsec_1",
          "customer_endpoint": "203.0.113.1",
          "cloudflare_endpoint": "203.0.113.2",
          "interface_address": "192.0.2.0/31",
          "description": "Tunnel for ISP X",
          "allow_null_cipher": true
        }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/ipsec_tunnels/c4a7362d577a6c3019a474fd6f485821", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitIPsecTunnel{
		ID:                 "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:          &createdOn,
		ModifiedOn:         &modifiedOn,
		Name:               "IPsec_1",
		CustomerEndpoint:   "203.0.113.1",
		CloudflareEndpoint: "203.0.113.2",
		InterfaceAddress:   "192.0.2.0/31",
		Description:        "Tunnel for ISP X",
		AllowNullCipher:    true,
	}

	actual, err := client.UpdateMagicTransitIPsecTunnel(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteMagicTransitIPsecTunnel(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "deleted": true,
        "deleted_ipsec_tunnel": {
          "id": "c4a7362d577a6c3019a474fd6f485821",
          "created_on": "2017-06-14T00:00:00Z",
          "modified_on": "2017-06-14T05:20:00Z",
          "name": "IPsec_1",
          "customer_endpoint": "203.0.113.1",
          "cloudflare_endpoint": "203.0.113.2",
          "interface_address": "192.0.2.0/31",
          "description": "Tunnel for ISP X"
        }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/ipsec_tunnels/c4a7362d577a6c3019a474fd6f485821", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitIPsecTunnel{
		ID:                 "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:          &createdOn,
		ModifiedOn:         &modifiedOn,
		Name:               "IPsec_1",
		CustomerEndpoint:   "203.0.113.1",
		CloudflareEndpoint: "203.0.113.2",
		InterfaceAddress:   "192.0.2.0/31",
		Description:        "Tunnel for ISP X",
	}

	actual, err := client.DeleteMagicTransitIPsecTunnel(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestMagicTransitIPsecTunnelGeneratePSK(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "psk": "itworks",
				"psk_metadata": {
		      "last_generated_on": "2017-06-14T05:20:00Z"
		    }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/ipsec_tunnels/c4a7362d577a6c3019a474fd6f485821/psk_generate", handler)

	lastGeneratedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitIPsecTunnelPskMetadata{
		LastGeneratedOn: &lastGeneratedOn,
	}

	want_psk := "itworks"

	psk, actual, err := client.GenerateMagicTransitIPsecTunnelPSK(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821")
	if assert.NoError(t, err) {
		assert.Equal(t, want, *actual)
		assert.Equal(t, want_psk, psk)
	}
}
