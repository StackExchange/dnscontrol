package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateKeylessSSL(t *testing.T) {
	setup()
	defer teardown()

	input := KeylessSSLCreateRequest{
		Host:         "example.com",
		Port:         24008,
		Certificate:  "-----BEGIN CERTIFICATE----- MIIDtTCCAp2g1v2tdw= -----END CERTIFICATE-----",
		Name:         "example.com Keyless SSL",
		BundleMethod: "optimal",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		var v KeylessSSLCreateRequest
		err := json.NewDecoder(r.Body).Decode(&v)
		require.NoError(t, err)
		assert.Equal(t, input, v)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "4d2844d2ce78891c34d0b6c0535a291e",
				"name": "example.com Keyless SSL",
				"host": "example.com",
				"port": 24008,
				"status": "active",
				"enabled": false,
				"permissions": [
					"#ssl:read",
					"#ssl:edit"
				],
				"created_on": "2014-01-01T05:20:00Z",
				"modified_on": "2014-01-01T05:20:00Z"
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/keyless_certificates", handler)

	actual, err := client.CreateKeylessSSL(context.Background(), testZoneID, input)
	require.NoError(t, err)

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := KeylessSSL{
		ID:          "4d2844d2ce78891c34d0b6c0535a291e",
		Name:        input.Name,
		Host:        input.Host,
		Port:        input.Port,
		Status:      "active",
		Enabled:     false,
		Permissions: []string{"#ssl:read", "#ssl:edit"},
		CreatedOn:   createdOn,
		ModifiedOn:  modifiedOn,
	}
	assert.Equal(t, want, actual)
}

func TestListKeylessSSL(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success":true,
			"errors":[],
			"messages":[],
			"result":[
			   {
				  "id":"4d2844d2ce78891c34d0b6c0535a291e",
				  "name":"example.com Keyless SSL",
				  "host":"example.com",
				  "port":24008,
				  "status":"active",
				  "enabled":false,
				  "permissions":[
					 "#ssl:read",
					 "#ssl:edit"
				  ],
				  "created_on":"2014-01-01T05:20:00Z",
				  "modified_on":"2014-01-01T05:20:00Z"
			   }
			]
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/keyless_certificates", handler)

	actual, err := client.ListKeylessSSL(context.Background(), testZoneID)
	require.NoError(t, err)

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := []KeylessSSL{{
		ID:          "4d2844d2ce78891c34d0b6c0535a291e",
		Name:        "example.com Keyless SSL",
		Host:        "example.com",
		Port:        24008,
		Status:      "active",
		Enabled:     false,
		Permissions: []string{"#ssl:read", "#ssl:edit"},
		CreatedOn:   createdOn,
		ModifiedOn:  modifiedOn,
	}}
	assert.Equal(t, want, actual)
}

func TestKeylessSSL(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success":true,
			"errors":[],
			"messages":[],
			"result":{
			   "id":"4d2844d2ce78891c34d0b6c0535a291e",
			   "name":"example.com Keyless SSL",
			   "host":"example.com",
			   "port":24008,
			   "status":"active",
			   "enabled":false,
			   "permissions":[
				  "#ssl:read",
				  "#ssl:edit"
			   ],
			   "created_on":"2014-01-01T05:20:00Z",
			   "modified_on":"2014-01-01T05:20:00Z"
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/keyless_certificates/"+"4d2844d2ce78891c34d0b6c0535a291e", handler)

	actual, err := client.KeylessSSL(context.Background(), testZoneID, "4d2844d2ce78891c34d0b6c0535a291e")
	require.NoError(t, err)

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := KeylessSSL{
		ID:          "4d2844d2ce78891c34d0b6c0535a291e",
		Name:        "example.com Keyless SSL",
		Host:        "example.com",
		Port:        24008,
		Status:      "active",
		Enabled:     false,
		Permissions: []string{"#ssl:read", "#ssl:edit"},
		CreatedOn:   createdOn,
		ModifiedOn:  modifiedOn,
	}
	assert.Equal(t, want, actual)
}

func TestUpdateKeylessSSL(t *testing.T) {
	setup()
	defer teardown()

	enabled := false
	input := KeylessSSLUpdateRequest{
		Host:    "example.com",
		Name:    "example.com Keyless SSL",
		Port:    24008,
		Enabled: &enabled,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		var v KeylessSSLUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&v)
		require.NoError(t, err)
		assert.Equal(t, input, v)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success":true,
			"errors":[],
			"messages":[],
			"result":{
			   "id":"4d2844d2ce78891c34d0b6c0535a291e",
			   "name":"example.com Keyless SSL",
			   "host":"example.com",
			   "port":24008,
			   "status":"active",
			   "enabled":false,
			   "permissions":[
				  "#ssl:read",
				  "#ssl:edit"
			   ],
			   "created_on":"2014-01-01T05:20:00Z",
			   "modified_on":"2014-01-01T05:20:00Z"
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/keyless_certificates/"+"4d2844d2ce78891c34d0b6c0535a291e", handler)

	actual, err := client.UpdateKeylessSSL(context.Background(), testZoneID, "4d2844d2ce78891c34d0b6c0535a291e", input)
	require.NoError(t, err)

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := KeylessSSL{
		ID:          "4d2844d2ce78891c34d0b6c0535a291e",
		Name:        input.Name,
		Host:        input.Host,
		Port:        input.Port,
		Status:      "active",
		Enabled:     enabled,
		Permissions: []string{"#ssl:read", "#ssl:edit"},
		CreatedOn:   createdOn,
		ModifiedOn:  modifiedOn,
	}
	assert.Equal(t, want, actual)
}

func TestDeleteKeylessSSL(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success":true,
			"errors":[],
			"messages":[],
			"result":{
			   "id":"4d2844d2ce78891c34d0b6c0535a291e"
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/keyless_certificates/"+"4d2844d2ce78891c34d0b6c0535a291e", handler)

	err := client.DeleteKeylessSSL(context.Background(), testZoneID, "4d2844d2ce78891c34d0b6c0535a291e")
	require.NoError(t, err)
}
