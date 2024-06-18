package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListMagicTransitStaticRoutes(t *testing.T) {
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
        "routes": [
          {
            "id": "c4a7362d577a6c3019a474fd6f485821",
            "created_on": "2017-06-14T00:00:00Z",
            "modified_on": "2017-06-14T05:20:00Z",
            "prefix": "192.0.2.0/24",
            "nexthop": "203.0.113.1",
            "priority": 100,
            "description": "New route for new prefix 203.0.113.1",
            "weight": 100,
            "scope": {
              "colo_regions": [
                "APAC"
              ],
              "colo_names": [
                "den01"
              ]
            }
          }
        ]
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/routes", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := []MagicTransitStaticRoute{
		{
			ID:          "c4a7362d577a6c3019a474fd6f485821",
			CreatedOn:   &createdOn,
			ModifiedOn:  &modifiedOn,
			Prefix:      "192.0.2.0/24",
			Nexthop:     "203.0.113.1",
			Priority:    100,
			Description: "New route for new prefix 203.0.113.1",
			Weight:      100,
			Scope: MagicTransitStaticRouteScope{
				ColoRegions: []string{
					"APAC",
				},
				ColoNames: []string{
					"den01",
				},
			},
		},
	}

	actual, err := client.ListMagicTransitStaticRoutes(context.Background(), testAccountID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetMagicTransitStaticRoute(t *testing.T) {
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
        "route": {
          "id": "c4a7362d577a6c3019a474fd6f485821",
          "created_on": "2017-06-14T00:00:00Z",
          "modified_on": "2017-06-14T05:20:00Z",
          "prefix": "192.0.2.0/24",
          "nexthop": "203.0.113.1",
          "priority": 100,
          "description": "New route for new prefix 203.0.113.1",
          "weight": 100,
          "scope": {
            "colo_regions": [
              "APAC"
            ],
            "colo_names": [
              "den01"
            ]
          }
        }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/routes/c4a7362d577a6c3019a474fd6f485821", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitStaticRoute{
		ID:          "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
		Prefix:      "192.0.2.0/24",
		Nexthop:     "203.0.113.1",
		Priority:    100,
		Description: "New route for new prefix 203.0.113.1",
		Weight:      100,
		Scope: MagicTransitStaticRouteScope{
			ColoRegions: []string{
				"APAC",
			},
			ColoNames: []string{
				"den01",
			},
		},
	}

	actual, err := client.GetMagicTransitStaticRoute(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateMagicTransitStaticRoutes(t *testing.T) {
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
        "routes": [
          {
            "id": "c4a7362d577a6c3019a474fd6f485821",
            "created_on": "2017-06-14T00:00:00Z",
            "modified_on": "2017-06-14T05:20:00Z",
            "prefix": "192.0.2.0/24",
            "nexthop": "203.0.113.1",
            "priority": 100,
            "description": "New route for new prefix 203.0.113.1",
            "weight": 100,
            "scope": {
              "colo_regions": [
                "APAC"
              ],
              "colo_names": [
                "den01"
              ]
            }
          }
        ]
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/routes", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitStaticRoute{
		ID:          "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
		Prefix:      "192.0.2.0/24",
		Nexthop:     "203.0.113.1",
		Priority:    100,
		Description: "New route for new prefix 203.0.113.1",
		Weight:      100,
		Scope: MagicTransitStaticRouteScope{
			ColoRegions: []string{
				"APAC",
			},
			ColoNames: []string{
				"den01",
			},
		},
	}

	actual, err := client.CreateMagicTransitStaticRoute(context.Background(), testAccountID, want)
	if assert.NoError(t, err) {
		assert.Equal(t, []MagicTransitStaticRoute{
			want,
		}, actual)
	}
}

func TestUpdateMagicTransitStaticRoute(t *testing.T) {
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
        "modified_route": {
          "id": "c4a7362d577a6c3019a474fd6f485821",
          "created_on": "2017-06-14T00:00:00Z",
          "modified_on": "2017-06-14T05:20:00Z",
          "prefix": "192.0.2.0/24",
          "nexthop": "203.0.113.1",
          "priority": 100,
          "description": "New route for new prefix 203.0.113.1",
          "weight": 100,
          "scope": {
            "colo_regions": [
              "APAC"
            ],
            "colo_names": [
              "den01"
            ]
          }
        }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/routes/c4a7362d577a6c3019a474fd6f485821", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitStaticRoute{
		ID:          "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
		Prefix:      "192.0.2.0/24",
		Nexthop:     "203.0.113.1",
		Priority:    100,
		Description: "New route for new prefix 203.0.113.1",
		Weight:      100,
		Scope: MagicTransitStaticRouteScope{
			ColoRegions: []string{
				"APAC",
			},
			ColoNames: []string{
				"den01",
			},
		},
	}

	actual, err := client.UpdateMagicTransitStaticRoute(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteMagicTransitStaticRoute(t *testing.T) {
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
        "deleted_route": {
          "id": "c4a7362d577a6c3019a474fd6f485821",
          "created_on": "2017-06-14T00:00:00Z",
          "modified_on": "2017-06-14T05:20:00Z",
          "prefix": "192.0.2.0/24",
          "nexthop": "203.0.113.1",
          "priority": 100,
          "description": "New route for new prefix 203.0.113.1",
          "weight": 100,
          "scope": {
            "colo_regions": [
              "APAC"
            ],
            "colo_names": [
              "den01"
            ]
          }
        }
      }
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/magic/routes/c4a7362d577a6c3019a474fd6f485821", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2017-06-14T00:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-06-14T05:20:00Z")

	want := MagicTransitStaticRoute{
		ID:          "c4a7362d577a6c3019a474fd6f485821",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
		Prefix:      "192.0.2.0/24",
		Nexthop:     "203.0.113.1",
		Priority:    100,
		Description: "New route for new prefix 203.0.113.1",
		Weight:      100,
		Scope: MagicTransitStaticRouteScope{
			ColoRegions: []string{
				"APAC",
			},
			ColoNames: []string{
				"den01",
			},
		},
	}

	actual, err := client.DeleteMagicTransitStaticRoute(context.Background(), testAccountID, "c4a7362d577a6c3019a474fd6f485821")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
