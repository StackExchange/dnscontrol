package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkersAccountSettings_CreateAccountSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/account-settings", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "default_usage_model": "unbound",
	"green_compute": false
  }
}`)
	})
	res, err := client.CreateWorkersAccountSettings(context.Background(), AccountIdentifier(testAccountID), CreateWorkersAccountSettingsParameters{DefaultUsageModel: "unbound", GreenCompute: false})
	want := WorkersAccountSettings{
		DefaultUsageModel: "unbound",
		GreenCompute:      false,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestWorkersAccountSettings_GetAccountSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/account-settings", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "default_usage_model": "unbound",
	"green_compute": true
  }
}`)
	})

	res, err := client.WorkersAccountSettings(context.Background(), AccountIdentifier(testAccountID), WorkersAccountSettingsParameters{})
	want := WorkersAccountSettings{
		DefaultUsageModel: "unbound",
		GreenCompute:      true,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}
