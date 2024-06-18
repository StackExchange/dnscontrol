package cloudflare

import (
	context "context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWAFOverride(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "a27cece9ec0e4af39ae9c58e3326e2b6",
				"paused": false,
				"urls": [
					"foo.bar.com"
				],
				"priority": 0,
				"groups": {
					"ea8687e59929c1fd05ba97574ad43f77": "default"
				},
				"rules": {
					"100015": "disable"
				},
				"rewrite_action": {
					"default": "simulate"
				}
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/firewall/waf/overrides/a27cece9ec0e4af39ae9c58e3326e2b6", handler)
	want := WAFOverride{
		ID:            "a27cece9ec0e4af39ae9c58e3326e2b6",
		Paused:        false,
		URLs:          []string{"foo.bar.com"},
		Priority:      0,
		Groups:        map[string]string{"ea8687e59929c1fd05ba97574ad43f77": "default"},
		Rules:         map[string]string{"100015": "disable"},
		RewriteAction: map[string]string{"default": "simulate"},
	}

	actual, err := client.WAFOverride(context.Background(), "01a7362d577a6c3019a474fd6f485823", "a27cece9ec0e4af39ae9c58e3326e2b6")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListWAFOverrides(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#waf-overrides-list-uri-controlled-waf-configurations
		fmt.Fprintf(w, `{
			"result": [
				{
					"id": "a27cece9ec0e4af39ae9c58e3326e2b6",
					"paused": false,
					"urls": [
						"foo.bar.com"
					],
					"priority": 0,
					"groups": {
						"ea8687e59929c1fd05ba97574ad43f77": "default"
					},
					"rules": {
						"100015": "disable"
					},
					"rewrite_action": {
						"default": "simulate"
					}
				}
			],
			"success": true,
			"errors": [],
			"messages": [],
			"result_info": {
				"page": 1,
				"per_page": 25,
				"count": 1,
				"total_count": 1,
				"total_pages": 1
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/overrides", handler)

	want := []WAFOverride{
		{
			ID:            "a27cece9ec0e4af39ae9c58e3326e2b6",
			Paused:        false,
			URLs:          []string{"foo.bar.com"},
			Priority:      0,
			Groups:        map[string]string{"ea8687e59929c1fd05ba97574ad43f77": "default"},
			Rules:         map[string]string{"100015": "disable"},
			RewriteAction: map[string]string{"default": "simulate"},
		},
	}

	d, err := client.ListWAFOverrides(context.Background(), testZoneID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}
}

func TestCreateWAFOverride(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "cf5b1db4ac454d7bad98ebec4533db57",
				"paused": false,
				"urls": [
					"foo.bar.com"
				],
				"priority": 0,
				"groups": {
					"ea8687e59929c1fd05ba97574ad43f77": "default"
				},
				"rules": {
					"100015": "disable"
				},
				"rewrite_action": {
					"default": "simulate"
				}
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/firewall/waf/overrides", handler)

	want := WAFOverride{
		ID:            "cf5b1db4ac454d7bad98ebec4533db57",
		URLs:          []string{"foo.bar.com"},
		Rules:         map[string]string{"100015": "disable"},
		Groups:        map[string]string{"ea8687e59929c1fd05ba97574ad43f77": "default"},
		RewriteAction: map[string]string{"default": "simulate"},
	}

	actual, err := client.CreateWAFOverride(context.Background(), "01a7362d577a6c3019a474fd6f485823", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteWAFOverride(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "18a9b91a93364593a8f41bd53bb2c02d"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/firewall/waf/overrides/18a9b91a93364593a8f41bd53bb2c02d", handler)

	err := client.DeleteWAFOverride(context.Background(), "01a7362d577a6c3019a474fd6f485823", "18a9b91a93364593a8f41bd53bb2c02d")
	assert.NoError(t, err)
}

func TestUpdateWAFOverride(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "e160a4fca2b346a7a418f49da049c566",
				"paused": false,
				"urls": [
					"foo.bar.com"
				],
				"priority": 0,
				"groups": {
					"ea8687e59929c1fd05ba97574ad43f77": "block"
				},
				"rules": {
					"100015": "disable"
				},
				"rewrite_action": {
					"default": "block"
				}
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/firewall/waf/overrides/e160a4fca2b346a7a418f49da049c566", handler)

	want := WAFOverride{
		ID:            "e160a4fca2b346a7a418f49da049c566",
		URLs:          []string{"foo.bar.com"},
		Rules:         map[string]string{"100015": "disable"},
		Groups:        map[string]string{"ea8687e59929c1fd05ba97574ad43f77": "block"},
		RewriteAction: map[string]string{"default": "block"},
	}

	actual, err := client.UpdateWAFOverride(context.Background(), "01a7362d577a6c3019a474fd6f485823", "e160a4fca2b346a7a418f49da049c566", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
