package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testNetworkID = "f174e90a-fafe-4643-bbbc-4a0ed4fc8415"

func TestGetDeviceManagedNetworks(t *testing.T) {
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
				"network_id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"type": "tls",
				"name": "managed-network-1",
				"config": {
				  "tls_sockaddr": "foobar:1234",
				  "sha256": "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"
				}
			  }
			]
		  }`)
	}

	want := []DeviceManagedNetwork{{
		NetworkID: testNetworkID,
		Type:      "tls",
		Name:      "managed-network-1",
		Config: &Config{
			TlsSockAddr: "foobar:1234",
			Sha256:      "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		},
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/networks", handler)

	actual, err := client.ListDeviceManagedNetworks(context.Background(), AccountIdentifier(testAccountID), ListDeviceManagedNetworksParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeviceManagedNetwork(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result":
			  {
				"network_id": "%s",
				"type": "tls",
				"name": "managed-network-1",
				"config": {
				  "tls_sockaddr": "foobar:1234",
				  "sha256": "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"
				}
			  }
		  }`, testNetworkID)
	}

	want := DeviceManagedNetwork{
		NetworkID: testNetworkID,
		Type:      "tls",
		Name:      "managed-network-1",
		Config: &Config{
			TlsSockAddr: "foobar:1234",
			Sha256:      "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/networks/"+testNetworkID, handler)

	actual, err := client.GetDeviceManagedNetwork(context.Background(), AccountIdentifier(testAccountID), testNetworkID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateDeviceManagedNetwork(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result":
			  {
				"network_id": "%s",
				"type": "tls",
				"name": "managed-network-1",
				"config": {
				  "tls_sockaddr": "foobar:1234",
				  "sha256": "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"
				}
			  }
		  }`, testNetworkID)
	}

	want := DeviceManagedNetwork{
		NetworkID: testNetworkID,
		Type:      "tls",
		Name:      "managed-network-1",
		Config: &Config{
			TlsSockAddr: "foobar:1234",
			Sha256:      "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/networks", handler)

	actual, err := client.CreateDeviceManagedNetwork(context.Background(), AccountIdentifier(testAccountID), CreateDeviceManagedNetworkParams{
		NetworkID: testNetworkID,
		Type:      "tls",
		Name:      "managed-network-1",
		Config: &Config{
			TlsSockAddr: "foobar:1234",
			Sha256:      "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateDeviceManagedNetwork(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result":
			  {
				"network_id": "%s",
				"type": "tls",
				"name": "managed-network-1",
				"config": {
				  "tls_sockaddr": "foobar:1234",
				  "sha256": "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"
				}
			  }
		  }`, testNetworkID)
	}

	want := DeviceManagedNetwork{
		NetworkID: testNetworkID,
		Type:      "tls",
		Name:      "managed-network-1",
		Config: &Config{
			TlsSockAddr: "foobar:1234",
			Sha256:      "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/networks/"+testNetworkID, handler)

	actual, err := client.UpdateDeviceManagedNetwork(context.Background(), AccountIdentifier(testAccountID), UpdateDeviceManagedNetworkParams{
		NetworkID: testNetworkID,
		Type:      "tls",
		Name:      "managed-network-1",
		Config: &Config{
			TlsSockAddr: "foobar:1234",
			Sha256:      "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteDeviceManagedNetwork(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
			  {
				"network_id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"type": "tls",
				"name": "managed-network-1",
				"config": {
				  "tls_sockaddr": "foobar:1234",
				  "sha256": "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"
				}
			  }
			]
		  }`)
	}

	want := []DeviceManagedNetwork{{
		NetworkID: testNetworkID,
		Type:      "tls",
		Name:      "managed-network-1",
		Config: &Config{
			TlsSockAddr: "foobar:1234",
			Sha256:      "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
		},
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/networks/"+testNetworkID, handler)

	actual, err := client.DeleteManagedNetworks(context.Background(), AccountIdentifier(testAccountID), testNetworkID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
