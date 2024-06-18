package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListManagedHeaders(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "managed_request_headers": [
          {
            "id": "add_true_client_ip_headers",
            "enabled": false,
            "has_conflict": false,
            "conflicts_with": ["remove_visitor_ip_headers"]
          },
          {
            "id": "add_visitor_location_headers",
            "enabled": true,
            "has_conflict": false
          }
        ],
        "managed_response_headers": [
          {
            "id": "add_security_headers",
            "enabled": false,
            "has_conflict": false
          },
          {
            "id": "remove_x-powered-by_header",
            "enabled": true,
            "has_conflict": false
          }
        ]
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/managed_headers", handler)

	want := ManagedHeaders{
		ManagedRequestHeaders: []ManagedHeader{
			{
				ID:            "add_true_client_ip_headers",
				Enabled:       false,
				HasCoflict:    false,
				ConflictsWith: []string{"remove_visitor_ip_headers"},
			},
			{
				ID:         "add_visitor_location_headers",
				Enabled:    true,
				HasCoflict: false,
			},
		},
		ManagedResponseHeaders: []ManagedHeader{
			{
				ID:         "add_security_headers",
				Enabled:    false,
				HasCoflict: false,
			},
			{
				ID:         "remove_x-powered-by_header",
				Enabled:    true,
				HasCoflict: false,
			},
		},
	}

	zoneActual, err := client.ListZoneManagedHeaders(context.Background(), ZoneIdentifier(testZoneID), ListManagedHeadersParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}
}

func TestFilterManagedHeaders(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, r.URL.Query().Get("status"), "enabled")
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "managed_request_headers": [
          {
            "id": "add_visitor_location_headers",
            "enabled": true,
            "has_conflict": false
          }
        ],
        "managed_response_headers": [
          {
            "id": "remove_x-powered-by_header",
            "enabled": true,
            "has_conflict": false
          }
        ]
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/managed_headers", handler)

	want := ManagedHeaders{
		ManagedRequestHeaders: []ManagedHeader{
			{
				ID:         "add_visitor_location_headers",
				Enabled:    true,
				HasCoflict: false,
			},
		},
		ManagedResponseHeaders: []ManagedHeader{
			{
				ID:         "remove_x-powered-by_header",
				Enabled:    true,
				HasCoflict: false,
			},
		},
	}

	zoneActual, err := client.ListZoneManagedHeaders(context.Background(), ZoneIdentifier(testZoneID), ListManagedHeadersParams{
		Status: "enabled",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}
}

func TestUpdateManagedHeaders(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "managed_request_headers": [
          {
            "id": "add_true_client_ip_headers",
            "enabled": true,
            "has_conflict": false,
            "conflicts_with": ["remove_visitor_ip_headers"]
          },
          {
            "id": "add_visitor_location_headers",
            "enabled": true,
            "has_conflict": false
          }
        ],
        "managed_response_headers": [
          {
            "id": "add_security_headers",
            "enabled": false,
            "has_conflict": false
          },
          {
            "id": "remove_x-powered-by_header",
            "enabled": false,
            "has_conflict": false
          }
        ]
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/managed_headers", handler)
	managedHeadersForUpdate := ManagedHeaders{
		ManagedRequestHeaders: []ManagedHeader{
			{
				ID:      "add_visitor_location_headers",
				Enabled: true,
			},
		},
		ManagedResponseHeaders: []ManagedHeader{
			{
				ID:      "remove_x-powered-by_header",
				Enabled: false,
			},
		},
	}
	want := ManagedHeaders{
		ManagedRequestHeaders: []ManagedHeader{
			{
				ID:            "add_true_client_ip_headers",
				Enabled:       true,
				HasCoflict:    false,
				ConflictsWith: []string{"remove_visitor_ip_headers"},
			},
			{
				ID:         "add_visitor_location_headers",
				Enabled:    true,
				HasCoflict: false,
			},
		},
		ManagedResponseHeaders: []ManagedHeader{
			{
				ID:         "add_security_headers",
				Enabled:    false,
				HasCoflict: false,
			},
			{
				ID:         "remove_x-powered-by_header",
				Enabled:    false,
				HasCoflict: false,
			},
		},
	}
	zoneActual, err := client.UpdateZoneManagedHeaders(context.Background(), ZoneIdentifier(testZoneID), UpdateManagedHeadersParams{
		ManagedHeaders: managedHeadersForUpdate,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}
}
