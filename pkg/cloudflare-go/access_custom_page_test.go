package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccessCustomPages(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodGet, "HTTP method")
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"name": "Forbidden",
					"app_count": 0,
					"type": "forbidden",
					"uid": "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc"
				}
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1
			}
		}`)
	}

	want := []AccessCustomPage{
		{
			Name:     "Forbidden",
			AppCount: 0,
			Type:     Forbidden,
			UID:      "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc",
		},
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/access/custom_pages", handler)
	actual, err := client.ListAccessCustomPages(context.Background(), AccountIdentifier(testAccountID), ListAccessCustomPagesParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAccessCustomPage(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodGet, "HTTP method")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"custom_html": "<html><body><h1>Forbidden</h1></body></html>",
				"name": "Forbidden",
				"app_count": 0,
				"type": "forbidden",
				"uid": "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc"
			}
		}`)
	}

	want := AccessCustomPage{
		Name:       "Forbidden",
		AppCount:   0,
		Type:       Forbidden,
		UID:        "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc",
		CustomHTML: "<html><body><h1>Forbidden</h1></body></html>",
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/access/custom_pages/480f4f69-1a28-4fdd-9240-1ed29f0ac1dc", handler)
	actual, err := client.GetAccessCustomPage(context.Background(), AccountIdentifier(testAccountID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateAccessCustomPage(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodPost, "HTTP method")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"custom_html": "<html><body><h1>Forbidden</h1></body></html>",
				"name": "Forbidden",
				"app_count": 0,
				"type": "forbidden",
				"uid": "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc"
			}
		}`)
	}

	customPage := AccessCustomPage{
		Name:       "Forbidden",
		AppCount:   0,
		Type:       Forbidden,
		UID:        "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc",
		CustomHTML: "<html><body><h1>Forbidden</h1></body></html>",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/custom_pages", handler)
	actual, err := client.CreateAccessCustomPage(context.Background(), AccountIdentifier(testAccountID), CreateAccessCustomPageParams{
		Name:       "Forbidden",
		Type:       Forbidden,
		CustomHTML: "<html><body><h1>Forbidden</h1></body></html>",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, customPage, actual)
	}
}

func TestUpdateAccessCustomPage(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodPut, "HTTP method")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"custom_html": "<html><body><h1>Forbidden</h1></body></html>",
				"name": "Forbidden",
				"app_count": 0,
				"type": "forbidden",
				"uid": "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc"
			}
		}`)
	}

	customPage := AccessCustomPage{
		Name:       "Forbidden",
		AppCount:   0,
		Type:       Forbidden,
		UID:        "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc",
		CustomHTML: "<html><body><h1>Forbidden</h1></body></html>",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/custom_pages/480f4f69-1a28-4fdd-9240-1ed29f0ac1dc", handler)
	actual, err := client.UpdateAccessCustomPage(context.Background(), AccountIdentifier(testAccountID), UpdateAccessCustomPageParams{
		UID:        "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc",
		Name:       "Forbidden",
		Type:       Forbidden,
		CustomHTML: "<html><body><h1>Forbidden</h1></body></html>",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, customPage, actual)
	}
}

func TestDeleteAccessCustomPage(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodDelete, "HTTP method")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			result: {
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc"
			}
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/custom_pages/480f4f69-1a28-4fdd-9240-1ed29f0ac1dc", handler)
	err := client.DeleteAccessCustomPage(context.Background(), AccountIdentifier(testAccountID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1dc")

	assert.NoError(t, err)
}
