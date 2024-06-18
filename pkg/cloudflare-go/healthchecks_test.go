package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

const (
	healthcheckID       = "314d6b003029433741b94a7c9284915a"
	healthcheckResponse = `{
		"id": "%s",
		"name": "example-healthcheck",
		"description": "Example Healthcheck",
		"suspended": false,
		"address": "www.example.com",
		"retries": 2,
		"timeout": 5,
		"consecutive_successes": 2,
		"consecutive_fails": 2,
		"interval": 60,
		"type": "HTTP",
		"check_regions": [
			"WNAM"
		],
		"http_config": {
			"method": "GET",
			"path": "/",
			"port": 8443,
			"expected_body": "",
			"expected_codes": [
				"200"
			],
			"follow_redirects": true,
			"allow_insecure": false,
			"header": {
				"Host": [
					"www.example.com"
				]
			}
		},
		"tcp_config": null,
		"created_on": "2019-01-13T12:20:00.12345Z",
		"modified_on": "2019-01-13T12:20:00.12345Z",
		"status": "unknown",
		"failure_reason": ""
	}`
)

var (
	createdOn, _  = time.Parse(time.RFC3339, "2019-01-13T12:20:00.12345Z")
	modifiedOn, _ = time.Parse(time.RFC3339, "2019-01-13T12:20:00.12345Z")

	expectedHealthcheck = Healthcheck{
		ID:                   "314d6b003029433741b94a7c9284915a",
		CreatedOn:            &createdOn,
		ModifiedOn:           &modifiedOn,
		Description:          "Example Healthcheck",
		Name:                 "example-healthcheck",
		Suspended:            false,
		Address:              "www.example.com",
		Retries:              2,
		ConsecutiveSuccesses: 2,
		ConsecutiveFails:     2,
		Timeout:              5,
		Interval:             60,
		Type:                 "HTTP",
		CheckRegions:         []string{"WNAM"},
		HTTPConfig: &HealthcheckHTTPConfig{
			Method:          http.MethodGet,
			Path:            "/",
			Port:            8443,
			ExpectedBody:    "",
			ExpectedCodes:   []string{"200"},
			FollowRedirects: true,
			AllowInsecure:   false,
			Header: map[string][]string{
				"Host": {"www.example.com"},
			},
		},
		Status:        "unknown",
		FailureReason: "",
	}
)

func TestHealthchecks(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"result": [
			%s
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
		}
		`, fmt.Sprintf(healthcheckResponse, healthcheckID))
	}

	mux.HandleFunc("/zones/"+testZoneID+"/healthchecks", handler)
	want := []Healthcheck{expectedHealthcheck}

	actual, err := client.Healthchecks(context.Background(), testZoneID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestHealthcheck(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"result": %s,
			"success": true,
			"errors": [],
			"messages": []
		}
		`, fmt.Sprintf(healthcheckResponse, healthcheckID))
	}

	mux.HandleFunc("/zones/"+testZoneID+"/healthchecks/"+healthcheckID, handler)
	want := expectedHealthcheck

	actual, err := client.Healthcheck(context.Background(), testZoneID, healthcheckID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateHealthcheck(t *testing.T) {
	setup()
	defer teardown()
	newHealthcheck := Healthcheck{
		Name:      "example-healthcheck",
		Address:   "www.example.com",
		Suspended: false,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": %s,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`, fmt.Sprintf(healthcheckResponse, healthcheckID))
	}

	mux.HandleFunc("/zones/"+testZoneID+"/healthchecks", handler)
	want := expectedHealthcheck

	actual, err := client.CreateHealthcheck(context.Background(), testZoneID, newHealthcheck)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateHealthcheck(t *testing.T) {
	setup()
	defer teardown()
	updatedHealthcheck := Healthcheck{
		Name:    "example-healthcheck",
		Address: "www.example.com",
		HTTPConfig: &HealthcheckHTTPConfig{
			Path: "/newpath",
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": %s,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`, fmt.Sprintf(healthcheckResponse, healthcheckID))
	}

	mux.HandleFunc("/zones/"+testZoneID+"/healthchecks/"+healthcheckID, handler)
	want := expectedHealthcheck

	actual, err := client.UpdateHealthcheck(context.Background(), testZoneID, healthcheckID, updatedHealthcheck)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteHealthcheck(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
		  "result": null,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/healthchecks/"+healthcheckID, handler)

	err := client.DeleteHealthcheck(context.Background(), testZoneID, healthcheckID)
	assert.NoError(t, err)
}

func TestCreateHealthcheckPreview(t *testing.T) {
	setup()
	defer teardown()
	newHealthcheck := Healthcheck{
		Name:      "example-healthcheck",
		Address:   "www.example.com",
		Suspended: false,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": %s,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`, fmt.Sprintf(healthcheckResponse, healthcheckID))
	}

	mux.HandleFunc("/zones/"+testZoneID+"/healthchecks/preview", handler)
	want := expectedHealthcheck

	actual, err := client.CreateHealthcheckPreview(context.Background(), testZoneID, newHealthcheck)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestHealthcheckPreview(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"result": %s,
			"success": true,
			"errors": [],
			"messages": []
		}
		`, fmt.Sprintf(healthcheckResponse, healthcheckID))
	}

	mux.HandleFunc("/zones/"+testZoneID+"/healthchecks/preview/"+healthcheckID, handler)
	want := expectedHealthcheck

	actual, err := client.HealthcheckPreview(context.Background(), testZoneID, healthcheckID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteHealthcheckPreview(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
		  "result": null,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/healthchecks/preview/"+healthcheckID, handler)

	err := client.DeleteHealthcheckPreview(context.Background(), testZoneID, healthcheckID)
	assert.NoError(t, err)
}
