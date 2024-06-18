package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAccessRules(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"id": "92f17202ed8bd63d69a66b86a49a8f6b",
					"notes": "This rule is on because of an event that occurred on date X",
					"allowed_modes": [
						"whitelist",
						"block",
						"challenge",
						"js_challenge"
					],
					"mode": "challenge",
					"configuration": {
						"target": "ip",
						"value": "198.51.100.4"
					},
					"created_on": "2014-01-01T05:20:00Z",
					"modified_on": "2014-01-01T05:20:00Z",
					"scope": {
						"id": "7c5dae5552338874e5053f2534d2767a",
						"email": "user@example.com",
						"type": "user"
					}
				}
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 200
			}
		}`)
	}

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := &AccessRuleListResponse{
		Result: []AccessRule{{
			ID:           "92f17202ed8bd63d69a66b86a49a8f6b",
			Notes:        "This rule is on because of an event that occurred on date X",
			AllowedModes: []string{"whitelist", "block", "challenge", "js_challenge"},
			Mode:         "challenge",
			Configuration: AccessRuleConfiguration{
				Target: "ip",
				Value:  "198.51.100.4",
			},
			CreatedOn:  createdOn,
			ModifiedOn: modifiedOn,
			Scope: AccessRuleScope{
				ID:    "7c5dae5552338874e5053f2534d2767a",
				Email: "user@example.com",
				Type:  "user",
			},
		}},
		ResultInfo: ResultInfo{
			Page:    1,
			PerPage: 20,
			Count:   1,
			Total:   200,
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	accessRule := AccessRule{
		Notes: "my note",
		Mode:  "challenge",
		Configuration: AccessRuleConfiguration{
			Target: "ip",
			Value:  "198.51.100.4",
		},
	}

	mux.HandleFunc("/user/firewall/access_rules/rules", handler)
	actual, err := client.ListUserAccessRules(context.Background(), accessRule, 1)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/access_rules/rules", handler)
	actual, err = client.ListZoneAccessRules(context.Background(), testZoneID, accessRule, 1)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/accounts/"+testAccountID+"/firewall/access_rules/rules", handler)
	actual, err = client.ListAccountAccessRules(context.Background(), testAccountID, accessRule, 1)
	require.NoError(t, err)
	assert.Equal(t, want, actual)
}

func TestCreateAccessRule(t *testing.T) {
	setup()
	defer teardown()

	input := AccessRule{
		Mode: "challenge",
		Configuration: AccessRuleConfiguration{
			Target: "ip",
			Value:  "198.51.100.4",
		},
		Notes: "This rule is on because of an event that occurred on date X",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		var v AccessRule
		err := json.NewDecoder(r.Body).Decode(&v)
		require.NoError(t, err)
		assert.Equal(t, input, v)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "92f17202ed8bd63d69a66b86a49a8f6b",
				"notes": "This rule is on because of an event that occurred on date X",
				"allowed_modes": [
					"whitelist",
					"block",
					"challenge",
					"js_challenge"
				],
				"mode": "challenge",
				"configuration": {
					"target": "ip",
					"value": "198.51.100.4"
				},
				"created_on": "2014-01-01T05:20:00Z",
				"modified_on": "2014-01-01T05:20:00Z",
				"scope": {
					"id": "7c5dae5552338874e5053f2534d2767a",
					"email": "user@example.com",
					"type": "user"
				}
			}
		}`)
	}

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := &AccessRuleResponse{
		Result: AccessRule{
			ID:            "92f17202ed8bd63d69a66b86a49a8f6b",
			Notes:         input.Notes,
			AllowedModes:  []string{"whitelist", "block", "challenge", "js_challenge"},
			Mode:          input.Mode,
			Configuration: input.Configuration,
			CreatedOn:     createdOn,
			ModifiedOn:    modifiedOn,
			Scope: AccessRuleScope{
				ID:    "7c5dae5552338874e5053f2534d2767a",
				Email: "user@example.com",
				Type:  "user",
			},
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	mux.HandleFunc("/user/firewall/access_rules/rules", handler)
	actual, err := client.CreateUserAccessRule(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/access_rules/rules", handler)
	actual, err = client.CreateZoneAccessRule(context.Background(), testZoneID, input)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/accounts/"+testAccountID+"/firewall/access_rules/rules", handler)
	actual, err = client.CreateAccountAccessRule(context.Background(), testAccountID, input)
	require.NoError(t, err)
	assert.Equal(t, want, actual)
}

func TestAccessRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "92f17202ed8bd63d69a66b86a49a8f6b",
				"notes": "This rule is on because of an event that occurred on date X",
				"allowed_modes": [
					"whitelist",
					"block",
					"challenge",
					"js_challenge"
				],
				"mode": "challenge",
				"configuration": {
					"target": "ip",
					"value": "198.51.100.4"
				},
				"created_on": "2014-01-01T05:20:00Z",
				"modified_on": "2014-01-01T05:20:00Z",
				"scope": {
					"id": "7c5dae5552338874e5053f2534d2767a",
					"email": "user@example.com",
					"type": "user"
				}
			}
		}`)
	}

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := &AccessRuleResponse{
		Result: AccessRule{
			ID:           "92f17202ed8bd63d69a66b86a49a8f6b",
			Notes:        "This rule is on because of an event that occurred on date X",
			AllowedModes: []string{"whitelist", "block", "challenge", "js_challenge"},
			Mode:         "challenge",
			Configuration: AccessRuleConfiguration{
				Target: "ip",
				Value:  "198.51.100.4",
			},
			CreatedOn:  createdOn,
			ModifiedOn: modifiedOn,
			Scope: AccessRuleScope{
				ID:    "7c5dae5552338874e5053f2534d2767a",
				Email: "user@example.com",
				Type:  "user",
			},
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	accessRuleID := "92f17202ed8bd63d69a66b86a49a8f6b"

	mux.HandleFunc("/user/firewall/access_rules/rules/"+accessRuleID, handler)
	actual, err := client.UserAccessRule(context.Background(), accessRuleID)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/access_rules/rules/"+accessRuleID, handler)
	actual, err = client.ZoneAccessRule(context.Background(), testZoneID, accessRuleID)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/accounts/"+testAccountID+"/firewall/access_rules/rules/"+accessRuleID, handler)
	actual, err = client.AccountAccessRule(context.Background(), testAccountID, accessRuleID)
	require.NoError(t, err)
	assert.Equal(t, want, actual)
}

func TestUpdateAccessRule(t *testing.T) {
	setup()
	defer teardown()

	input := AccessRule{
		Mode:  "challenge",
		Notes: "This rule is on because of an event that occurred on date X",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		var v AccessRule
		err := json.NewDecoder(r.Body).Decode(&v)
		require.NoError(t, err)
		assert.Equal(t, input, v)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "92f17202ed8bd63d69a66b86a49a8f6b",
				"notes": "This rule is on because of an event that occurred on date X",
				"allowed_modes": [
					"whitelist",
					"block",
					"challenge",
					"js_challenge"
				],
				"mode": "challenge",
				"configuration": {
					"target": "ip",
					"value": "198.51.100.4"
				},
				"created_on": "2014-01-01T05:20:00Z",
				"modified_on": "2014-01-01T05:20:00Z",
				"scope": {
					"id": "7c5dae5552338874e5053f2534d2767a",
					"email": "user@example.com",
					"type": "user"
				}
			}
		}`)
	}

	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	want := &AccessRuleResponse{
		Result: AccessRule{
			ID:           "92f17202ed8bd63d69a66b86a49a8f6b",
			Notes:        "This rule is on because of an event that occurred on date X",
			AllowedModes: []string{"whitelist", "block", "challenge", "js_challenge"},
			Mode:         "challenge",
			Configuration: AccessRuleConfiguration{
				Target: "ip",
				Value:  "198.51.100.4",
			},
			CreatedOn:  createdOn,
			ModifiedOn: modifiedOn,
			Scope: AccessRuleScope{
				ID:    "7c5dae5552338874e5053f2534d2767a",
				Email: "user@example.com",
				Type:  "user",
			},
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	accessRuleID := "92f17202ed8bd63d69a66b86a49a8f6b"

	mux.HandleFunc("/user/firewall/access_rules/rules/"+accessRuleID, handler)
	actual, err := client.UpdateUserAccessRule(context.Background(), accessRuleID, input)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/access_rules/rules/"+accessRuleID, handler)
	actual, err = client.UpdateZoneAccessRule(context.Background(), testZoneID, accessRuleID, input)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/accounts/"+testAccountID+"/firewall/access_rules/rules/"+accessRuleID, handler)
	actual, err = client.UpdateAccountAccessRule(context.Background(), testAccountID, accessRuleID, input)
	require.NoError(t, err)
	assert.Equal(t, want, actual)
}

func TestDeleteAccessRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "92f17202ed8bd63d69a66b86a49a8f6b"
			}
		}`)
	}

	want := &AccessRuleResponse{
		Result: AccessRule{
			ID: "92f17202ed8bd63d69a66b86a49a8f6b",
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	accessRuleID := "92f17202ed8bd63d69a66b86a49a8f6b"

	mux.HandleFunc("/user/firewall/access_rules/rules/"+accessRuleID, handler)
	actual, err := client.DeleteUserAccessRule(context.Background(), accessRuleID)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/access_rules/rules/"+accessRuleID, handler)
	actual, err = client.DeleteZoneAccessRule(context.Background(), testZoneID, accessRuleID)
	require.NoError(t, err)
	assert.Equal(t, want, actual)

	mux.HandleFunc("/accounts/"+testAccountID+"/firewall/access_rules/rules/"+accessRuleID, handler)
	actual, err = client.DeleteAccountAccessRule(context.Background(), testAccountID, accessRuleID)
	require.NoError(t, err)
	assert.Equal(t, want, actual)
}
