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

func TestTeamsLocations(t *testing.T) {
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
					"networks": [
						{
							"network": "198.51.100.1/32",
							"id": "8e4c7835436345f0ab395429b187a076"
						}
					],
					"policy_ids": [],
					"ip": "2a06:98c1:54::2419",
					"doh_subdomain": "q15l7x2lbw",
					"anonymized_logs_enabled": false,
					"ipv4_destination": null,
					"client_default": false,
					"ecs_support": false,
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

	want := []TeamsLocation{{
		ID:                    "0f8185414dec4a5e9034f3d917c17890",
		Name:                  "home",
		Networks:              []TeamsLocationNetwork{{ID: "8e4c7835436345f0ab395429b187a076", Network: "198.51.100.1/32"}},
		PolicyIDs:             []string{},
		Ip:                    "2a06:98c1:54::2419",
		Subdomain:             "q15l7x2lbw",
		AnonymizedLogsEnabled: false,
		IPv4Destination:       "",
		ClientDefault:         false,
		ECSSupport:            BoolPtr(false),
		CreatedAt:             &createdAt,
		UpdatedAt:             &updatedAt,
	}}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/locations", testAccountID), handler)

	actual, _, err := client.TeamsLocations(context.Background(), testAccountID)
	require.Nil(t, err)
	assert.Equal(t, want, actual)
}

func TestTeamsLocation(t *testing.T) {
	setup()
	defer teardown()

	id := "0f8185414dec4a5e9034f3d917c17890"
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err := fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "%s",
				"name": "home",
				"networks": [
					{
						"network": "198.51.100.1/32",
						"id": "8e4c7835436345f0ab395429b187a076"
					}
				],
				"policy_ids": [],
				"ip": "2a06:98c1:54::2419",
				"doh_subdomain": "q15l7x2lbw",
				"anonymized_logs_enabled": false,
				"ipv4_destination": null,
				"client_default": false,
				"ecs_support": false,
				"created_at": "2020-05-18T22:07:03Z",
				"updated_at": "2020-05-18T22:07:05Z"
			}
		}`, id)
		require.Nil(t, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:03Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:05Z")

	want := TeamsLocation{
		ID:                    id,
		Name:                  "home",
		Networks:              []TeamsLocationNetwork{{ID: "8e4c7835436345f0ab395429b187a076", Network: "198.51.100.1/32"}},
		PolicyIDs:             []string{},
		Ip:                    "2a06:98c1:54::2419",
		Subdomain:             "q15l7x2lbw",
		AnonymizedLogsEnabled: false,
		IPv4Destination:       "",
		ClientDefault:         false,
		ECSSupport:            BoolPtr(false),
		CreatedAt:             &createdAt,
		UpdatedAt:             &updatedAt,
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/locations/%s", testAccountID, id), handler)

	actual, err := client.TeamsLocation(context.Background(), testAccountID, id)
	require.Nil(t, err)
	assert.Equal(t, want, actual)
}

func TestCreateTeamsLocation(t *testing.T) {
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
				"networks": [
					{
						"network": "198.51.100.1/32",
						"id": "8e4c7835436345f0ab395429b187a076"
					}
				],
				"policy_ids": [],
				"ip": "2a06:98c1:54::2419",
				"doh_subdomain": "q15l7x2lbw",
				"anonymized_logs_enabled": false,
				"ipv4_destination": null,
				"client_default": false,
				"ecs_support": false,
				"created_at": "2020-05-18T22:07:03Z",
				"updated_at": "2020-05-18T22:07:05Z"
			}
		}`, id)
		require.Nil(t, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:03Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:05Z")

	want := TeamsLocation{
		ID:                    id,
		Name:                  "test",
		Networks:              []TeamsLocationNetwork{{ID: "8e4c7835436345f0ab395429b187a076", Network: "198.51.100.1/32"}},
		PolicyIDs:             []string{},
		Ip:                    "2a06:98c1:54::2419",
		Subdomain:             "q15l7x2lbw",
		AnonymizedLogsEnabled: false,
		IPv4Destination:       "",
		ClientDefault:         false,
		ECSSupport:            BoolPtr(false),
		CreatedAt:             &createdAt,
		UpdatedAt:             &updatedAt,
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/locations", testAccountID), handler)

	actual, err := client.CreateTeamsLocation(context.Background(), testAccountID, TeamsLocation{
		Name:          "test",
		ClientDefault: true,
		Networks:      []TeamsLocationNetwork{},
	})
	require.Nil(t, err)
	assert.Equal(t, want, actual)
}

func TestUpdateTeamsLocation(t *testing.T) {
	setup()
	defer teardown()

	id := "0f8185414dec4a5e9034f3d917c17890"
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err := fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "%s",
				"name": "new",
				"networks": [
					{
						"network": "198.51.100.1/32",
						"id": "8e4c7835436345f0ab395429b187a076"
					}
				],
				"policy_ids": [],
				"ip": "2a06:98c1:54::2419",
				"doh_subdomain": "q15l7x2lbw",
				"anonymized_logs_enabled": false,
				"ipv4_destination": null,
				"client_default": false,
				"ecs_support": false,
				"created_at": "2020-05-18T22:07:03Z",
				"updated_at": "2020-05-18T22:07:05Z"
			}
		}`, id)
		require.Nil(t, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:03Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2020-05-18T22:07:05Z")

	want := TeamsLocation{
		ID:                    id,
		Name:                  "new",
		Networks:              []TeamsLocationNetwork{{ID: "8e4c7835436345f0ab395429b187a076", Network: "198.51.100.1/32"}},
		PolicyIDs:             []string{},
		Ip:                    "2a06:98c1:54::2419",
		Subdomain:             "q15l7x2lbw",
		AnonymizedLogsEnabled: false,
		IPv4Destination:       "",
		ClientDefault:         false,
		ECSSupport:            BoolPtr(false),
		CreatedAt:             &createdAt,
		UpdatedAt:             &updatedAt,
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/locations/%s", testAccountID, id), handler)

	actual, err := client.UpdateTeamsLocation(context.Background(), testAccountID, TeamsLocation{
		ID:            id,
		Name:          "new",
		ClientDefault: false,
		Networks:      []TeamsLocationNetwork{{ID: "", Network: "1.2.3.4"}},
	})
	require.Nil(t, err)
	assert.Equal(t, want, actual)
}

func TestDeleteTeamsLocation(t *testing.T) {
	setup()
	defer teardown()

	id := "0f8185414dec4a5e9034f3d917c17890"
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/gateway/locations/%s", testAccountID, id), handler)
	err := client.DeleteTeamsLocation(context.Background(), testAccountID, id)
	require.Nil(t, err)
}
