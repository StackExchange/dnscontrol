package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSecondaryDNSZone(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "269d8f4853475ca241c4e730be286b20",
				"name": "www.example.com.",
				"primaries": [
					"23ff594956f20c2a721606e94745a8aa",
					"00920f38ce07c2e2f4df50b1f61d4194"
				],
				"auto_refresh_seconds": 86400,
				"soa_serial": 2019102400,
				"created_time": "2019-10-24T17:09:42.883908+01:00",
				"checked_time": "2019-10-24T17:09:42.883908+01:00",
				"modified_time": "2019-10-24T17:09:42.883908+01:00"
			}
		}
		`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/secondary_dns", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2019-10-24T17:09:42.883908+01:00")
	modifiedOn, _ := time.Parse(time.RFC3339, "2019-10-24T17:09:42.883908+01:00")
	checkedOn, _ := time.Parse(time.RFC3339, "2019-10-24T17:09:42.883908+01:00")
	want := SecondaryDNSZone{
		ID:   "269d8f4853475ca241c4e730be286b20",
		Name: "www.example.com.",
		Primaries: []string{
			"23ff594956f20c2a721606e94745a8aa",
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		AutoRefreshSeconds: 86400,
		SoaSerial:          2019102400,
		CreatedTime:        createdOn,
		CheckedTime:        checkedOn,
		ModifiedTime:       modifiedOn,
	}

	actual, err := client.GetSecondaryDNSZone(context.Background(), "01a7362d577a6c3019a474fd6f485823")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateSecondaryDNSZone(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "269d8f4853475ca241c4e730be286b20",
				"name": "www.example.com.",
				"primaries": [
					"23ff594956f20c2a721606e94745a8aa",
					"00920f38ce07c2e2f4df50b1f61d4194"
				],
				"auto_refresh_seconds": 86400,
				"soa_serial": 2019102400,
				"created_time": "2019-10-24T17:09:42.883908+01:00",
				"checked_time": "2019-10-24T17:09:42.883908+01:00",
				"modified_time": "2019-10-24T17:09:42.883908+01:00"
			}
		}
		`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/secondary_dns", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2019-10-24T17:09:42.883908+01:00")
	modifiedOn, _ := time.Parse(time.RFC3339, "2019-10-24T17:09:42.883908+01:00")
	checkedOn, _ := time.Parse(time.RFC3339, "2019-10-24T17:09:42.883908+01:00")
	want := SecondaryDNSZone{
		ID:   "269d8f4853475ca241c4e730be286b20",
		Name: "www.example.com.",
		Primaries: []string{
			"23ff594956f20c2a721606e94745a8aa",
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		AutoRefreshSeconds: 86400,
		SoaSerial:          2019102400,
		CreatedTime:        createdOn,
		CheckedTime:        checkedOn,
		ModifiedTime:       modifiedOn,
	}

	newSecondaryZone := SecondaryDNSZone{
		Name: "www.example.com.",
		Primaries: []string{
			"23ff594956f20c2a721606e94745a8aa",
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		AutoRefreshSeconds: 86400,
	}

	actual, err := client.CreateSecondaryDNSZone(context.Background(), "01a7362d577a6c3019a474fd6f485823", newSecondaryZone)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateSecondaryDNSZone(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "269d8f4853475ca241c4e730be286b20",
				"name": "www.example.com.",
				"primaries": [
					"23ff594956f20c2a721606e94745a8aa",
					"00920f38ce07c2e2f4df50b1f61d4194"
				],
				"auto_refresh_seconds": 86400,
				"soa_serial": 2019102400,
				"created_time": "2019-10-24T17:09:42.883908+01:00",
				"checked_time": "2019-10-24T17:09:42.883908+01:00",
				"modified_time": "2019-10-24T17:09:42.883908+01:00"
			}
		}
		`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/secondary_dns", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2019-10-24T17:09:42.883908+01:00")
	modifiedOn, _ := time.Parse(time.RFC3339, "2019-10-24T17:09:42.883908+01:00")
	checkedOn, _ := time.Parse(time.RFC3339, "2019-10-24T17:09:42.883908+01:00")
	want := SecondaryDNSZone{
		ID:   "269d8f4853475ca241c4e730be286b20",
		Name: "www.example.com.",
		Primaries: []string{
			"23ff594956f20c2a721606e94745a8aa",
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		AutoRefreshSeconds: 86400,
		SoaSerial:          2019102400,
		CreatedTime:        createdOn,
		CheckedTime:        checkedOn,
		ModifiedTime:       modifiedOn,
	}

	updatedZone := SecondaryDNSZone{
		Name: "www.example.com.",
		Primaries: []string{
			"23ff594956f20c2a721606e94745a8aa",
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		AutoRefreshSeconds: 86400,
	}

	actual, err := client.UpdateSecondaryDNSZone(context.Background(), "01a7362d577a6c3019a474fd6f485823", updatedZone)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteSecondaryDNSZone(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "269d8f4853475ca241c4e730be286b20"
			}
		}
		`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/secondary_dns", handler)

	err := client.DeleteSecondaryDNSZone(context.Background(), "01a7362d577a6c3019a474fd6f485823")
	assert.NoError(t, err)
}

func TestForceSecondaryDNSZoneAXFR(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": "OK"
		}
		`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/secondary_dns/force_axfr", handler)

	err := client.ForceSecondaryDNSZoneAXFR(context.Background(), "01a7362d577a6c3019a474fd6f485823")
	assert.NoError(t, err)
}

func TestValidateRequiredSecondaryDNSZoneValues(t *testing.T) {
	z1 := SecondaryDNSZone{}
	err1 := validateRequiredSecondaryDNSZoneValues(z1)
	assert.EqualError(t, err1, errSecondaryDNSInvalidZoneName)

	z2 := SecondaryDNSZone{Name: "example.com."}
	err2 := validateRequiredSecondaryDNSZoneValues(z2)
	assert.EqualError(t, err2, errSecondaryDNSInvalidAutoRefreshValue)

	z3 := SecondaryDNSZone{Name: "example.com.", AutoRefreshSeconds: 1}
	err3 := validateRequiredSecondaryDNSZoneValues(z3)
	assert.EqualError(t, err3, errSecondaryDNSInvalidPrimaries)

	z4 := SecondaryDNSZone{Name: "example.com.", AutoRefreshSeconds: 1, Primaries: []string{"a", "b"}}
	err4 := validateRequiredSecondaryDNSZoneValues(z4)
	assert.NoError(t, err4)
}
