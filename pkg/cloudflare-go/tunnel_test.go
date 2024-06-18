package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListTunnels(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("tunnel", "multiple_full"))
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/cfd_tunnel", handler)

	createdAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	deletedAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	want := []Tunnel{{
		ID:        testTunnelID,
		Name:      "blog",
		CreatedAt: &createdAt,
		DeletedAt: &deletedAt,
		Connections: []TunnelConnection{{
			ColoName:           "DFW",
			ID:                 testTunnelID,
			IsPendingReconnect: false,
			ClientID:           "dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
			ClientVersion:      "2022.2.0",
			OpenedAt:           "2021-01-25T18:22:34.317854Z",
			OriginIP:           "198.51.100.1",
		}},
	}}

	actual, _, err := client.ListTunnels(context.Background(), AccountIdentifier(testAccountID), TunnelListParams{UUID: testTunnelID})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListTunnelsPagination(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		qry, _ := url.Parse(r.RequestURI)
		assert.Equal(t, "blog", qry.Query().Get("name"))
		assert.Equal(t, "2", qry.Query().Get("page"))
		assert.Equal(t, "1", qry.Query().Get("per_page"))
		fmt.Fprint(w, loadFixture("tunnel", "multiple_full"))
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/cfd_tunnel", handler)

	createdAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	deletedAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	want := []Tunnel{
		{
			ID:        testTunnelID,
			Name:      "blog",
			CreatedAt: &createdAt,
			DeletedAt: &deletedAt,
			Connections: []TunnelConnection{{
				ColoName:           "DFW",
				ID:                 testTunnelID,
				IsPendingReconnect: false,
				ClientID:           "dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
				ClientVersion:      "2022.2.0",
				OpenedAt:           "2021-01-25T18:22:34.317854Z",
				OriginIP:           "198.51.100.1",
			}},
		},
	}

	actual, _, err := client.ListTunnels(context.Background(), AccountIdentifier(testAccountID),
		TunnelListParams{
			Name: "blog",
			ResultInfo: ResultInfo{
				Page:    2,
				PerPage: 1,
			},
		})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetTunnel(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("tunnel", "single_full"))
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/cfd_tunnel/"+testTunnelID, handler)

	createdAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	deletedAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	want := Tunnel{
		ID:        testTunnelID,
		Name:      "blog",
		CreatedAt: &createdAt,
		DeletedAt: &deletedAt,
		Connections: []TunnelConnection{{
			ColoName:           "DFW",
			ID:                 testTunnelID,
			IsPendingReconnect: false,
			ClientID:           "dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
			ClientVersion:      "2022.2.0",
			OpenedAt:           "2021-01-25T18:22:34.317854Z",
			OriginIP:           "198.51.100.1",
		}},
		TunnelType:   "cfd_tunnel",
		Status:       "healthy",
		RemoteConfig: true,
	}

	actual, err := client.GetTunnel(context.Background(), AccountIdentifier(testAccountID), testTunnelID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateTunnel(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("tunnel", "single_full"))
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/cfd_tunnel", handler)

	createdAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	deletedAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	want := Tunnel{
		ID:        testTunnelID,
		Name:      "blog",
		CreatedAt: &createdAt,
		DeletedAt: &deletedAt,
		Connections: []TunnelConnection{{
			ColoName:           "DFW",
			ID:                 testTunnelID,
			IsPendingReconnect: false,
			ClientID:           "dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
			ClientVersion:      "2022.2.0",
			OpenedAt:           "2021-01-25T18:22:34.317854Z",
			OriginIP:           "198.51.100.1",
		}},
		TunnelType:   "cfd_tunnel",
		Status:       "healthy",
		RemoteConfig: true,
	}

	actual, err := client.CreateTunnel(context.Background(), AccountIdentifier(testAccountID), TunnelCreateParams{Name: "blog", Secret: "notarealsecret", ConfigSrc: "cloudflare"})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateTunnelConfiguration(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method '%s', got %s", http.MethodPut, r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("tunnel", "configuration"))
	}

	timeout, _ := time.ParseDuration("10s")
	mux.HandleFunc(fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/configurations", testAccountID, testTunnelID), handler)
	want := TunnelConfigurationResult{
		TunnelID: testTunnelID,
		Version:  5,
		Config: TunnelConfiguration{
			Ingress: []UnvalidatedIngressRule{
				{
					Hostname: "test.example.com",
					Service:  "https://localhost:8000",
					OriginRequest: &OriginRequestConfig{
						NoTLSVerify: BoolPtr(true),
					},
				},
				{
					Service: "http_status:404",
				},
			},
			WarpRouting: &WarpRoutingConfig{
				Enabled: true,
			},
			OriginRequest: OriginRequestConfig{
				ConnectTimeout: &TunnelDuration{timeout},
			},
		}}

	actual, err := client.UpdateTunnelConfiguration(context.Background(), AccountIdentifier(testAccountID), TunnelConfigurationParams{
		TunnelID: testTunnelID,
		Config: TunnelConfiguration{
			Ingress: []UnvalidatedIngressRule{
				{
					Hostname: "test.example.com",
					Service:  "https://localhost:8000",
					OriginRequest: &OriginRequestConfig{
						NoTLSVerify: BoolPtr(true),
					},
				},
				{
					Service: "http_status:404",
				},
			},
			WarpRouting: &WarpRoutingConfig{
				Enabled: true,
			},
			OriginRequest: OriginRequestConfig{
				ConnectTimeout: &TunnelDuration{10},
			},
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetTunnelConfiguration(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method '%s', got %s", http.MethodGet, r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("tunnel", "configuration"))
	}

	timeout, _ := time.ParseDuration("10s")
	mux.HandleFunc(fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/configurations", testAccountID, testTunnelID), handler)
	want := TunnelConfigurationResult{
		TunnelID: testTunnelID,
		Version:  5,
		Config: TunnelConfiguration{
			Ingress: []UnvalidatedIngressRule{
				{
					Hostname: "test.example.com",
					Service:  "https://localhost:8000",
					OriginRequest: &OriginRequestConfig{
						NoTLSVerify: BoolPtr(true),
					},
				},
				{
					Service: "http_status:404",
				},
			},
			WarpRouting: &WarpRoutingConfig{
				Enabled: true,
			},
			OriginRequest: OriginRequestConfig{
				ConnectTimeout: &TunnelDuration{timeout},
			},
		}}

	actual, err := client.GetTunnelConfiguration(context.Background(), AccountIdentifier(testAccountID), testTunnelID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTunnelConnections(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success":true,
			"errors":[],
			"messages":[],
			"result":[{
				"id":"dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
				"features": [
					"allow_remote_config",
					"serialized_headers",
					"ha-origin"
				],
				"version": "2022.2.0",
            	"arch": "linux_amd64",
				"config_version": 15,
				"run_at":"2009-11-10T23:00:00Z",
				"conns": [
					{
						"colo_name": "DFW",
						"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
						"is_pending_reconnect": false,
						"client_id": "dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
						"client_version": "2022.2.0",
						"opened_at": "2021-01-25T18:22:34.317854Z",
						"origin_ip": "198.51.100.1"
					}
				]
			}]
		}
		`)
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/connections", testAccountID, testTunnelID), handler)

	runAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	want := []Connection{
		{
			ID: "dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
			Features: []string{
				"allow_remote_config",
				"serialized_headers",
				"ha-origin",
			},
			Version: "2022.2.0",
			Arch:    "linux_amd64",
			RunAt:   &runAt,
			Connections: []TunnelConnection{{
				ColoName:           "DFW",
				ID:                 testTunnelID,
				IsPendingReconnect: false,
				ClientID:           "dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
				ClientVersion:      "2022.2.0",
				OpenedAt:           "2021-01-25T18:22:34.317854Z",
				OriginIP:           "198.51.100.1",
			}},
			ConfigVersion: 15,
		},
	}

	actual, err := client.ListTunnelConnections(context.Background(), AccountIdentifier(testAccountID), testTunnelID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteTunnel(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("tunnel", "single_full"))
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/cfd_tunnel/%s", testAccountID, testTunnelID), handler)

	err := client.DeleteTunnel(context.Background(), AccountIdentifier(testAccountID), testTunnelID)
	assert.NoError(t, err)
}

func TestCleanupTunnelConnections(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("tunnel", "empty"))
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/connections", testAccountID, testTunnelID), handler)

	err := client.CleanupTunnelConnections(context.Background(), AccountIdentifier(testAccountID), testTunnelID)
	assert.NoError(t, err)
}

func TestTunnelToken(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("tunnel", "token"))
	}

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/token", testAccountID, testTunnelID), handler)

	token, err := client.GetTunnelToken(context.Background(), AccountIdentifier(testAccountID), testTunnelID)
	assert.NoError(t, err)
	assert.Equal(t, "ZHNraGdhc2RraGFza2hqZGFza2poZGFza2poYXNrZGpoYWtzamRoa2FzZGpoa2FzamRoa2Rhc2po\na2FzamRoa2FqCg==", token)
}
