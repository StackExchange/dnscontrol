package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	accountCreatedOn, _   = time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	expectedAccountStruct = Account{
		ID:        "01a7362d577a6c3019a474fd6f485823",
		Name:      "Cloudflare Demo",
		CreatedOn: accountCreatedOn,
		Settings: &AccountSettings{
			EnforceTwoFactor: false,
		},
	}
)

func TestAccounts(t *testing.T) {
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
					"id": "01a7362d577a6c3019a474fd6f485823",
					"name": "Cloudflare Demo",
					"created_on": "2014-01-01T05:20:00.12345Z",
					"settings": {
						"enforce_twofactor": false
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

	mux.HandleFunc("/accounts", handler)
	want := []Account{expectedAccountStruct}

	actual, _, err := client.Accounts(context.Background(), AccountsListParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAccount(t *testing.T) {
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
				"id": "01a7362d577a6c3019a474fd6f485823",
				"name": "Cloudflare Demo",
				"created_on": "2014-01-01T05:20:00.12345Z",
				"settings": {
					"enforce_twofactor": false
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

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823", handler)
	want := expectedAccountStruct

	actual, _, err := client.Account(context.Background(), "01a7362d577a6c3019a474fd6f485823")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateAccount(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
				"id":"01a7362d577a6c3019a474fd6f485823",
				"name":"Cloudflare Demo - New",
				"created_on": "2014-01-01T05:20:00.12345Z",
				"settings":{
					"enforce_twofactor":false
					}
				}`, string(b), "JSON payload not equal")
		}

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "01a7362d577a6c3019a474fd6f485823",
				"name": "Cloudflare Demo - New",
				"created_on": "2014-01-01T05:20:00.12345Z",
				"settings": {
					"enforce_twofactor": false
				}
			}
		}`)
	})

	oldAccountDetails := Account{
		ID:        "01a7362d577a6c3019a474fd6f485823",
		Name:      "Cloudflare Demo - Old",
		CreatedOn: accountCreatedOn,
		Settings: &AccountSettings{
			EnforceTwoFactor: false,
		},
	}

	newAccountDetails := Account{
		ID:        "01a7362d577a6c3019a474fd6f485823",
		Name:      "Cloudflare Demo - New",
		CreatedOn: accountCreatedOn,
		Settings: &AccountSettings{
			EnforceTwoFactor: false,
		},
	}

	account, err := client.UpdateAccount(context.Background(), newAccountDetails.ID, newAccountDetails)
	if assert.NoError(t, err) {
		assert.NotEqual(t, oldAccountDetails.Name, account.Name)
		assert.Equal(t, account.Name, "Cloudflare Demo - New")
	}
}

func TestCreateAccount(t *testing.T) {
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
				"name": "Cloudflare Demo",
				"type": "standard"
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

	mux.HandleFunc("/accounts", handler)
	newAccount := Account{
		Name: "Cloudflare Demo",
		Type: "standard",
	}

	actual, err := client.CreateAccount(context.Background(), newAccount)

	if assert.NoError(t, err) {
		assert.Equal(t, newAccount, actual)
	}
}

func TestDeleteAccount(t *testing.T) {
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
				"id": "1b16db169c9cb7853009857198fae1b9"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID, handler)
	err := client.DeleteAccount(context.Background(), testAccountID)

	assert.NoError(t, err)
}
