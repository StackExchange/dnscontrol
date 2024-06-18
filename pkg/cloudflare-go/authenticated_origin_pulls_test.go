package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticatedOriginPullsDetails(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": {
		    "id": "tls_client_auth",
		    "value": "on",
		    "editable": true,
		    "modified_on": "2014-01-01T05:20:00.12345Z"
		  }
		}`)
	}
	mux.HandleFunc("/zones/023e105f4ecef8ad9ca31a8372d0c353/settings/tls_client_auth", handler)

	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	want := AuthenticatedOriginPulls{
		ID:         "tls_client_auth",
		Value:      "on",
		Editable:   true,
		ModifiedOn: modifiedOn,
	}

	actual, err := client.GetAuthenticatedOriginPullsStatus(context.Background(), "023e105f4ecef8ad9ca31a8372d0c353")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestSetAuthenticatedOriginPullsStatusEnabled(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
			  "id": "tls_client_auth",
			  "value": "on",
			  "editable": true,
			  "modified_on": "2014-01-01T05:20:00.12345Z"
			}
		  }`)
	}
	mux.HandleFunc("/zones/023e105f4ecef8ad9ca31a8372d0c353/settings/tls_client_auth", handler)
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	want := AuthenticatedOriginPulls{
		ID:         "tls_client_auth",
		Value:      "on",
		Editable:   true,
		ModifiedOn: modifiedOn,
	}

	actual, err := client.SetAuthenticatedOriginPullsStatus(context.Background(), "023e105f4ecef8ad9ca31a8372d0c353", true)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestSetAuthenticatedOriginPullsStatusDisabled(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
			  "id": "tls_client_auth",
			  "value": "off",
			  "editable": true,
			  "modified_on": "2014-01-01T05:20:00.12345Z"
			}
		  }`)
	}
	mux.HandleFunc("/zones/023e105f4ecef8ad9ca31a8372d0c353/settings/tls_client_auth", handler)
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	want := AuthenticatedOriginPulls{
		ID:         "tls_client_auth",
		Value:      "off",
		Editable:   true,
		ModifiedOn: modifiedOn,
	}

	actual, err := client.SetAuthenticatedOriginPullsStatus(context.Background(), "023e105f4ecef8ad9ca31a8372d0c353", false)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
