package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTeamsLists(t *testing.T) {
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
					"name": "My Serial List",
					"description": "My Description",
					"type": "SERIAL",
					"count": 1,
					"created_at": "2014-01-01T05:20:00.12345Z",
					"updated_at": "2014-01-01T05:20:00.12345Z"
				}
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1,
				"total_pages": 1
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := []TeamsList{{
		ID:          "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:        "My Serial List",
		Description: "My Description",
		Type:        "SERIAL",
		Count:       1,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/lists", handler)

	actual, _, err := client.ListTeamsLists(context.Background(), AccountIdentifier(testAccountID), ListTeamListsParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTeamsList(t *testing.T) {
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
				"name": "My Serial List",
				"description": "My Description",
				"type": "SERIAL",
				"count": 1,
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z"
			},
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1,
				"total_pages": 1
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := TeamsList{
		ID:          "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:        "My Serial List",
		Description: "My Description",
		Type:        "SERIAL",
		Count:       1,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/lists/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err := client.GetTeamsList(context.Background(), AccountIdentifier(testAccountID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTeamsListItems(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodGet, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"value": "val1",
					"created_at": "2014-01-01T05:20:00.12345Z"
				},
				{
					"value": "val2",
					"created_at": "2014-01-01T05:20:00.12345Z"
				}
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1,
				"total_pages": 1
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := []TeamsListItem{
		{
			Value:     "val1",
			CreatedAt: &createdAt,
		},
		{
			Value:     "val2",
			CreatedAt: &createdAt,
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/lists/480f4f69-1a28-4fdd-9240-1ed29f0ac1db/items", handler)

	actual, _, err := client.ListTeamsListItems(context.Background(), AccountIdentifier(testAccountID), ListTeamsListItemsParams{ListID: "480f4f69-1a28-4fdd-9240-1ed29f0ac1db"})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateTeamsList(t *testing.T) {
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
				"name": "My Serial List",
				"description": "My Description",
				"type": "SERIAL",
				"count": 1,
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z"
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	teamsList := TeamsList{
		ID:          "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:        "My Serial List",
		Description: "My Description",
		Type:        "SERIAL",
		Count:       1,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/lists", handler)

	actual, err := client.CreateTeamsList(context.Background(), AccountIdentifier(testAccountID), CreateTeamsListParams{
		Name:        "My Serial List",
		Description: "My Description",
		Type:        "SERIAL",
		Count:       1,
	})

	if assert.NoError(t, err) {
		assert.Equal(t, teamsList, actual)
	}
}

func TestUpdateTeamsList(t *testing.T) {
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
				"name": "My Serial List",
				"description": "My Updated Description",
				"type": "SERIAL",
				"count": 1,
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z"
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	teamsList := TeamsList{
		ID:          "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:        "My Serial List",
		Description: "My Updated Description",
		Type:        "SERIAL",
		Count:       1,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/lists/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err := client.UpdateTeamsList(context.Background(), AccountIdentifier(testAccountID), UpdateTeamsListParams{
		ID:          "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:        "My Serial List",
		Description: "My Updated Description",
		Type:        "SERIAL",
		Count:       1,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	})

	if assert.NoError(t, err) {
		assert.Equal(t, teamsList, actual)
	}
}

func TestUpdateTeamsListWithMissingID(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateTeamsList(context.Background(), AccountIdentifier(testAccountID), UpdateTeamsListParams{})
	assert.EqualError(t, err, "teams list ID cannot be empty")
}

func TestPatchTeamsList(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"name": "My Serial List",
				"description": "My Updated Description",
				"type": "SERIAL",
				"count": 1,
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z"
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	teamsList := TeamsList{
		ID:          "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:        "My Serial List",
		Description: "My Updated Description",
		Type:        "SERIAL",
		Count:       1,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/lists/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err := client.PatchTeamsList(context.Background(), AccountIdentifier(testAccountID), PatchTeamsListParams{
		ID:     "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Append: []TeamsListItem{{Value: "abcd-1234"}},
		Remove: []string{"def-5678"},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, teamsList, actual)
	}
}

func TestDeleteTeamsList(t *testing.T) {
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

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/lists/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)
	err := client.DeleteTeamsList(context.Background(), AccountIdentifier(testAccountID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	assert.NoError(t, err)
}
