package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testASNNumber = "13335"

func TestIntelligence_ASNOverview(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel/asn/"+testASNNumber, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "asn": 13335,
      "description": "CLOUDFLARENET",
      "country": "US",
      "type": "hosting_provider",
      "domain_count": 1,
      "top_domains": [
        "example.com"
      ]
    }
  ]
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligenceASNOverview(context.Background(), IntelligenceASNOverviewParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}
	// Make sure missing ASN is thrown
	_, err = client.IntelligenceASNOverview(context.Background(), IntelligenceASNOverviewParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingASN, err)
	}
	want := ASNInfo{
		ASN:         13335,
		Description: "CLOUDFLARENET",
		Country:     "US",
		Type:        "hosting_provider",
		DomainCount: 1,
		TopDomains:  []string{"example.com"},
	}

	out, err := client.IntelligenceASNOverview(context.Background(), IntelligenceASNOverviewParameters{AccountID: testAccountID, ASN: 13335})
	if assert.NoError(t, err) {
		assert.Equal(t, len(out), 1, "Length of ASN overview not expected")
		assert.Equal(t, out[0], want, "structs not equal")
	}
}

func TestIntelligence_ASNSubnet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel/asn/"+testASNNumber+"/subnets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "asn": 13335,
  "ip_count_total": 1,
  "subnets": [
    "192.0.2.0/24",
    "2001:DB8::/32"
  ],
  "count": 1,
  "page": 1,
  "per_page": 20
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligenceASNSubnets(context.Background(), IntelligenceASNSubnetsParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure missing ASN is thrown
	_, err = client.IntelligenceASNSubnets(context.Background(), IntelligenceASNSubnetsParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingASN, err)
	}

	want := IntelligenceASNSubnetResponse{
		ASN:          13335,
		IPCountTotal: 1,
		Subnets:      []string{"192.0.2.0/24", "2001:DB8::/32"},
		Count:        1,
		Page:         1,
		PerPage:      20,
	}

	out, err := client.IntelligenceASNSubnets(context.Background(), IntelligenceASNSubnetsParameters{AccountID: testAccountID, ASN: 13335})
	if assert.NoError(t, err) {
		assert.Equal(t, out, want, "structs not equal")
	}
}
