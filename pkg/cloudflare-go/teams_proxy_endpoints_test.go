package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyEndpoint(t *testing.T) {
	setup()
	defer teardown()

	id := "0f8185414dec4a5e9034f3d917c17890"
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err := fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				 "id": "0f8185414dec4a5e9034f3d917c17890",
				 "name": "home",
				 "ips": ["192.0.2.1/32"],
				 "subdomain": "q15l7x2lbw",
				 "created_at": "2020-05-18T22:07:03Z",
				 "updated_at": "2020-05-18T22:07:05Z"
			}
		}`)
		require.Nil(t, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:03Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:05Z")

	want := TeamsProxyEndpoint{
		ID:        "0f8185414dec4a5e9034f3d917c17890",
		Name:      "home",
		IPs:       []string{"192.0.2.1/32"},
		Subdomain: "q15l7x2lbw",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/proxy_endpoints/%s", testAccountID, id), handler)

	actual, err := client.TeamsProxyEndpoint(context.Background(), testAccountID, id)
	require.Nil(t, err)
	assert.Equal(t, want, actual)
}

func TestProxyEndpoints(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err := fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"id": "0f8185414dec4a5e9034f3d917c17890",
					"name": "home",
					"ips": ["192.0.2.1/32"],
					"subdomain": "q15l7x2lbw",
					"created_at": "2020-05-18T22:07:03Z",
					"updated_at": "2020-05-18T22:07:05Z"
				}
			]
		}
		`)
		require.Nil(t, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:03Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:05Z")

	want := []TeamsProxyEndpoint{{
		ID:        "0f8185414dec4a5e9034f3d917c17890",
		Name:      "home",
		IPs:       []string{"192.0.2.1/32"},
		Subdomain: "q15l7x2lbw",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/proxy_endpoints", testAccountID), handler)

	actual, _, err := client.TeamsProxyEndpoints(context.Background(), testAccountID)
	require.Nil(t, err)
	assert.Equal(t, want, actual)
}

func TestCreateProxyEndpoint(t *testing.T) {
	setup()
	defer teardown()

	id := "0f8185414dec4a5e9034f3d917c17890"
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err := fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "%s",
				"name": "test",
				"ips": ["192.0.2.1/32"],
				"subdomain": "q15l7x2lbw",
				"created_at": "2020-05-18T22:07:03Z",
				"updated_at": "2020-05-18T22:07:05Z"
			}
		}`, id)
		require.Nil(t, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:03Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:05Z")

	want := TeamsProxyEndpoint{
		ID:        id,
		Name:      "test",
		IPs:       []string{"192.0.2.1/32"},
		Subdomain: "q15l7x2lbw",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/proxy_endpoints", testAccountID), handler)

	actual, err := client.CreateTeamsProxyEndpoint(context.Background(), testAccountID, TeamsProxyEndpoint{})
	require.Nil(t, err)
	assert.Equal(t, want, actual)
}

func TestUpdateProxyEndpoint(t *testing.T) {
	setup()
	defer teardown()

	id := "0f8185414dec4a5e9034f3d917c17890"
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err := fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "%s",
				"name": "new",
				"ips": ["192.0.2.1/32"],
				"subdomain": "q15l7x2lbw",
				"created_at": "2020-05-18T22:07:03Z",
				"updated_at": "2020-05-18T22:07:05Z"
			}
		}`, id)
		require.Nil(t, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:03Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:05Z")

	want := TeamsProxyEndpoint{
		ID:        id,
		Name:      "new",
		IPs:       []string{"192.0.2.1/32"},
		Subdomain: "q15l7x2lbw",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/proxy_endpoints/%s", testAccountID, id), handler)

	actual, err := client.UpdateTeamsProxyEndpoint(context.Background(), testAccountID, TeamsProxyEndpoint{
		ID: id,
	})
	require.Nil(t, err)
	assert.Equal(t, want, actual)
}

func TestDeleteProxyEndpoint(t *testing.T) {
	setup()
	defer teardown()

	id := "0f8185414dec4a5e9034f3d917c17890"
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/proxy_endpoints/%s", testAccountID, id), handler)
	err := client.DeleteTeamsProxyEndpoint(context.Background(), testAccountID, id)
	require.Nil(t, err)
}
