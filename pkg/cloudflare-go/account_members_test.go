package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var expectedAccountMemberStruct = AccountMember{
	ID:   "4536bcfad5faccb111b47003c79917fa",
	Code: "05dd05cce12bbed97c0d87cd78e89bc2fd41a6cee72f27f6fc84af2e45c0fac0",
	User: AccountMemberUserDetails{
		ID:                             "7c5dae5552338874e5053f2534d2767a",
		FirstName:                      "John",
		LastName:                       "Appleseed",
		Email:                          "user@example.com",
		TwoFactorAuthenticationEnabled: false,
	},
	Status: "accepted",
	Roles: []AccountRole{
		{
			ID:          "3536bcfad5faccb999b47003c79917fb",
			Name:        "Account Administrator",
			Description: "Administrative access to the entire Account",
			Permissions: map[string]AccountRolePermission{
				"analytics": {Read: true, Edit: true},
				"billing":   {Read: true, Edit: false},
			},
		},
	},
}

var expectedNewAccountMemberStruct = AccountMember{
	ID:   "4536bcfad5faccb111b47003c79917fa",
	Code: "05dd05cce12bbed97c0d87cd78e89bc2fd41a6cee72f27f6fc84af2e45c0fac0",
	User: AccountMemberUserDetails{
		Email:                          "user@example.com",
		TwoFactorAuthenticationEnabled: false,
	},
	Status: "pending",
	Roles: []AccountRole{
		{
			ID:          "3536bcfad5faccb999b47003c79917fb",
			Name:        "Account Administrator",
			Description: "Administrative access to the entire Account",
			Permissions: map[string]AccountRolePermission{
				"analytics": {Read: true, Edit: true},
				"billing":   {Read: true, Edit: true},
			},
		},
	},
}

var expectedNewAccountMemberAcceptedStruct = AccountMember{
	ID:   "4536bcfad5faccb111b47003c79917fa",
	Code: "05dd05cce12bbed97c0d87cd78e89bc2fd41a6cee72f27f6fc84af2e45c0fac0",
	User: AccountMemberUserDetails{
		Email:                          "user@example.com",
		TwoFactorAuthenticationEnabled: false,
	},
	Status: "accepted",
	Roles: []AccountRole{
		{
			ID:          "3536bcfad5faccb999b47003c79917fb",
			Name:        "Account Administrator",
			Description: "Administrative access to the entire Account",
			Permissions: map[string]AccountRolePermission{
				"analytics": {Read: true, Edit: true},
				"billing":   {Read: true, Edit: true},
			},
		},
	},
}

var newUpdatedAccountMemberStruct = AccountMember{
	ID:   "4536bcfad5faccb111b47003c79917fa",
	Code: "05dd05cce12bbed97c0d87cd78e89bc2fd41a6cee72f27f6fc84af2e45c0fac0",
	User: AccountMemberUserDetails{
		ID:                             "7c5dae5552338874e5053f2534d2767a",
		FirstName:                      "John",
		LastName:                       "Appleseeds",
		Email:                          "new-user@example.com",
		TwoFactorAuthenticationEnabled: false,
	},
	Status: "accepted",
	Roles: []AccountRole{
		{
			ID:          "3536bcfad5faccb999b47003c79917fb",
			Name:        "Account Administrator",
			Description: "Administrative access to the entire Account",
			Permissions: map[string]AccountRolePermission{
				"analytics": {Read: true, Edit: true},
				"billing":   {Read: true, Edit: true},
			},
		},
	},
}

var mockPolicy = Policy{
	ID: "mock-policy-id",
	PermissionGroups: []PermissionGroup{{
		ID:   "mock-permission-group-id",
		Name: "mock-permission-group-name",
		Permissions: []Permission{{
			ID:  "mock-permission-id",
			Key: "mock-permission-key",
		}},
	}},
	ResourceGroups: []ResourceGroup{{
		ID:   "mock-resource-group-id",
		Name: "mock-resource-group-name",
		Scope: Scope{
			Key: "mock-resource-group-name",
			ScopeObjects: []ScopeObject{{
				Key: "*",
			}},
		},
	}},
	Access: "allow",
}

var expectedNewAccountMemberWithPoliciesStruct = AccountMember{
	ID:   "new-member-with-polcies-id",
	Code: "new-member-with-policies-code",
	User: AccountMemberUserDetails{
		ID:                             "new-member-with-policies-user-id",
		FirstName:                      "John",
		LastName:                       "Appleseed",
		Email:                          "user@example.com",
		TwoFactorAuthenticationEnabled: false,
	},
	Status:   "accepted",
	Policies: []Policy{mockPolicy},
}

func TestAccountMembers(t *testing.T) {
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
					"id": "4536bcfad5faccb111b47003c79917fa",
					"code": "05dd05cce12bbed97c0d87cd78e89bc2fd41a6cee72f27f6fc84af2e45c0fac0",
					"user": {
						"id": "7c5dae5552338874e5053f2534d2767a",
						"first_name": "John",
						"last_name": "Appleseed",
						"email": "user@example.com",
						"two_factor_authentication_enabled": false
					},
					"status": "accepted",
					"roles": [
						{
							"id": "3536bcfad5faccb999b47003c79917fb",
							"name": "Account Administrator",
							"description": "Administrative access to the entire Account",
							"permissions": {
								"analytics": {
									"read": true,
									"edit": true
								},
								"billing": {
									"read": true,
									"edit": false
								}
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

	mux.HandleFunc("/accounts/"+testAccountID+"/members", handler)
	want := []AccountMember{expectedAccountMemberStruct}

	actual, _, err := client.AccountMembers(context.Background(), "01a7362d577a6c3019a474fd6f485823", PaginationOptions{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAccountMembersWithoutAccountID(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.AccountMembers(context.Background(), "", PaginationOptions{})

	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}
}

func TestCreateAccountMemberWithStatus(t *testing.T) {
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
				"id": "4536bcfad5faccb111b47003c79917fa",
				"code": "05dd05cce12bbed97c0d87cd78e89bc2fd41a6cee72f27f6fc84af2e45c0fac0",
				"user": {
					"id": null,
					"first_name": null,
					"last_name": null,
					"email": "user@example.com",
					"two_factor_authentication_enabled": false
				},
				"status": "accepted",
				"roles": [{
					"id": "3536bcfad5faccb999b47003c79917fb",
					"name": "Account Administrator",
					"description": "Administrative access to the entire Account",
					"permissions": {
						"analytics": {
							"read": true,
							"edit": true
						},
						"billing": {
							"read": true,
							"edit": true
						}
					}
				}]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/members", handler)

	actual, err := client.CreateAccountMemberWithStatus(context.Background(), "01a7362d577a6c3019a474fd6f485823", "user@example.com", []string{"3536bcfad5faccb999b47003c79917fb"}, "accepted")

	if assert.NoError(t, err) {
		assert.Equal(t, expectedNewAccountMemberAcceptedStruct, actual)
	}
}

func TestCreateAccountMember(t *testing.T) {
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
				"id": "4536bcfad5faccb111b47003c79917fa",
				"code": "05dd05cce12bbed97c0d87cd78e89bc2fd41a6cee72f27f6fc84af2e45c0fac0",
				"user": {
					"id": null,
					"first_name": null,
					"last_name": null,
					"email": "user@example.com",
					"two_factor_authentication_enabled": false
				},
				"status": "pending",
				"roles": [{
					"id": "3536bcfad5faccb999b47003c79917fb",
					"name": "Account Administrator",
					"description": "Administrative access to the entire Account",
					"permissions": {
						"analytics": {
							"read": true,
							"edit": true
						},
						"billing": {
							"read": true,
							"edit": true
						}
					}
				}]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/members", handler)

	createAccountParams := CreateAccountMemberParams{
		EmailAddress: "user@example.com",
		Roles:        []string{"3536bcfad5faccb999b47003c79917fb"},
	}

	actual, err := client.CreateAccountMember(context.Background(), AccountIdentifier("01a7362d577a6c3019a474fd6f485823"), createAccountParams)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedNewAccountMemberStruct, actual)
	}
}

func TestCreateAccountMemberWithPolicies(t *testing.T) {
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
				"id": "new-member-with-polcies-id",
				"code": "new-member-with-policies-code",
				"user": {
					"id": "new-member-with-policies-user-id",
					"first_name": "John",
					"last_name": "Appleseed",
					"email": "user@example.com",
					"two_factor_authentication_enabled": false
				},
				"status": "accepted",
				"policies": [{
					"id": "mock-policy-id",
					"permission_groups": [{
						"id": "mock-permission-group-id",
						"name": "mock-permission-group-name",
						"permissions": [{
							"id": "mock-permission-id",
							"key": "mock-permission-key"
						}]
					}],
					"resource_groups": [{
						"id": "mock-resource-group-id",
						"name": "mock-resource-group-name",
						"scope": {
							"key": "mock-resource-group-name",
							"objects": [{
								"key": "*"
							}]
						}
					}],
					"access": "allow"
				}]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/members", handler)
	actual, err := client.CreateAccountMember(context.Background(), AccountIdentifier("01a7362d577a6c3019a474fd6f485823"), CreateAccountMemberParams{
		EmailAddress: "user@example.com",
		Roles:        nil,
		Policies:     []Policy{mockPolicy},
		Status:       "",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, expectedNewAccountMemberWithPoliciesStruct, actual)
	}
}

func TestCreateAccountMemberWithRolesAndPoliciesErr(t *testing.T) {
	setup()
	defer teardown()

	accountResource := &ResourceContainer{
		Level:      AccountRouteLevel,
		Identifier: "01a7362d577a6c3019a474fd6f485823",
	}
	_, err := client.CreateAccountMember(context.Background(), accountResource, CreateAccountMemberParams{
		EmailAddress: "user@example.com",
		Roles:        []string{"fake-role-id"},
		Policies:     []Policy{mockPolicy},
		Status:       "active",
	})

	if assert.Error(t, err) {
		assert.Equal(t, err, ErrMissingMemberRolesOrPolicies)
	}
}

func TestCreateAccountMemberWithoutAccountID(t *testing.T) {
	setup()
	defer teardown()

	accountResource := &ResourceContainer{
		Level:      AccountRouteLevel,
		Identifier: "",
	}
	_, err := client.CreateAccountMember(context.Background(), accountResource, CreateAccountMemberParams{
		EmailAddress: "user@example.com",
		Roles:        []string{"fake-role-id"},
		Status:       "active",
	})

	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}
}

func TestUpdateAccountMember(t *testing.T) {
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
				"id": "4536bcfad5faccb111b47003c79917fa",
				"code": "05dd05cce12bbed97c0d87cd78e89bc2fd41a6cee72f27f6fc84af2e45c0fac0",
				"user": {
					"id": "7c5dae5552338874e5053f2534d2767a",
					"first_name": "John",
					"last_name": "Appleseeds",
					"email": "new-user@example.com",
					"two_factor_authentication_enabled": false
				},
				"status": "accepted",
				"roles": [{
					"id": "3536bcfad5faccb999b47003c79917fb",
					"name": "Account Administrator",
					"description": "Administrative access to the entire Account",
					"permissions": {
						"analytics": {
							"read": true,
							"edit": true
						},
						"billing": {
							"read": true,
							"edit": true
						}
					}
				}]
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

	mux.HandleFunc("/accounts/"+testAccountID+"/members/4536bcfad5faccb111b47003c79917fa", handler)

	actual, err := client.UpdateAccountMember(context.Background(), "01a7362d577a6c3019a474fd6f485823", "4536bcfad5faccb111b47003c79917fa", newUpdatedAccountMemberStruct)

	if assert.NoError(t, err) {
		assert.Equal(t, newUpdatedAccountMemberStruct, actual)
	}
}

func TestUpdateAccountMemberWithPolicies(t *testing.T) {
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
				"id": "new-member-with-polcies-id",
				"code": "new-member-with-policies-code",
				"user": {
					"id": "new-member-with-policies-user-id",
					"first_name": "John",
					"last_name": "Appleseed",
					"email": "user@example.com",
					"two_factor_authentication_enabled": false
				},
				"status": "accepted",
				"policies": [{
					"id": "mock-policy-id",
					"permission_groups": [{
						"id": "mock-permission-group-id",
						"name": "mock-permission-group-name",
						"permissions": [{
							"id": "mock-permission-id",
							"key": "mock-permission-key"
						}]
					}],
					"resource_groups": [{
						"id": "mock-resource-group-id",
						"name": "mock-resource-group-name",
						"scope": {
							"key": "mock-resource-group-name",
							"objects": [{
								"key": "*"
							}]
						}
					}],
					"access": "allow"
				}]
			},
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1
			}
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/members/new-member-with-polcies-id", handler)

	actual, err := client.UpdateAccountMember(context.Background(), "01a7362d577a6c3019a474fd6f485823", "new-member-with-polcies-id", expectedNewAccountMemberWithPoliciesStruct)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedNewAccountMemberWithPoliciesStruct, actual)
	}
}

func UpdateAccountMemberWithRolesAndPoliciesErr(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateAccountMember(context.Background(),
		"01a7362d577a6c3019a474fd6f485823",
		"",
		AccountMember{
			Roles: []AccountRole{{
				ID: "some-role",
			}},
			Policies: []Policy{mockPolicy},
		},
	)

	if assert.Error(t, err) {
		assert.Equal(t, err, ErrMissingMemberRolesOrPolicies)
	}
}

func TestUpdateAccountMemberWithoutAccountID(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateAccountMember(context.Background(), "", "4536bcfad5faccb111b47003c79917fa", newUpdatedAccountMemberStruct)

	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}
}

func TestAccountMember(t *testing.T) {
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
				"id": "4536bcfad5faccb111b47003c79917fa",
				"code": "05dd05cce12bbed97c0d87cd78e89bc2fd41a6cee72f27f6fc84af2e45c0fac0",
				"user": {
					"id": "7c5dae5552338874e5053f2534d2767a",
					"first_name": "John",
					"last_name": "Appleseed",
					"email": "user@example.com",
					"two_factor_authentication_enabled": false
				},
				"status": "accepted",
				"roles": [
					{
						"id": "3536bcfad5faccb999b47003c79917fb",
						"name": "Account Administrator",
						"description": "Administrative access to the entire Account",
						"permissions": {
							"analytics": {
								"read": true,
								"edit": true
							},
							"billing": {
								"read": true,
								"edit": false
							}
						}
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/members/4536bcfad5faccb111b47003c79917fa", handler)

	actual, err := client.AccountMember(context.Background(), "01a7362d577a6c3019a474fd6f485823", "4536bcfad5faccb111b47003c79917fa")

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAccountMemberStruct, actual)
	}
}

func TestAccountMemberWithoutAccountID(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.AccountMember(context.Background(), "", "4536bcfad5faccb111b47003c79917fa")

	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}
}
