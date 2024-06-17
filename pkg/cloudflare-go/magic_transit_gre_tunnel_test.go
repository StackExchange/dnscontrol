package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListMagicTransitGRETunnels(t *testing.T) {
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
        "gre_tunnels": [
          {
            "id": "c4a7362d577a6c3019a474fd6f485821",
            "created_on": "2017-06-14T00:00:00Z",
            "modified_on": "2017-06-14T05:20:00Z",
            "name": "GRE_1",
            "customer_gre_endpoint": "203.0.113.1",
            "cloudflare_gre_endpoint": "203.0.113.2",
            "interface_address": "192.0.2.0/31",
            "description": "Tunnel for ISP X",
            "ttl": 64,
            "mtu": 1476,
            "health_check": {
              "enabled": true,
              "target": "203.0.113.1",
              "type": "request"
            }
          }
        ]
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/gre_tunnels", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := []MagicTransitGRETunnel{
		{
			ID:                    "c4a7362d577a6c3019a474fd6f485821",
			CreatedOn:             &createdOn,
			ModifiedOn:            &modifiedOn,
			Name:                  "GRE_1",
			CustomerGREEndpoint:   "203.0.113.1",
			CloudflareGREEndpoint: "203.0.113.2",
			InterfaceAddress:      "192.0.2.0/31",
			Description:           "Tunnel for ISP X",
			TTL:                   64,
			MTU:                   1476,
			HealthCheck: &MagicTransitGRETunnelHealthcheck{
				Enabled: true,
				Target:  "203.0.113.1",
				Type:    "request",
			},
		},
	}

	actual, err := client.ListMagicTransitGRETunnels(context.Background(), testAccountID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetMagicTransitGRETunnel(t *testing.T) {
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
        "gre_tunnel": {
          "id": "c4a7362d577a6c3019a474fd6f485821",
          "created_on": "2017-06-14T00:00:00Z",
          "modified_on": "2017-06-14T05:20:00Z",
          "name": "GRE_1",
          "customer_gre_endpoint": "203.0.113.1",
          "cloudflare_gre_endpoint": "203.0.113.2",
          "interface_address": "192.0.2.0/31",
          "description": "Tunnel for ISP X",
          "ttl": 64,
          "mtu": 1476,
          "health_check": {
            "enabled": true,
            "target": "203.0.113.1",
            "type": "request"
          }
        }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/gre_tunnels/c4a7362d577a6c3019a474fd6f485821", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitGRETunnel{
		ID:                    "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:             &createdOn,
		ModifiedOn:            &modifiedOn,
		Name:                  "GRE_1",
		CustomerGREEndpoint:   "203.0.113.1",
		CloudflareGREEndpoint: "203.0.113.2",
		InterfaceAddress:      "192.0.2.0/31",
		Description:           "Tunnel for ISP X",
		TTL:                   64,
		MTU:                   1476,
		HealthCheck: &MagicTransitGRETunnelHealthcheck{
			Enabled: true,
			Target:  "203.0.113.1",
			Type:    "request",
		},
	}

	actual, err := client.GetMagicTransitGRETunnel(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateMagicTransitGRETunnels(t *testing.T) {
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
        "gre_tunnels": [
          {
            "id": "c4a7362d577a6c3019a474fd6f485821",
            "created_on": "2017-06-14T00:00:00Z",
            "modified_on": "2017-06-14T05:20:00Z",
            "name": "GRE_1",
            "customer_gre_endpoint": "203.0.113.1",
            "cloudflare_gre_endpoint": "203.0.113.2",
            "interface_address": "192.0.2.0/31",
            "description": "Tunnel for ISP X",
            "ttl": 64,
            "mtu": 1476,
            "health_check": {
                  "enabled": true,
                  "target": "203.0.113.1",
                  "type": "request"
            }
          }
        ]
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/gre_tunnels", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := []MagicTransitGRETunnel{
		{
			ID:                    "c4a7362d577a6c3019a474fd6f485821",
			CreatedOn:             &createdOn,
			ModifiedOn:            &modifiedOn,
			Name:                  "GRE_1",
			CustomerGREEndpoint:   "203.0.113.1",
			CloudflareGREEndpoint: "203.0.113.2",
			InterfaceAddress:      "192.0.2.0/31",
			Description:           "Tunnel for ISP X",
			TTL:                   64,
			MTU:                   1476,
			HealthCheck: &MagicTransitGRETunnelHealthcheck{
				Enabled: true,
				Target:  "203.0.113.1",
				Type:    "request",
			},
		},
	}

	actual, err := client.CreateMagicTransitGRETunnels(context.Background(), testAccountID, want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateMagicTransitGRETunnel(t *testing.T) {
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
        "modified_gre_tunnel": {
          "id": "c4a7362d577a6c3019a474fd6f485821",
          "created_on": "2017-06-14T00:00:00Z",
          "modified_on": "2017-06-14T05:20:00Z",
          "name": "GRE_1",
          "customer_gre_endpoint": "203.0.113.1",
          "cloudflare_gre_endpoint": "203.0.113.2",
          "interface_address": "192.0.2.0/31",
          "description": "Tunnel for ISP X",
          "ttl": 64,
          "mtu": 1476,
          "health_check": {
            "enabled": true,
            "target": "203.0.113.1",
            "type": "request"
          }
        }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/gre_tunnels/c4a7362d577a6c3019a474fd6f485821", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitGRETunnel{
		ID:                    "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:             &createdOn,
		ModifiedOn:            &modifiedOn,
		Name:                  "GRE_1",
		CustomerGREEndpoint:   "203.0.113.1",
		CloudflareGREEndpoint: "203.0.113.2",
		InterfaceAddress:      "192.0.2.0/31",
		Description:           "Tunnel for ISP X",
		TTL:                   64,
		MTU:                   1476,
		HealthCheck: &MagicTransitGRETunnelHealthcheck{
			Enabled: true,
			Target:  "203.0.113.1",
			Type:    "request",
		},
	}

	actual, err := client.UpdateMagicTransitGRETunnel(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteMagicTransitGRETunnel(t *testing.T) {
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
        "deleted_gre_tunnel": {
          "id": "c4a7362d577a6c3019a474fd6f485821",
          "created_on": "2017-06-14T00:00:00Z",
          "modified_on": "2017-06-14T05:20:00Z",
          "name": "GRE_1",
          "customer_gre_endpoint": "203.0.113.1",
          "cloudflare_gre_endpoint": "203.0.113.2",
          "interface_address": "192.0.2.0/31",
          "description": "Tunnel for ISP X",
          "ttl": 64,
          "mtu": 1476,
          "health_check": {
            "enabled": true,
            "target": "203.0.113.1",
            "type": "request"
          }
        }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/gre_tunnels/c4a7362d577a6c3019a474fd6f485821", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitGRETunnel{
		ID:                    "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:             &createdOn,
		ModifiedOn:            &modifiedOn,
		Name:                  "GRE_1",
		CustomerGREEndpoint:   "203.0.113.1",
		CloudflareGREEndpoint: "203.0.113.2",
		InterfaceAddress:      "192.0.2.0/31",
		Description:           "Tunnel for ISP X",
		TTL:                   64,
		MTU:                   1476,
		HealthCheck: &MagicTransitGRETunnelHealthcheck{
			Enabled: true,
			Target:  "203.0.113.1",
			Type:    "request",
		},
	}

	actual, err := client.DeleteMagicTransitGRETunnel(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
