package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamsDevicesList(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
        {
          "success": true,
          "errors": [],
          "messages": [],
          "result": [
            {
              "id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
              "user": {
                "id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
                "name": "John Appleseed",
                "email": "user@example.com"
              },
              "key": "yek0SUYoOQ10vMGsIYAevozXUQpQtNFJFfFGqER/BGc=",
              "device_type": "windows",
              "name": "My mobile device",
              "model": "MyPhone(pro-X)",
              "manufacturer": "My phone corp",
              "deleted": true,
              "version": "1.0.0",
              "serial_number": "EXAMPLEHMD6R",
              "os_version": "10.0.0",
              "os_distro_name": "ubuntu",
              "os_distro_revision": "1.0.0",
			  "os_version_extra": "(a)",
              "mac_address": "00-00-5E-00-53-00",
              "ip": "192.0.2.1",
              "created": "2017-06-14T00:00:00Z",
              "updated": "2017-06-14T00:00:00Z",
              "last_seen": "2017-06-14T00:00:00Z",
              "revoked_at": "2017-06-14T00:00:00Z"
            }
          ]
        }
    `)
	}

	want := []TeamsDeviceListItem{{
		ID: "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		User: UserItem{
			ID:    "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
			Name:  "John Appleseed",
			Email: "user@example.com",
		},
		Key:              "yek0SUYoOQ10vMGsIYAevozXUQpQtNFJFfFGqER/BGc=",
		DeviceType:       "windows",
		Name:             "My mobile device",
		Model:            "MyPhone(pro-X)",
		Manufacturer:     "My phone corp",
		Deleted:          true,
		Version:          "1.0.0",
		SerialNumber:     "EXAMPLEHMD6R",
		OSVersion:        "10.0.0",
		OSDistroName:     "ubuntu",
		OsDistroRevision: "1.0.0",
		OSVersionExtra:   "(a)",
		MacAddress:       "00-00-5E-00-53-00",
		IP:               "192.0.2.1",
		Created:          "2017-06-14T00:00:00Z",
		Updated:          "2017-06-14T00:00:00Z",
		LastSeen:         "2017-06-14T00:00:00Z",
		RevokedAt:        "2017-06-14T00:00:00Z",
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices", handler)

	actual, err := client.ListTeamsDevices(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestRevokeTeamsDevices(t *testing.T) {
	setup()
	defer teardown()

	deviceIds := []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415", "g174e90a-fafe-4643-bbbc-4a0ed4fc8415"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "result": null,
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/revoke", handler)

	want := Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}}

	actual, err := client.RevokeTeamsDevices(context.Background(), testAccountID, deviceIds)
	require.NoError(t, err)
	assert.Equal(t, want, actual)
}

func TestGetTeamsDeviceDetails(t *testing.T) {
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
        "id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
        "user": {
          "id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
          "name": "John Appleseed",
          "email": "user@example.com"
        },
        "key": "yek0SUYoOQ10vMGsIYAevozXUQpQtNFJFfFGqER/BGc=",
        "device_type": "windows",
        "name": "My mobile device",
        "model": "MyPhone(pro-X)",
        "manufacturer": "My phone corp",
        "deleted": true,
        "version": "1.0.0",
        "serial_number": "EXAMPLEHMD6R",
        "os_version": "10.0.0",
        "mac_address": "00-00-5E-00-53-00",
        "ip": "192.0.2.1",
        "created": "2017-06-14T00:00:00Z",
        "updated": "2017-06-14T00:00:00Z",
        "last_seen": "2017-06-14T00:00:00Z",
        "revoked_at": "2017-06-14T00:00:00Z"
      }
    }
    `)
	}

	want := TeamsDeviceListItem{
		ID: "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
		User: UserItem{
			ID:    "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
			Name:  "John Appleseed",
			Email: "user@example.com",
		},
		Key:          "yek0SUYoOQ10vMGsIYAevozXUQpQtNFJFfFGqER/BGc=",
		DeviceType:   "windows",
		Name:         "My mobile device",
		Model:        "MyPhone(pro-X)",
		Manufacturer: "My phone corp",
		Deleted:      true,
		Version:      "1.0.0",
		SerialNumber: "EXAMPLEHMD6R",
		OSVersion:    "10.0.0",
		MacAddress:   "00-00-5E-00-53-00",
		IP:           "192.0.2.1",
		Created:      "2017-06-14T00:00:00Z",
		Updated:      "2017-06-14T00:00:00Z",
		LastSeen:     "2017-06-14T00:00:00Z",
		RevokedAt:    "2017-06-14T00:00:00Z",
	}

	deviceID := "f174e90a-fafe-4643-bbbc-4a0ed4fc8415"

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/"+deviceID, handler)

	actual, err := client.GetTeamsDeviceDetails(context.Background(), testAccountID, deviceID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
