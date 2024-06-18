package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetAuditSSHSettings(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"public_key": "1pyl6I1tL7xfJuFYVzXlUW8uXXlpxegHXBzGCBKaSFA=",
				"seed_id": "f1f968a9-83e7-401a-8abc-e0efe128425c",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z"
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := AuditSSHSettings{
		PublicKey: "1pyl6I1tL7xfJuFYVzXlUW8uXXlpxegHXBzGCBKaSFA=",
		SeedUUID:  "f1f968a9-83e7-401a-8abc-e0efe128425c",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/audit_ssh_settings", handler)

	actual, _, err := client.GetAuditSSHSettings(context.Background(), AccountIdentifier(testAccountID), GetAuditSSHSettingsParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateAuditSSHSettings(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"public_key": "updated1tL7xfJuFYVzXlUW8uXXlpxegHXBzGCBKaSFA=",
				"seed_id": "f1f968a9-83e7-401a-8abc-e0efe128425c",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z"
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := AuditSSHSettings{
		PublicKey: "updated1tL7xfJuFYVzXlUW8uXXlpxegHXBzGCBKaSFA=",
		SeedUUID:  "f1f968a9-83e7-401a-8abc-e0efe128425c",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/audit_ssh_settings", handler)

	actual, err := client.UpdateAuditSSHSettings(context.Background(), AccountIdentifier(testAccountID), UpdateAuditSSHSettingsParams{PublicKey: "updated1tL7xfJuFYVzXlUW8uXXlpxegHXBzGCBKaSFA="})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
