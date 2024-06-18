package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestArgoTunnels(t *testing.T) {
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
					"id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
					"name": "blog",
					"created_at": "2009-11-10T23:00:00Z",
					"deleted_at": "2009-11-10T23:00:00Z",
					"connections": [
						{
							"colo_name": "DFW",
							"uuid": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
							"is_pending_reconnect": false
						}
					]
				}
			]
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/cfd_tunnel", handler)

	createdAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	deletedAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	want := []ArgoTunnel{{
		ID:        "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name:      "blog",
		CreatedAt: &createdAt,
		DeletedAt: &deletedAt,
		Connections: []ArgoTunnelConnection{{
			ColoName:           "DFW",
			UUID:               "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
			IsPendingReconnect: false,
		}},
	}}

	actual, err := client.ArgoTunnels(context.Background(), "01a7362d577a6c3019a474fd6f485823")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestArgoTunnel(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success":true,
			"errors":[],
			"messages":[],
			"result":{
				"id":"f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name":"blog",
				"created_at":"2009-11-10T23:00:00Z",
				"deleted_at":"2009-11-10T23:00:00Z",
				"connections":[
					{
						"colo_name":"DFW",
						"uuid":"f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
						"is_pending_reconnect":false
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/cfd_tunnel/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	createdAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	deletedAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	want := ArgoTunnel{
		ID:        "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name:      "blog",
		CreatedAt: &createdAt,
		DeletedAt: &deletedAt,
		Connections: []ArgoTunnelConnection{{
			ColoName:           "DFW",
			UUID:               "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
			IsPendingReconnect: false,
		}},
	}

	actual, err := client.ArgoTunnel(context.Background(), "01a7362d577a6c3019a474fd6f485823", "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateArgoTunnel(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success":true,
			"errors":[],
			"messages":[],
			"result":{
				"id":"f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name":"blog",
				"created_at":"2009-11-10T23:00:00Z",
				"deleted_at":"2009-11-10T23:00:00Z",
				"connections":[
					{
						"colo_name":"DFW",
						"uuid":"f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
						"is_pending_reconnect":false
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/cfd_tunnel", handler)

	createdAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	deletedAt, _ := time.Parse(time.RFC3339, "2009-11-10T23:00:00Z")
	want := ArgoTunnel{
		ID:        "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		Name:      "blog",
		CreatedAt: &createdAt,
		DeletedAt: &deletedAt,
		Connections: []ArgoTunnelConnection{{
			ColoName:           "DFW",
			UUID:               "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
			IsPendingReconnect: false,
		}},
	}

	actual, err := client.CreateArgoTunnel(context.Background(), "01a7362d577a6c3019a474fd6f485823", "blog", "notarealsecret")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteArgoTunnel(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success":true,
			"errors":[],
			"messages":[],
			"result":{
				"id":"f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name":"blog",
				"created_at":"2009-11-10T23:00:00Z",
				"deleted_at":"2009-11-10T23:00:00Z",
				"connections":[
					{
						"colo_name":"DFW",
						"uuid":"f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
						"is_pending_reconnect":false
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/cfd_tunnel/f174e90a-fafe-4643-bbbc-4a0ed4fc8415", handler)

	err := client.DeleteArgoTunnel(context.Background(), "01a7362d577a6c3019a474fd6f485823", "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")
	assert.NoError(t, err)
}

func TestCleanupArgoTunnelConnections(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": []
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/cfd_tunnel/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/connections", handler)

	err := client.CleanupArgoTunnelConnections(context.Background(), "01a7362d577a6c3019a474fd6f485823", "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")
	assert.NoError(t, err)
}
