package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntelligence_WHOIS(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel/whois", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "domain": "cloudflare.com",
    "created_date": "2009-02-17",
    "updated_date": "2017-05-24",
    "registrant": "DATA REDACTED",
    "registrant_org": "DATA REDACTED",
    "registrant_country": "United States",
    "registrant_email": "https://domaincontact.cloudflareregistrar.com/cloudflare.com",
    "registrar": "Cloudflare, Inc.",
    "nameservers": [
      "ns3.cloudflare.com",
      "ns4.cloudflare.com",
      "ns5.cloudflare.com",
      "ns6.cloudflare.com",
      "ns7.cloudflare.com"
    ]
  }
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligenceWHOIS(context.Background(), WHOISParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure missing domain is thrown
	_, err = client.IntelligenceWHOIS(context.Background(), WHOISParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingDomain, err)
	}

	want := WHOIS{
		Domain:            "cloudflare.com",
		CreatedDate:       "2009-02-17",
		UpdatedDate:       "2017-05-24",
		Registrant:        "DATA REDACTED",
		RegistrantOrg:     "DATA REDACTED",
		RegistrantCountry: "United States",
		RegistrantEmail:   "https://domaincontact.cloudflareregistrar.com/cloudflare.com",
		Registrar:         "Cloudflare, Inc.",
		Nameservers: []string{
			"ns3.cloudflare.com",
			"ns4.cloudflare.com",
			"ns5.cloudflare.com",
			"ns6.cloudflare.com",
			"ns7.cloudflare.com",
		},
	}

	out, err := client.IntelligenceWHOIS(context.Background(), WHOISParameters{AccountID: testAccountID, Domain: "cloudflare.com"})
	if assert.NoError(t, err) {
		assert.Equal(t, out, want, "structs not equal")
	}
}
