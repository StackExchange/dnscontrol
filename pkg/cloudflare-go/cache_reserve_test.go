package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var cacheReserveTimestampString = "2019-02-20T22:37:07.107449Z"
var cacheReserveTimestamp, _ = time.Parse(time.RFC3339Nano, cacheReserveTimestampString)

func TestCacheReserve(t *testing.T) {
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
				"id": "cache_reserve",
				"value": "on",
				"modified_on": "%s"
			}
		}
		`, cacheReserveTimestampString)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/cache/cache_reserve", handler)
	want := CacheReserve{
		ID:         "cache_reserve",
		Value:      "on",
		ModifiedOn: cacheReserveTimestamp,
	}

	actual, err := client.GetCacheReserve(context.Background(), ZoneIdentifier("01a7362d577a6c3019a474fd6f485823"), GetCacheReserveParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateCacheReserve(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "cache_reserve",
				"value": "off",
				"modified_on": "%s"
			}
		}
		`, cacheReserveTimestampString)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/cache/cache_reserve", handler)
	want := CacheReserve{
		ID:         "cache_reserve",
		Value:      "off",
		ModifiedOn: cacheReserveTimestamp,
	}

	actual, err := client.UpdateCacheReserve(context.Background(), ZoneIdentifier("01a7362d577a6c3019a474fd6f485823"), UpdateCacheReserveParams{Value: "off"})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
