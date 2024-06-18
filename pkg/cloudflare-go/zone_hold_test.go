package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateZoneHold(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/hold", testZoneID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("include_subdomains"))
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		fmt.Fprint(w, `
	{
	  "success": true,
	  "errors": [],
	  "messages": [],
	  "result": {
		"hold": true,
		"hold_after": "2023-03-20T15:12:32Z",
		"include_subdomains": true
	  }
	}`)
	})

	_, err := client.CreateZoneHold(context.Background(), AccountIdentifier(""), CreateZoneHoldParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrRequiredZoneLevelResourceContainer, err)
	}

	out, err := client.CreateZoneHold(context.Background(), ZoneIdentifier(testZoneID), CreateZoneHoldParams{IncludeSubdomains: BoolPtr(true)})
	holdAfter, _ := time.Parse(time.RFC3339Nano, "2023-03-20T15:12:32Z")
	want := ZoneHold{
		IncludeSubdomains: BoolPtr(true),
		Hold:              BoolPtr(true),
		HoldAfter:         &holdAfter,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, out)
	}
}

func TestDeleteZoneHold(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/hold", testZoneID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2023-03-20T15:12:32Z", r.URL.Query().Get("hold_after"))
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		fmt.Fprint(w, `
	{
	  "success": true,
	  "errors": [],
	  "messages": [],
	  "result": {
		"hold": false,
		"hold_after": "2023-03-20T15:12:32.025799Z",
		"include_subdomains": false
	  }
	}`)
	})

	_, err := client.DeleteZoneHold(context.Background(), AccountIdentifier(""), DeleteZoneHoldParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrRequiredZoneLevelResourceContainer, err)
	}

	holdAfter, _ := time.Parse(time.RFC3339, "2023-03-20T15:12:32.025799Z")
	out, err := client.DeleteZoneHold(context.Background(), ZoneIdentifier(testZoneID), DeleteZoneHoldParams{HoldAfter: &holdAfter})
	want := ZoneHold{
		IncludeSubdomains: BoolPtr(false),
		Hold:              BoolPtr(false),
		HoldAfter:         &holdAfter,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, out)
	}
}

func TestGetZoneHold(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/hold", testZoneID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		fmt.Fprint(w, `
	{
	  "success": true,
	  "errors": [],
	  "messages": [],
	  "result": {
		"hold": false,
		"hold_after": "2023-03-20T15:12:32Z",
		"include_subdomains": false
	  }
	}`)
	})

	_, err := client.GetZoneHold(context.Background(), AccountIdentifier(""), GetZoneHoldParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrRequiredZoneLevelResourceContainer, err)
	}

	out, err := client.GetZoneHold(context.Background(), ZoneIdentifier(testZoneID), GetZoneHoldParams{})
	holdAfter, _ := time.Parse(time.RFC3339Nano, "2023-03-20T15:12:32Z")
	want := ZoneHold{
		IncludeSubdomains: BoolPtr(false),
		Hold:              BoolPtr(false),
		HoldAfter:         &holdAfter,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, out)
	}
}
