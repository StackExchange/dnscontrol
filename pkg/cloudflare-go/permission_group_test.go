package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockPermissionGroup = PermissionGroup{
	ID:   "f08020434ba14a0bb46bd9ff52f23b04",
	Name: "Fake Permission Group",
	Meta: map[string]string{
		"description": "Can represent a permission group",
	},
	Permissions: []Permission{{
		ID:  "7d42552322884d19bb63ed7f69b5ac21",
		Key: "com.cloudflare.api.account.fake.permission",
		Attributes: map[string]string{
			"legacy_name": "#fake:permission",
		},
	}},
}

func TestListPermissionGroup_ByName(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "result": [
				{
				  "id": "f08020434ba14a0bb46bd9ff52f23b04",
				  "name": "Fake Permission Group",
				  "status": "V",
				  "meta": {
					"description": "Can represent a permission group"
				  },
				  "created_on": "2022-08-29T16:58:29.745574Z",
				  "modified_on": "2022-09-12T08:26:37.255907Z",
				  "permissions": [
					{
					  "id": "7d42552322884d19bb63ed7f69b5ac21",
					  "key": "com.cloudflare.api.account.fake.permission",
					  "attributes": {
						"legacy_name": "#fake:permission"
					  }
					}
				  ]
				}
			  ],
			  "success": true,
			  "errors": [],
			  "messages": []
			}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/iam/permission_groups", handler)

	result, err := client.ListPermissionGroups(context.Background(), AccountIdentifier(testAccountID), ListPermissionGroupParams{RoleName: "fake-permission-group-name"})
	if assert.NoError(t, err) {
		assert.Equal(t, result, []PermissionGroup{mockPermissionGroup})
	}
}

func TestGetPermissionGroup(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"result": {
				"id": "f08020434ba14a0bb46bd9ff52f23b04",
				"name": "Fake Permission Group",
				"status": "V",
				"meta": {
					"description": "Can represent a permission group"
				},
				"created_on": "2022-08-29T16:58:29.745574Z",
				"modified_on": "2022-09-12T08:26:37.255907Z",
				"permissions": [
					{
						"id": "7d42552322884d19bb63ed7f69b5ac21",
						"key": "com.cloudflare.api.account.fake.permission",
						"attributes": {
							"legacy_name": "#fake:permission"
						}
					}
				]
			},
			"success": true,
			"errors": [],
			"messages": []
			}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/iam/permission_groups/fake-permission-group-id", handler)

	result, err := client.GetPermissionGroup(context.Background(), AccountIdentifier(testAccountID), "fake-permission-group-id")
	if assert.NoError(t, err) {
		assert.Equal(t, result, mockPermissionGroup)
	}
}

func TestPermissionGroups(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "result": [
				{
				  "id": "f08020434ba14a0bb46bd9ff52f23b04",
				  "name": "Fake Permission Group",
				  "status": "V",
				  "meta": {
					"description": "Can represent a permission group"
				  },
				  "created_on": "2022-08-29T16:58:29.745574Z",
				  "modified_on": "2022-09-12T08:26:37.255907Z",
				  "permissions": [
					{
					  "id": "7d42552322884d19bb63ed7f69b5ac21",
					  "key": "com.cloudflare.api.account.fake.permission",
					  "attributes": {
						"legacy_name": "#fake:permission"
					  }
					}
				  ]
				}
			  ],
			  "success": true,
			  "errors": [],
			  "messages": []
			}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/iam/permission_groups", handler)

	result, err := client.ListPermissionGroups(context.Background(), AccountIdentifier(testAccountID), ListPermissionGroupParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, result, []PermissionGroup{mockPermissionGroup})
	}
}
