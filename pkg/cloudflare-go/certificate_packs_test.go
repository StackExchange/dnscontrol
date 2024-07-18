package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	uploadedOn, _ = time.Parse(time.RFC3339, "2014-01-01T05:20:00Z")
	expiresOn, _  = time.Parse(time.RFC3339, "2016-01-01T05:20:00Z")

	desiredCertificatePack = CertificatePack{
		ID:                   "3822ff90-ea29-44df-9e55-21300bb9419b",
		Type:                 "advanced",
		Hosts:                []string{"example.com", "*.example.com", "www.example.com"},
		PrimaryCertificate:   "b2cfa4183267af678ea06c7407d4d6d8",
		ValidationMethod:     "txt",
		ValidityDays:         90,
		CertificateAuthority: "lets_encrypt",
		Certificates: []CertificatePackCertificate{{
			ID:              "3822ff90-ea29-44df-9e55-21300bb9419b",
			Hosts:           []string{"example.com"},
			Issuer:          "GlobalSign",
			Signature:       "SHA256WithRSA",
			Status:          "active",
			BundleMethod:    "ubiquitous",
			GeoRestrictions: CertificatePackGeoRestrictions{Label: "us"},
			ZoneID:          testZoneID,
			UploadedOn:      uploadedOn,
			ModifiedOn:      uploadedOn,
			ExpiresOn:       expiresOn,
			Priority:        1,
		}},
	}

	pendingCertificatePack = CertificatePack{
		ID:                   "3822ff90-ea29-44df-9e55-21300bb9419b",
		Type:                 "advanced",
		Hosts:                []string{"example.com", "*.example.com", "www.example.com"},
		Status:               "initializing",
		ValidationMethod:     "txt",
		ValidityDays:         90,
		CertificateAuthority: "lets_encrypt",
	}
)

func TestListCertificatePacks(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "id": "3822ff90-ea29-44df-9e55-21300bb9419b",
      "type": "advanced",
      "hosts": [
        "example.com",
        "*.example.com",
        "www.example.com"
      ],
      "validity_days": 90,
      "validation_method": "txt",
      "certificate_authority": "lets_encrypt",
      "certificates": [
        {
          "id": "3822ff90-ea29-44df-9e55-21300bb9419b",
          "hosts": [
            "example.com"
          ],
          "issuer": "GlobalSign",
          "signature": "SHA256WithRSA",
          "status": "active",
          "bundle_method": "ubiquitous",
          "geo_restrictions": {
            "label": "us"
          },
          "zone_id": "%[1]s",
          "uploaded_on": "2014-01-01T05:20:00Z",
          "modified_on": "2014-01-01T05:20:00Z",
          "expires_on": "2016-01-01T05:20:00Z",
          "priority": 1
        }
      ],
      "primary_certificate": "b2cfa4183267af678ea06c7407d4d6d8"
    }
  ]
}
		`, testZoneID)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/ssl/certificate_packs", handler)

	want := []CertificatePack{desiredCertificatePack}
	actual, err := client.ListCertificatePacks(context.Background(), testZoneID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListCertificatePack(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "3822ff90-ea29-44df-9e55-21300bb9419b",
    "type": "advanced",
    "validity_days": 90,
    "validation_method": "txt",
    "certificate_authority": "lets_encrypt",
    "hosts": [
      "example.com",
      "*.example.com",
      "www.example.com"
    ],
    "certificates": [
      {
        "id": "3822ff90-ea29-44df-9e55-21300bb9419b",
        "hosts": [
          "example.com"
        ],
        "issuer": "GlobalSign",
        "signature": "SHA256WithRSA",
        "status": "active",
        "bundle_method": "ubiquitous",
        "geo_restrictions": {
          "label": "us"
        },
        "zone_id": "%[1]s",
        "uploaded_on": "2014-01-01T05:20:00Z",
        "modified_on": "2014-01-01T05:20:00Z",
        "expires_on": "2016-01-01T05:20:00Z",
        "priority": 1
      }
    ],
    "primary_certificate": "b2cfa4183267af678ea06c7407d4d6d8"
  }
}
		`, testZoneID)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/ssl/certificate_packs/3822ff90-ea29-44df-9e55-21300bb9419b", handler)

	actual, err := client.CertificatePack(context.Background(), testZoneID, "3822ff90-ea29-44df-9e55-21300bb9419b")

	if assert.NoError(t, err) {
		assert.Equal(t, desiredCertificatePack, actual)
	}
}

func TestCreateCertificatePack(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "id": "3822ff90-ea29-44df-9e55-21300bb9419b",
        "type": "advanced",
        "hosts": [
          "example.com",
          "*.example.com",
          "www.example.com"
        ],
        "status": "initializing",
        "validation_method": "txt",
        "validity_days": 90,
        "certificate_authority": "lets_encrypt",
        "cloudflare_branding": false
      }
    }
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/ssl/certificate_packs/order", handler)

	certificate := CertificatePackRequest{
		Type:                 "advanced",
		Hosts:                []string{"example.com", "*.example.com", "www.example.com"},
		ValidationMethod:     "txt",
		ValidityDays:         90,
		CertificateAuthority: "lets_encrypt",
	}
	actual, err := client.CreateCertificatePack(context.Background(), testZoneID, certificate)

	if assert.NoError(t, err) {
		assert.Equal(t, pendingCertificatePack, actual)
	}
}

func TestRestartAdvancedCertificateValidation(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "3822ff90-ea29-44df-9e55-21300bb9419b",
    "type": "advanced",
    "hosts": [
      "example.com",
      "*.example.com",
      "www.example.com"
    ],
    "status": "initializing",
    "validation_method": "txt",
    "validity_days": 365,
    "certificate_authority": "lets_encrypt",
    "cloudflare_branding": false
  }
}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/ssl/certificate_packs/3822ff90-ea29-44df-9e55-21300bb9419b", handler)

	certificate := CertificatePack{
		ID:                   "3822ff90-ea29-44df-9e55-21300bb9419b",
		Type:                 "advanced",
		Hosts:                []string{"example.com", "*.example.com", "www.example.com"},
		Status:               "initializing",
		ValidityDays:         365,
		ValidationMethod:     "txt",
		CertificateAuthority: "lets_encrypt",
		CloudflareBranding:   false,
	}

	actual, err := client.RestartCertificateValidation(context.Background(), testZoneID, "3822ff90-ea29-44df-9e55-21300bb9419b")

	if assert.NoError(t, err) {
		assert.Equal(t, certificate, actual)
	}
}

func TestDeleteCertificatePack(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "3822ff90-ea29-44df-9e55-21300bb9419b"
  }
}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/ssl/certificate_packs/3822ff90-ea29-44df-9e55-21300bb9419b", handler)

	err := client.DeleteCertificatePack(context.Background(), testZoneID, "3822ff90-ea29-44df-9e55-21300bb9419b")

	assert.NoError(t, err)
}
