package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const regionalHostname = "eu.example.com"

func TestListRegions(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"result": [
				{
				  "key": "ca",
				  "label": "Canada"
				},
				{
				  "key": "eu",
				  "label": "Europe"
				}
			],
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/regional_hostnames/regions", handler)

	want := []Region{
		{
			Key:   "ca",
			Label: "Canada",
		},
		{
			Key:   "eu",
			Label: "Europe",
		},
	}

	actual, err := client.ListDataLocalizationRegions(context.Background(), AccountIdentifier(testAccountID), ListDataLocalizationRegionsParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListRegionalHostnames(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": [
			{
			  "hostname": "%s",
			  "region_key": "ca",
			  "created_on": "2023-01-14T00:47:57.060267Z"
			}
		  ],
		  "success": true,
		  "errors": [],
		  "messages": []
		}`, regionalHostname)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/addressing/regional_hostnames", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2023-01-14T00:47:57.060267Z")
	want := []RegionalHostname{
		{
			Hostname:  regionalHostname,
			RegionKey: "ca",
			CreatedOn: &createdOn,
		},
	}

	actual, err := client.ListDataLocalizationRegionalHostnames(context.Background(), ZoneIdentifier(testZoneID), ListDataLocalizationRegionalHostnamesParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateRegionalHostname(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": {
			  "hostname": "%s",
			  "region_key": "ca",
			  "created_on": "2023-01-14T00:47:57.060267Z"
		  },
		  "success": true,
		  "errors": [],
		  "messages": []
		}`, regionalHostname)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/addressing/regional_hostnames", handler)

	params := CreateDataLocalizationRegionalHostnameParams{
		RegionKey: "ca",
		Hostname:  regionalHostname,
	}

	want := RegionalHostname{
		RegionKey: "ca",
		Hostname:  regionalHostname,
	}

	actual, err := client.CreateDataLocalizationRegionalHostname(context.Background(), ZoneIdentifier(testZoneID), params)
	createdOn, _ := time.Parse(time.RFC3339, "2023-01-14T00:47:57.060267Z")
	want.CreatedOn = &createdOn
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetRegionalHostname(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": {
			  "hostname": "%s",
			  "region_key": "ca",
			  "created_on": "2023-01-14T00:47:57.060267Z"
		  },
		  "success": true,
		  "errors": [],
		  "messages": []
		}`, regionalHostname)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/addressing/regional_hostnames/"+regionalHostname, handler)

	actual, err := client.GetDataLocalizationRegionalHostname(context.Background(), ZoneIdentifier(testZoneID), regionalHostname)
	createdOn, _ := time.Parse(time.RFC3339, "2023-01-14T00:47:57.060267Z")
	want := RegionalHostname{
		RegionKey: "ca",
		Hostname:  regionalHostname,
		CreatedOn: &createdOn,
	}
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateRegionalHostname(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": {
			  "hostname": "%s",
			  "region_key": "eu",
			  "created_on": "2023-01-14T00:47:57.060267Z"
		  },
		  "success": true,
		  "errors": [],
		  "messages": []
		}`, regionalHostname)
	}

	params := UpdateDataLocalizationRegionalHostnameParams{
		RegionKey: "eu",
		Hostname:  regionalHostname,
	}

	want := RegionalHostname{
		RegionKey: "eu",
		Hostname:  regionalHostname,
	}

	mux.HandleFunc("/zones/"+testZoneID+"/addressing/regional_hostnames/"+regionalHostname, handler)

	actual, err := client.UpdateDataLocalizationRegionalHostname(context.Background(), ZoneIdentifier(testZoneID), params)
	createdOn, _ := time.Parse(time.RFC3339, "2023-01-14T00:47:57.060267Z")
	want.CreatedOn = &createdOn
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteRegionalHostname(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/addressing/regional_hostnames/"+regionalHostname, handler)

	err := client.DeleteDataLocalizationRegionalHostname(context.Background(), ZoneIdentifier(testZoneID), regionalHostname)
	assert.NoError(t, err)
}
