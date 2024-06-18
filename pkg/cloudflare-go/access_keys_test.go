package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

func TestAccessKeysConfig(t *testing.T) {
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
				"key_rotation_interval_days": 42,
				"last_key_rotation_at": "2014-01-01T05:20:00.12345Z",
				"days_until_next_rotation": 3
			}
		}`)
	}

	wantLastRotationAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	want := AccessKeysConfig{
		KeyRotationIntervalDays: 42,
		LastKeyRotationAt:       wantLastRotationAt,
		DaysUntilNextRotation:   3,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/keys", handler)

	actual, err := client.AccessKeysConfig(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateAccessKeysConfig(t *testing.T) {
	setup()
	defer teardown()

	wantRequest := AccessKeysConfigUpdateRequest{
		KeyRotationIntervalDays: 33,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		var gotRequest AccessKeysConfigUpdateRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&gotRequest))
		assert.Equal(t, wantRequest, gotRequest)

		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"key_rotation_interval_days": 42,
				"last_key_rotation_at": "2014-01-01T05:20:00.12345Z",
				"days_until_next_rotation": 3
			}
		}`)
	}

	wantLastRotationAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	want := AccessKeysConfig{
		KeyRotationIntervalDays: 42,
		LastKeyRotationAt:       wantLastRotationAt,
		DaysUntilNextRotation:   3,
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/access/keys", handler)

	actual, err := client.UpdateAccessKeysConfig(context.Background(), testAccountID, wantRequest)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestRotateAccessKeys(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"key_rotation_interval_days": 42,
				"last_key_rotation_at": "2014-01-01T05:20:00.12345Z",
				"days_until_next_rotation": 3
			}
		}`)
	}

	wantLastRotationAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	want := AccessKeysConfig{
		KeyRotationIntervalDays: 42,
		LastKeyRotationAt:       wantLastRotationAt,
		DaysUntilNextRotation:   3,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/keys/rotate", handler)

	actual, err := client.RotateAccessKeys(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
