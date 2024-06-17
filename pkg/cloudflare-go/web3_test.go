package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testWeb3HostnameID = "9a7806061c88ada191ed06f989cc3dac"

func createTestWeb3Hostname() Web3Hostname {
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	return Web3Hostname{
		ID:          testWeb3HostnameID,
		Name:        "gateway.example.com",
		Description: "This is my IPFS gateway.",
		Status:      "active",
		Target:      "ipfs",
		Dnslink:     "/ipns/onboarding.ipfs.cloudflare.com",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
	}
}

func TestListWeb3Hostnames(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/web3/hostnames", testZoneID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		fmt.Fprint(w, `
	{
	  "success": true,
	  "errors": [],
	  "messages": [],
	  "result": [
		{
		  "id": "9a7806061c88ada191ed06f989cc3dac",
		  "name": "gateway.example.com",
		  "description": "This is my IPFS gateway.",
		  "status": "active",
		  "target": "ipfs",
		  "dnslink": "/ipns/onboarding.ipfs.cloudflare.com",
		  "created_on": "2014-01-01T05:20:00.12345Z",
		  "modified_on": "2014-01-01T05:20:00.12345Z"
		}
	  ]
	}`)
	})

	_, err := client.ListWeb3Hostnames(context.Background(), Web3HostnameListParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}
	out, err := client.ListWeb3Hostnames(context.Background(), Web3HostnameListParameters{ZoneID: testZoneID})
	assert.NoError(t, err, "Got error listing web3 hostnames")
	want := createTestWeb3Hostname()
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(out), "length of web3hosts is wrong")
		assert.Equal(t, want, out[0], "structs not equal")
	}
}

func TestCreateWeb3Hostname(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/web3/hostnames", testZoneID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		fmt.Fprint(w, `
	{
	  "success": true,
	  "errors": [],
	  "messages": [],
	  "result":
		{
		  "id": "9a7806061c88ada191ed06f989cc3dac",
		  "name": "gateway.example.com",
		  "description": "This is my IPFS gateway.",
		  "status": "active",
		  "target": "ipfs",
		  "dnslink": "/ipns/onboarding.ipfs.cloudflare.com",
		  "created_on": "2014-01-01T05:20:00.12345Z",
		  "modified_on": "2014-01-01T05:20:00.12345Z"
		}
	}`)
	})

	_, err := client.CreateWeb3Hostname(context.Background(), Web3HostnameCreateParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}

	_, err = client.CreateWeb3Hostname(context.Background(), Web3HostnameCreateParameters{ZoneID: testZoneID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingName, err)
	}

	_, err = client.CreateWeb3Hostname(context.Background(), Web3HostnameCreateParameters{ZoneID: testZoneID, Name: "gateway.example.com"})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingTarget, err)
	}

	out, err := client.CreateWeb3Hostname(context.Background(), Web3HostnameCreateParameters{ZoneID: testZoneID, Name: "gateway.example.com", Target: "ipfs"})
	assert.NoError(t, err, "Got error creating web3 hostname")
	want := createTestWeb3Hostname()
	if assert.NoError(t, err) {
		assert.Equal(t, want, out, "structs not equal")
	}
}

func TestGetWeb3Hostname(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/web3/hostnames/%s", testZoneID, testWeb3HostnameID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		fmt.Fprint(w, `
	{
	  "success": true,
	  "errors": [],
	  "messages": [],
	  "result": {
		"id": "9a7806061c88ada191ed06f989cc3dac",
		"name": "gateway.example.com",
		"description": "This is my IPFS gateway.",
		"status": "active",
		"target": "ipfs",
		"dnslink": "/ipns/onboarding.ipfs.cloudflare.com",
		"created_on": "2014-01-01T05:20:00.12345Z",
		"modified_on": "2014-01-01T05:20:00.12345Z"
	  }
	}`)
	})
	_, err := client.GetWeb3Hostname(context.Background(), Web3HostnameDetailsParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}
	_, err = client.GetWeb3Hostname(context.Background(), Web3HostnameDetailsParameters{ZoneID: testZoneID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingIdentifier, err)
	}

	out, err := client.GetWeb3Hostname(context.Background(), Web3HostnameDetailsParameters{ZoneID: testZoneID, Identifier: testWeb3HostnameID})
	assert.NoError(t, err, "Got error getting web3 hostname")
	want := createTestWeb3Hostname()
	if assert.NoError(t, err) {
		assert.Equal(t, want, out, "structs not equal")
	}
}

func TestUpdateWeb3Hostname(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/web3/hostnames/%s", testZoneID, testWeb3HostnameID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		fmt.Fprint(w, `
	{
	  "success": true,
	  "errors": [],
	  "messages": [],
	  "result": {
		"id": "9a7806061c88ada191ed06f989cc3dac",
		"name": "gateway.example.com",
		"description": "This is my IPFS gateway.",
		"status": "active",
		"target": "ipfs",
		"dnslink": "/ipns/onboarding.ipfs.cloudflare.com",
		"created_on": "2014-01-01T05:20:00.12345Z",
		"modified_on": "2014-01-01T05:20:00.12345Z"
	  }
	}`)
	})
	_, err := client.UpdateWeb3Hostname(context.Background(), Web3HostnameUpdateParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}
	_, err = client.UpdateWeb3Hostname(context.Background(), Web3HostnameUpdateParameters{ZoneID: testZoneID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingIdentifier, err)
	}

	out, err := client.UpdateWeb3Hostname(context.Background(), Web3HostnameUpdateParameters{ZoneID: testZoneID, Identifier: testWeb3HostnameID})
	assert.NoError(t, err, "Got error getting web3 hostname")
	want := createTestWeb3Hostname()
	if assert.NoError(t, err) {
		assert.Equal(t, want, out, "structs not equal")
	}
}

func TestDeleteWeb3Hostname(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/zones/%s/web3/hostnames/%s", testZoneID, testWeb3HostnameID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		fmt.Fprint(w, `
	{
	  "success": true,
	  "errors": [],
	  "messages": [],
	  "result": {
		"id": "9a7806061c88ada191ed06f989cc3dac"
	  }
	}
	`)
	})
	_, err := client.DeleteWeb3Hostname(context.Background(), Web3HostnameDetailsParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}
	_, err = client.DeleteWeb3Hostname(context.Background(), Web3HostnameDetailsParameters{ZoneID: testZoneID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingIdentifier, err)
	}

	out, err := client.DeleteWeb3Hostname(context.Background(), Web3HostnameDetailsParameters{ZoneID: testZoneID, Identifier: testWeb3HostnameID})
	assert.NoError(t, err, "Got error deleting web3 hostname")
	if assert.NoError(t, err) {
		assert.Equal(t, testWeb3HostnameID, out.ID, "delete web3 response incorrect")
	}
}
