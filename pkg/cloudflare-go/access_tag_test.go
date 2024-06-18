package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccessTags(t *testing.T) {
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
					"name": "engineers",
					"app_count": 0
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

	want := []AccessTag{
		{
			Name:     "engineers",
			AppCount: 0,
		},
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/access/tags", handler)
	actual, err := client.ListAccessTags(context.Background(), AccountIdentifier(testAccountID), ListAccessTagsParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAccessTag(t *testing.T) {
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
				"name": "engineers",
				"app_count": 0
			}
		}`)
	}

	want := AccessTag{
		Name:     "engineers",
		AppCount: 0,
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/access/tags/engineers", handler)
	actual, err := client.GetAccessTag(context.Background(), AccountIdentifier(testAccountID), "engineers")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateAccessTag(t *testing.T) {
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
				"name": "sales",
				"app_count": 0
			}
		}`)
	}

	Tag := AccessTag{
		Name:     "sales",
		AppCount: 0,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/tags", handler)
	actual, err := client.CreateAccessTag(context.Background(), AccountIdentifier(testAccountID), CreateAccessTagParams{
		Name: "sales",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, Tag, actual)
	}
}

func TestDeleteAccessTag(t *testing.T) {
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
				"id": "engineers"
			}
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/tags/engineers", handler)
	err := client.DeleteAccessTag(context.Background(), AccountIdentifier(testAccountID), "engineers")

	assert.NoError(t, err)
}
