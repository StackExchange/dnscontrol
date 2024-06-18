package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccessBookmarks(t *testing.T) {
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
					"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
					"name": "Example Site",
					"domain": "example.com",
					"logo_url": "https://www.example.com/example.png",
					"app_launcher_visible": true,
					"created_at": "2014-01-01T05:20:00.12345Z",
					"updated_at": "2014-01-01T05:20:00.12345Z"
				}
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := []AccessBookmark{{
		ID:                 "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:               "Example Site",
		Domain:             "example.com",
		LogoURL:            "https://www.example.com/example.png",
		AppLauncherVisible: BoolPtr(true),
		CreatedAt:          &createdAt,
		UpdatedAt:          &updatedAt,
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/bookmarks", handler)

	actual, _, err := client.AccessBookmarks(context.Background(), testAccountID, PaginationOptions{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/bookmarks", handler)

	actual, _, err = client.ZoneLevelAccessBookmarks(context.Background(), testZoneID, PaginationOptions{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAccessBookmark(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"name": "Example Site",
				"domain": "example.com",
				"logo_url": "https://www.example.com/example.png",
				"app_launcher_visible": true,
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z"
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := AccessBookmark{
		ID:                 "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:               "Example Site",
		Domain:             "example.com",
		LogoURL:            "https://www.example.com/example.png",
		AppLauncherVisible: BoolPtr(true),
		CreatedAt:          &createdAt,
		UpdatedAt:          &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/bookmarks/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err := client.AccessBookmark(context.Background(), testAccountID, "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/bookmarks/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err = client.ZoneLevelAccessBookmark(context.Background(), testZoneID, "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateAccessBookmarks(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"name": "Example Site",
				"domain": "example.com",
				"logo_url": "https://www.example.com/example.png",
				"app_launcher_visible": true,
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z"
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	fullAccessBookmark := AccessBookmark{
		ID:                 "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:               "Example Site",
		Domain:             "example.com",
		LogoURL:            "https://www.example.com/example.png",
		AppLauncherVisible: BoolPtr(true),
		CreatedAt:          &createdAt,
		UpdatedAt:          &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/bookmarks", handler)

	actual, err := client.CreateAccessBookmark(context.Background(), testAccountID, AccessBookmark{
		Name:               "Example Site",
		Domain:             "example.com",
		LogoURL:            "https://www.example.com/example.png",
		AppLauncherVisible: BoolPtr(true),
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessBookmark, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/bookmarks", handler)

	actual, err = client.CreateZoneLevelAccessBookmark(context.Background(), testZoneID, AccessBookmark{
		Name:               "Example Site",
		Domain:             "example.com",
		LogoURL:            "https://www.example.com/example.png",
		AppLauncherVisible: BoolPtr(true),
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessBookmark, actual)
	}
}

func TestUpdateAccessBookmark(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"name": "Example Site",
				"domain": "example.com",
				"logo_url": "https://www.example.com/example.png",
				"app_launcher_visible": true,
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z"
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	fullAccessBookmark := AccessBookmark{
		ID:                 "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:               "Example Site",
		Domain:             "example.com",
		LogoURL:            "https://www.example.com/example.png",
		AppLauncherVisible: BoolPtr(true),
		CreatedAt:          &createdAt,
		UpdatedAt:          &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/bookmarks/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err := client.UpdateAccessBookmark(context.Background(), testAccountID, fullAccessBookmark)

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessBookmark, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/bookmarks/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err = client.UpdateZoneLevelAccessBookmark(context.Background(), testZoneID, fullAccessBookmark)

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessBookmark, actual)
	}
}

func TestUpdateAccessBookmarkWithMissingID(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateAccessBookmark(context.Background(), testZoneID, AccessBookmark{})
	assert.EqualError(t, err, "access bookmark ID cannot be empty")

	_, err = client.UpdateZoneLevelAccessBookmark(context.Background(), testZoneID, AccessBookmark{})
	assert.EqualError(t, err, "access bookmark ID cannot be empty")
}

func TestDeleteAccessBookmark(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db"
      }
    }
    `)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/bookmarks/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)
	err := client.DeleteAccessBookmark(context.Background(), testAccountID, "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	assert.NoError(t, err)

	mux.HandleFunc("/zones/"+testZoneID+"/access/bookmarks/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)
	err = client.DeleteZoneLevelAccessBookmark(context.Background(), testZoneID, "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	assert.NoError(t, err)
}
