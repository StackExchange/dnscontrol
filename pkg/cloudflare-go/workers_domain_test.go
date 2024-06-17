package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testWorkerDomainID = "dbe10b4bc17c295377eabd600e1787fd"

var expectedWorkerDomain = WorkersDomain{
	ID:          testWorkerDomainID,
	ZoneID:      "593c9c94de529bbbfaac7c53ced0447d",
	ZoneName:    "example.com",
	Hostname:    "foo.example.com",
	Service:     "foo",
	Environment: "production",
}

func TestWorkersDomain_GetDomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/domains/%s", testAccountID, testWorkerDomainID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": {
			"id": "dbe10b4bc17c295377eabd600e1787fd",
			"zone_id": "593c9c94de529bbbfaac7c53ced0447d",
			"zone_name": "example.com",
			"hostname": "foo.example.com",
			"service": "foo",
			"environment": "production"
		  }
		}`)
	})
	_, err := client.GetWorkersDomain(context.Background(), AccountIdentifier(""), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}
	res, err := client.GetWorkersDomain(context.Background(), AccountIdentifier(testAccountID), testWorkerDomainID)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedWorkerDomain, res)
	}
}

func TestWorkersDomain_ListDomains(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/domains", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": [
			{
			  "id": "dbe10b4bc17c295377eabd600e1787fd",
			  "zone_id": "593c9c94de529bbbfaac7c53ced0447d",
			  "zone_name": "example.com",
			  "hostname": "foo.example.com",
			  "service": "foo",
			  "environment": "production"
			}
		  ]
		}`)
	})
	_, err := client.ListWorkersDomains(context.Background(), AccountIdentifier(""), ListWorkersDomainParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}
	res, err := client.ListWorkersDomains(context.Background(), AccountIdentifier(testAccountID), ListWorkersDomainParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(res))
		assert.Equal(t, expectedWorkerDomain, res[0])
	}
}

func TestWorkersDomain_AttachDomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/domains", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": {
			"id": "dbe10b4bc17c295377eabd600e1787fd",
			"zone_id": "593c9c94de529bbbfaac7c53ced0447d",
			"zone_name": "example.com",
			"hostname": "foo.example.com",
			"service": "foo",
			"environment": "production"
		  }
		}`)
	})
	_, err := client.AttachWorkersDomain(context.Background(), AccountIdentifier(""), AttachWorkersDomainParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.AttachWorkersDomain(context.Background(), AccountIdentifier(testAccountID), AttachWorkersDomainParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingZoneID, err)
	}

	_, err = client.AttachWorkersDomain(context.Background(), AccountIdentifier(testAccountID), AttachWorkersDomainParams{
		ZoneID: testZoneID,
	})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingHostname, err)
	}

	_, err = client.AttachWorkersDomain(context.Background(), AccountIdentifier(testAccountID), AttachWorkersDomainParams{
		ZoneID:   testZoneID,
		Hostname: "foo.example.com",
	})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingService, err)
	}

	_, err = client.AttachWorkersDomain(context.Background(), AccountIdentifier(testAccountID), AttachWorkersDomainParams{
		ZoneID:   testZoneID,
		Hostname: "foo.example.com",
		Service:  "foo",
	})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingEnvironment, err)
	}

	res, err := client.AttachWorkersDomain(context.Background(), AccountIdentifier(testAccountID), AttachWorkersDomainParams{
		ZoneID:      testZoneID,
		Hostname:    "foo.example.com",
		Service:     "foo",
		Environment: "production",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, expectedWorkerDomain, res)
	}
}

func TestWorkersDomain_DetachDomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/domains/%s", testAccountID, testWorkerDomainID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": {
			"id": "dbe10b4bc17c295377eabd600e1787fd",
			"zone_id": "593c9c94de529bbbfaac7c53ced0447d",
			"zone_name": "example.com",
			"hostname": "foo.example.com",
			"service": "foo",
			"environment": "production"
		  }
		}`)
	})
	err := client.DetachWorkersDomain(context.Background(), AccountIdentifier(""), testWorkerDomainID)
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}
	err = client.DetachWorkersDomain(context.Background(), AccountIdentifier(testAccountID), testWorkerDomainID)
	assert.NoError(t, err)
}
