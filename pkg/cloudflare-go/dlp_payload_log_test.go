package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDLPPayloadLogSettings(t *testing.T) {
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
				"public_key": "3NP5MGKjzBLLceVxNZrF+LyithbWX+AVFBMRAA0Xl2A=",
				"updated_at": "2022-12-22T21:02:39Z"
			}
		}`)
	}

	updatedAt, _ := time.Parse(time.RFC3339, "2022-12-22T21:02:39Z")

	want := DLPPayloadLogSettings{
		PublicKey: "3NP5MGKjzBLLceVxNZrF+LyithbWX+AVFBMRAA0Xl2A=",
		UpdatedAt: &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/payload_log", handler)

	actual, err := client.GetDLPPayloadLogSettings(context.Background(), AccountIdentifier(testAccountID), GetDLPPayloadLogSettingsParams{})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestPutDLPPayloadLogSettings(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		var requestSettings DLPPayloadLogSettings
		err := json.NewDecoder(r.Body).Decode(&requestSettings)
		require.Nil(t, err)

		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"public_key": "`+requestSettings.PublicKey+`",
				"updated_at": "2022-12-22T21:02:39Z"
			}
		}`)
	}

	updatedAt, _ := time.Parse(time.RFC3339, "2022-12-22T21:02:39Z")

	want := DLPPayloadLogSettings{
		PublicKey: "3NP5MGKjzBLLceVxNZrF+LyithbWX+AVFBMRAA0Xl2A=",
		UpdatedAt: &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/payload_log", handler)

	actual, err := client.UpdateDLPPayloadLogSettings(context.Background(), AccountIdentifier(testAccountID), DLPPayloadLogSettings{
		PublicKey: "3NP5MGKjzBLLceVxNZrF+LyithbWX+AVFBMRAA0Xl2A=",
	})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}
