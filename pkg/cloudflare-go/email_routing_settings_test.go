package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTestEmailRoutingSettings() EmailRoutingSettings {
	created, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")
	modified, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")
	return EmailRoutingSettings{
		Tag:        "75610dab9e69410a82cf7e400a09ecec",
		Name:       "example.net",
		Enabled:    true,
		Created:    &created,
		Modified:   &modified,
		SkipWizard: BoolPtr(true),
		Status:     "read",
	}
}

func TestEmailRouting_GetSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/email/routing", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "tag": "75610dab9e69410a82cf7e400a09ecec",
    "name": "example.net",
    "enabled": true,
    "created": "2014-01-02T02:20:00Z",
    "modified": "2014-01-02T02:20:00Z",
    "skip_wizard": true,
    "status": "read"
  }
}`)
	})

	_, err := client.GetEmailRoutingSettings(context.Background(), AccountIdentifier(""))
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}

	want := createTestEmailRoutingSettings()

	res, err := client.GetEmailRoutingSettings(context.Background(), AccountIdentifier(testZoneID))
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestEmailRouting_Enable(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/email/routing/enable", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "tag": "75610dab9e69410a82cf7e400a09ecec",
    "name": "example.net",
    "enabled": true,
    "created": "2014-01-02T02:20:00Z",
    "modified": "2014-01-02T02:20:00Z",
    "skip_wizard": true,
    "status": "read"
  }
}`)
	})

	_, err := client.EnableEmailRouting(context.Background(), AccountIdentifier(""))
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}

	want := createTestEmailRoutingSettings()

	res, err := client.EnableEmailRouting(context.Background(), AccountIdentifier(testZoneID))
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestEmailRouting_Disabled(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/email/routing/disable", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "tag": "75610dab9e69410a82cf7e400a09ecec",
    "name": "example.net",
    "enabled": true,
    "created": "2014-01-02T02:20:00Z",
    "modified": "2014-01-02T02:20:00Z",
    "skip_wizard": true,
    "status": "read"
  }
}`)
	})

	_, err := client.DisableEmailRouting(context.Background(), AccountIdentifier(""))
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}

	want := createTestEmailRoutingSettings()

	res, err := client.DisableEmailRouting(context.Background(), AccountIdentifier(testZoneID))
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestEmailRouting_DNSSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/email/routing/dns", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "type": "A",
      "name": "example.com",
      "content": "192.0.2.1",
      "ttl": 3600,
      "priority": 10
    }
  ]
}`)
	})

	_, err := client.GetEmailRoutingDNSSettings(context.Background(), AccountIdentifier(""))
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}

	want := []DNSRecord{
		{
			Type:     "A",
			Name:     "example.com",
			Content:  "192.0.2.1",
			TTL:      3600,
			Priority: Uint16Ptr(10),
		},
	}

	res, err := client.GetEmailRoutingDNSSettings(context.Background(), AccountIdentifier(testZoneID))
	if assert.NoError(t, err) {
		assert.Len(t, res, 1)
		assert.Equal(t, want, res)
	}
}
