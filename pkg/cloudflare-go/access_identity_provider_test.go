package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAccessIdentityProviders(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "25", r.URL.Query().Get("per_page"))
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
					"name": "Widget Corps OTP",
					"type": "github",
					"config": {
						"client_id": "example_id",
						"client_secret": "a-secret-key"
					}
				}
			],
			"result_info": {
				"count": 1,
				"page": 1,
				"per_page": 20,
				"total_count": 1
		  	}
		}
		`)
	}

	want := []AccessIdentityProvider{
		{
			ID:   "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
			Name: "Widget Corps OTP",
			Type: "github",
			Config: AccessIdentityProviderConfiguration{
				ClientID:     "example_id",
				ClientSecret: "a-secret-key",
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/identity_providers", handler)

	actual, _, err := client.ListAccessIdentityProviders(context.Background(), testAccountRC, ListAccessIdentityProvidersParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/identity_providers", handler)

	actual, _, err = client.ListAccessIdentityProviders(context.Background(), testZoneRC, ListAccessIdentityProvidersParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAccessIdentityProviderDetails(t *testing.T) {
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
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "Widget Corps OTP",
				"type": "github",
				"config": {
					"client_id": "example_id",
					"client_secret": "a-secret-key"
				}
			}
		}
		`)
	}

	want := AccessIdentityProvider{
		ID:   "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name: "Widget Corps OTP",
		Type: "github",
		Config: AccessIdentityProviderConfiguration{
			ClientID:     "example_id",
			ClientSecret: "a-secret-key",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc841", handler)

	actual, err := client.GetAccessIdentityProvider(context.Background(), testAccountRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc841")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc841", handler)

	actual, err = client.GetAccessIdentityProvider(context.Background(), testZoneRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc841")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateAccessIdentityProvider(t *testing.T) {
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
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "Widget Corps OTP",
				"type": "github",
				"config": {
					"client_id": "example_id",
					"client_secret": "a-secret-key",
					"conditional_access_enabled": true
				}
			}
		}
		`)
	}

	newIdentityProvider := CreateAccessIdentityProviderParams{
		Name: "Widget Corps OTP",
		Type: "github",
		Config: AccessIdentityProviderConfiguration{
			ClientID:                 "example_id",
			ClientSecret:             "a-secret-key",
			ConditionalAccessEnabled: true,
		},
	}

	want := AccessIdentityProvider{
		ID:   "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name: "Widget Corps OTP",
		Type: "github",
		Config: AccessIdentityProviderConfiguration{
			ClientID:                 "example_id",
			ClientSecret:             "a-secret-key",
			ConditionalAccessEnabled: true,
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/identity_providers", handler)

	actual, err := client.CreateAccessIdentityProvider(context.Background(), testAccountRC, newIdentityProvider)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/identity_providers", handler)

	actual, err = client.CreateAccessIdentityProvider(context.Background(), testZoneRC, newIdentityProvider)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
func TestUpdateAccessIdentityProvider(t *testing.T) {
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
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "Widget Corps OTP",
				"type": "github",
				"config": {
					"client_id": "example_id",
					"client_secret": "a-secret-key"
				}
			}
		}
		`)
	}

	updatedIdentityProvider := UpdateAccessIdentityProviderParams{
		ID:   "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name: "Widget Corps OTP",
		Type: "github",
		Config: AccessIdentityProviderConfiguration{
			ClientID:     "example_id",
			ClientSecret: "a-secret-key",
		},
	}

	want := AccessIdentityProvider{
		ID:   "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name: "Widget Corps OTP",
		Type: "github",
		Config: AccessIdentityProviderConfiguration{
			ClientID:     "example_id",
			ClientSecret: "a-secret-key",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	actual, err := client.UpdateAccessIdentityProvider(context.Background(), testAccountRC, updatedIdentityProvider)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	actual, err = client.UpdateAccessIdentityProvider(context.Background(), testZoneRC, updatedIdentityProvider)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteAccessIdentityProvider(t *testing.T) {
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
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "Widget Corps OTP",
				"type": "github",
				"config": {
					"client_id": "example_id",
					"client_secret": "a-secret-key"
				}
			}
		}
		`)
	}

	want := AccessIdentityProvider{
		ID:   "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name: "Widget Corps OTP",
		Type: "github",
		Config: AccessIdentityProviderConfiguration{
			ClientID:     "example_id",
			ClientSecret: "a-secret-key",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	actual, err := client.DeleteAccessIdentityProvider(context.Background(), testAccountRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	actual, err = client.DeleteAccessIdentityProvider(context.Background(), testZoneRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListAccessIdentityProviderAuthContexts(t *testing.T) {
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
					"id": "04709095-568a-40c4-bf23-5d9edbefe21e",
					"uid": "04709095-568a-40c4-bf23-5d9edbefe21e",
					"ac_id": "c1",
					"display_name": "test_c1",
					"description": ""
				},
				{
					"id": "a6c9b024-8fd1-48b7-9a05-8bca3a43f758",
					"uid": "a6c9b024-8fd1-48b7-9a05-8bca3a43f758",
					"ac_id": "c25",
					"display_name": "test_c25",
					"description": ""
				}
			]
		}
		`)
	}

	want := []AccessAuthContext{
		{
			ID:          "04709095-568a-40c4-bf23-5d9edbefe21e",
			UID:         "04709095-568a-40c4-bf23-5d9edbefe21e",
			ACID:        "c1",
			DisplayName: "test_c1",
			Description: "",
		},
		{
			ID:          "a6c9b024-8fd1-48b7-9a05-8bca3a43f758",
			UID:         "a6c9b024-8fd1-48b7-9a05-8bca3a43f758",
			ACID:        "c25",
			DisplayName: "test_c25",
			Description: "",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/auth_context", handler)

	actual, err := client.ListAccessIdentityProviderAuthContexts(context.Background(), testAccountRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/auth_context", handler)

	actual, err = client.ListAccessIdentityProviderAuthContexts(context.Background(), testAccountRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateAccessIdentityProviderAuthContext(t *testing.T) {
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
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "Widget Corps",
				"type": "AzureAD",
				"config": {
					"client_id": "example_id",
					"client_secret": "a-secret-key",
					"conditional_access_enabled": true
				}
			}
		}
		`)
	}

	want := AccessIdentityProvider{
		ID:   "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name: "Widget Corps",
		Type: "AzureAD",
		Config: AccessIdentityProviderConfiguration{
			ClientID:                 "example_id",
			ClientSecret:             "a-secret-key",
			ConditionalAccessEnabled: true,
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/auth_context", handler)

	actual, err := client.UpdateAccessIdentityProviderAuthContexts(context.Background(), testAccountRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/identity_providers/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/auth_context", handler)

	actual, err = client.UpdateAccessIdentityProviderAuthContexts(context.Background(), testAccountRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
