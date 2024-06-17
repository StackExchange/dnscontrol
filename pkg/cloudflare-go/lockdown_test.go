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

func TestCreateZoneLockdown(t *testing.T) {
	setup()
	defer teardown()

	input := ZoneLockdownCreateParams{
		Description: "Restrict access to these endpoints to requests from a known IP address",
		URLs:        []string{"api.mysite.com/some/endpoint*"},
		Configurations: []ZoneLockdownConfig{{
			Target: "ip",
			Value:  "198.51.100.4",
		}},
		Paused: false,
	}

	want := ZoneLockdown{
		Description: "Restrict access to these endpoints to requests from a known IP address",
		URLs:        []string{"api.mysite.com/some/endpoint*"},
		Configurations: []ZoneLockdownConfig{{
			Target: "ip",
			Value:  "198.51.100.4",
		}},
		Paused: false,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s")

		var v ZoneLockdown
		err := json.NewDecoder(r.Body).Decode(&v)
		require.NoError(t, err)
		assert.Equal(t, want, v)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "372e67954025e0ba6aaa6d586b9e0b59",
				"created_on": "2014-01-01T05:20:00Z",
				"modified_on": "2014-01-01T05:20:00Z",
				"paused": false,
				"description": "Restrict access to these endpoints to requests from a known IP address",
				"urls": [
					"api.mysite.com/some/endpoint*"
				],
				"configurations": [
				  {
					  "target": "ip",
					  "value": "198.51.100.4"
				  }
				]
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/lockdowns", handler)

	actual, err := client.CreateZoneLockdown(context.Background(), ZoneIdentifier(testZoneID), input)
	require.NoError(t, err)

	want.ID = "372e67954025e0ba6aaa6d586b9e0b59"
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want.CreatedOn = &createdOn
	want.ModifiedOn = &modifiedOn

	assert.Equal(t, want, actual)
}

func TestUpdateZoneLockdown(t *testing.T) {
	setup()
	defer teardown()

	input := ZoneLockdownUpdateParams{
		ID:          "372e67954025e0ba6aaa6d586b9e0b59",
		Description: "Restrict access to these endpoints to requests from a known IP address",
		URLs:        []string{"api.mysite.com/some/endpoint*"},
		Configurations: []ZoneLockdownConfig{{
			Target: "ip",
			Value:  "198.51.100.4",
		}},
		Paused: false,
	}

	want := ZoneLockdown{
		ID:          "372e67954025e0ba6aaa6d586b9e0b59",
		Description: "Restrict access to these endpoints to requests from a known IP address",
		URLs:        []string{"api.mysite.com/some/endpoint*"},
		Configurations: []ZoneLockdownConfig{{
			Target: "ip",
			Value:  "198.51.100.4",
		}},
		Paused: false,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s")

		var v ZoneLockdown
		err := json.NewDecoder(r.Body).Decode(&v)
		require.NoError(t, err)
		assert.Equal(t, want, v)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "372e67954025e0ba6aaa6d586b9e0b59",
				"created_on": "2014-01-01T05:20:00Z",
				"modified_on": "2014-01-01T05:20:00Z",
				"paused": false,
				"description": "Restrict access to these endpoints to requests from a known IP address",
				"urls": [
					"api.mysite.com/some/endpoint*"
				],
				"configurations": [
				  {
					  "target": "ip",
					  "value": "198.51.100.4"
				  }
				]
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/lockdowns/372e67954025e0ba6aaa6d586b9e0b59", handler)

	actual, err := client.UpdateZoneLockdown(context.Background(), ZoneIdentifier(testZoneID), input)
	require.NoError(t, err)

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want.CreatedOn = &createdOn
	want.ModifiedOn = &modifiedOn

	assert.Equal(t, want, actual)
}

func TestDeleteZoneLockdown(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s")

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "372e67954025e0ba6aaa6d586b9e0b59"
			}
		}`)
	}

	zoneLockdownID := "372e67954025e0ba6aaa6d586b9e0b59"

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/lockdowns/"+zoneLockdownID, handler)

	actual, err := client.DeleteZoneLockdown(context.Background(), ZoneIdentifier(testZoneID), zoneLockdownID)
	require.NoError(t, err)

	want := ZoneLockdown{
		ID: zoneLockdownID,
	}
	assert.Equal(t, want, actual)
}

func TestZoneLockdown(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s")

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "372e67954025e0ba6aaa6d586b9e0b59",
				"created_on": "2014-01-01T05:20:00Z",
				"modified_on": "2014-01-01T05:20:00Z",
				"paused": false,
				"description": "Restrict access to these endpoints to requests from a known IP address",
				"urls": [
					"api.mysite.com/some/endpoint*"
				],
				"configurations": [
				  {
					  "target": "ip",
					  "value": "198.51.100.4"
				  }
				]
			}
		}`)
	}

	zoneLockdownID := "372e67954025e0ba6aaa6d586b9e0b59"

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/lockdowns/"+zoneLockdownID, handler)

	actual, err := client.ZoneLockdown(context.Background(), ZoneIdentifier(testZoneID), zoneLockdownID)
	require.NoError(t, err)

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := ZoneLockdown{
		ID:          zoneLockdownID,
		Description: "Restrict access to these endpoints to requests from a known IP address",
		URLs:        []string{"api.mysite.com/some/endpoint*"},
		Configurations: []ZoneLockdownConfig{{
			Target: "ip",
			Value:  "198.51.100.4",
		}},
		Paused:     false,
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
	}
	assert.Equal(t, want, actual)
}

func TestListZoneLockdowns(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s")

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
			  {
				  "id": "372e67954025e0ba6aaa6d586b9e0b59",
				  "created_on": "2014-01-01T05:20:00Z",
				  "modified_on": "2014-01-01T05:20:00Z",
				  "paused": false,
				  "description": "Restrict access to these endpoints to requests from a known IP address",
				  "urls": [
					  "api.mysite.com/some/endpoint*"
				  ],
				  "configurations": [
					{
						"target": "ip",
						"value": "198.51.100.4"
					}
				  ]
			  }
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/lockdowns", handler)

	actual, _, err := client.ListZoneLockdowns(context.Background(), ZoneIdentifier(testZoneID), LockdownListParams{})
	require.NoError(t, err)

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := []ZoneLockdown{{
		ID:          "372e67954025e0ba6aaa6d586b9e0b59",
		Description: "Restrict access to these endpoints to requests from a known IP address",
		URLs:        []string{"api.mysite.com/some/endpoint*"},
		Configurations: []ZoneLockdownConfig{{
			Target: "ip",
			Value:  "198.51.100.4",
		}},
		Paused:     false,
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
	},
	}
	assert.Equal(t, want, actual)
}
