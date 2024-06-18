package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccessAuditLogs(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "user_email": "michelle@example.com",
      "ip_address": "198.51.100.1",
      "app_uid": "df7e2w5f-02b7-4d9d-af26-8d1988fca630",
      "app_domain": "test.example.com/admin",
      "action": "login",
      "connection": "saml",
      "allowed": false,
      "created_at": "2014-01-01T05:20:00.12345Z",
      "ray_id": "187d944c61940c77"
    }
  ]
}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/access/logs/access-requests", handler)
	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := []AccessAuditLogRecord{{
		UserEmail:  "michelle@example.com",
		IPAddress:  "198.51.100.1",
		AppUID:     "df7e2w5f-02b7-4d9d-af26-8d1988fca630",
		AppDomain:  "test.example.com/admin",
		Action:     "login",
		Connection: "saml",
		Allowed:    false,
		CreatedAt:  &createdAt,
		RayID:      "187d944c61940c77",
	}}

	actual, err := client.AccessAuditLogs(context.Background(), "01a7362d577a6c3019a474fd6f485823", AccessAuditLogFilterOptions{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAccessAuditLogsEncodeAllParametersDefined(t *testing.T) {
	since, _ := time.Parse(time.RFC3339, "2020-07-01T00:00:00Z")
	until, _ := time.Parse(time.RFC3339, "2020-07-02T00:00:00Z")

	opts := AccessAuditLogFilterOptions{
		Direction: "desc",
		Since:     &since,
		Until:     &until,
		Limit:     10,
	}

	assert.Equal(t, "direction=desc&limit=10&since=2020-07-01T00%3A00%3A00Z&until=2020-07-02T00%3A00%3A00Z", opts.Encode())
}

func TestAccessAuditLogsEncodeOnlyDatesDefined(t *testing.T) {
	since, _ := time.Parse(time.RFC3339, "2020-07-01T00:00:00Z")
	until, _ := time.Parse(time.RFC3339, "2020-07-02T00:00:00Z")

	opts := AccessAuditLogFilterOptions{
		Since: &since,
		Until: &until,
	}

	assert.Equal(t, "since=2020-07-01T00%3A00%3A00Z&until=2020-07-02T00%3A00%3A00Z", opts.Encode())
}

func TestAccessAuditLogsEncodeEmptyValues(t *testing.T) {
	opts := AccessAuditLogFilterOptions{}

	assert.Equal(t, "", opts.Encode())
}
