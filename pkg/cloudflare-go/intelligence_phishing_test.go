package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntelligence_PhishingScan(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel-phishing/predict", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "url": "https://www.cloudflare.com",
    "phishing": false,
    "verified": false,
    "score": 0.99,
    "classifier": "MACHINE_LEARNING_v2"
  }
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligencePhishingScan(context.Background(), PhishingScanParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	want := PhishingScan{
		URL:        "https://www.cloudflare.com",
		Phishing:   false,
		Verified:   false,
		Score:      0.99,
		Classifier: "MACHINE_LEARNING_v2",
	}

	out, err := client.IntelligencePhishingScan(context.Background(), PhishingScanParameters{AccountID: testAccountID, URL: "https://www.cloudflare.com"})
	if assert.NoError(t, err) {
		assert.Equal(t, out, want, "structs not equal")
	}
}
