package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	accessGroupID = "699d98642c564d2e855e9661899b7252"

	expectedAccessGroup = AccessGroup{
		ID:        "699d98642c564d2e855e9661899b7252",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		Name:      "Allow devs",
		Include: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		Exclude: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		Require: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
	}

	expectedAccessGroupIpList = AccessGroup{
		ID:        "899d98642c564d2e855e9661899b7252",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		Name:      "Allow devs",
		Include: []interface{}{
			map[string]interface{}{"ip_list": map[string]interface{}{"id": "989d98642c564d2e855e9661899b7252"}},
		},
	}

	expectedAccessGroupEmailList = AccessGroup{
		ID:        "a99d98642c564d2e855e9661899b7252",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		Name:      "Allow devs",
		Include: []interface{}{
			map[string]interface{}{"email_list": map[string]interface{}{"id": "8a9d98642c564d2e855e9661899b7252"}},
		},
	}
)

func TestAccessGroups(t *testing.T) {
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
					"id": "699d98642c564d2e855e9661899b7252",
					"created_at": "2014-01-01T05:20:00.12345Z",
					"updated_at": "2014-01-01T05:20:00.12345Z",
					"name": "Allow devs",
					"include": [
						{
							"email": {
								"email": "test@example.com"
							}
						}
					],
					"exclude": [
						{
							"email": {
								"email": "test@example.com"
							}
						}
					],
					"require": [
						{
							"email": {
								"email": "test@example.com"
							}
						}
					]
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

	mux.HandleFunc("/accounts/"+testAccountID+"/access/groups", handler)

	actual, _, err := client.ListAccessGroups(context.Background(), testAccountRC, ListAccessGroupsParams{ResultInfo{}})

	if assert.NoError(t, err) {
		assert.Equal(t, []AccessGroup{expectedAccessGroup}, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/groups", handler)

	actual, _, err = client.ListAccessGroups(context.Background(), testZoneRC, ListAccessGroupsParams{ResultInfo{}})

	if assert.NoError(t, err) {
		assert.Equal(t, []AccessGroup{expectedAccessGroup}, actual)
	}
}

func TestAccessGroup(t *testing.T) {
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
				"id": "699d98642c564d2e855e9661899b7252",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"name": "Allow devs",
				"include": [
					{
						"email": {
							"email": "test@example.com"
						}
					}
				],
				"exclude": [
					{
						"email": {
							"email": "test@example.com"
						}
					}
				],
				"require": [
					{
						"email": {
							"email": "test@example.com"
						}
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/groups/"+accessGroupID, handler)

	actual, err := client.GetAccessGroup(context.Background(), testAccountRC, accessGroupID)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroup, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/groups/"+accessGroupID, handler)

	actual, err = client.GetAccessGroup(context.Background(), testZoneRC, accessGroupID)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroup, actual)
	}
}

func TestCreateAccessGroup(t *testing.T) {
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
				"id": "699d98642c564d2e855e9661899b7252",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"name": "Allow devs",
				"include": [
					{
						"email": {
							"email": "test@example.com"
						}
					}
				],
				"exclude": [
					{
						"email": {
							"email": "test@example.com"
						}
					}
				],
				"require": [
					{
						"email": {
							"email": "test@example.com"
						}
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/groups", handler)

	params := CreateAccessGroupParams{
		Name: "Allow devs",
		Include: []interface{}{
			AccessGroupEmail{struct {
				Email string `json:"email"`
			}{Email: "test@example.com"}},
		},
		Exclude: []interface{}{
			AccessGroupEmail{struct {
				Email string `json:"email"`
			}{Email: "test@example.com"}},
		},
		Require: []interface{}{
			AccessGroupEmail{struct {
				Email string `json:"email"`
			}{Email: "test@example.com"}},
		},
	}

	actual, err := client.CreateAccessGroup(context.Background(), testAccountRC, params)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroup, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/groups", handler)

	actual, err = client.CreateAccessGroup(context.Background(), testZoneRC, params)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroup, actual)
	}
}

func TestUpdateAccessGroup(t *testing.T) {
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
				"id": "699d98642c564d2e855e9661899b7252",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"name": "Allow devs",
				"include": [
					{
						"email": {
							"email": "test@example.com"
						}
					}
				],
				"exclude": [
					{
						"email": {
							"email": "test@example.com"
						}
					}
				],
				"require": [
					{
						"email": {
							"email": "test@example.com"
						}
					}
				]
			}
		}
		`)
	}

	params := UpdateAccessGroupParams{
		ID:   "699d98642c564d2e855e9661899b7252",
		Name: "Allow devs",
		Include: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		Exclude: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		Require: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/access/groups/"+accessGroupID, handler)
	actual, err := client.UpdateAccessGroup(context.Background(), testAccountRC, params)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroup, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/groups/"+accessGroupID, handler)
	actual, err = client.UpdateAccessGroup(context.Background(), testZoneRC, params)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroup, actual)
	}
}

func TestUpdateAccessGroupWithMissingID(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateAccessGroup(context.Background(), testAccountRC, UpdateAccessGroupParams{})
	assert.EqualError(t, err, "access group ID cannot be empty")

	_, err = client.UpdateAccessGroup(context.Background(), testZoneRC, UpdateAccessGroupParams{})
	assert.EqualError(t, err, "access group ID cannot be empty")
}

func TestDeleteAccessGroup(t *testing.T) {
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
				"id": "699d98642c564d2e855e9661899b7252"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/groups/"+accessGroupID, handler)
	err := client.DeleteAccessGroup(context.Background(), testAccountRC, accessGroupID)

	assert.NoError(t, err)

	mux.HandleFunc("/zones/"+testZoneID+"/access/groups/"+accessGroupID, handler)
	err = client.DeleteAccessGroup(context.Background(), testZoneRC, accessGroupID)

	assert.NoError(t, err)
}

func TestCreateIPListAccessGroup(t *testing.T) {
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
				"id": "899d98642c564d2e855e9661899b7252",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"name": "Allow devs",
				"include": [
					{
						"ip_list": {
							"id": "989d98642c564d2e855e9661899b7252"
						}
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/groups", handler)

	accessGroup := CreateAccessGroupParams{
		Name: "Allow devs by iplist",
		Include: []interface{}{
			AccessGroupIPList{struct {
				ID string `json:"id"`
			}{ID: "989d98642c564d2e855e9661899b7252"}},
		},
	}

	actual, err := client.CreateAccessGroup(context.Background(), testAccountRC, accessGroup)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroupIpList, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/groups", handler)

	actual, err = client.CreateAccessGroup(context.Background(), testZoneRC, accessGroup)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroupIpList, actual)
	}
}

func TestCreateEmailListAccessGroup(t *testing.T) {
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
				"id": "a99d98642c564d2e855e9661899b7252",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"name": "Allow devs",
				"include": [
					{
						"email_list": {
							"id": "8a9d98642c564d2e855e9661899b7252"
						}
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/groups", handler)

	accessGroup := CreateAccessGroupParams{
		Name: "Allow devs by email_list",
		Include: []interface{}{
			AccessGroupEmailList{struct {
				ID string `json:"id"`
			}{ID: "989d98642c564d2e855e9661899b7252"}},
		},
	}

	actual, err := client.CreateAccessGroup(context.Background(), testAccountRC, accessGroup)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroupEmailList, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/groups", handler)

	actual, err = client.CreateAccessGroup(context.Background(), testZoneRC, accessGroup)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessGroupEmailList, actual)
	}
}
