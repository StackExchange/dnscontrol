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

func TestListTunnelRoutes(t *testing.T) {
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
			    "network": "ff01::/32",
				"tunnel_id": "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
				"tunnel_name": "blog",
				"comment": "Example comment for this route",
				"created_at": "2021-01-25T18:22:34.317854Z",
				"deleted_at": "2021-01-25T18:22:34.317854Z",
				"virtual_network_id": "9f322de4-5988-4945-b770-f1d6ac200f86"
              }
            ]
          }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/teamnet/routes", handler)

	ts, _ := time.Parse(time.RFC3339Nano, "2021-01-25T18:22:34.317854Z")
	want := []TunnelRoute{
		{
			"ff01::/32",
			"f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
			"blog",
			"Example comment for this route",
			&ts,
			&ts,
			"9f322de4-5988-4945-b770-f1d6ac200f86",
		},
	}

	params := TunnelRoutesListParams{}
	got, err := client.ListTunnelRoutes(context.Background(), AccountIdentifier(testAccountID), params)

	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestTunnelRouteForIP(t *testing.T) {
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
			  "network": "ff01::/32",
			  "tunnel_id": "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
			  "tunnel_name": "blog",
			  "comment": "Example comment for this route",
			  "created_at": "2021-01-25T18:22:34.317854Z",
			  "deleted_at": "2021-01-25T18:22:34.317854Z",
			  "virtual_network_id": "9f322de4-5988-4945-b770-f1d6ac200f86"
            }
          }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/teamnet/routes/ip/10.1.0.137", handler)

	ts, _ := time.Parse(time.RFC3339Nano, "2021-01-25T18:22:34.317854Z")
	want := TunnelRoute{
		Network:          "ff01::/32",
		TunnelID:         "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
		TunnelName:       "blog",
		Comment:          "Example comment for this route",
		CreatedAt:        &ts,
		DeletedAt:        &ts,
		VirtualNetworkID: "9f322de4-5988-4945-b770-f1d6ac200f86",
	}

	got, err := client.GetTunnelRouteForIP(context.Background(), AccountIdentifier(testAccountID), TunnelRoutesForIPParams{Network: "10.1.0.137"})

	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestCreateTunnelRoute(t *testing.T) {
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
			  "network": "10.0.0.0/16",
			  "tunnel_id": "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
			  "tunnel_name": "blog",
			  "comment": "Example comment for this route",
			  "created_at": "2021-01-25T18:22:34.317854Z",
			  "deleted_at": "2021-01-25T18:22:34.317854Z",
			  "virtual_network_id": "9f322de4-5988-4945-b770-f1d6ac200f86"
            }
          }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/teamnet/routes/network/10.0.0.0/16", handler)

	ts, _ := time.Parse(time.RFC3339Nano, "2021-01-25T18:22:34.317854Z")
	want := TunnelRoute{
		Network:          "10.0.0.0/16",
		TunnelID:         "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
		TunnelName:       "blog",
		Comment:          "Example comment for this route",
		CreatedAt:        &ts,
		DeletedAt:        &ts,
		VirtualNetworkID: "9f322de4-5988-4945-b770-f1d6ac200f86",
	}

	tunnel, err := client.CreateTunnelRoute(context.Background(), AccountIdentifier(testAccountID), TunnelRoutesCreateParams{TunnelID: testTunnelID, Network: "10.0.0.0/16", Comment: "foo", VirtualNetworkID: "9f322de4-5988-4945-b770-f1d6ac200f86"})
	if assert.NoError(t, err) {
		assert.Equal(t, want, tunnel)
	}
}

func TestUpdateTunnelRoute(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		_, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
			  "network": "10.0.0.0/16",
			  "tunnel_id": "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
			  "tunnel_name": "blog",
			  "comment": "Example comment for this route",
			  "created_at": "2021-01-25T18:22:34.317854Z",
			  "deleted_at": "2021-01-25T18:22:34.317854Z",
			  "virtual_network_id": "9f322de4-5988-4945-b770-f1d6ac200f86"
            }
          }`)
	}

	ts, _ := time.Parse(time.RFC3339Nano, "2021-01-25T18:22:34.317854Z")
	want := TunnelRoute{
		Network:          "10.0.0.0/16",
		TunnelID:         "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
		TunnelName:       "blog",
		Comment:          "Example comment for this route",
		CreatedAt:        &ts,
		DeletedAt:        &ts,
		VirtualNetworkID: "9f322de4-5988-4945-b770-f1d6ac200f86",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/teamnet/routes/network/10.0.0.0/16", handler)
	tunnel, err := client.UpdateTunnelRoute(context.Background(), AccountIdentifier(testAccountID), TunnelRoutesUpdateParams{TunnelID: testTunnelID, Network: "10.0.0.0/16", Comment: "foo", VirtualNetworkID: "9f322de4-5988-4945-b770-f1d6ac200f86"})

	if assert.NoError(t, err) {
		assert.Equal(t, want, tunnel)
	}
}

func TestDeleteTunnelRoute(t *testing.T) {
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
			  "network": "ff01::/32",
			  "tunnel_id": "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
			  "tunnel_name": "blog",
			  "comment": "Example comment for this route",
			  "created_at": "2021-01-25T18:22:34.317854Z",
			  "deleted_at": "2021-01-25T18:22:34.317854Z",
			  "virtual_network_id": "9f322de4-5988-4945-b770-f1d6ac200f86"
            }
          }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/teamnet/routes/network/10.0.0.0/16", handler)
	err := client.DeleteTunnelRoute(context.Background(), AccountIdentifier(testAccountID), TunnelRoutesDeleteParams{Network: "10.0.0.0/16", VirtualNetworkID: "9f322de4-5988-4945-b770-f1d6ac200f86"})
	assert.NoError(t, err)
}
