package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var expectedAccountRole = AccountRole{
	ID:          "3536bcfad5faccb999b47003c79917fb",
	Name:        "Account Administrator",
	Description: "Administrative access to the entire Account",
	Permissions: map[string]AccountRolePermission{
		"dns_records": {Read: true, Edit: true},
		"lb":          {Read: true, Edit: false},
	},
}

func TestAccountRoles(t *testing.T) {
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
					"id": "3536bcfad5faccb999b47003c79917fb",
					"name": "Account Administrator",
					"description": "Administrative access to the entire Account",
					"permissions": {
						"dns_records": {
							"read": true,
							"edit": true
						},
						"lb": {
							"read": true,
							"edit": false
						}
					}
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

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/roles", testAccountID), handler)
	want := []AccountRole{expectedAccountRole}
	_, err := client.ListAccountRoles(context.Background(), AccountIdentifier(""), ListAccountRolesParams{})
	if assert.Error(t, err, "Expected error when no account ID is provided") {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	actual, err := client.ListAccountRoles(context.Background(), testAccountRC, ListAccountRolesParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAccountRole(t *testing.T) {
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
				"id": "3536bcfad5faccb999b47003c79917fb",
				"name": "Account Administrator",
				"description": "Administrative access to the entire Account",
				"permissions": {
					"dns_records": {
						"read": true,
						"edit": true
					},
					"lb": {
						"read": true,
						"edit": false
					}
				}
			},
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1
			}
		}
		`)
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/roles/3536bcfad5faccb999b47003c79917fb", testAccountID), handler)
	_, err := client.GetAccountRole(context.Background(), AccountIdentifier(""), "3536bcfad5faccb999b47003c79917fb")
	if assert.Error(t, err, "Expected error when no account ID is provided") {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	actual, err := client.GetAccountRole(context.Background(), testAccountRC, "3536bcfad5faccb999b47003c79917fb")

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccountRole, actual)
	}
}
