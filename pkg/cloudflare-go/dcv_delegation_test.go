package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDCVDelegation(t *testing.T) {
	setup()
	defer teardown()

	testUuid := "b9ab465427f949ed"

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
	"success" : true,
	"errors": [],
	"messages": [],
	"result": {
		"uuid": "%s"
	}
}
		`, testUuid)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/dcv_delegation/uuid", handler)

	want := DCVDelegation{
		UUID: testUuid,
	}

	actual, _, err := client.GetDCVDelegation(context.Background(), ZoneIdentifier(testZoneID), GetDCVDelegationParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
