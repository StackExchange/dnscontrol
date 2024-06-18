package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogpullRetentionFlag(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		"errors": [],
		"messages": [],
		"result": {
			"flag": true
		},
		"success": true
	}`)
	}

	mux.HandleFunc("/zones/d56084adb405e0b7e32c52321bf07be6/logs/control/retention/flag", handler)
	want := &LogpullRetentionConfiguration{Flag: true}

	actual, err := client.GetLogpullRetentionFlag(context.Background(), "d56084adb405e0b7e32c52321bf07be6")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestSetLogpullRetentionFlag(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		"errors": [],
		"messages": [],
		"result": {
			"flag": false
		},
		"success": true
	}`)
	}

	mux.HandleFunc("/zones/d56084adb405e0b7e32c52321bf07be6/logs/control/retention/flag", handler)
	want := &LogpullRetentionConfiguration{Flag: false}

	actual, err := client.SetLogpullRetentionFlag(context.Background(), "d56084adb405e0b7e32c52321bf07be6", false)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
