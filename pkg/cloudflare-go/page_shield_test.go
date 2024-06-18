package cloudflare

import (
	"context"
	"net/http"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

// Mock PageShieldSettings data.
var mockPageShieldSettings = PageShieldSettings{
	PageShield: PageShield{
		Enabled:                        BoolPtr(true),
		UseCloudflareReportingEndpoint: BoolPtr(true),
		UseConnectionURLPath:           BoolPtr(true),
	},
	UpdatedAt: "2022-10-12T17:56:52.083582+01:00",
}

func TestGetPageShieldSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/page_shield", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		response := PageShieldSettingsResponse{
			PageShield: mockPageShieldSettings,
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatal(err)
		}
	})

	result, err := client.GetPageShieldSettings(context.Background(), ZoneIdentifier(testZoneID), GetPageShieldSettingsParams{})
	assert.NoError(t, err)
	assert.Equal(t, &PageShieldSettingsResponse{PageShield: mockPageShieldSettings}, result)
}

func TestUpdatePageShieldSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/page_shield", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		var params PageShield
		err := json.NewDecoder(r.Body).Decode(&params)
		assert.NoError(t, err)

		response := PageShieldSettingsResponse{
			PageShield: PageShieldSettings{
				PageShield: params,
				UpdatedAt:  "2022-10-13T10:00:00.000Z",
			},
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatal(err)
		}
	})

	newSettings := UpdatePageShieldSettingsParams{
		Enabled:                        BoolPtr(false),
		UseCloudflareReportingEndpoint: BoolPtr(false),
		UseConnectionURLPath:           BoolPtr(false),
	}
	result, err := client.UpdatePageShieldSettings(context.Background(), ZoneIdentifier(testZoneID), newSettings)
	assert.NoError(t, err)
	assert.Equal(t, false, *result.PageShield.Enabled)
	assert.Equal(t, false, *result.PageShield.UseCloudflareReportingEndpoint)
	assert.Equal(t, false, *result.PageShield.UseConnectionURLPath)
}
