package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCustomHostname_DeleteCustomHostname(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/bar", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
{
  "id": "bar"
}`)
	})

	err := client.DeleteCustomHostname(context.Background(), "foo", "bar")

	assert.NoError(t, err)
}

func TestCustomHostname_CreateCustomHostname(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
	"success": true,
	"errors": [],
	"messages": [],
	"result": {
		"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"hostname": "app.example.com",
		"custom_origin_server": "example.app.com",
		"ssl": {
			"status": "pending_validation",
			"method": "cname",
			"type": "dv",
			"cname_target": "dcv.digicert.com",
			"cname": "810b7d5f01154524b961ba0cd578acc2.app.example.com",
			"validation_records": [{
				"cname_target": "dcv.digicert.com",
				"cname": "810b7d5f01154524b961ba0cd578acc2.app.example.com"
			}],
			"settings": {
			"http2": "on"
			}
		},
		"status": "pending",
		"verification_errors": [
		  "None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."
		],
		"ownership_verification": {
		  "type": "txt",
		  "name": "_cf-custom-hostname.app.example.com",
		  "value": "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0"
		},
		"ownership_verification_http": {
		  "http_url": "http://app.example.com/.well-known/cf-custom-hostname-challenge/37c82d20-99fb-490e-ba0a-489fa483b776",
		  "http_body": "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0"
		},
		"created_at": "2020-02-06T18:11:23.531995Z"
	}
}`)
	})

	response, err := client.CreateCustomHostname(context.Background(), "foo", CustomHostname{Hostname: "app.example.com", SSL: &CustomHostnameSSL{Method: "cname", Type: "dv"}})

	createdAt, _ := time.Parse(time.RFC3339, "2020-02-06T18:11:23.531995Z")

	validationRec := SSLValidationRecord{
		CnameTarget: "dcv.digicert.com",
		CnameName:   "810b7d5f01154524b961ba0cd578acc2.app.example.com",
	}
	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Hostname:           "app.example.com",
			CustomOriginServer: "example.app.com",
			SSL: &CustomHostnameSSL{
				Type:                "dv",
				Method:              "cname",
				Status:              "pending_validation",
				SSLValidationRecord: validationRec,
				ValidationRecords:   []SSLValidationRecord{validationRec},
				Settings: CustomHostnameSSLSettings{
					HTTP2: "on",
				},
			},
			Status:             "pending",
			VerificationErrors: []string{"None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."},
			OwnershipVerification: CustomHostnameOwnershipVerification{
				Type:  "txt",
				Name:  "_cf-custom-hostname.app.example.com",
				Value: "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0",
			},
			OwnershipVerificationHTTP: CustomHostnameOwnershipVerificationHTTP{
				HTTPUrl:  "http://app.example.com/.well-known/cf-custom-hostname-challenge/37c82d20-99fb-490e-ba0a-489fa483b776",
				HTTPBody: "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0",
			},
			CreatedAt: &createdAt,
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_CreateCustomHostname_MethodTxt(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
	"success": true,
	"errors": [],
	"messages": [],
	"result": {
		"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"hostname": "app.example.com",
		"custom_origin_server": "example.app.com",
		"ssl": {
			"status": "pending_validation",
			"method": "txt",
			"type": "dv",
			"txt_name": "app.example.com",
			"txt_value": "ca3-f8db94da174g4c409b17fcaa5470deb2",
			"settings": {
			"http2": "on"
			}
		},
		"status": "pending",
		"verification_errors": [
		  "None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."
		],
		"ownership_verification": {
		  "type": "txt",
		  "name": "_cf-custom-hostname.app.example.com",
		  "value": "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0"
		},
		"ownership_verification_http": {
		  "http_url": "http://app.example.com/.well-known/cf-custom-hostname-challenge/37c82d20-99fb-490e-ba0a-489fa483b776",
		  "http_body": "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0"
		},
		"created_at": "2020-02-06T18:11:23.531995Z"
	}
}`)
	})

	response, err := client.CreateCustomHostname(context.Background(), "foo", CustomHostname{Hostname: "app.example.com", SSL: &CustomHostnameSSL{Method: "txt", Type: "dv"}})

	createdAt, _ := time.Parse(time.RFC3339, "2020-02-06T18:11:23.531995Z")

	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Hostname:           "app.example.com",
			CustomOriginServer: "example.app.com",
			SSL: &CustomHostnameSSL{
				Type:   "dv",
				Method: "txt",
				Status: "pending_validation",
				SSLValidationRecord: SSLValidationRecord{
					TxtName:  "app.example.com",
					TxtValue: "ca3-f8db94da174g4c409b17fcaa5470deb2",
				},
				Settings: CustomHostnameSSLSettings{
					HTTP2: "on",
				},
			},
			Status:             "pending",
			VerificationErrors: []string{"None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."},
			OwnershipVerification: CustomHostnameOwnershipVerification{
				Type:  "txt",
				Name:  "_cf-custom-hostname.app.example.com",
				Value: "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0",
			},
			OwnershipVerificationHTTP: CustomHostnameOwnershipVerificationHTTP{
				HTTPUrl:  "http://app.example.com/.well-known/cf-custom-hostname-challenge/37c82d20-99fb-490e-ba0a-489fa483b776",
				HTTPBody: "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0",
			},
			CreatedAt: &createdAt,
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_CreateCustomHostname_CustomOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
	"success": true,
	"errors": [],
	"messages": [],
	"result": {
		"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"hostname": "app.example.com",
		"custom_origin_server": "example.app.com",
		"ssl": {
			"status": "pending_validation",
			"method": "cname",
			"type": "dv",
			"cname_target": "dcv.digicert.com",
			"cname": "810b7d5f01154524b961ba0cd578acc2.app.example.com",
			"settings": {
			"http2": "on"
			}
		}
  	}
}`)
	})

	response, err := client.CreateCustomHostname(context.Background(), "foo", CustomHostname{Hostname: "app.example.com", CustomOriginServer: "example.app.com", SSL: &CustomHostnameSSL{Method: "cname", Type: "dv"}})

	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Hostname:           "app.example.com",
			CustomOriginServer: "example.app.com",
			SSL: &CustomHostnameSSL{
				Type:   "dv",
				Method: "cname",
				Status: "pending_validation",
				SSLValidationRecord: SSLValidationRecord{
					CnameTarget: "dcv.digicert.com",
					CnameName:   "810b7d5f01154524b961ba0cd578acc2.app.example.com",
				},
				Settings: CustomHostnameSSLSettings{
					HTTP2: "on",
				},
			},
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_CreateCustomHostname_No_SSL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
	"success": true,
	"errors": [],
	"messages": [],
	"result": {
		"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"hostname": "app.example.com",
		"custom_origin_server": "example.app.com",
		"status": "pending",
		"verification_errors": [
		  "None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."
		],
		"ownership_verification": {
		  "type": "txt",
		  "name": "_cf-custom-hostname.app.example.com",
		  "value": "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0"
		},
		"ownership_verification_http": {
		  "http_url": "http://app.example.com/.well-known/cf-custom-hostname-challenge/37c82d20-99fb-490e-ba0a-489fa483b776",
		  "http_body": "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0"
		},
		"created_at": "2020-02-06T18:11:23.531995Z"
	}
}`)
	})

	response, err := client.CreateCustomHostname(context.Background(), "foo", CustomHostname{Hostname: "app.example.com"})

	createdAt, _ := time.Parse(time.RFC3339, "2020-02-06T18:11:23.531995Z")

	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Hostname:           "app.example.com",
			CustomOriginServer: "example.app.com",
			Status:             "pending",
			VerificationErrors: []string{"None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."},
			OwnershipVerification: CustomHostnameOwnershipVerification{
				Type:  "txt",
				Name:  "_cf-custom-hostname.app.example.com",
				Value: "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0",
			},
			OwnershipVerificationHTTP: CustomHostnameOwnershipVerificationHTTP{
				HTTPUrl:  "http://app.example.com/.well-known/cf-custom-hostname-challenge/37c82d20-99fb-490e-ba0a-489fa483b776",
				HTTPBody: "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0",
			},
			CreatedAt: &createdAt,
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_CreateCustomHostname_CustomOriginSNI(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
	"success": true,
	"errors": [],
	"messages": [],
	"result": {
		"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"hostname": "app.example.com",
		"custom_origin_server": "example.app.com",
		"custom_origin_sni": "app.example.com",
		"status": "pending",
		"verification_errors": [
		  "None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."
		],
		"ownership_verification": {
		  "type": "txt",
		  "name": "_cf-custom-hostname.app.example.com",
		  "value": "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0"
		},
		"ownership_verification_http": {
		  "http_url": "http://app.example.com/.well-known/cf-custom-hostname-challenge/37c82d20-99fb-490e-ba0a-489fa483b776",
		  "http_body": "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0"
		},
		"created_at": "2020-02-06T18:11:23.531995Z"
	}
}`)
	})

	response, err := client.CreateCustomHostname(context.Background(), "foo", CustomHostname{Hostname: "app.example.com", CustomOriginSNI: "app.example.com"})

	createdAt, _ := time.Parse(time.RFC3339, "2020-02-06T18:11:23.531995Z")

	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Hostname:           "app.example.com",
			CustomOriginServer: "example.app.com",
			CustomOriginSNI:    "app.example.com",
			Status:             "pending",
			VerificationErrors: []string{"None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."},
			OwnershipVerification: CustomHostnameOwnershipVerification{
				Type:  "txt",
				Name:  "_cf-custom-hostname.app.example.com",
				Value: "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0",
			},
			OwnershipVerificationHTTP: CustomHostnameOwnershipVerificationHTTP{
				HTTPUrl:  "http://app.example.com/.well-known/cf-custom-hostname-challenge/37c82d20-99fb-490e-ba0a-489fa483b776",
				HTTPBody: "38ddbedc-6cc3-4a4c-af67-9c5b02344ce0",
			},
			CreatedAt: &createdAt,
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_CustomHostnames(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
	"success": true,
	"result": [
		{
			"id": "custom_host_1",
			"hostname": "custom.host.one",
			"ssl": {
				"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
				"type": "dv",
				"method": "cname",
				"status": "pending_validation",
				"cname_target": "dcv.digicert.com",
				"cname": "810b7d5f01154524b961ba0cd578acc2.app.example.com",
				"issuer": "DigiCertInc",
				"serial_number": "6743787633689793699141714808227354901",
				"http_url": "http://app.example.com/.well-known/pki-validation/ca3-da12a1c25e7b48cf80408c6c1763b8a2.txt",
      			"http_body": "ca3-574923932a82475cb8592200f1a2a23d"
			},
			"custom_metadata": {
				"a_random_field": "random field value"
			},
			"status": "pending",
			"verification_errors": [
				"None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."
    		],
			"ownership_verification": {
				"type": "txt",
      			"name": "_cf-custom-hostname.app.example.com",
      			"value": "5cc07c04-ea62-4a5a-95f0-419334a875a4"
    		}
		}
	],
	"result_info": {
		"page": 1,
		"per_page": 20,
		"count": 5,
		"total_count": 5
	}
}`)
	})

	customHostnames, _, err := client.CustomHostnames(context.Background(), "foo", 1, CustomHostname{})

	want := []CustomHostname{
		{
			ID:       "custom_host_1",
			Hostname: "custom.host.one",
			SSL: &CustomHostnameSSL{
				ID:     "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
				Type:   "dv",
				Method: "cname",
				Status: "pending_validation",
				SSLValidationRecord: SSLValidationRecord{
					CnameTarget: "dcv.digicert.com",
					CnameName:   "810b7d5f01154524b961ba0cd578acc2.app.example.com",
					HTTPUrl:     "http://app.example.com/.well-known/pki-validation/ca3-da12a1c25e7b48cf80408c6c1763b8a2.txt",
					HTTPBody:    "ca3-574923932a82475cb8592200f1a2a23d",
				},
				Issuer:       "DigiCertInc",
				SerialNumber: "6743787633689793699141714808227354901",
			},
			CustomMetadata: &CustomMetadata{"a_random_field": "random field value"},
			Status:         PENDING,
			VerificationErrors: []string{"None of the A or AAAA records are owned " +
				"by this account and the pre-generated ownership verification token was not found."},
			OwnershipVerification: CustomHostnameOwnershipVerification{
				Type:  "txt",
				Name:  "_cf-custom-hostname.app.example.com",
				Value: "5cc07c04-ea62-4a5a-95f0-419334a875a4",
			},
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, customHostnames)
	}
}

func TestCustomHostname_CustomHostname(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/bar", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
"success": true,
"result": {
    "id": "bar",
    "hostname": "foo.bar.com",
    "ssl": {
      "id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
      "type": "dv",
      "method": "http",
      "status": "active",
      "issuer": "DigiCertInc",
      "serial_number": "6743787633689793699141714808227354901",
      "settings": {
        "ciphers": ["ECDHE-RSA-AES128-GCM-SHA256","AES128-SHA"],
        "http2": "on",
        "min_tls_version": "1.2"
      }
    },
    "custom_metadata": {
      "origin": "a.custom.origin"
    },
	"status": "pending",
	"verification_errors": [
		"None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."
    ],
	"ownership_verification": {
		"type": "txt",
     	"name": "_cf-custom-hostname.app.example.com",
    	"value": "5cc07c04-ea62-4a5a-95f0-419334a875a4"
    }
  }
}`)
	})

	customHostname, err := client.CustomHostname(context.Background(), "foo", "bar")

	want := CustomHostname{
		ID:       "bar",
		Hostname: "foo.bar.com",
		SSL: &CustomHostnameSSL{
			ID:           "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Status:       "active",
			Method:       "http",
			Type:         "dv",
			Issuer:       "DigiCertInc",
			SerialNumber: "6743787633689793699141714808227354901",
			Settings: CustomHostnameSSLSettings{
				HTTP2:         "on",
				MinTLSVersion: "1.2",
				Ciphers:       []string{"ECDHE-RSA-AES128-GCM-SHA256", "AES128-SHA"},
			},
		},
		CustomMetadata: &CustomMetadata{"origin": "a.custom.origin"},
		Status:         PENDING,
		VerificationErrors: []string{"None of the A or AAAA records are owned " +
			"by this account and the pre-generated ownership verification token was not found."},
		OwnershipVerification: CustomHostnameOwnershipVerification{
			Type:  "txt",
			Name:  "_cf-custom-hostname.app.example.com",
			Value: "5cc07c04-ea62-4a5a-95f0-419334a875a4",
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, customHostname)
	}
}

func TestCustomHostname_CustomHostname_WithSSLError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/bar", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
"success": true,
"result": {
	"id": "bar",
	"hostname": "example.com",
	"ssl": {
		"type": "dv",
		"method": "cname",
		"status": "pending_validation",
		"cname_target": "dcv.digicert.com",
		"cname": "810b7d5f01154524b961ba0cd578acc2.example.com",
		"validation_errors": [{
			"message": "SERVFAIL looking up CAA for example.com"
		}]
	},
	"status": "pending",
	"verification_errors": [
		"None of the A or AAAA records are owned by this account and the pre-generated ownership verification token was not found."
	],
	"ownership_verification_http": {
		"http_url": "http://example.com/.well-known/cf-custom-hostname-challenge/0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"http_body": "5cc07c04-ea62-4a5a-95f0-419334a875a4"
	}
}
}`)
	})

	customHostname, err := client.CustomHostname(context.Background(), "foo", "bar")

	want := CustomHostname{
		ID:       "bar",
		Hostname: "example.com",
		SSL: &CustomHostnameSSL{
			Type:   "dv",
			Method: "cname",
			Status: "pending_validation",
			SSLValidationRecord: SSLValidationRecord{
				CnameName:   "810b7d5f01154524b961ba0cd578acc2.example.com",
				CnameTarget: "dcv.digicert.com",
			},
			ValidationErrors: []SSLValidationError{
				{
					Message: "SERVFAIL looking up CAA for example.com",
				},
			},
		},
		Status: PENDING,
		VerificationErrors: []string{"None of the A or AAAA records are owned " +
			"by this account and the pre-generated ownership verification token was not found."},
		OwnershipVerificationHTTP: CustomHostnameOwnershipVerificationHTTP{
			HTTPBody: "5cc07c04-ea62-4a5a-95f0-419334a875a4",
			HTTPUrl:  "http://example.com/.well-known/cf-custom-hostname-challenge/0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, customHostname)
	}
}

func TestCustomHostname_UpdateCustomHostnameSSL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/0d89c70d-ad9f-4843-b99f-6cc0252067e9", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
	"success": true,
	"errors": [],
	"messages": [],
	"result": {
		"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"hostname": "app.example.com",
		"custom_origin_server": "example.app.com",
		"ssl": {
			"status": "pending_validation",
			"method": "cname",
			"type": "dv",
			"cname_target": "dcv.digicert.com",
			"cname": "810b7d5f01154524b961ba0cd578acc2.app.example.com",
			"settings": {
			"http2": "off",
			"tls_1_3": "on"
			}
		}
  	}
}`)
	})

	response, err := client.UpdateCustomHostnameSSL(context.Background(), "foo", "0d89c70d-ad9f-4843-b99f-6cc0252067e9", &CustomHostnameSSL{Method: "cname", Type: "dv", Settings: CustomHostnameSSLSettings{HTTP2: "off", TLS13: "on"}})

	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Hostname:           "app.example.com",
			CustomOriginServer: "example.app.com",
			SSL: &CustomHostnameSSL{
				Type:   "dv",
				Method: "cname",
				Status: "pending_validation",
				SSLValidationRecord: SSLValidationRecord{
					CnameTarget: "dcv.digicert.com",
					CnameName:   "810b7d5f01154524b961ba0cd578acc2.app.example.com",
				},
				Settings: CustomHostnameSSLSettings{
					HTTP2: "off",
					TLS13: "on",
				},
			},
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_UpdateCustomHostname(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/0d89c70d-ad9f-4843-b99f-6cc0252067e9", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		defer r.Body.Close()
		reqBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err, "Reading request body")
		assert.JSONEq(t, `
{
	"hostname": "app.example.com",
	"custom_origin_server": "example.app.com",
	"ssl": {
		"method": "cname",
		"type": "dv",
		"wildcard": false,
		"settings": {}
	},
	"ownership_verification": {},
	"ownership_verification_http": {}
}`, string(reqBody), "Unexpected request body")

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
	"success": true,
	"errors": [],
	"messages": [],
	"result": {
		"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"hostname": "app.example.com",
		"custom_origin_server": "example.app.com",
		"ssl": {
			"status": "pending_validation",
			"method": "cname",
			"type": "dv",
			"cname_target": "dcv.digicert.com",
			"cname": "810b7d5f01154524b961ba0cd578acc2.app.example.com",
			"settings": {
				"http2": "off",
				"tls_1_3": "on"
			}
		}
  	}
}`)
	})

	wildcard := false
	response, err := client.UpdateCustomHostname(context.Background(), "foo", "0d89c70d-ad9f-4843-b99f-6cc0252067e9", CustomHostname{Hostname: "app.example.com", CustomOriginServer: "example.app.com", SSL: &CustomHostnameSSL{Method: "cname", Type: "dv", Wildcard: &wildcard}})

	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Hostname:           "app.example.com",
			CustomOriginServer: "example.app.com",
			SSL: &CustomHostnameSSL{
				Type:   "dv",
				Method: "cname",
				Status: "pending_validation",
				SSLValidationRecord: SSLValidationRecord{
					CnameTarget: "dcv.digicert.com",
					CnameName:   "810b7d5f01154524b961ba0cd578acc2.app.example.com",
				},
				Settings: CustomHostnameSSLSettings{
					HTTP2: "off",
					TLS13: "on",
				},
			},
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_UpdateCustomHostnameWithCustomMetadata(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/0d89c70d-ad9f-4843-b99f-6cc0252067e9", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		defer r.Body.Close()
		reqBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err, "Reading request body")
		assert.JSONEq(t, `
{
	"hostname": "app.example.com",
	"custom_origin_server": "example.app.com",
	"ssl": {
		"method": "cname",
		"type": "dv",
		"wildcard": false,
		"settings": {}
	},
	"custom_metadata": {
		"a_random_field": "updated field value"
	},
	"ownership_verification": {},
	"ownership_verification_http": {}
}`, string(reqBody), "Unexpected request body")

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
	"success": true,
	"errors": [],
	"messages": [],
	"result": {
		"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"hostname": "app.example.com",
		"custom_origin_server": "example.app.com",
		"ssl": {
			"status": "pending_validation",
			"method": "cname",
			"type": "dv",
			"cname_target": "dcv.digicert.com",
			"cname": "810b7d5f01154524b961ba0cd578acc2.app.example.com",
			"settings": {
				"http2": "off",
				"tls_1_3": "on"
			}
		},
		"custom_metadata": {
			"a_random_field": "updated field value"
		}
	}
}`)
	})

	wildcard := false
	response, err := client.UpdateCustomHostname(context.Background(), "foo", "0d89c70d-ad9f-4843-b99f-6cc0252067e9", CustomHostname{Hostname: "app.example.com", CustomOriginServer: "example.app.com", SSL: &CustomHostnameSSL{Method: "cname", Type: "dv", Wildcard: &wildcard}, CustomMetadata: &CustomMetadata{"a_random_field": "updated field value"}})

	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Hostname:           "app.example.com",
			CustomOriginServer: "example.app.com",
			SSL: &CustomHostnameSSL{
				Type:   "dv",
				Method: "cname",
				Status: "pending_validation",
				SSLValidationRecord: SSLValidationRecord{
					CnameTarget: "dcv.digicert.com",
					CnameName:   "810b7d5f01154524b961ba0cd578acc2.app.example.com",
				},
				Settings: CustomHostnameSSLSettings{
					HTTP2: "off",
					TLS13: "on",
				},
			},
			CustomMetadata: &CustomMetadata{
				"a_random_field": "updated field value",
			},
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_UpdateCustomHostnameWithEmptyCustomMetadata(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/0d89c70d-ad9f-4843-b99f-6cc0252067e9", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		defer r.Body.Close()
		reqBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err, "Reading request body")
		assert.JSONEq(t, `
{
	"hostname": "app.example.com",
	"custom_origin_server": "example.app.com",
	"ssl": {
		"method": "cname",
		"type": "dv",
		"wildcard": false,
		"settings": {}
	},
	"custom_metadata": {},
	"ownership_verification": {},
	"ownership_verification_http": {}
}`, string(reqBody), "Unexpected request body")

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
	"success": true,
	"errors": [],
	"messages": [],
	"result": {
		"id": "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
		"hostname": "app.example.com",
		"custom_origin_server": "example.app.com",
		"ssl": {
			"status": "pending_validation",
			"method": "cname",
			"type": "dv",
			"cname_target": "dcv.digicert.com",
			"cname": "810b7d5f01154524b961ba0cd578acc2.app.example.com",
			"settings": {
				"http2": "off",
				"tls_1_3": "on"
			}
		},
		"custom_metadata": {}
	}
}`)
	})

	wildcard := false
	response, err := client.UpdateCustomHostname(context.Background(), "foo", "0d89c70d-ad9f-4843-b99f-6cc0252067e9", CustomHostname{Hostname: "app.example.com", CustomOriginServer: "example.app.com", SSL: &CustomHostnameSSL{Method: "cname", Type: "dv", Wildcard: &wildcard}, CustomMetadata: &CustomMetadata{}})

	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "0d89c70d-ad9f-4843-b99f-6cc0252067e9",
			Hostname:           "app.example.com",
			CustomOriginServer: "example.app.com",
			SSL: &CustomHostnameSSL{
				Type:   "dv",
				Method: "cname",
				Status: "pending_validation",
				SSLValidationRecord: SSLValidationRecord{
					CnameTarget: "dcv.digicert.com",
					CnameName:   "810b7d5f01154524b961ba0cd578acc2.app.example.com",
				},
				Settings: CustomHostnameSSLSettings{
					HTTP2: "off",
					TLS13: "on",
				},
			},
			CustomMetadata: &CustomMetadata{},
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_CustomHostnameFallbackOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/fallback_origin", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "origin": "fallback.example.com",
    "status": "pending_deployment",
    "errors": [
      "DNS records are not setup correctly. Origin should be a proxied A/AAAA/CNAME dns record"
    ],
    "created_at": "2019-10-28T18:11:23.37411Z",
    "updated_at": "2020-03-16T18:11:23.531995Z"
  }
}`)
	})

	customHostnameFallbackOrigin, err := client.CustomHostnameFallbackOrigin(context.Background(), "foo")

	want := CustomHostnameFallbackOrigin{
		Origin: "fallback.example.com",
		Status: "pending_deployment",
		Errors: []string{"DNS records are not setup correctly. Origin should be a proxied A/AAAA/CNAME dns record"},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, customHostnameFallbackOrigin)
	}
}

func TestCustomHostname_DeleteCustomHostnameFallbackOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/fallback_origin", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
{
  "id": "bar"
}`)
	})

	err := client.DeleteCustomHostnameFallbackOrigin(context.Background(), "foo")

	assert.NoError(t, err)
}

func TestCustomHostname_UpdateCustomHostnameFallbackOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames/fallback_origin", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "origin": "fallback.example.com",
    "status": "pending_deployment",
    "errors": [
      "DNS records are not setup correctly. Origin should be a proxied A/AAAA/CNAME dns record"
    ],
    "created_at": "2019-10-28T18:11:23.37411Z",
    "updated_at": "2020-03-16T18:11:23.531995Z"
  }
}`)
	})

	response, err := client.UpdateCustomHostnameFallbackOrigin(context.Background(), "foo", CustomHostnameFallbackOrigin{Origin: "fallback.example.com"})

	want := &CustomHostnameFallbackOriginResponse{
		Result: CustomHostnameFallbackOrigin{
			Origin: "fallback.example.com",
			Status: "pending_deployment",
			Errors: []string{"DNS records are not setup correctly. Origin should be a proxied A/AAAA/CNAME dns record"},
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func TestCustomHostname_CreateCustomHostnameCustomCertificateAuthority(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/custom_hostnames", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
  "result": {
    "id": "614b3124-cd57-42f0-8307-000000000000",
    "hostname": "app.example.com",
    "ssl": {
      "id": "d9ae4881-34d2-4820-8e28-000000000000",
      "type": "dv",
      "method": "http",
      "status": "initializing",
      "settings": {
        "min_tls_version": "1.2"
      },
      "wildcard": false,
      "certificate_authority": "lets_encrypt"
    },
    "custom_origin_server": "origin.example.com",
    "created_at": "2020-06-30T21:37:36.563495Z"
  },
  "success": true,
  "errors": [],
  "messages": []
}`)
	})

	response, err := client.CreateCustomHostname(context.Background(), "foo", CustomHostname{Hostname: "app.example.com", SSL: &CustomHostnameSSL{Method: "cname", Type: "dv", CertificateAuthority: "lets_encrypt"}})

	createdAt, _ := time.Parse(time.RFC3339, "2020-06-30T21:37:36.563495Z")

	wildcard := false
	want := &CustomHostnameResponse{
		Result: CustomHostname{
			ID:                 "614b3124-cd57-42f0-8307-000000000000",
			Hostname:           "app.example.com",
			CustomOriginServer: "origin.example.com",
			SSL: &CustomHostnameSSL{
				ID:     "d9ae4881-34d2-4820-8e28-000000000000",
				Type:   "dv",
				Method: "http",
				Status: "initializing",
				Settings: CustomHostnameSSLSettings{
					MinTLSVersion: "1.2",
				},
				Wildcard:             &wildcard,
				CertificateAuthority: "lets_encrypt",
			},
			CreatedAt: &createdAt,
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}
