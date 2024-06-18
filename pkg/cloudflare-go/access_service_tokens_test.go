package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccessServiceTokens(t *testing.T) {
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
					"created_at": "2014-01-01T05:20:00.12345Z",
					"updated_at": "2014-01-01T05:20:00.12345Z",
					"expires_at": "2015-01-01T05:20:00.12345Z",
					"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
					"name": "CI/CD token",
					"client_id": "88bf3b6d86161464f6509f7219099e57.access.example.com",
					"duration": "8760h"
				}
			]
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	expiresAt, _ := time.Parse(time.RFC3339, "2015-01-01T05:20:00.12345Z")

	want := []AccessServiceToken{
		{
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
			ExpiresAt: &expiresAt,
			ID:        "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
			Name:      "CI/CD token",
			ClientID:  "88bf3b6d86161464f6509f7219099e57.access.example.com",
			Duration:  "8760h",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/service_tokens", handler)

	actual, _, err := client.ListAccessServiceTokens(context.Background(), testAccountRC, ListAccessServiceTokensParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/service_tokens", handler)

	actual, _, err = client.ListAccessServiceTokens(context.Background(), testZoneRC, ListAccessServiceTokensParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateAccessServiceToken(t *testing.T) {
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
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"expires_at": "2015-01-01T05:20:00.12345Z",
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "CI/CD token",
				"client_id": "88bf3b6d86161464f6509f7219099e57.access.example.com",
				"client_secret": "bdd31cbc4dec990953e39163fbbb194c93313ca9f0a6e420346af9d326b1d2a5",
				"duration": "8760h"
			}
		}
		`)
	}

	expected := AccessServiceTokenCreateResponse{
		CreatedAt:    &createdAt,
		UpdatedAt:    &updatedAt,
		ExpiresAt:    &expiresAt,
		ID:           "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name:         "CI/CD token",
		ClientID:     "88bf3b6d86161464f6509f7219099e57.access.example.com",
		ClientSecret: "bdd31cbc4dec990953e39163fbbb194c93313ca9f0a6e420346af9d326b1d2a5",
		Duration:     "8760h",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/service_tokens", handler)

	actual, err := client.CreateAccessServiceToken(context.Background(), testAccountRC, CreateAccessServiceTokenParams{Name: "CI/CD token"})

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/service_tokens", handler)

	actual, err = client.CreateAccessServiceToken(context.Background(), testZoneRC, CreateAccessServiceTokenParams{Name: "CI/CD token"})

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestUpdateAccessServiceToken(t *testing.T) {
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
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"expires_at": "2015-01-01T05:20:00.12345Z",
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "CI/CD token",
				"client_id": "88bf3b6d86161464f6509f7219099e57.access.example.com",
				"duration": "8760h"
			}
		}
		`)
	}

	expected := AccessServiceTokenUpdateResponse{
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		ExpiresAt: &expiresAt,
		ID:        "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name:      "CI/CD token",
		ClientID:  "88bf3b6d86161464f6509f7219099e57.access.example.com",
		Duration:  "8760h",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/service_tokens/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	actual, err := client.UpdateAccessServiceToken(context.Background(), testAccountRC, UpdateAccessServiceTokenParams{UUID: "f174e90a-fafe-4643-bbbc-4a0ed4fc8415", Name: "CI/CD token"})

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/service_tokens/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	actual, err = client.UpdateAccessServiceToken(context.Background(), testZoneRC, UpdateAccessServiceTokenParams{UUID: "f174e90a-fafe-4643-bbbc-4a0ed4fc8415", Name: "CI/CD token"})

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestRefreshAccessServiceToken(t *testing.T) {
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
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"expires_at": "2015-01-01T05:20:00.12345Z",
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "CI/CD token",
				"client_id": "88bf3b6d86161464f6509f7219099e57.access.example.com",
				"duration": "8760h"
			}
		}
		`)
	}

	expected := AccessServiceTokenRefreshResponse{
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		ExpiresAt: &expiresAt,
		ID:        "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name:      "CI/CD token",
		ClientID:  "88bf3b6d86161464f6509f7219099e57.access.example.com",
		Duration:  "8760h",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/service_tokens/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/refresh", handler)

	actual, err := client.RefreshAccessServiceToken(context.Background(), testAccountRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestRotateAccessServiceToken(t *testing.T) {
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
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"expires_at": "2015-01-01T05:20:00.12345Z",
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "CI/CD token",
				"client_id": "88bf3b6d86161464f6509f7219099e57.access.example.com",
				"client_secret": "bdd31cbc4dec990953e39163fbbb194c93313ca9f0a6e420346af9d326b1d2a5",
				"duration": "8760h"
			}
		}
		`)
	}

	expected := AccessServiceTokenRotateResponse{
		CreatedAt:    &createdAt,
		UpdatedAt:    &updatedAt,
		ExpiresAt:    &expiresAt,
		ID:           "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name:         "CI/CD token",
		ClientID:     "88bf3b6d86161464f6509f7219099e57.access.example.com",
		ClientSecret: "bdd31cbc4dec990953e39163fbbb194c93313ca9f0a6e420346af9d326b1d2a5",
		Duration:     "8760h",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/service_tokens/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/rotate", handler)

	actual, err := client.RotateAccessServiceToken(context.Background(), testAccountRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestDeleteAccessServiceToken(t *testing.T) {
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
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"expires_at": "2015-01-01T05:20:00.12345Z",
				"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "CI/CD token",
				"client_id": "88bf3b6d86161464f6509f7219099e57.access.example.com",
				"duration": "8760h"
			}
		}
		`)
	}

	expected := AccessServiceTokenUpdateResponse{
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		ExpiresAt: &expiresAt,
		ID:        "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name:      "CI/CD token",
		ClientID:  "88bf3b6d86161464f6509f7219099e57.access.example.com",
		Duration:  "8760h",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/service_tokens/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	actual, err := client.DeleteAccessServiceToken(context.Background(), testAccountRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/service_tokens/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	actual, err = client.DeleteAccessServiceToken(context.Background(), testZoneRC, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
