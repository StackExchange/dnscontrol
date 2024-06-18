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

func TestTunnelVirtualNetworks(t *testing.T) {
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
			    "id": "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
				"name": "us-east-1-vpc",
				"is_default_network": true,
				"comment": "Staging VPC for data science",
				"created_at": "2021-01-25T18:22:34.317854Z",
				"deleted_at": "2021-01-25T18:22:34.317854Z"
              }
            ]
          }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/teamnet/virtual_networks", handler)

	ts, _ := time.Parse(time.RFC3339Nano, "2021-01-25T18:22:34.317854Z")

	want := []TunnelVirtualNetwork{
		{
			ID:               "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
			Name:             "us-east-1-vpc",
			IsDefaultNetwork: true,
			Comment:          "Staging VPC for data science",
			CreatedAt:        &ts,
			DeletedAt:        &ts,
		},
	}

	got, err := client.ListTunnelVirtualNetworks(context.Background(), AccountIdentifier(testAccountID), TunnelVirtualNetworksListParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestCreateTunnelVirtualNetwork(t *testing.T) {
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
			  "id": "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
			  "name": "us-east-1-vpc",
			  "is_default_network": true,
			  "comment": "Staging VPC for data science",
			  "created_at": "2021-01-25T18:22:34.317854Z",
			  "deleted_at": "2021-01-25T18:22:34.317854Z"
            }
          }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/teamnet/virtual_networks", handler)

	ts, _ := time.Parse(time.RFC3339Nano, "2021-01-25T18:22:34.317854Z")
	want := TunnelVirtualNetwork{
		ID:               "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
		Name:             "us-east-1-vpc",
		IsDefaultNetwork: true,
		Comment:          "Staging VPC for data science",
		CreatedAt:        &ts,
		DeletedAt:        &ts,
	}

	tunnel, err := client.CreateTunnelVirtualNetwork(context.Background(), AccountIdentifier(testAccountID), TunnelVirtualNetworkCreateParams{
		Name:      "us-east-1-vpc",
		IsDefault: true,
		Comment:   "Staging VPC for data science",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, tunnel)
	}
}

func TestUpdateTunnelVirtualNetwork(t *testing.T) {
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
			  "id": "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
			  "name": "us-east-1-vpc",
			  "is_default_network": true,
			  "comment": "Staging VPC for data science",
			  "created_at": "2021-01-25T18:22:34.317854Z",
			  "deleted_at": "2021-01-25T18:22:34.317854Z"
            }
          }`)
	}

	ts, _ := time.Parse(time.RFC3339Nano, "2021-01-25T18:22:34.317854Z")
	want := TunnelVirtualNetwork{
		ID:               "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
		Name:             "us-east-1-vpc",
		IsDefaultNetwork: true,
		Comment:          "Staging VPC for data science",
		CreatedAt:        &ts,
		DeletedAt:        &ts,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/teamnet/virtual_networks/f70ff985-a4ef-4643-bbbc-4a0ed4fc8415", handler)

	tunnel, err := client.UpdateTunnelVirtualNetwork(context.Background(), AccountIdentifier(testAccountID), TunnelVirtualNetworkUpdateParams{
		VnetID:           "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
		Name:             "us-east-1-vpc",
		IsDefaultNetwork: BoolPtr(true),
		Comment:          "Staging VPC for data science",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, tunnel)
	}
}

func TestDeleteTunnelVirtualNetwork(t *testing.T) {
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
			  "id": "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
			  "name": "us-east-1-vpc",
			  "is_default_network": true,
			  "comment": "Staging VPC for data science",
			  "created_at": "2021-01-25T18:22:34.317854Z",
			  "deleted_at": "2021-01-25T18:22:34.317854Z"
            }
          }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/teamnet/virtual_networks/f70ff985-a4ef-4643-bbbc-4a0ed4fc8415", handler)

	err := client.DeleteTunnelVirtualNetwork(context.Background(), AccountIdentifier(testAccountID), "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415")

	assert.NoError(t, err)
}
