package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkersSubdomain_CreateSubdomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/subdomain", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "name": "example-subdomain"
  }
}`)
	})

	_, err := client.WorkersCreateSubdomain(context.Background(), AccountIdentifier(""), WorkersSubdomain{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	res, err := client.WorkersCreateSubdomain(context.Background(), AccountIdentifier(testAccountID), WorkersSubdomain{Name: "example-subdomain"})

	want := WorkersSubdomain{Name: "example-subdomain"}

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestWorkersSubdomain_GetSubdomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/subdomain", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "name": "example-subdomain"
  }
}`)
	})

	_, err := client.WorkersGetSubdomain(context.Background(), AccountIdentifier(""))
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	res, err := client.WorkersGetSubdomain(context.Background(), AccountIdentifier(testAccountID))

	want := WorkersSubdomain{Name: "example-subdomain"}

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}
