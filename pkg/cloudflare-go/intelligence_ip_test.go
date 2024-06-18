package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntelligence_GetIPOverview(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel/ip", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "ip": "192.0.2.0",
      "belongs_to_ref": {
        "id": "autonomous-system--2fa28d71-3549-5a38-af05-770b79ad6ea8",
        "value": 13335,
        "type": "hosting_provider",
        "country": "US",
        "description": "CLOUDFLARENET"
      },
      "risk_types": [
        {
          "id": 131,
          "super_category_id": 21,
          "name": "Phishing"
        }
      ]
    }
  ]
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligenceGetIPOverview(context.Background(), IPIntelligenceParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	want := IPIntelligence{
		IP: "192.0.2.0",
		BelongsToRef: BelongsToRef{
			ID:          "autonomous-system--2fa28d71-3549-5a38-af05-770b79ad6ea8",
			Value:       13335,
			Type:        "hosting_provider",
			Country:     "US",
			Description: "CLOUDFLARENET",
		},
		RiskTypes: []RiskTypes{
			{
				ID:              131,
				SuperCategoryID: 21,
				Name:            "Phishing",
			},
		},
	}

	out, err := client.IntelligenceGetIPOverview(context.Background(), IPIntelligenceParameters{AccountID: testAccountID, IPv4: "192.0.2.0", IPv6: "2001:0DB8::"})
	if assert.NoError(t, err) {
		assert.Equal(t, len(out), 1, "IP overview length mismatch")
		assert.Equal(t, out[0], want, "structs not equal")
	}
}

func TestIntelligence_GetIPLists(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel/ip-list", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "id": 3,
      "name": "Malware"
    }
]
}
  `)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligenceGetIPList(context.Background(), IPIntelligenceListParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	want := []IPIntelligenceItem{
		{
			ID:   3,
			Name: "Malware",
		},
	}

	out, err := client.IntelligenceGetIPList(context.Background(), IPIntelligenceListParameters{AccountID: testAccountID})
	if assert.NoError(t, err) {
		assert.Equal(t, out, want, "structs not equal")
	}
}

func TestIntelligence_PassiveDNS(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/intel/dns", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "reverse_records": [
      {
        "first_seen": "2021-04-01",
        "last_seen": "2021-04-30",
        "hostname": "cloudflare.com"
      }
    ],
    "count": 1,
    "page": 1,
    "per_page": 20
  }
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.IntelligencePassiveDNS(context.Background(), IPIntelligencePassiveDNSParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	want := IPPassiveDNS{
		Count:   1,
		Page:    1,
		PerPage: 20,
		ReverseRecords: []ReverseRecords{{
			FirstSeen: "2021-04-01",
			LastSeen:  "2021-04-30",
			Hostname:  "cloudflare.com",
		}},
	}

	out, err := client.IntelligencePassiveDNS(context.Background(), IPIntelligencePassiveDNSParameters{AccountID: testAccountID})
	if assert.NoError(t, err) {
		assert.Equal(t, out, want, "structs not equal")
	}
}
