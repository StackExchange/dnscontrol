package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createSmartTieredCacheHandler(val string, lastModified string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"editable": true,
				"id": "tiered_cache_smart_topology_enable",
				"modified_on": "%s",
				"value": "%s"
            }
          }`, lastModified, val)
	}
}

func nonexistentSmartTieredCacheHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(404)
		fmt.Fprintf(w, `{
			"result": null,
			"success": false,
			"errors": [
				{
					"code": 1142,
					"message": "Unable to retrieve tiered_cache_smart_topology_enable setting value. The zone setting does not exist."
				}
			],
			"messages": []
		}`)
	}
}

func createGenericTieredCacheHandler(val string, lastModified string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "tiered_caching",
				"value": "%s",
				"modified_on": "%s",
				"editable": false
            }
          }`, val, lastModified)
	}
}

func TestGetTieredCache(t *testing.T) {
	t.Run("can identify when Smart Tiered Cache", func(t *testing.T) {
		t.Run("is disabled", func(t *testing.T) {
			setup()
			defer teardown()

			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("on", lastModified))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", createSmartTieredCacheHandler("off", lastModified))

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheGeneric,
				LastModified: wanted,
			}

			got, err := client.GetTieredCache(context.Background(), ZoneIdentifier(testZoneID))

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})

		t.Run("is enabled", func(t *testing.T) {
			setup()
			defer teardown()

			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("on", lastModified))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", createSmartTieredCacheHandler("on", lastModified))

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheSmart,
				LastModified: wanted,
			}

			got, err := client.GetTieredCache(context.Background(), ZoneIdentifier(testZoneID))

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})

		t.Run("zone setting does not exist", func(t *testing.T) {
			setup()
			defer teardown()

			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("on", lastModified))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", nonexistentSmartTieredCacheHandler())

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheGeneric,
				LastModified: wanted,
			}

			got, err := client.GetTieredCache(context.Background(), ZoneIdentifier(testZoneID))

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})
	})

	t.Run("can identify when generic tiered cache", func(t *testing.T) {
		t.Run("is disabled", func(t *testing.T) {
			setup()
			defer teardown()

			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("off", lastModified))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", createSmartTieredCacheHandler("on", lastModified))

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheOff,
				LastModified: wanted,
			}

			got, err := client.GetTieredCache(context.Background(), ZoneIdentifier(testZoneID))

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})

		t.Run("is enabled", func(t *testing.T) {
			setup()
			defer teardown()

			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("on", lastModified))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", nonexistentSmartTieredCacheHandler())

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheGeneric,
				LastModified: wanted,
			}

			got, err := client.GetTieredCache(context.Background(), ZoneIdentifier(testZoneID))

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})
	})

	t.Run("determines the latest last modified when", func(t *testing.T) {
		t.Run("smart tiered cache zone setting does not exist", func(t *testing.T) {
			setup()
			defer teardown()

			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("on", lastModified))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", nonexistentSmartTieredCacheHandler())

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheGeneric,
				LastModified: wanted,
			}

			got, err := client.GetTieredCache(context.Background(), ZoneIdentifier(testZoneID))

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})

		t.Run("generic tiered cache was modified more recently", func(t *testing.T) {
			setup()
			defer teardown()

			earlier := time.Now().Add(time.Minute * -5).Format(time.RFC3339)
			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("on", earlier))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", createSmartTieredCacheHandler("on", lastModified))

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheSmart,
				LastModified: wanted,
			}

			got, err := client.GetTieredCache(context.Background(), ZoneIdentifier(testZoneID))

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})

		t.Run("smart tiered cache was modified more recently", func(t *testing.T) {
			setup()
			defer teardown()

			earlier := time.Now().Add(time.Minute * -5).Format(time.RFC3339)
			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("on", lastModified))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", createSmartTieredCacheHandler("on", earlier))

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheSmart,
				LastModified: wanted,
			}

			got, err := client.GetTieredCache(context.Background(), ZoneIdentifier(testZoneID))

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})
	})
}

func TestSetTieredCache(t *testing.T) {
	t.Run("can enable tiered caching", func(t *testing.T) {
		t.Run("using smart caching", func(t *testing.T) {
			setup()
			defer teardown()

			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("on", lastModified))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", createSmartTieredCacheHandler("on", lastModified))

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheSmart,
				LastModified: wanted,
			}

			got, err := client.SetTieredCache(context.Background(), ZoneIdentifier(testZoneID), TieredCacheSmart)

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})

		t.Run("use generic caching", func(t *testing.T) {
			setup()
			defer teardown()

			lastModified := time.Now().Format(time.RFC3339)

			mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("on", lastModified))
			mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", nonexistentSmartTieredCacheHandler())

			wanted, _ := time.Parse(time.RFC3339, lastModified)
			want := TieredCache{
				Type:         TieredCacheGeneric,
				LastModified: wanted,
			}

			got, err := client.SetTieredCache(context.Background(), ZoneIdentifier(testZoneID), TieredCacheGeneric)

			if assert.NoError(t, err) {
				assert.Equal(t, want, got)
			}
		})
	})
}

func TestDeleteTieredCache(t *testing.T) {
	t.Run("can disable tiered caching", func(t *testing.T) {
		setup()
		defer teardown()

		lastModified := time.Now().Format(time.RFC3339)

		mux.HandleFunc("/zones/"+testZoneID+"/argo/tiered_caching", createGenericTieredCacheHandler("off", lastModified))
		mux.HandleFunc("/zones/"+testZoneID+"/cache/tiered_cache_smart_topology_enable", nonexistentSmartTieredCacheHandler())

		wanted, _ := time.Parse(time.RFC3339, lastModified)
		want := TieredCache{
			Type:         TieredCacheOff,
			LastModified: wanted,
		}

		got, err := client.GetTieredCache(context.Background(), ZoneIdentifier(testZoneID))

		if assert.NoError(t, err) {
			assert.Equal(t, want, got)
		}
	})
}
