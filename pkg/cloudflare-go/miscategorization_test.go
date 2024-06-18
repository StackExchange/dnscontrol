package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMiscategorization(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/intel/miscategorization", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": []
}`)
	})
	ctx := context.Background()
	err := client.CreateMiscategorization(ctx, MisCategorizationParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	err = client.CreateMiscategorization(ctx, MisCategorizationParameters{AccountID: testAccountID, IndicatorType: "ipv4"})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingIP, err)
	}

	err = client.CreateMiscategorization(ctx, MisCategorizationParameters{AccountID: testAccountID, IndicatorType: "ipv6"})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingIP, err)
	}

	err = client.CreateMiscategorization(ctx, MisCategorizationParameters{AccountID: testAccountID, IndicatorType: "url"})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingURL, err)
	}

	err = client.CreateMiscategorization(ctx, MisCategorizationParameters{AccountID: testAccountID, IndicatorType: "domain"})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingURL, err)
	}

	err = client.CreateMiscategorization(ctx, MisCategorizationParameters{AccountID: testAccountID, IndicatorType: "ipv4", IP: "192.0.2.0"})
	assert.NoError(t, err, "Got error for creating miscategorization for ipv4")

	err = client.CreateMiscategorization(ctx, MisCategorizationParameters{AccountID: testAccountID, IndicatorType: "ipv6", IP: "2400:cb00::/32"})
	assert.NoError(t, err, "Got error for creating miscategorization for ipv6")

	err = client.CreateMiscategorization(ctx, MisCategorizationParameters{AccountID: testAccountID, IndicatorType: "domain", URL: "example.com"})
	assert.NoError(t, err, "Got error for creating miscategorization for domain")

	err = client.CreateMiscategorization(ctx, MisCategorizationParameters{AccountID: testAccountID, IndicatorType: "url", URL: "https://example.com/news/"})
	assert.NoError(t, err, "Got error for creating miscategorization for url")
}
