package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTotalTLS_GetSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/acm/total_tls", testZoneID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
	  "success": true,
	  "errors": [],
	  "messages": [],
	  "result": {
		"enabled": true,
		"certificate_authority": "google",
		"validity_days": 90
	  }
	}`)
	})

	_, err := client.GetTotalTLS(context.Background(), ZoneIdentifier(""))
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}

	result, err := client.GetTotalTLS(context.Background(), ZoneIdentifier(testZoneID))
	if assert.NoError(t, err) {
		assert.Equal(t, BoolPtr(true), result.Enabled)
		assert.Equal(t, "google", result.CertificateAuthority)
		assert.Equal(t, 90, result.ValidityDays)
	}
}

func TestTotalTLS_SetSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/acm/total_tls", testZoneID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "enabled": true,
    "certificate_authority": "google",
    "validity_days": 90
  }
}`)
	})

	_, err := client.SetTotalTLS(context.Background(), ZoneIdentifier(""), TotalTLS{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}

	result, err := client.SetTotalTLS(context.Background(), ZoneIdentifier(testZoneID), TotalTLS{CertificateAuthority: "google", Enabled: BoolPtr(true)})
	if assert.NoError(t, err) {
		assert.Equal(t, BoolPtr(true), result.Enabled)
		assert.Equal(t, "google", result.CertificateAuthority)
		assert.Equal(t, 90, result.ValidityDays)
	}
}
