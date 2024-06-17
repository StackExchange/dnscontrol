package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

var mockPageShieldScripts = []PageShieldScript{
	{
		AddedAt:                 "2021-08-18T10:51:10.09615Z",
		DomainReportedMalicious: BoolPtr(false),
		FetchedAt:               "2021-09-02T10:17:54Z",
		FirstPageURL:            "blog.cloudflare.com/page",
		FirstSeenAt:             "2021-08-18T10:51:08Z",
		Hash:                    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Host:                    "blog.cloudflare.com",
		ID:                      "c9ef84a6bf5e47138c75d95e2f933e8f",
		JSIntegrityScore:        10,
		LastSeenAt:              "2021-09-02T09:57:54Z",
		PageURLs:                []string{"blog.cloudflare.com/page1", "blog.cloudflare.com/page2"},
		URL:                     "https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.6.0/js/bootstrap.min.js",
		URLContainsCdnCgiPath:   BoolPtr(false),
	},
}

var mockVersions = []PageShieldScriptVersion{
	{
		FetchedAt:        "2021-09-02T10:17:54Z",
		Hash:             "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		JSIntegrityScore: 10,
	},
}

func TestListPageShieldScripts(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/"+testZoneID+"/page_shield/scripts", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		response := PageShieldScriptsResponse{Results: mockPageShieldScripts}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatal(err)
		}
	})
	result, _, err := client.ListPageShieldScripts(context.Background(), ZoneIdentifier(testZoneID), ListPageShieldScriptsParams{})
	assert.NoError(t, err)
	assert.Equal(t, mockPageShieldScripts, result)
}

func TestGetPageShieldScript(t *testing.T) {
	setup()
	defer teardown()

	scriptID := "c9ef84a6bf5e47138c75d95e2f933e8f"
	mux.HandleFunc(fmt.Sprintf("/zones/"+testZoneID+"/page_shield/scripts/%s", scriptID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		response := PageShieldScriptResponse{
			Result:   mockPageShieldScripts[0],
			Versions: mockVersions,
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Fatal(err)
		}
	})
	result, versions, err := client.GetPageShieldScript(context.Background(), ZoneIdentifier(testZoneID), scriptID)
	assert.NoError(t, err)
	assert.Equal(t, &mockPageShieldScripts[0], result)
	assert.Equal(t, mockVersions, versions)
}
