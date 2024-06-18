package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testID = "f174e90a-fafe-4643-bbbc-4a0ed4fc8415"

var dexTimestamp, _ = time.Parse(time.RFC3339, "2023-01-30T19:59:44.401278Z")

func TestGetDeviceDexTests(t *testing.T) {
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
				"dex_tests": [
					{
						"test_id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
						"name": "http test dash",
						"description": "dex test description",
						"interval": "0h30m0s",
						"enabled": true,
						"data": {
						"host": "https://dash.cloudflare.com",
						"kind": "http",
						"method": "GET"
						},
						"updated": "2023-01-30T19:59:44.401278Z",
						"created": "2023-01-30T19:59:44.401278Z"
					}
				]
			}
		  }`)
	}

	dexTest := []DeviceDexTest{{
		TestID:      testID,
		Name:        "http test dash",
		Description: "dex test description",
		Interval:    "0h30m0s",
		Enabled:     true,
		Data: &DeviceDexTestData{
			"kind":   "http",
			"method": "GET",
			"host":   "https://dash.cloudflare.com",
		},
		Updated: dexTimestamp,
		Created: dexTimestamp,
	}}

	want := DeviceDexTests{
		DexTests: dexTest,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/dex_tests", handler)

	actual, err := client.ListDexTests(context.Background(), AccountIdentifier(testAccountID), ListDeviceDexTestParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeviceDexTest(t *testing.T) {
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
				"test_id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "http test dash",
				"description": "dex test description",
				"interval": "0h30m0s",
				"enabled": true,
				"data": {
				"host": "https://dash.cloudflare.com",
				"kind": "http",
				"method": "GET"
				},
				"updated": "2023-01-30T19:59:44.401278Z",
				"created": "2023-01-30T19:59:44.401278Z"
			}
		  }`)
	}

	want := DeviceDexTest{
		TestID:      testID,
		Name:        "http test dash",
		Description: "dex test description",
		Interval:    "0h30m0s",
		Enabled:     true,
		Data: &DeviceDexTestData{
			"kind":   "http",
			"method": "GET",
			"host":   "https://dash.cloudflare.com",
		},
		Updated: dexTimestamp,
		Created: dexTimestamp,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/dex_tests/"+testID, handler)

	actual, err := client.GetDeviceDexTest(context.Background(), AccountIdentifier(testAccountID), testID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateDeviceDexTest(t *testing.T) {
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
				"test_id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "http test dash",
				"description": "dex test description",
				"interval": "0h30m0s",
				"enabled": true,
				"data": {
				"host": "https://dash.cloudflare.com",
				"kind": "http",
				"method": "GET"
				},
				"updated": "2023-01-30T19:59:44.401278Z",
				"created": "2023-01-30T19:59:44.401278Z"
			}
		  }`)
	}

	want := DeviceDexTest{
		TestID:      testID,
		Name:        "http test dash",
		Description: "dex test description",
		Interval:    "0h30m0s",
		Enabled:     true,
		Data: &DeviceDexTestData{
			"kind":   "http",
			"method": "GET",
			"host":   "https://dash.cloudflare.com",
		},
		Updated: dexTimestamp,
		Created: dexTimestamp,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/dex_tests", handler)

	actual, err := client.CreateDeviceDexTest(context.Background(), AccountIdentifier(testAccountID), CreateDeviceDexTestParams{
		TestID:      testID,
		Name:        "http test dash",
		Description: "dex test description",
		Interval:    "0h30m0s",
		Enabled:     true,
		Data: &DeviceDexTestData{
			"kind":   "http",
			"method": "GET",
			"host":   "https://dash.cloudflare.com",
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateDeviceDexTest(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"test_id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"name": "http test dash",
				"description": "dex test description",
				"interval": "0h30m0s",
				"enabled": true,
				"data": {
				"host": "https://dash.cloudflare.com",
				"kind": "http",
				"method": "GET"
				},
				"updated": "2023-01-30T19:59:44.401278Z",
				"created": "2023-01-30T19:59:44.401278Z"
			}
		  }`)
	}

	want := DeviceDexTest{
		TestID:      testID,
		Name:        "http test dash",
		Description: "dex test description",
		Interval:    "0h30m0s",
		Enabled:     true,
		Data: &DeviceDexTestData{
			"kind":   "http",
			"method": "GET",
			"host":   "https://dash.cloudflare.com",
		},
		Updated: dexTimestamp,
		Created: dexTimestamp,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/dex_tests/"+testID, handler)

	actual, err := client.UpdateDeviceDexTest(context.Background(), AccountIdentifier(testAccountID), UpdateDeviceDexTestParams{
		TestID:      testID,
		Name:        "http test dash",
		Description: "dex test description",
		Interval:    "0h30m0s",
		Enabled:     true,
		Data: &DeviceDexTestData{
			"kind":   "http",
			"method": "GET",
			"host":   "https://dash.cloudflare.com",
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteDeviceDexTest(t *testing.T) {
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
				"dex_tests": []
			}
		  }`)
	}

	want := DeviceDexTests{
		DexTests: []DeviceDexTest{},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/dex_tests/"+testID, handler)

	actual, err := client.DeleteDexTest(context.Background(), AccountIdentifier(testAccountID), testID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
