package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntelligence_DomainDetails(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel/domain", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "domain": "cloudflare.com",
    "resolves_to_refs": [
      {
        "id": "ipv4-addr--baa568ec-6efe-5902-be55-0663833db537",
        "value": "192.0.2.0"
      }
    ],
    "popularity_rank": 18,
    "application": {
      "id": 1370,
      "name": "CLOUDFLARE"
    },
    "risk_types": [],
    "content_categories": [
      {
        "id": 155,
        "super_category_id": 26,
        "name": "Technology"
      }
    ],
    "additional_information": {
      "suspected_malware_family": ""
    }
  }
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligenceDomainDetails(context.Background(), GetDomainDetailsParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}
	// Make sure missing domain is thrown
	_, err = client.IntelligenceDomainDetails(context.Background(), GetDomainDetailsParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingDomain, err)
	}
	want := DomainDetails{
		Domain: "cloudflare.com",
		ResolvesToRefs: []ResolvesToRefs{
			{
				ID:    "ipv4-addr--baa568ec-6efe-5902-be55-0663833db537",
				Value: "192.0.2.0",
			},
		},
		PopularityRank: 18,
		Application: Application{
			ID:   1370,
			Name: "CLOUDFLARE",
		},
		RiskTypes: []interface{}{},
		ContentCategories: []ContentCategories{
			{
				ID:              155,
				SuperCategoryID: 26,
				Name:            "Technology",
			},
		},
		AdditionalInformation: AdditionalInformation{
			SuspectedMalwareFamily: "",
		},
	}

	out, err := client.IntelligenceDomainDetails(context.Background(), GetDomainDetailsParameters{AccountID: testAccountID, Domain: "cloudflare.com"})
	if assert.NoError(t, err) {
		assert.Equal(t, out, want, "structs not equal")
	}
}

func TestIntelligence_BulkDomainDetails(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel/domain/bulk", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "domain": "cloudflare.com",
      "popularity_rank": 18,
      "application": {
        "id": 1370,
        "name": "CLOUDFLARE"
      },
      "risk_types": [],
      "content_categories": [
        {
          "id": 155,
          "super_category_id": 26,
          "name": "Technology"
        }
      ],
      "additional_information": {
        "suspected_malware_family": ""
      }
    }
  ]
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligenceBulkDomainDetails(context.Background(), GetBulkDomainDetailsParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}
	// Make sure missing domain is thrown
	_, err = client.IntelligenceBulkDomainDetails(context.Background(), GetBulkDomainDetailsParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingDomain, err)
	}
	want := DomainDetails{
		Domain:         "cloudflare.com",
		PopularityRank: 18,
		Application: Application{
			ID:   1370,
			Name: "CLOUDFLARE",
		},
		RiskTypes: []interface{}{},
		ContentCategories: []ContentCategories{
			{
				ID:              155,
				SuperCategoryID: 26,
				Name:            "Technology",
			},
		},
		AdditionalInformation: AdditionalInformation{
			SuspectedMalwareFamily: "",
		},
	}

	out, err := client.IntelligenceBulkDomainDetails(context.Background(), GetBulkDomainDetailsParameters{AccountID: testAccountID, Domains: []string{"cloudflare.com"}})
	if assert.NoError(t, err) {
		assert.Equal(t, len(out), 1, "Length of ASN overview not expected")
		assert.Equal(t, out[0], want, "structs not equal")
	}
}

func TestIntelligence_DomainHistory(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel/domain-history", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "domain": "cloudflare.com",
      "categorizations": [
        {
          "categories": [
            {
              "id": 155,
              "name": "Technology"
            }
          ],
          "start": "2021-04-01",
          "end": "2021-04-30"
        }
      ]
    }
  ]
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligenceDomainHistory(context.Background(), GetDomainHistoryParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}
	// Make sure missing domain is thrown
	_, err = client.IntelligenceDomainHistory(context.Background(), GetDomainHistoryParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingDomain, err)
	}
	want := DomainHistory{
		Domain: "cloudflare.com",
		Categorizations: []Categorizations{
			{
				Categories: []Categories{
					{
						ID:   155,
						Name: "Technology",
					},
				},
				Start: "2021-04-01",
				End:   "2021-04-30",
			},
		},
	}

	out, err := client.IntelligenceDomainHistory(context.Background(), GetDomainHistoryParameters{AccountID: testAccountID, Domain: "cloudflare.com"})
	if assert.NoError(t, err) {
		assert.Equal(t, len(out), 1, "Length of Domain History not expected")
		assert.Equal(t, out[0], want, "structs not equal")
	}
}
