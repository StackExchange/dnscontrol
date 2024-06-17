package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPIShield(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
	"success" : true,
	"errors": [],
	"messages": [],
	"result": {
		"auth_id_characteristics": [
			{
				"type": "header",
				"name": "test-header"
			},
			{
				"type": "cookie",
				"name": "test-cookie"
			}
		]
	}
}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/api_gateway/configuration", handler)

	var authChars []AuthIdCharacteristics
	authChars = append(authChars, AuthIdCharacteristics{Type: "header", Name: "test-header"})
	authChars = append(authChars, AuthIdCharacteristics{Type: "cookie", Name: "test-cookie"})

	want := APIShield{
		AuthIdCharacteristics: authChars,
	}

	actual, _, err := client.GetAPIShieldConfiguration(context.Background(), ZoneIdentifier(testZoneID))

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestPutAPIShield(t *testing.T) {
	setup()
	defer teardown()

	// now lets do a PUT
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
	"success" : true,
	"errors": [],
	"messages": [],
	"result": []
}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/api_gateway/configuration", handler)

	apiShieldData := UpdateAPIShieldParams{AuthIdCharacteristics: []AuthIdCharacteristics{{Type: "header", Name: "different-header"}, {Type: "cookie", Name: "different-cookie"}}}

	want := Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}}

	actual, err := client.UpdateAPIShieldConfiguration(context.Background(), ZoneIdentifier(testZoneID), apiShieldData)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
