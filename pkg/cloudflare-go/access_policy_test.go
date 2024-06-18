package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	pageOptions         = PaginationOptions{}
	accessApplicationID = "6e1c88f1-6b06-4b8a-a9e9-3ec7da2ee0c1"
	accessPolicyID      = "699d98642c564d2e855e9661899b7252"

	createdAt, _ = time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ = time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	expiresAt, _ = time.Parse(time.RFC3339, "2015-01-01T05:20:00.12345Z")

	isolationRequired            = true
	purposeJustificationRequired = true
	purposeJustificationPrompt   = "Please provide a business reason for your need to access before continuing."
	approvalRequired             = true

	expectedAccessPolicy = AccessPolicy{
		ID:         "699d98642c564d2e855e9661899b7252",
		Precedence: 1,
		Decision:   "allow",
		CreatedAt:  &createdAt,
		UpdatedAt:  &updatedAt,
		Name:       "Allow devs",
		Include: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		Exclude: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		Require: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		IsolationRequired:            &isolationRequired,
		SessionDuration:              StringPtr("12h"),
		PurposeJustificationRequired: &purposeJustificationRequired,
		ApprovalRequired:             &approvalRequired,
		PurposeJustificationPrompt:   &purposeJustificationPrompt,
		ApprovalGroups: []AccessApprovalGroup{
			{
				EmailListUuid:   "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
				ApprovalsNeeded: 3,
			},
			{
				EmailAddresses:  []string{"email1@example.com", "email2@example.com"},
				ApprovalsNeeded: 1,
			},
		},
	}
)

func TestAccessPolicies(t *testing.T) {
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
				"precedence": 1,
				"decision": "allow",
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
				],
				"isolation_required": true,
				"purpose_justification_required": true,
				"purpose_justification_prompt": "Please provide a business reason for your need to access before continuing.",
				"approval_required": true,
				"session_duration": "12h",
				"approval_groups": [
				  {
					"email_list_uuid": "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
					"approvals_needed": 3
				  },
				  {
					"email_addresses": [
					  "email1@example.com",
					  "email2@example.com"
					],
					"approvals_needed": 1
				  }
				]
			  }
			],
			"result_info": {
			  "page": 1,
			  "per_page": 20,
			  "count": 1,
			  "total_count": 20
			}
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/"+accessApplicationID+"/policies", handler)

	actual, _, err := client.ListAccessPolicies(context.Background(), testAccountRC, ListAccessPoliciesParams{ApplicationID: accessApplicationID})

	if assert.NoError(t, err) {
		assert.Equal(t, []AccessPolicy{expectedAccessPolicy}, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/"+accessApplicationID+"/policies", handler)

	actual, _, err = client.ListAccessPolicies(context.Background(), testZoneRC, ListAccessPoliciesParams{ApplicationID: accessApplicationID})

	if assert.NoError(t, err) {
		assert.Equal(t, []AccessPolicy{expectedAccessPolicy}, actual)
	}

	// Test Listing reusable policies
	mux.HandleFunc("/accounts/"+testAccountID+"/access/policies", handler)

	actual, _, err = client.ListAccessPolicies(context.Background(), testAccountRC, ListAccessPoliciesParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, []AccessPolicy{expectedAccessPolicy}, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/policies", handler)

	actual, _, err = client.ListAccessPolicies(context.Background(), testZoneRC, ListAccessPoliciesParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, []AccessPolicy{expectedAccessPolicy}, actual)
	}
}

func TestAccessPolicy(t *testing.T) {
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
				"precedence": 1,
				"decision": "allow",
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
				],
				"isolation_required": true,
				"purpose_justification_required": true,
				"purpose_justification_prompt": "Please provide a business reason for your need to access before continuing.",
				"approval_required": true,
				"session_duration": "12h",
				"approval_groups": [
					{
						"email_list_uuid": "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
						"approvals_needed": 3
					},
					{
						"email_addresses": ["email1@example.com", "email2@example.com"],
						"approvals_needed": 1
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/"+accessApplicationID+"/policies/"+accessPolicyID, handler)

	actual, err := client.GetAccessPolicy(context.Background(), testAccountRC, GetAccessPolicyParams{ApplicationID: accessApplicationID, PolicyID: accessPolicyID})

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/"+accessApplicationID+"/policies/"+accessPolicyID, handler)

	actual, err = client.GetAccessPolicy(context.Background(), testZoneRC, GetAccessPolicyParams{ApplicationID: accessApplicationID, PolicyID: accessPolicyID})

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}

	// Test getting a reusable policy
	mux.HandleFunc("/accounts/"+testAccountID+"/access/policies/"+accessPolicyID, handler)

	actual, err = client.GetAccessPolicy(context.Background(), testAccountRC, GetAccessPolicyParams{PolicyID: accessPolicyID})

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/policies/"+accessPolicyID, handler)

	actual, err = client.GetAccessPolicy(context.Background(), testZoneRC, GetAccessPolicyParams{PolicyID: accessPolicyID})

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}
}

func TestCreateAccessPolicy(t *testing.T) {
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
				"precedence": 1,
				"decision": "allow",
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
				],
				"isolation_required": true,
				"purpose_justification_required": true,
				"purpose_justification_prompt": "Please provide a business reason for your need to access before continuing.",
				"approval_required": true,
				"session_duration": "12h",
				"approval_groups": [
					{
						"email_list_uuid": "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
						"approvals_needed": 3
					},
					{
						"email_addresses": ["email1@example.com", "email2@example.com"],
						"approvals_needed": 1
					}
				]
			}
		}
		`)
	}

	accessPolicy := CreateAccessPolicyParams{
		ApplicationID: accessApplicationID,
		Name:          "Allow devs",
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
		Decision:                     "allow",
		PurposeJustificationRequired: &purposeJustificationRequired,
		PurposeJustificationPrompt:   &purposeJustificationPrompt,
		SessionDuration:              StringPtr("12h"),
		ApprovalGroups: []AccessApprovalGroup{
			{
				EmailListUuid:   "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
				ApprovalsNeeded: 3,
			},
			{
				EmailAddresses:  []string{"email1@example.com", "email2@example.com"},
				ApprovalsNeeded: 1,
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/"+accessApplicationID+"/policies", handler)

	actual, err := client.CreateAccessPolicy(context.Background(), testAccountRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/"+accessApplicationID+"/policies", handler)

	actual, err = client.CreateAccessPolicy(context.Background(), testZoneRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}

	// Test creating a reusable policy
	accessPolicy.ApplicationID = ""
	mux.HandleFunc("/accounts/"+testAccountID+"/access/policies", handler)

	actual, err = client.CreateAccessPolicy(context.Background(), testAccountRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/policies", handler)

	actual, err = client.CreateAccessPolicy(context.Background(), testZoneRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}
}

func TestCreateAccessPolicyAuthContextRule(t *testing.T) {
	setup()
	defer teardown()

	expectedAccessPolicyAuthContext := AccessPolicy{
		ID:         "699d98642c564d2e855e9661899b7252",
		Precedence: 1,
		Decision:   "allow",
		CreatedAt:  &createdAt,
		UpdatedAt:  &updatedAt,
		Name:       "Allow devs",
		Include: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		Exclude: []interface{}{},
		Require: []interface{}{
			map[string]interface{}{"auth_context": map[string]interface{}{"id": "authContextID123", "identity_provider_id": "IDPIDtest123", "ac_id": "c1"}},
		},
		IsolationRequired:            &isolationRequired,
		PurposeJustificationRequired: &purposeJustificationRequired,
		ApprovalRequired:             &approvalRequired,
		PurposeJustificationPrompt:   &purposeJustificationPrompt,
		SessionDuration:              StringPtr("12h"),
		ApprovalGroups: []AccessApprovalGroup{
			{
				EmailListUuid:   "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
				ApprovalsNeeded: 3,
			},
			{
				EmailAddresses:  []string{"email1@example.com", "email2@example.com"},
				ApprovalsNeeded: 1,
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "699d98642c564d2e855e9661899b7252",
				"precedence": 1,
				"decision": "allow",
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
				"exclude": [],
				"require": [
					{
						"auth_context": {
							"id": "authContextID123",
							"identity_provider_id": "IDPIDtest123",
							"ac_id": "c1"
						}
					}
				],
				"isolation_required": true,
				"purpose_justification_required": true,
				"purpose_justification_prompt": "Please provide a business reason for your need to access before continuing.",
				"approval_required": true,
				"session_duration": "12h",
				"approval_groups": [
					{
						"email_list_uuid": "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
						"approvals_needed": 3
					},
					{
						"email_addresses": ["email1@example.com", "email2@example.com"],
						"approvals_needed": 1
					}
				]
			}
		}
		`)
	}

	accessPolicy := CreateAccessPolicyParams{
		ApplicationID: accessApplicationID,
		Name:          "Allow devs",
		Include: []interface{}{
			AccessGroupEmail{struct {
				Email string `json:"email"`
			}{Email: "test@example.com"}},
		},
		Exclude: []interface{}{},
		Require: []interface{}{
			AccessGroupAzureAuthContext{struct {
				ID                 string `json:"id"`
				IdentityProviderID string `json:"identity_provider_id"`
				ACID               string `json:"ac_id"`
			}{
				ID:                 "authContextID123",
				IdentityProviderID: "IDPIDtest123",
				ACID:               "c1",
			}},
		},
		Decision:                     "allow",
		PurposeJustificationRequired: &purposeJustificationRequired,
		PurposeJustificationPrompt:   &purposeJustificationPrompt,
		ApprovalGroups: []AccessApprovalGroup{
			{
				EmailListUuid:   "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
				ApprovalsNeeded: 3,
			},
			{
				EmailAddresses:  []string{"email1@example.com", "email2@example.com"},
				ApprovalsNeeded: 1,
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/"+accessApplicationID+"/policies", handler)

	actual, err := client.CreateAccessPolicy(context.Background(), testAccountRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicyAuthContext, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/"+accessApplicationID+"/policies", handler)

	actual, err = client.CreateAccessPolicy(context.Background(), testZoneRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicyAuthContext, actual)
	}
}

func TestUpdateAccessPolicy(t *testing.T) {
	setup()
	defer teardown()

	accessPolicy := UpdateAccessPolicyParams{
		ApplicationID: accessApplicationID,
		PolicyID:      accessPolicyID,
		Precedence:    1,
		Decision:      "allow",
		Name:          "Allow devs",
		Include: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		Exclude: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		Require: []interface{}{
			map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
		},
		IsolationRequired:            &isolationRequired,
		PurposeJustificationRequired: &purposeJustificationRequired,
		ApprovalRequired:             &approvalRequired,
		PurposeJustificationPrompt:   &purposeJustificationPrompt,
		SessionDuration:              StringPtr("12h"),
		ApprovalGroups: []AccessApprovalGroup{
			{
				EmailListUuid:   "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
				ApprovalsNeeded: 3,
			},
			{
				EmailAddresses:  []string{"email1@example.com", "email2@example.com"},
				ApprovalsNeeded: 1,
			},
		},
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "699d98642c564d2e855e9661899b7252",
				"precedence": 1,
				"decision": "allow",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"name": "Allow devs",
				"session_duration": "12h",
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
				],
				"isolation_required": true,
				"purpose_justification_required": true,
				"purpose_justification_prompt": "Please provide a business reason for your need to access before continuing.",
				"approval_required": true,
				"approval_groups": [
					{
						"email_list_uuid": "2413b6d7-bbe5-48bd-8fbb-e52069c85561",
						"approvals_needed": 3
					},
					{
						"email_addresses": ["email1@example.com", "email2@example.com"],
						"approvals_needed": 1
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/"+accessApplicationID+"/policies/"+accessPolicyID, handler)
	actual, err := client.UpdateAccessPolicy(context.Background(), testAccountRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/"+accessApplicationID+"/policies/"+accessPolicyID, handler)
	actual, err = client.UpdateAccessPolicy(context.Background(), testZoneRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}

	// Test updating reusable policies
	accessPolicy.ApplicationID = ""
	mux.HandleFunc("/accounts/"+testAccountID+"/access/policies/"+accessPolicyID, handler)
	actual, err = client.UpdateAccessPolicy(context.Background(), testAccountRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/policies/"+accessPolicyID, handler)
	actual, err = client.UpdateAccessPolicy(context.Background(), testZoneRC, accessPolicy)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccessPolicy, actual)
	}
}

func TestUpdateAccessPolicyWithMissingID(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateAccessPolicy(context.Background(), testAccountRC, UpdateAccessPolicyParams{ApplicationID: accessApplicationID})
	assert.EqualError(t, err, "access policy ID cannot be empty")

	_, err = client.UpdateAccessPolicy(context.Background(), testZoneRC, UpdateAccessPolicyParams{ApplicationID: accessApplicationID})
	assert.EqualError(t, err, "access policy ID cannot be empty")
}

func TestDeleteAccessPolicy(t *testing.T) {
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

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/"+accessApplicationID+"/policies/"+accessPolicyID, handler)
	err := client.DeleteAccessPolicy(context.Background(), testAccountRC, DeleteAccessPolicyParams{ApplicationID: accessApplicationID, PolicyID: accessPolicyID})

	assert.NoError(t, err)

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/"+accessApplicationID+"/policies/"+accessPolicyID, handler)
	err = client.DeleteAccessPolicy(context.Background(), testZoneRC, DeleteAccessPolicyParams{ApplicationID: accessApplicationID, PolicyID: accessPolicyID})

	assert.NoError(t, err)

	// Test deleting a reusable policy
	mux.HandleFunc("/accounts/"+testAccountID+"/access/policies/"+accessPolicyID, handler)
	err = client.DeleteAccessPolicy(context.Background(), testAccountRC, DeleteAccessPolicyParams{PolicyID: accessPolicyID})

	assert.NoError(t, err)

	mux.HandleFunc("/zones/"+testZoneID+"/access/policies/"+accessPolicyID, handler)
	err = client.DeleteAccessPolicy(context.Background(), testZoneRC, DeleteAccessPolicyParams{PolicyID: accessPolicyID})

	assert.NoError(t, err)
}
