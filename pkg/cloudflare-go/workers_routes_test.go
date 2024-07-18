package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWorkersRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/workers/routes", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, createWorkerRouteResponse)
	})

	res, err := client.CreateWorkerRoute(context.Background(), ZoneIdentifier(testZoneID), CreateWorkerRouteParams{
		Pattern: "app1.example.com/*",
		Script:  "example",
	})

	want := WorkerRouteResponse{successResponse, WorkerRoute{ID: "e7a57d8746e74ae49c25994dadb421b1"}}
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestDeleteWorkersRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/workers/routes/e7a57d8746e74ae49c25994dadb421b1", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, deleteWorkerRouteResponseData)
	})
	res, err := client.DeleteWorkerRoute(context.Background(), ZoneIdentifier(testZoneID), "e7a57d8746e74ae49c25994dadb421b1")
	want := WorkerRouteResponse{successResponse,
		WorkerRoute{
			ID: "e7a57d8746e74ae49c25994dadb421b1",
		}}
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestListWorkersRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/workers/routes", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, listWorkerRouteResponse)
	})

	res, err := client.ListWorkerRoutes(context.Background(), ZoneIdentifier(testZoneID), ListWorkerRoutesParams{})
	want := WorkerRoutesResponse{successResponse,
		[]WorkerRoute{
			{ID: "e7a57d8746e74ae49c25994dadb421b1", Pattern: "app1.example.com/*", ScriptName: "test_script_1"},
			{ID: "f8b68e9857f85bf59c25994dadb421b1", Pattern: "app2.example.com/*", ScriptName: "test_script_2"},
			{ID: "2b5bf4240cd34c77852fac70b1bf745a", Pattern: "app3.example.com/*", ScriptName: "test_script_3"},
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestGetWorkersRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/workers/routes/1234", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, getRouteResponseData)
	})

	res, err := client.GetWorkerRoute(context.Background(), ZoneIdentifier(testZoneID), "1234")
	want := WorkerRouteResponse{successResponse,
		WorkerRoute{
			ID:         "e7a57d8746e74ae49c25994dadb421b1",
			Pattern:    "app1.example.com/*",
			ScriptName: "script-name"},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestUpdateWorkersRoute(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/workers/routes/e7a57d8746e74ae49c25994dadb421b1", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, updateWorkerRouteEntResponse)
	})

	res, err := client.UpdateWorkerRoute(context.Background(), ZoneIdentifier(testZoneID), UpdateWorkerRouteParams{
		ID:      "e7a57d8746e74ae49c25994dadb421b1",
		Pattern: "app3.example.com/*",
		Script:  "test_script_1",
	})
	want := WorkerRouteResponse{successResponse,
		WorkerRoute{
			ID:         "e7a57d8746e74ae49c25994dadb421b1",
			Pattern:    "app3.example.com/*",
			ScriptName: "test_script_1",
		}}
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}
