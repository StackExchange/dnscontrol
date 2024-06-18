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

func TestZoneCacheVariants(t *testing.T) {
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
          "result": {
            "id": "variants",
            "modified_on": "2014-01-01T05:20:00.12345Z",
            "value": {
              "avif": [
                "image/webp",
                "image/jpeg"
              ],
              "bmp": [
                "image/webp",
                "image/jpeg"
              ]
            }
          }
        }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/cache/variants", handler)

	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	want := ZoneCacheVariants{
		ModifiedOn: modifiedOn,
		Value: ZoneCacheVariantsValues{
			Avif: []string{
				"image/webp",
				"image/jpeg",
			},
			Bmp: []string{
				"image/webp",
				"image/jpeg",
			},
		},
	}

	actual, err := client.ZoneCacheVariants(context.Background(), testZoneID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestZoneCacheVariantsUpdate(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if assert.NoError(t, err) {
			assert.JSONEq(t, `{"value":{"avif":["image/webp","image/jpeg"],"bmp":["image/webp","image/jpeg"]}}`, string(b))
		}

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
        {
		  "success": true,
          "errors": [],
          "messages": [],
          "result": {
		    "id": "variants",
            "modified_on": "2014-01-01T05:20:00.12345Z",
            "value": {
			  "avif": [
			    "image/webp",
                "image/jpeg"
              ],
			  "bmp": [
				"image/webp",
				"image/jpeg"
			  ]
            }
          }
        }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/cache/variants", handler)

	modifiedOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := ZoneCacheVariants{
		ModifiedOn: modifiedOn,
		Value: ZoneCacheVariantsValues{
			Avif: []string{
				"image/webp",
				"image/jpeg",
			},
			Bmp: []string{
				"image/webp",
				"image/jpeg",
			},
		},
	}

	zoneCacheVariants := ZoneCacheVariantsValues{
		Avif: []string{
			"image/webp",
			"image/jpeg",
		},
		Bmp: []string{
			"image/webp",
			"image/jpeg",
		}}

	actual, err := client.UpdateZoneCacheVariants(context.Background(), testZoneID, zoneCacheVariants)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestZoneCacheVariantsDelete(t *testing.T) {
	setup()
	defer teardown()

	apiCalled := false

	handler := func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
        {
		  "success": true,
          "errors": [],
          "messages": [],
          "result": {
            "id": "variants",
            "modified_on": "2014-01-01T05:20:00.12345Z"
          }
        }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/cache/variants", handler)

	err := client.DeleteZoneCacheVariants(context.Background(), testZoneID)

	assert.NoError(t, err)
	assert.True(t, apiCalled)
}
