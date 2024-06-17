package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniversalSSLSettingDetails(t *testing.T) {
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
			  "enabled": true
			}
		  }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/ssl/universal/settings", handler)

	want := UniversalSSLSetting{
		Enabled: true,
	}

	got, err := client.UniversalSSLSettingDetails(context.Background(), testZoneID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestEditUniversalSSLSetting(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		assert.Equal(t, `{"enabled":true}`, string(body))

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
			  "enabled": true
			}
		  }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/ssl/universal/settings", handler)

	want := UniversalSSLSetting{
		Enabled: true,
	}

	got, err := client.EditUniversalSSLSetting(context.Background(), testZoneID, want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestUniversalSSLVerificationDetails(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"result": [{
				"certificate_status": "active",
				"verification_type": "cname",
				"verification_status": true,
				"verification_info": [
				{
					"status": "pending",
					"http_url": "http://example.com/.well-known/acme-challenge/Km-ycWoOVh10cLfL4pRPppGt6jU_mGz8xgvNOxudMiA",
					"http_body": "Km-ycWoOVh10cLfL4pRPppGt6jU_mGz8xgvNOxudMiA.Jckzm7Z9uOFls_MXPYibNRz6koY5a8qpI_BeHtDtf-g"
				}
			],
				"brand_check": false,
				"validation_method": "txt",
				"cert_pack_uuid": "a77f8bd7-3b47-46b4-a6f1-75cf98109948"
			}]
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/ssl/verification", handler)

	want := []UniversalSSLVerificationDetails{
		{
			CertificateStatus:  "active",
			VerificationType:   "cname",
			ValidationMethod:   "txt",
			CertPackUUID:       "a77f8bd7-3b47-46b4-a6f1-75cf98109948",
			VerificationStatus: true,
			BrandCheck:         false,
			VerificationInfo: []SSLValidationRecord{{
				HTTPUrl:  "http://example.com/.well-known/acme-challenge/Km-ycWoOVh10cLfL4pRPppGt6jU_mGz8xgvNOxudMiA",
				HTTPBody: "Km-ycWoOVh10cLfL4pRPppGt6jU_mGz8xgvNOxudMiA.Jckzm7Z9uOFls_MXPYibNRz6koY5a8qpI_BeHtDtf-g",
			}},
		},
	}

	got, err := client.UniversalSSLVerificationDetails(context.Background(), testZoneID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestUpdateSSLCertificatePackValidationMethod(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		assert.Equal(t, `{"validation_method":"txt"}`, string(body))

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"validation_method": "txt",
				"status": "pending_validation"
			}
		  }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/ssl/verification/"+testCertPackUUID, handler)

	want := UniversalSSLCertificatePackValidationMethodSetting{
		ValidationMethod: "txt",
	}

	got, err := client.UpdateUniversalSSLCertificatePackValidationMethod(context.Background(), testZoneID, testCertPackUUID, want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}
